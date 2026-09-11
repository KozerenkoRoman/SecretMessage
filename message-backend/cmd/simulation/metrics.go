// =============================================================================
// simulation/metrics.go
//
// Агрегація телеметрії headless Bot-vs-Bot симуляції.
//
// MetricsCollector акумулює результати кожної гри/ходу у потокобезпечний спосіб
// і будує підсумковий звіт: win-rate по стратегіях, розподіл зіграних карт,
// вибір цілей, середня тривалість/кількість ходів, невалідні ходи та fallback.
// =============================================================================

package simulation

import (
	"errors"
	"sync"

	"secret-message/cmd/engine"
)

// Помилки рівня стратегій (використовуються у strategy.go).
var (
	errUnexpectedDecision = errors.New("strategy returned unexpected decision type")
	errNoLegalMove        = errors.New("no legal move available for bot")
)

// CardName повертає людсько-читабельну назву карти для звітів.
func CardName(c engine.CardType) string {
	switch c {
	case engine.CardSpy:
		return "Spy"
	case engine.CardGuard:
		return "Guard"
	case engine.CardPriest:
		return "Priest"
	case engine.CardBaron:
		return "Baron"
	case engine.CardHandmaid:
		return "Handmaid"
	case engine.CardPrince:
		return "Prince"
	case engine.CardChancellor:
		return "Chancellor"
	case engine.CardKing:
		return "King"
	case engine.CardCountess:
		return "Countess"
	case engine.CardPrincess:
		return "Princess"
	default:
		return "Unknown"
	}
}

// StrategyStats — агрегати по одній стратегії/особистості бота.
type StrategyStats struct {
	Strategy       string         `json:"strategy"`
	GamesPlayed    int            `json:"games_played"`
	Wins           int            `json:"wins"`
	SpyBonuses     int            `json:"spy_bonuses"`
	WinRate        float64        `json:"win_rate"`
	TotalMoves     int            `json:"total_moves"`
	InvalidMoves   int            `json:"invalid_moves"`
	FallbackMoves  int            `json:"fallback_moves"`
	CardPlays      map[string]int `json:"card_plays"`       // назва карти -> скільки разів зіграно
	TargetedMoves  int            `json:"targeted_moves"`   // ходи, що мали ціль
	FirstMoveGames int            `json:"first_move_games"` // у скількох іграх ходив першим
	FirstMoveWins  int            `json:"first_move_wins"`  // з них скільки виграв
}

// Summary — фінальний агрегований звіт батча.
type Summary struct {
	TotalGames     int                       `json:"total_games"`
	CompletedGames int                       `json:"completed_games"`
	DrawGames      int                       `json:"draw_games"`
	AvgTurns       float64                   `json:"avg_turns"`
	AvgRounds      float64                   `json:"avg_rounds"`
	AvgDurationMs  float64                   `json:"avg_duration_ms"`
	TotalInvalid   int                       `json:"total_invalid_moves"`
	TotalFallback  int                       `json:"total_fallback_moves"`
	PerStrategy    map[string]*StrategyStats `json:"per_strategy"`
	CardPlayTotals map[string]int            `json:"card_play_totals"`
	EliminationsBy map[string]int            `json:"eliminations_by_reason"`
}

// MetricsCollector — потокобезпечний акумулятор.
type MetricsCollector struct {
	mu sync.Mutex

	totalGames     int
	completedGames int
	drawGames      int
	sumTurns       int
	sumRounds      int
	sumDurationMs  int64
	totalInvalid   int
	totalFallback  int

	perStrategy    map[string]*StrategyStats
	cardPlayTotals map[string]int
	eliminationsBy map[string]int
}

func NewMetricsCollector(totalGames int) *MetricsCollector {
	return &MetricsCollector{
		totalGames:     totalGames,
		perStrategy:    make(map[string]*StrategyStats),
		cardPlayTotals: make(map[string]int),
		eliminationsBy: make(map[string]int),
	}
}

func (m *MetricsCollector) strategyStats(name string) *StrategyStats {
	st, ok := m.perStrategy[name]
	if !ok {
		st = &StrategyStats{
			Strategy:  name,
			CardPlays: make(map[string]int),
		}
		m.perStrategy[name] = st
	}
	return st
}

