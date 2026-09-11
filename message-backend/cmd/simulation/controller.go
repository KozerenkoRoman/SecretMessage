// =============================================================================
// simulation/controller.go
//
// Batch-контролер: керує запуском N ігор, агрегує метрики, публікує живий
// прогрес і (опційно) персистить телеметрію у сховище.
//
// Контролер тримає реєстр активних/завершених батчів у пам'яті, тож фронтенд
// може поллити прогрес за batchID навіть якщо БД-персист вимкнено.
// =============================================================================

package simulation

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TelemetryStore — мінімальний інтерфейс персистенції, який потрібен контролеру.
// Реалізується *storage.Storage (див. simulation_store.go). Виноситься в
// інтерфейс, щоб unit-тести могли підставити no-op/фейкове сховище.
type TelemetryStore interface {
	CreateSimBatch(ctx context.Context, id uuid.UUID, totalGames int, config json.RawMessage) error
	UpdateSimBatchProgress(ctx context.Context, id uuid.UUID, playedGames int) error
	FinishSimBatch(ctx context.Context, id uuid.UUID, status string, summary json.RawMessage, simErr string) error
	SaveSimGame(ctx context.Context, batchID uuid.UUID, game StorageGame) error
}

// StorageGame — DTO для персисту (дзеркалить storage.SimGameRecord). Визначений
// тут, щоб пакет simulation не імпортував storage (уникаємо import-циклу).
type StorageGame struct {
	GameIndex    int
	Seed         int64
	WinnerSlot   *int
	WinnerBot    string
	SpyWinnerBot string
	TotalTurns   int
	TotalRounds  int
	FirstMoveBot string
	DurationMs   int64
	Moves        []SimMoveRecord
}

