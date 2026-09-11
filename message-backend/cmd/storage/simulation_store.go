package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// =============================================================================
// storage/simulation_store.go
//
// Персистенція телеметрії Bot-vs-Bot симуляцій (таблиці sim_batches / sim_games
// / sim_moves з міграції 0005).
//
// Методи написані вручну на pgx (як SaveGameResult у storage.go), щоб не
// залежати від повторного `sqlc generate`. Це узгоджено з наявним стилем
// проєкту, де складніші агрегатні запити тримаються поза згенерованим шаром.
// =============================================================================

// SimBatchRecord — рядок таблиці sim_batches у Go-формі.
type SimBatchRecord struct {
	ID          uuid.UUID       `json:"id"`
	Status      string          `json:"status"`
	TotalGames  int             `json:"total_games"`
	PlayedGames int             `json:"played_games"`
	Config      json.RawMessage `json:"config"`
	Summary     json.RawMessage `json:"summary"`
	Error       string          `json:"error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	FinishedAt  *time.Time      `json:"finished_at,omitempty"`
}

// SimGameRecord — один рядок sim_games для батч-вставки.
type SimGameRecord struct {
	GameIndex    int    `json:"game_index"`
	Seed         int64  `json:"seed"`
	WinnerSlot   *int   `json:"winner_slot"`
	WinnerBot    string `json:"winner_bot"`
	SpyWinnerBot string `json:"spy_winner_bot"`
	TotalTurns   int    `json:"total_turns"`
	TotalRounds  int    `json:"total_rounds"`
	FirstMoveBot string `json:"first_move_bot"`
	DurationMs   int64  `json:"duration_ms"`
	Moves        []SimMoveRecord
}

// SimMoveRecord — один рядок sim_moves.
type SimMoveRecord struct {
	TurnIndex        int    `json:"turn_index"`
	RoundIndex       int    `json:"round_index"`
	BotID            string `json:"bot_id"`
	BotStrategy      string `json:"bot_strategy"`
	PlayedCard       int    `json:"played_card"`
	TargetID         string `json:"target_id"`
	GuessCard        *int   `json:"guess_card"`
	DecisionMs       int64  `json:"decision_ms"`
	InvalidAttempt   bool   `json:"invalid_attempt"`
	UsedFallback     bool   `json:"used_fallback"`
	EliminatedReason string `json:"eliminated_reason"`
}

// CreateSimBatch створює запис батча у статусі "running" і повертає його ID.
func (s *Storage) CreateSimBatch(ctx context.Context, id uuid.UUID, totalGames int, config json.RawMessage) error {
	if len(config) == 0 {
		config = json.RawMessage("{}")
	}
	_, err := s.db.Exec(ctx,
		`INSERT INTO sim_batches (id, status, total_games, played_games, config)
		 VALUES ($1, 'running', $2, 0, $3)`,
		id, totalGames, config,
	)
	if err != nil {
		return fmt.Errorf("failed to create sim batch: %w", err)
	}
	return nil
}

// UpdateSimBatchProgress оновлює лічильник зіграних ігор (для real-time прогресу).
func (s *Storage) UpdateSimBatchProgress(ctx context.Context, id uuid.UUID, playedGames int) error {
	_, err := s.db.Exec(ctx,
		`UPDATE sim_batches SET played_games = $2 WHERE id = $1`,
		id, playedGames,
	)
	if err != nil {
		return fmt.Errorf("failed to update sim batch progress: %w", err)
	}
	return nil
}

// FinishSimBatch фіксує фінальний статус, підсумкову аналітику та час завершення.
func (s *Storage) FinishSimBatch(ctx context.Context, id uuid.UUID, status string, summary json.RawMessage, simErr string) error {
	if len(summary) == 0 {
		summary = json.RawMessage("{}")
	}
	_, err := s.db.Exec(ctx,
		`UPDATE sim_batches
		 SET status = $2, summary = $3, error = NULLIF($4, ''), finished_at = now(),
		     played_games = total_games
		 WHERE id = $1`,
		id, status, summary, simErr,
	)
	if err != nil {
		return fmt.Errorf("failed to finish sim batch: %w", err)
	}
	return nil
}

// SaveSimGame атомарно зберігає одну зіграну гру разом з усіма її ходами.
func (s *Storage) SaveSimGame(ctx context.Context, batchID uuid.UUID, game SimGameRecord) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin sim game tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var gameID int64
	err = tx.QueryRow(ctx,
		`INSERT INTO sim_games
		   (batch_id, game_index, seed, winner_slot, winner_bot, spy_winner_bot,
		    total_turns, total_rounds, first_move_bot, duration_ms)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 RETURNING id`,
		batchID, game.GameIndex, game.Seed, game.WinnerSlot, game.WinnerBot,
		nullifyEmpty(game.SpyWinnerBot), game.TotalTurns, game.TotalRounds,
		nullifyEmpty(game.FirstMoveBot), game.DurationMs,
	).Scan(&gameID)
	if err != nil {
		return fmt.Errorf("failed to insert sim game: %w", err)
	}

	if len(game.Moves) > 0 {
		rows := make([][]any, 0, len(game.Moves))
		for _, m := range game.Moves {
			rows = append(rows, []any{
				gameID, m.TurnIndex, m.RoundIndex, m.BotID, m.BotStrategy,
				m.PlayedCard, nullifyEmpty(m.TargetID), m.GuessCard, m.DecisionMs,
				m.InvalidAttempt, m.UsedFallback, nullifyEmpty(m.EliminatedReason),
			})
		}
		_, err = tx.CopyFrom(ctx,
			pgx.Identifier{"sim_moves"},
			[]string{
				"game_id", "turn_index", "round_index", "bot_id", "bot_strategy",
				"played_card", "target_id", "guess_card", "decision_ms",
				"invalid_attempt", "used_fallback", "eliminated_reason",
			},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return fmt.Errorf("failed to bulk insert sim moves: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetSimBatch зчитує один батч за ID (для polling прогресу та результатів з UI).
func (s *Storage) GetSimBatch(ctx context.Context, id uuid.UUID) (SimBatchRecord, error) {
	var rec SimBatchRecord
	var simErr *string
	err := s.db.QueryRow(ctx,
		`SELECT id, status, total_games, played_games, config, summary, error, created_at, finished_at
		 FROM sim_batches WHERE id = $1`,
		id,
	).Scan(&rec.ID, &rec.Status, &rec.TotalGames, &rec.PlayedGames,
		&rec.Config, &rec.Summary, &simErr, &rec.CreatedAt, &rec.FinishedAt)
	if err != nil {
		return SimBatchRecord{}, fmt.Errorf("failed to get sim batch: %w", err)
	}
	if simErr != nil {
		rec.Error = *simErr
	}
	return rec, nil
}

// ListSimBatches повертає останні батчі для списку в адмін-панелі.
func (s *Storage) ListSimBatches(ctx context.Context, limit int) ([]SimBatchRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.Query(ctx,
		`SELECT id, status, total_games, played_games, config, summary, error, created_at, finished_at
		 FROM sim_batches ORDER BY created_at DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list sim batches: %w", err)
	}
	defer rows.Close()

	out := make([]SimBatchRecord, 0, limit)
	for rows.Next() {
		var rec SimBatchRecord
		var simErr *string
		if err := rows.Scan(&rec.ID, &rec.Status, &rec.TotalGames, &rec.PlayedGames,
			&rec.Config, &rec.Summary, &simErr, &rec.CreatedAt, &rec.FinishedAt); err != nil {
			return nil, fmt.Errorf("failed to scan sim batch: %w", err)
		}
		if simErr != nil {
			rec.Error = *simErr
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// nullifyEmpty повертає nil для порожнього рядка, щоб у БД лягав NULL замість "".
func nullifyEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
