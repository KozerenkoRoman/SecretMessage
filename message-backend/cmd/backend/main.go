package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"secret-message/cmd/config"
	"secret-message/cmd/network"
	"secret-message/cmd/simulation"
	"secret-message/cmd/storage"
	"secret-message/cmd/tlsutil"

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

	// Якщо задано FRONTEND_PROXY_URL — Go-бекенд стає єдиною публічною
	// точкою входу (порти 80/443): усі запити, що не підпадають під
	// /api/* чи /ws (зареєстровані вище через InitRoutes), проксіюються
	// на внутрішній SPA-контейнер (nginx зі статикою фронтенду).
	if cfg.FrontendProxyURL != "" {
		proxyHandler, err := network.NewFrontendProxyHandler(cfg.FrontendProxyURL, logger)
		if err != nil {
			logger.Fatalf("Некоректний FRONTEND_PROXY_URL %q: %v", cfg.FrontendProxyURL, err)
		}
		mux.Handle("/", proxyHandler)
		logger.Infof("Реверс-проксі на фронтенд увімкнено: %s", cfg.FrontendProxyURL)
	}

	handler := network.CORSMiddleware(mux, cfg.CORSAllowedOrigins)

	// 5. Запускаємо HTTP(S) сервер(и) у окремих горутинах.
	servers := startServers(cfg, logger, handler)

	logger.Info("Сервер повністю готовий до роботи. Очікування сигналів зупинки...")

	// Очікуємо Ctrl+C / SIGTERM
	<-ctx.Done()

	// Наводимо порядок перед виходом: плавно вимикаємо усі HTTP(S) сервери.
	logger.Info("Початок плавної зупинки сервера(ів)...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for _, s := range servers {
		wg.Add(1)
		go func(s *http.Server) {
			defer wg.Done()
			if err := s.Shutdown(shutdownCtx); err != nil {
				logger.Errorf("Помилка під час зупинки сервера %s: %v", s.Addr, err)
			}
		}(s)
	}
	wg.Wait()

	logger.Info("Сервер успішно зупинено.")
}

// startServers піднімає HTTP(S) сервер(и) відповідно до конфігурації і
// повертає їх список для подальшого graceful shutdown.
//
// Дизайн:
//   - TLSEnabled=false (за замовчуванням, сумісно з існуючою топологією,
//     де TLS термінується на nginx) — один звичайний HTTP-сервер на
//     cfg.SocketPort, поведінка ідентична попередній реалізації.
//   - TLSEnabled=true — два сервери:
//   - :443 (HTTPS) з autocert.Manager.GetCertificate у TLSConfig,
//     обслуговує основний застосунок (handler);
//   - :80 (HTTP) з autocert.Manager.HTTPHandler(nil), що вирішує
//     ACME HTTP-01 challenge і редіректить решту трафіку на HTTPS.
func startServers(cfg *config.Config, logger *logrus.Logger, handler http.Handler) []*http.Server {
	if !cfg.TLSEnabled {
		addr := fmt.Sprintf(":%d", cfg.SocketPort)
		srv := &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		go func() {
			logger.Infof("Сервер Love Letter успішно піднято на http://localhost%s", addr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Fatalf("Помилка роботи HTTP-сервера: %v", err)
			}
		}()

		return []*http.Server{srv}
	}

	logger.Infof("TLS увімкнено (Let's Encrypt/autocert). Білий список доменів: %v", cfg.TLSDomains)

	certManager, err := tlsutil.NewManager(cfg.TLSDomains, cfg.TLSCacheDir, cfg.TLSEmail)
	if err != nil {
		logger.Fatalf("Не вдалося ініціалізувати autocert.Manager: %v", err)
	}

	httpsSrv := &http.Server{
		Addr:         ":443",
		Handler:      handler,
		TLSConfig:    certManager.TLSConfig(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// HTTP-сервер на :80 вирішує ACME HTTP-01 challenge (шлях
	// /.well-known/acme-challenge/...) і автоматично редіректить решту
	// запитів на https:// — саме так влаштований autocert.Manager.HTTPHandler.
	httpSrv := &http.Server{
		Addr:         ":80",
		Handler:      certManager.HTTPHandler(nil),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("HTTP-сервер (ACME challenge + редірект на HTTPS) піднято на :80")
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Помилка роботи HTTP-сервера (:80): %v", err)
		}
	}()

	go func() {
		logger.Info("HTTPS-сервер успішно піднято на :443 (сертифікат отримується/оновлюється автоматично через Let's Encrypt)")
		if err := httpsSrv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Помилка роботи HTTPS-сервера: %v", err)
		}
	}()

	return []*http.Server{httpsSrv, httpSrv}
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