// BatchProgress — знімок прогресу для фронтенду.
type BatchProgress struct {
	BatchID     string    `json:"batch_id"`
	Status      string    `json:"status"` // running | completed | failed | cancelled
	TotalGames  int       `json:"total_games"`
	PlayedGames int       `json:"played_games"`
	Config      Config    `json:"config"`
	Summary     Summary   `json:"summary"`
	Error       string    `json:"error,omitempty"`
	StartedAt   time.Time `json:"started_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type batchState struct {
	mu        sync.RWMutex
	progress  BatchProgress
	collector *MetricsCollector
	cancel    context.CancelFunc
}

// Controller керує життєвим циклом батчів симуляції.
type Controller struct {
	log   *logrus.Logger
	store TelemetryStore

	mu      sync.RWMutex
	batches map[string]*batchState
}

// NewController створює контролер. store може бути nil — тоді персист вимкнено
// (метрики все одно доступні у пам'яті через прогрес).
func NewController(store TelemetryStore, logger *logrus.Logger) *Controller {
	return &Controller{
		log:     logger,
		store:   store,
		batches: make(map[string]*batchState),
	}
}

// StartBatch валідовує конфіг і запускає батч у фоні. Повертає batchID одразу.
func (c *Controller) StartBatch(cfg Config) (string, error) {
	norm, err := cfg.Normalize()
	if err != nil {
		return "", err
	}
	if norm.Seed == 0 {
		norm.Seed = time.Now().UnixNano()
	}

	id := uuid.New()
	collector := NewMetricsCollector(norm.Count)

	bs := &batchState{
		collector: collector,
		progress: BatchProgress{
			BatchID:    id.String(),
			Status:     "running",
			TotalGames: norm.Count,
			Config:     norm,
			StartedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	c.mu.Lock()
	c.batches[id.String()] = bs
	c.mu.Unlock()

	// Персист-запис батча (best-effort; не блокуємо старт при помилці БД).
	if c.store != nil && norm.Persist {
		cfgJSON, _ := json.Marshal(norm)
		if err := c.store.CreateSimBatch(context.Background(), id, norm.Count, cfgJSON); err != nil {
			c.log.WithError(err).Warn("[sim] failed to persist batch record; continuing in-memory")
		}
	}

	runCtx, cancel := context.WithCancel(context.Background())
	bs.mu.Lock()
	bs.cancel = cancel
	bs.mu.Unlock()

	go c.runBatch(runCtx, id, norm, bs)

	return id.String(), nil
}

// runBatch виконує усі ігри послідовно (детерміновано за seed) і оновлює прогрес.
func (c *Controller) runBatch(ctx context.Context, id uuid.UUID, cfg Config, bs *batchState) {
	slotStrategies := make(map[string]string, len(cfg.Slots))
	for _, slot := range cfg.Slots {
		slotStrategies[slot.BotID] = slot.Strategy
	}

	var finalErr string
	status := "completed"

	for i := 0; i < cfg.Count; i++ {
		select {
		case <-ctx.Done():
			status = "cancelled"
			finalErr = ctx.Err().Error()
			goto finish
		default:
		}

		gameSeed := cfg.Seed + int64(i)
		runner := NewRunner(cfg, gameSeed, c.log)
		rec, err := runner.RunGame(ctx, i, gameSeed)
		if err != nil {
			c.log.WithError(err).WithField("game_index", i).Warn("[sim] game aborted")
			// Не валимо весь батч через одну гру; продовжуємо.
			continue
		}

		bs.collector.RecordGame(rec, slotStrategies)

		// Персист однієї гри (best-effort).
		if c.store != nil && cfg.Persist {
			if err := c.store.SaveSimGame(context.Background(), id, toStorageGame(rec)); err != nil {
				c.log.WithError(err).Warn("[sim] failed to persist game telemetry")
			}
		}

		// Оновлюємо живий прогрес.
		snap := bs.collector.Snapshot()
		bs.mu.Lock()
		bs.progress.PlayedGames = i + 1
		bs.progress.Summary = snap
		bs.progress.UpdatedAt = time.Now()
		bs.mu.Unlock()

		// Персист прогресу зрідка (кожні 25 ігор), щоб не спамити БД.
		if c.store != nil && cfg.Persist && (i+1)%25 == 0 {
			_ = c.store.UpdateSimBatchProgress(context.Background(), id, i+1)
		}
	}

finish:
	snap := bs.collector.Snapshot()
	bs.mu.Lock()
	bs.progress.Status = status
	bs.progress.Summary = snap
	bs.progress.Error = finalErr
	bs.progress.UpdatedAt = time.Now()
	bs.mu.Unlock()

	if c.store != nil && cfg.Persist {
		summaryJSON, _ := json.Marshal(snap)
		if err := c.store.FinishSimBatch(context.Background(), id, status, summaryJSON, finalErr); err != nil {
			c.log.WithError(err).Warn("[sim] failed to finalize batch record")
		}
	}

	c.log.WithFields(logrus.Fields{
		"batch_id": id.String(),
		"status":   status,
		"games":    snap.CompletedGames,
	}).Info("[sim] batch finished")
}

// Progress повертає поточний знімок прогресу батча.
func (c *Controller) Progress(batchID string) (BatchProgress, bool) {
	c.mu.RLock()
	bs, ok := c.batches[batchID]
	c.mu.RUnlock()
	if !ok {
		return BatchProgress{}, false
	}
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.progress, true
}

// List повертає знімки всіх відомих контролеру батчів (найновіші перші).
func (c *Controller) List() []BatchProgress {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]BatchProgress, 0, len(c.batches))
	for _, bs := range c.batches {
		bs.mu.RLock()
		out = append(out, bs.progress)
		bs.mu.RUnlock()
	}
	return out
}

// Cancel зупиняє активний батч (best-effort).
func (c *Controller) Cancel(batchID string) bool {
	c.mu.RLock()
	bs, ok := c.batches[batchID]
	c.mu.RUnlock()
	if !ok {
		return false
	}
	bs.mu.RLock()
	cancel := bs.cancel
	bs.mu.RUnlock()
	if cancel != nil {
		cancel()
		return true
	}
	return false
}

// RunBatchSync проганяє батч синхронно й повертає підсумок (для CLI/тестів,
// без реєстрації в реєстрі й без БД).
func RunBatchSync(ctx context.Context, cfg Config, logger *logrus.Logger) (Summary, error) {
	norm, err := cfg.Normalize()
	if err != nil {
		return Summary{}, err
	}
	if norm.Seed == 0 {
		norm.Seed = time.Now().UnixNano()
	}

	collector := NewMetricsCollector(norm.Count)
	slotStrategies := make(map[string]string, len(norm.Slots))
	for _, slot := range norm.Slots {
		slotStrategies[slot.BotID] = slot.Strategy
	}

	for i := 0; i < norm.Count; i++ {
		select {
		case <-ctx.Done():
			return collector.Snapshot(), ctx.Err()
		default:
		}
		gameSeed := norm.Seed + int64(i)
		runner := NewRunner(norm, gameSeed, logger)
		rec, err := runner.RunGame(ctx, i, gameSeed)
		if err != nil {
			continue
		}
		collector.RecordGame(rec, slotStrategies)
	}
	return collector.Snapshot(), nil
}

func toStorageGame(rec SimGameRecord) StorageGame {
	return StorageGame{
		GameIndex:    rec.GameIndex,
		Seed:         rec.Seed,
		WinnerSlot:   rec.WinnerSlot,
		WinnerBot:    rec.WinnerBot,
		SpyWinnerBot: rec.SpyWinnerBot,
		TotalTurns:   rec.TotalTurns,
		TotalRounds:  rec.TotalRounds,
		FirstMoveBot: rec.FirstMoveBot,
		DurationMs:   rec.DurationMs,
		Moves:        rec.Moves,
	}
}
