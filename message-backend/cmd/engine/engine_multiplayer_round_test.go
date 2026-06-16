package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMultiplayerRoundSuite створює початковий ігровий стан.
// Ми даємо гравцям правильну кількість карт, щоб уникнути помилки порівняння.
func setupMultiplayerRoundSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:     5678,
		Sequence: 1,
		Phase:    engine.PhaseMainAction,
		// Колода містить карти, які рушій може автоматично видавати під час переходів ходу
		Deck:        []engine.CardType{engine.CardGuard, engine.CardPriest},
		TurnOrder:   []string{"p1", "p2", "p3"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			// p1 має Барона на індексі 0 та Вартового на індексі 1 (Вартовий піде на порівняння)
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardBaron, engine.CardGuard}, IsOut: false, Score: 0},
			// p2 має Принцесу, яка сильніша за Вартового p1
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false, Score: 0},
			"p3": {ID: "p3", Hand: []engine.CardType{engine.CardHandmaid}, IsOut: false, Score: 0},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestMultiplayerRoundSimulation перевіряє логіку роботи порівняння карт та зміни станів.
func TestMultiplayerRoundSimulation(t *testing.T) {
	state, clock, rng := setupMultiplayerRoundSuite()

	// Крок 1: Гравець p1 розігрує Барона (індекс 0) проти гравця p2
	action1 := engine.Action{
		PlayerID:  "p1",
		HandIndex: 0,
		TargetID:  "p2",
	}
	result1, err := engine.Apply(state, action1, rng, clock)
	require.NoError(t, err)

	// Перевірка результату першого ходу:
	// p1 має вибути, бо його Вартовий (1) слабший за Принцесу (8) гравця p2
	assert.True(t, result1.NewState.Players["p1"].IsOut, "p1 має вибути після невдалого порівняння Бароном")
	assert.False(t, result1.NewState.Players["p2"].IsOut, "p2 має залишитися в грі")

	// Перевіряємо події першого ходу
	require.GreaterOrEqual(t, len(result1.DomainEvents), 1, "Рушій мав згенерувати подію розіграшу")
	assert.Equal(t, engine.EventCardPlayed, result1.DomainEvents[0].Type)
}
