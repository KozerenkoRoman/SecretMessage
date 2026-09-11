// =============================================================================
// simulation/types.go
//
// Внутрішні телеметричні типи пакета симуляції. Вони НЕ залежать від пакета
// storage (щоб уникнути import-циклу: storage вже не імпортує simulation, а
// runner при persist мапить ці типи у storage.SimGameRecord).
// =============================================================================

package simulation

// SimMoveRecord — телеметрія одного ходу бота у headless-матчі.
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

// SimGameRecord — телеметрія однієї зіграної гри.
type SimGameRecord struct {
	GameIndex    int             `json:"game_index"`
	Seed         int64           `json:"seed"`
	WinnerSlot   *int            `json:"winner_slot"`
	WinnerBot    string          `json:"winner_bot"`
	SpyWinnerBot string          `json:"spy_winner_bot"`
	TotalTurns   int             `json:"total_turns"`
	TotalRounds  int             `json:"total_rounds"`
	FirstMoveBot string          `json:"first_move_bot"`
	DurationMs   int64           `json:"duration_ms"`
	Moves        []SimMoveRecord `json:"moves"`
}
