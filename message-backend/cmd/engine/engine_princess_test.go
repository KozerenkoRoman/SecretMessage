package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPrinceSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2034,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard}, // запасна карта для добору
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardPrince}, IsOut: false, Score: 0},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false, Score: 0},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

func TestApply_PrinceForcesPrincessDiscard(t *testing.T) {
	state, clock, rng := setupPrinceSuite()

	// Хід: p1 грає Prince проти p2
	action := engine.Action{PlayerID: "p1", HandIndex: 0, TargetID: "p2"}
	result, err := engine.Apply(state, action, rng, clock)
	require.NoError(t, err)

	// Перевірка: p2 має вибути, бо Princess була скинута
	assert.True(t, result.NewState.Players["p2"].IsOut, "p2 має вибути після примусового скидання Princess")

	// Перевірка: Princess у discard p2
	assert.Contains(t, result.NewState.Players["p2"].DiscardPile, engine.CardPrincess, "Princess має бути у discard")

	// Перевірка: події
	require.GreaterOrEqual(t, len(result.DomainEvents), 2, "Має бути щонайменше 2 події")
	assert.Equal(t, engine.EventCardPlayed, result.DomainEvents[0].Type)
	assert.Equal(t, engine.EventPlayerEliminated, result.DomainEvents[1].Type)
}

// TestApplyPrincess_RemainingHandGoesToDiscard: коли гравець вибуває, зігравши
// Принцесу, решта його руки МАЄ потрапити у DiscardPile (а не зникнути), так
// само як при вибутті через Guard/Baron. Використовуємо 3 гравців, щоб раунд
// НЕ завершився й ми могли перевірити суто механіку відбою.
func TestApplyPrincess_RemainingHandGoesToDiscard(t *testing.T) {
	clock := &mockClock{fixedTime: time.Now()}
	state := engine.GameState{
		Phase:     engine.PhaseMainAction,
		TurnOrder: []string{"p1", "p2", "p3"},
		Players: map[string]engine.Player{
			// p1 грає Принцесу (індекс 0), у руці лишається Шпигун (індекс 1).
			// engine.go зазвичай кладе зіграну карту у DiscardPile ще до ефекту,
			// але тут ми викликаємо ApplyPrincess напряму, тож імітуємо це:
			// Принцеса вже у DiscardPile, у руці лишився тільки Шпигун.
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardSpy},
				DiscardPile: []engine.CardType{engine.CardPrincess}, IsOut: false, Score: 0},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardGuard}, IsOut: false, Score: 0},
			"p3": {ID: "p3", Hand: []engine.CardType{engine.CardBaron}, IsOut: false, Score: 0},
		},
	}

	result, err := engine.ApplyPrincess(state, engine.Action{PlayerID: "p1", HandIndex: 0}, clock)
	require.NoError(t, err)

	p1 := result.NewState.Players["p1"]
	assert.True(t, p1.IsOut, "p1 має вибути після Принцеси")
	assert.Empty(t, p1.Hand, "Рука p1 має бути порожньою")
	// КРИТИЧНО: Шпигун із руки має опинитися у відбої, а не зникнути.
	assert.Contains(t, p1.DiscardPile, engine.CardSpy, "Шпигун із руки має піти у DiscardPile при вибутті")
	assert.Contains(t, p1.DiscardPile, engine.CardPrincess, "Принцеса залишається у DiscardPile")
	assert.Len(t, p1.DiscardPile, 2, "У відбої мають бути рівно Принцеса + Шпигун")
}

