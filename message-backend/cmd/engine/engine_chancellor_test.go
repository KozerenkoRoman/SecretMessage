package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupChancellorSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2026,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard, engine.CardPriest, engine.CardBaron},
		TurnOrder:   []string{"player_1", "player_2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			// ПРАВИЛЬНО: У гравця 1 має бути ДВІ карти в момент ходу (наприклад, якась інша карта + Канцлер)
			"player_1": {
				ID:    "player_1",
				Hand:  []engine.CardType{engine.CardChancellor, engine.CardPriest}, // Канцлер на позиції 0, Принц/Священник на позиції 1
				IsOut: false,
			},
			"player_2": {
				ID:    "player_2",
				Hand:  []engine.CardType{engine.CardPrince},
				IsOut: false,
			},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

func TestApply_ChancellorResolve(t *testing.T) {
	state, clock, rng := setupChancellorSuite()

	// Етап 1: гравець грає Канцлера (він під індексом 0)
	action := engine.Action{
		PlayerID:  "player_1",
		HandIndex: 0,
	}
	result, err := engine.Apply(state, action, rng, clock)
	require.NoError(t, err)

	// Перевірка: тепер у руці справді 3 карти! (1 яка залишалась [CardPriest] + 2 нові з колоди)
	assert.Len(t, result.NewState.Players["player_1"].Hand, 3, "Після добору Канцлера у руці має бути 3 карти")
	assert.Equal(t, engine.PhaseResolveChancellor, result.NewState.Phase, "FSM має перейти у фазу RESOLVE_CHANCELLOR")

	// Етап 2: гравець резолвить Канцлера
	resolve := engine.ChancellorResolveAction{
		PlayerID:      "player_1",
		KeepHandIndex: 1, // залишаємо карту під індексом 1
		BottomOrder: []engine.CardType{
			result.NewState.Players["player_1"].Hand[0],
			result.NewState.Players["player_1"].Hand[2],
		},
	}
	resolved, err := engine.ResolveChancellor(result.NewState, resolve, clock)
	require.NoError(t, err)

	// Перевірка: у руці лишилася 1 карта
	assert.Len(t, resolved.NewState.Players["player_1"].Hand, 1, "Після резолву у руці має бути 1 карта")

	// Перевірка: дві карти повернулися вниз колоди
	deck := resolved.NewState.Deck
	require.GreaterOrEqual(t, len(deck), 2, "У колоді має бути принаймні 2 карти після резолву")
	assert.Equal(t, resolve.BottomOrder[0], deck[len(deck)-2], "Перша карта має бути другою знизу")
	assert.Equal(t, resolve.BottomOrder[1], deck[len(deck)-1], "Друга карта має бути останньою")
}
