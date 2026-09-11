package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"secret-message/cmd/config"
	"secret-message/cmd/network"
	"secret-message/cmd/simulation"
	"secret-message/cmd/storage"

	"github.com/sirupsen/logrus"
)

func main() {
	// CLI прапорці для headless-режиму симуляції (без HTTP-сервера).
	var (
		simCount      = flag.Int("sim", 0, "run N headless bot-vs-bot games and exit (0 = normal server mode)")
		simStrategies = flag.String("sim-strategies", "smart,smart", "comma-separated strategies per slot (smart|random)")
		simFirstMove  = flag.String("sim-first-move", "", "bot slot id forced to move first every game (e.g. sim-bot-0)")
		simSeed       = flag.Int64("sim-seed", 0, "base RNG seed (0 = random)")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	logger.SetLevel(logrus.DebugLevel)

	// Headless CLI-режим: проганяємо батч і виходимо, не піднімаючи сервер/БД.
	if *simCount > 0 {
		logger.SetLevel(logrus.InfoLevel)
		runHeadlessCLI(ctx, logger, *simCount, *simStrategies, *simFirstMove, *simSeed)
		return
	}

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

	store.SeedUsers(ctx, logger)

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

// runHeadlessCLI проганяє батч Bot-vs-Bot ігор без HTTP-сервера й БД і друкує
// JSON-підсумок у stdout. Приклад:
//
//	go run ./cmd/backend -sim=500 -sim-strategies=smart,random -sim-first-move=sim-bot-0
func runHeadlessCLI(ctx context.Context, logger *logrus.Logger, count int, strategiesCSV, firstMove string, seed int64) {
	slots := make([]simulation.SlotConfig, 0)
	for i, strat := range splitCSV(strategiesCSV) {
		slots = append(slots, simulation.SlotConfig{
			BotID:    fmt.Sprintf("sim-bot-%d", i),
			Strategy: strat,
			Username: fmt.Sprintf("Bot-%d(%s)", i, strat),
		})
	}

	cfg := simulation.Config{
		Count:          count,
		Slots:          slots,
		Seed:           seed,
		FirstMoveBotID: firstMove,
	}

	logger.Infof("Headless-симуляція: %d ігор, слоти=%v, first-move=%q", count, strategiesCSV, firstMove)
	start := time.Now()
	summary, err := simulation.RunBatchSync(ctx, cfg, logger)
	if err != nil {
		logger.Errorf("Симуляцію перервано: %v", err)
	}
	logger.Infof("Завершено за %s", time.Since(start))

	out, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(out))
}

// splitCSV розбиває "a,b,c" на слайс, ігноруючи порожні елементи/пробіли.
func splitCSV(s string) []string {
	out := make([]string, 0)
	for _, part := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