// TestApplyPrincess_EliminatedKeepsSpyBonusEligibility: наскрізний сценарій —
// гравець вибуває, зігравши Принцесу з Шпигуном у руці, і все одно отримує
// бонус Шпигуна в кінці раунду (він єдиний власник Шпигуна).
func TestApplyPrincess_EliminatedKeepsSpyBonusEligibility(t *testing.T) {
	clock := &mockClock{fixedTime: time.Now()}
	state := engine.GameState{
		Phase:       engine.PhaseMainAction,
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		// Deck порожній: раунд завершиться через вибуття p1 (лишиться 1 живий).
		Players: map[string]engine.Player{
			// p1 у руці: Принцеса (грає) + Шпигун (має піти у відбій).
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardPrincess, engine.CardSpy},
				DiscardPile: []engine.CardType{}, IsOut: false, Score: 0},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardGuard},
				DiscardPile: []engine.CardType{}, IsOut: false, Score: 0},
		},
	}

	// Повний шлях через Apply: engine.go покладе Принцесу у відбій, ApplyPrincess
	// вибиває p1 і переносить Шпигуна у відбій, після чого раунд резолвиться.
	result, err := engine.Apply(state, engine.Action{PlayerID: "p1", HandIndex: 0}, &mockRNG{}, clock)
	require.NoError(t, err)

	final := result.NewState
	assert.True(t, final.Players["p1"].IsOut, "p1 вибув через Принцесу")
	assert.Contains(t, final.Players["p1"].DiscardPile, engine.CardSpy,
		"Шпигун вибулого p1 має бути у відбої для підрахунку бонусу")

	// p1 — єдиний власник Шпигуна -> отримує бонус, попри вибуття.
	assert.True(t, final.Players["p1"].SpyPointsAwarded,
		"Вибулий p1 має отримати бонус Шпигуна")
	// p2 виграє раунд (last_standing) -> +1; p1 отримує +1 лише за Шпигуна.
	assert.Equal(t, "p2", final.WinnerID)
	assert.Equal(t, 1, final.Players["p1"].Score, "p1: 0 + 1 бонус Шпигуна")
	assert.Equal(t, 1, final.Players["p2"].Score, "p2: 0 + 1 за перемогу")
}

// TestApplyPrincess_TwoPlayersRoundEndsExactlyOnce фіксує регресію з event-логу:
// у 2-гравцевій грі вибуття через Принцесу завершує раунд, і переможець отримує
// РІВНО +1 бал (а не +2), а подія ROUND_END генерується РІВНО один раз.
// Раніше ApplyPrincess і engine.Apply викликали ResolveRoundEnd двічі, через що
// у логу з'являлися дві ROUND_END і бот отримував подвійне очко.
func TestApplyPrincess_TwoPlayersRoundEndsExactlyOnce(t *testing.T) {
	clock := &mockClock{fixedTime: time.Now()}
	state := engine.GameState{
		Phase:       engine.PhaseMainAction,
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Deck:        []engine.CardType{}, // не важливо: раунд закінчиться вибуттям
		Players: map[string]engine.Player{
			// p1 добровільно грає Принцесу і вибуває.
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardPrincess, engine.CardGuard},
				DiscardPile: []engine.CardType{}, IsOut: false, Score: 4},
			// p2 лишається єдиним живим -> переможець раунду.
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardKing},
				DiscardPile: []engine.CardType{}, IsOut: false, Score: 1},
		},
	}

	result, err := engine.Apply(state, engine.Action{PlayerID: "p1", HandIndex: 0}, &mockRNG{}, clock)
	require.NoError(t, err)

	final := result.NewState

	// Переможець отримує РІВНО +1 (1 -> 2), без подвоєння.
	assert.Equal(t, "p2", final.WinnerID)
	assert.Equal(t, 2, final.Players["p2"].Score, "Переможець має отримати рівно +1 бал за раунд")
	assert.Equal(t, 4, final.Players["p1"].Score, "Бали вибулого не змінюються (немає Шпигуна)")

	// Подія ROUND_END має бути РІВНО одна.
	roundEndCount := 0
	for _, ev := range result.DomainEvents {
		if ev.Type == engine.EventRoundEnd {
			roundEndCount++
		}
	}
	assert.Equal(t, 1, roundEndCount, "ROUND_END має генеруватися рівно один раз (без дубля)")

	// Фаза має бути завершенням раунду.
	assert.Equal(t, engine.PhaseRoundEnd, final.Phase)
}
