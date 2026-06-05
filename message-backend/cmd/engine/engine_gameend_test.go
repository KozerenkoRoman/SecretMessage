package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGameEndSuite створює початковий стан гри.
func setupGameEndSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2030,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard, engine.CardPriest},
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			"p1": {
				ID:    "p1",
				Score: 6,
				IsOut: false,
				Hand:  []engine.CardType{engine.CardSpy, engine.CardPrincess},
			},
			"p2": {
				ID:    "p2",
				Score: 3,
				IsOut: false,
				Hand:  []engine.CardType{engine.CardPrincess},
			},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestGameEndCondition перевіряє завершення всієї гри при досягненні ліміту очок.
func TestGameEndCondition(t *testing.T) {
	state, clock, rng := setupGameEndSuite()

	// Хід 1: p1 грає Spy (індекс 0)
	action1 := engine.Action{PlayerID: "p1", HandIndex: 0}
	result1, err := engine.Apply(state, action1, rng, clock, 100)
	require.NoError(t, err)

	// Хід 2: p2 грає Princess (індекс 0) → вибуває
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0}
	result2, err := engine.Apply(result1.NewState, action2, rng, clock, 200)
	require.NoError(t, err)

	// Завершення раунду через ResolveRoundEnd
	finalResult := engine.ResolveRoundEnd(result2.NewState, clock, 300)
	final := finalResult.NewState

	// Перевірка: 6 (старт) + 1 (перемога) + 1 (бонус Spy) = 8 очок
	assert.Equal(t, 8, final.Players["p1"].Score, "p1 має отримати 8 очок (6 початкових + 1 за перемогу + 1 бонус за Шпигунку)")
	assert.True(t, final.IsGameOver, "Гра має завершитися після досягнення цільових очок")
	assert.Equal(t, "p1", final.WinnerID, "Переможцем гри має бути p1")
}