// RecordGame інтегрує результат однієї завершеної гри (з її ходами) у агрегати.
// slotStrategies відображає botID -> назва стратегії цього слота.
func (m *MetricsCollector) RecordGame(game SimGameRecord, slotStrategies map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.completedGames++
	m.sumTurns += game.TotalTurns
	m.sumRounds += game.TotalRounds
	m.sumDurationMs += game.DurationMs
	if game.WinnerSlot == nil {
		m.drawGames++
	}

	// Унікальні стратегії, що брали участь у цій грі (для GamesPlayed рахуємо гру раз на стратегію).
	seen := make(map[string]bool)
	for _, strat := range slotStrategies {
		if !seen[strat] {
			m.strategyStats(strat).GamesPlayed++
			seen[strat] = true
		}
	}

	// Переможець.
	if game.WinnerBot != "" {
		if strat, ok := slotStrategies[game.WinnerBot]; ok {
			m.strategyStats(strat).Wins++
		}
	}
	if game.SpyWinnerBot != "" {
		if strat, ok := slotStrategies[game.SpyWinnerBot]; ok {
			m.strategyStats(strat).SpyBonuses++
		}
	}

	// First-move аналіз.
	if game.FirstMoveBot != "" {
		if strat, ok := slotStrategies[game.FirstMoveBot]; ok {
			st := m.strategyStats(strat)
			st.FirstMoveGames++
			if game.WinnerBot == game.FirstMoveBot {
				st.FirstMoveWins++
			}
		}
	}

	// Ходи.
	for _, mv := range game.Moves {
		st := m.strategyStats(mv.BotStrategy)
		st.TotalMoves++
		if mv.InvalidAttempt {
			st.InvalidMoves++
			m.totalInvalid++
		}
		if mv.UsedFallback {
			st.FallbackMoves++
			m.totalFallback++
		}
		if mv.TargetID != "" {
			st.TargetedMoves++
		}
		cardName := CardName(engine.CardType(mv.PlayedCard))
		st.CardPlays[cardName]++
		m.cardPlayTotals[cardName]++
		if mv.EliminatedReason != "" {
			m.eliminationsBy[mv.EliminatedReason]++
		}
	}
}

// Snapshot будує підсумковий звіт (обчислює win-rate тощо). Безпечно викликати
// у будь-який момент — корисно для проміжного прогресу.
func (m *MetricsCollector) Snapshot() Summary {
	m.mu.Lock()
	defer m.mu.Unlock()

	sum := Summary{
		TotalGames:     m.totalGames,
		CompletedGames: m.completedGames,
		DrawGames:      m.drawGames,
		TotalInvalid:   m.totalInvalid,
		TotalFallback:  m.totalFallback,
		PerStrategy:    make(map[string]*StrategyStats, len(m.perStrategy)),
		CardPlayTotals: make(map[string]int, len(m.cardPlayTotals)),
		EliminationsBy: make(map[string]int, len(m.eliminationsBy)),
	}

	if m.completedGames > 0 {
		sum.AvgTurns = float64(m.sumTurns) / float64(m.completedGames)
		sum.AvgRounds = float64(m.sumRounds) / float64(m.completedGames)
		sum.AvgDurationMs = float64(m.sumDurationMs) / float64(m.completedGames)
	}

	for name, st := range m.perStrategy {
		// Копіюємо, щоб не віддавати внутрішні вказівники назовні.
		cp := *st
		cp.CardPlays = make(map[string]int, len(st.CardPlays))
		for k, v := range st.CardPlays {
			cp.CardPlays[k] = v
		}
		if cp.GamesPlayed > 0 {
			cp.WinRate = float64(cp.Wins) / float64(cp.GamesPlayed)
		}
		sum.PerStrategy[name] = &cp
	}
	for k, v := range m.cardPlayTotals {
		sum.CardPlayTotals[k] = v
	}
	for k, v := range m.eliminationsBy {
		sum.EliminationsBy[k] = v
	}
	return sum
}
