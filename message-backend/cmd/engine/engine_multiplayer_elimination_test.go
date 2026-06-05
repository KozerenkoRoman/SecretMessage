package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupEliminationSuite налаштовує коректну кількість карт для симуляції порівняння Бароном.
func setupEliminationSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:     1234,
		Sequence: 1,
		Phase:    engine.PhaseMainAction,
		// Порожня колода або мінімальний набір, щоб уникнути побічних ефектів зміщення карт
		Deck:        []engine.CardType{engine.CardPriest},
		TurnOrder:   []string{"p1", "p2", "p3"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			// Даємо p1 дві карти: Барона для розіграшу та Вартового для порівняння з p2
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardBaron, engine.CardGuard}, IsOut: false},
			// p2 має Принцесу — вона порівнюватиметься з Вартовим гравця p1
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false},
			"p3": {ID: "p3", Hand: []engine.CardType{engine.CardGuard}, IsOut: false},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestMultiplayerRoundWithElimination перевіряє ланцюжок ходів з вибуванням гравців.
func TestMultiplayerRoundWithElimination(t *testing.T) {
	state, clock, rng := setupEliminationSuite()

	// Хід 1: p1 грає Baron (індекс 0) проти p2
	action1 := engine.Action{
		PlayerID:  "p1",
		HandIndex: 0,
		TargetID:  "p2",
	}
	result1, err := engine.Apply(state, action1, rng, clock, 100)
	require.NoError(t, err)

	// Перевірка порівняння: p1 вибуває, бо його Вартовий (1) менший за Принцесу (8) гравця p2
	assert.True(t, result1.NewState.Players["p1"].IsOut, "p1 має вибути, бо його карта менша за Princess")
	assert.False(t, result1.NewState.Players["p2"].IsOut, "p2 лишається в грі з Princess")

	// Перехід ходу до наступного активного гравця — p2
	nextState := result1.NewState
	assert.Equal(t, "p2", nextState.TurnOrder[nextState.CurrentTurn], "Хід переходить до p2")

	// Хід 2: p2 добровільно або змушено грає Принцесу (індекс 0), що призводить до її вибування
	action2 := engine.Action{
		PlayerID:  "p2",
		HandIndex: 0,
	}
	result2, err := engine.Apply(nextState, action2, rng, clock, 200)
	require.NoError(t, err)
	assert.True(t, result2.NewState.Players["p2"].IsOut, "p2 має вибути після розіграшу Princess")

	// Перевірка фіналу раунду: єдиним активним гравцем має залишитися p3
	activeCount := 0
	for _, p := range result2.NewState.Players {
		if !p.IsOut {
			activeCount++
		}
	}
	assert.Equal(t, 1, activeCount, "У кінці раунду має лишитися один активний гравець")
	assert.False(t, result2.NewState.Players["p3"].IsOut, "p3 лишається переможцем раунду")
}
