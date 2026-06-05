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
	result, err := engine.Apply(state, action, rng, clock, 100)
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
