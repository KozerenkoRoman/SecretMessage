package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCountessSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2035,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard},
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			// p1 має Countess і King → мусить зіграти Countess
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardCountess, engine.CardKing}, IsOut: false, Score: 0},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPriest}, IsOut: false, Score: 0},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

func TestApply_CountessForcedPlay(t *testing.T) {
	state, clock, rng := setupCountessSuite()

	// Хід: p1 грає Countess (примусово, бо в руці King)
	action := engine.Action{PlayerID: "p1", HandIndex: 0}
	result, err := engine.Apply(state, action, rng, clock, 100)
	require.NoError(t, err)

	// Перевірка: Countess має бути у discard
	assert.Contains(t, result.NewState.Players["p1"].DiscardPile, engine.CardCountess, "Countess має бути у discard")

	// Перевірка: King лишається в руці
	assert.Equal(t, []engine.CardType{engine.CardKing}, result.NewState.Players["p1"].Hand, "King має лишитися в руці")

	// Перевірка: подія CARD_PLAYED
	require.Len(t, result.DomainEvents, 1, "Має бути одна подія")
	assert.Equal(t, engine.EventCardPlayed, result.DomainEvents[0].Type)
}
