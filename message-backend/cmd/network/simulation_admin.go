// =============================================================================
// network/simulation_admin.go
//
// Адмін-API для headless Bot-vs-Bot симуляцій:
//   POST /api/admin/simulations         — запустити батч (повертає batch_id).
//   GET  /api/admin/simulations         — список останніх батчів (in-memory).
//   GET  /api/admin/simulations/{id}    — прогрес/результат конкретного батча.
//   POST /api/admin/simulations/{id}/cancel — скасувати активний батч.
//
// Персист телеметрії у БД відбувається через simStoreAdapter, який мостить
// *storage.Storage до інтерфейсу simulation.TelemetryStore (розв'язує різницю
// у типах record між пакетами й уникає import-циклу).
// =============================================================================

package network

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"secret-message/cmd/simulation"
	"secret-message/cmd/storage"

	"github.com/google/uuid"
)

// simStoreAdapter адаптує *storage.Storage під simulation.TelemetryStore.
type simStoreAdapter struct {
	store *storage.Storage
}

func newSimStoreAdapter(store *storage.Storage) *simStoreAdapter {
	if store == nil {
		return nil
	}
	return &simStoreAdapter{store: store}
}

func (a *simStoreAdapter) CreateSimBatch(ctx context.Context, id uuid.UUID, totalGames int, config json.RawMessage) error {
	return a.store.CreateSimBatch(ctx, id, totalGames, config)
}

func (a *simStoreAdapter) UpdateSimBatchProgress(ctx context.Context, id uuid.UUID, playedGames int) error {
	return a.store.UpdateSimBatchProgress(ctx, id, playedGames)
}

func (a *simStoreAdapter) FinishSimBatch(ctx context.Context, id uuid.UUID, status string, summary json.RawMessage, simErr string) error {
	return a.store.FinishSimBatch(ctx, id, status, summary, simErr)
}

func (a *simStoreAdapter) SaveSimGame(ctx context.Context, batchID uuid.UUID, game simulation.StorageGame) error {
	// Мапимо simulation.StorageGame -> storage.SimGameRecord.
	moves := make([]storage.SimMoveRecord, 0, len(game.Moves))
	for _, m := range game.Moves {
		moves = append(moves, storage.SimMoveRecord{
			TurnIndex:        m.TurnIndex,
			RoundIndex:       m.RoundIndex,
			BotID:            m.BotID,
			BotStrategy:      m.BotStrategy,
			PlayedCard:       m.PlayedCard,
			TargetID:         m.TargetID,
			GuessCard:        m.GuessCard,
			DecisionMs:       m.DecisionMs,
			InvalidAttempt:   m.InvalidAttempt,
			UsedFallback:     m.UsedFallback,
			EliminatedReason: m.EliminatedReason,
		})
	}
	return a.store.SaveSimGame(ctx, batchID, storage.SimGameRecord{
		GameIndex:    game.GameIndex,
		Seed:         game.Seed,
		WinnerSlot:   game.WinnerSlot,
		WinnerBot:    game.WinnerBot,
		SpyWinnerBot: game.SpyWinnerBot,
		TotalTurns:   game.TotalTurns,
		TotalRounds:  game.TotalRounds,
		FirstMoveBot: game.FirstMoveBot,
		DurationMs:   game.DurationMs,
		Moves:        moves,
	})
}

// HandleStartSimulation обробляє POST /api/admin/simulations.
func (s *Server) HandleStartSimulation(w http.ResponseWriter, r *http.Request) {
	var cfg simulation.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		s.sendJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// За замовчуванням персистимо телеметрію (можна вимкнути persist:false у тілі).
	if !cfg.Persist {
		cfg.Persist = true
	}

	batchID, err := s.simController.StartBatch(cfg)
	if err != nil {
		s.sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.log.WithField("batch_id", batchID).Info("[sim] batch started via admin API")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "started",
		"batch_id": batchID,
	})
}

// HandleListSimulations обробляє GET /api/admin/simulations.
func (s *Server) HandleListSimulations(w http.ResponseWriter, r *http.Request) {
	batches := s.simController.List()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"batches": batches,
	})
}

// HandleGetSimulation обробляє GET /api/admin/simulations/{id}.
func (s *Server) HandleGetSimulation(w http.ResponseWriter, r *http.Request) {
	batchID := simulationIDFromPath(r)
	if batchID == "" {
		s.sendJSONError(w, http.StatusBadRequest, "missing batch id")
		return
	}

	progress, ok := s.simController.Progress(batchID)
	if !ok {
		s.sendJSONError(w, http.StatusNotFound, "batch not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(progress)
}

// HandleCancelSimulation обробляє POST /api/admin/simulations/{id}/cancel.
func (s *Server) HandleCancelSimulation(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("id")
	if batchID == "" {
		s.sendJSONError(w, http.StatusBadRequest, "missing batch id")
		return
	}
	if !s.simController.Cancel(batchID) {
		s.sendJSONError(w, http.StatusNotFound, "batch not found or already finished")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "cancelling"})
}

// simulationIDFromPath дістає {id} з /api/admin/simulations/{id}.
func simulationIDFromPath(r *http.Request) string {
	if v := r.PathValue("id"); v != "" {
		return v
	}
	// Резерв: ручний парсинг, якщо шаблон маршруту зміниться.
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}

// sendJSONError — уніфікована JSON-помилка (узгоджено зі стилем admin.go).
func (s *Server) sendJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
