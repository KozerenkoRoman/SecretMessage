package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"secret-message/cmd/config"
	"secret-message/cmd/network"
	"secret-message/cmd/storage"

	"github.com/sirupsen/logrus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	logger.SetLevel(logrus.DebugLevel)

	logger.Info("Запуск сервера Love Letter...")

	cfg := config.Load()
	store, err := storage.New(ctx, cfg, logger)
	if err != nil {
		logger.Fatalf("Критична помилка ініціалізації сховища: %v", err)
	}
	defer func() {
		logger.Info("Закриття пулу з'єднань із базою даних...")
		_ = store.Close(ctx)
	}()

	if err := store.SeedAdmin(ctx, logger); err != nil {
		logger.Errorf("Попередження: не вдалося виконати початкове заповнення адміна: %v", err)
	}

	logger.Infof("Успішно підключено до PostgreSQL на порту %d! Міграції перевірено/застосовано.", cfg.DBPort)

	hub := network.NewHub(store, logger)
	gateway := network.NewGateway(hub, logger)
	hub.OnLobbyChanged = func(lobbyRooms []network.RoomLobbyInfo) {
		gateway.BroadcastToDesktop(lobbyRooms)
	}
	hub.Start(context.Background())
	defer hub.Stop()

	wsServer := network.NewServer(hub, gateway, logger, store)

	// 3. Створюємо ізольований HTTP маршрутизатор (Mux) замість глобального DefaultServeMux
	mux := http.NewServeMux()

	// 4. Викликаємо винесену функцію для реєстрації всіх маршрутів
	network.InitRoutes(mux, wsServer)

	// Налаштовуємо параметри HTTP-сервера, передаючи туди наш mux
	addr := fmt.Sprintf(":%d", cfg.SocketPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      network.CORSMiddleware(mux),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// 5. Запускаємо HTTP сервер у окремій горутині
	go func() {
		logger.Infof("Сервер Love Letter успішно піднято на http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Помилка роботи HTTP-сервера: %v", err)
		}
	}()

	logger.Info("Сервер повністю готовий до роботи. Очікування сигналів зупинки...")

	// Очікуємо Ctrl+C
	<-ctx.Done()

	// Наводимо порядок перед виходом: плавно вимикаємо HTTP сервер
	logger.Info("Початок плавної зупинки HTTP сервера...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Помилка під час зупинки HTTP сервера: %v", err)
	}

	logger.Info("Сервер успішно зупинено.")
}
