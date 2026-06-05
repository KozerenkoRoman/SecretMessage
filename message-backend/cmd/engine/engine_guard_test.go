package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGuardSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2037,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardPriest},
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardGuard}, IsOut: false, Score: 0},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false, Score: 0},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

func TestApply_GuardCorrectGuessEliminatesOpponent(t *testing.T) {
	state, clock, rng := setupGuardSuite()

	// Хід: p1 грає Guard проти p2 і правильно вгадує Princess
	action := engine.Action{
		PlayerID:  "p1",
		HandIndex: 0,
		TargetID:  "p2",
		Guess:     engine.CardPrincess,
	}
	result, err := engine.Apply(state, action, rng, clock, 100)
	require.NoError(t, err)

	// Перевірка: p2 має вибути
	assert.True(t, result.NewState.Players["p2"].IsOut, "p2 має вибути після правильного вгадування Princess")

	// Перевірка: події (CARD_PLAYED -> GUARD_HIT -> PLAYER_ELIMINATED)
	require.GreaterOrEqual(t, len(result.DomainEvents), 3, "Має бути щонайменше 3 події")
	assert.Equal(t, engine.EventCardPlayed, result.DomainEvents[0].Type)
	assert.Equal(t, engine.EventGuardHit, result.DomainEvents[1].Type)
	assert.Equal(t, engine.EventPlayerEliminated, result.DomainEvents[2].Type)

	// Перевіряємо суворо типізований payload GUARD_HIT
	hitPayload, ok := result.DomainEvents[1].Payload.(engine.GuardHitPayload)
	require.True(t, ok, "Payload[1] має бути engine.GuardHitPayload")
	assert.Equal(t, "p1", hitPayload.PlayerID)
	assert.Equal(t, "p2", hitPayload.TargetID)
	assert.Equal(t, engine.CardPrincess, hitPayload.Guess)

	// Перевіряємо payload PLAYER_ELIMINATED з полем Reason
	elimPayload, ok := result.DomainEvents[2].Payload.(engine.PlayerEliminatedPayload)
	require.True(t, ok, "Payload[2] має бути engine.PlayerEliminatedPayload")
	assert.Equal(t, "p2", elimPayload.PlayerID)
	assert.Equal(t, engine.ReasonGuardHit, elimPayload.Reason)
}

func TestApply_GuardWrongGuessKeepsOpponentAlive(t *testing.T) {
	state, clock, rng := setupGuardSuite()

	// Хід: p1 грає Guard проти p2 і помиляється (вгадує Baron)
	action := engine.Action{
		PlayerID:  "p1",
		HandIndex: 0,
		TargetID:  "p2",
		Guess:     engine.CardBaron,
	}
	result, err := engine.Apply(state, action, rng, clock, 200)
	require.NoError(t, err)

	// Перевірка: p2 лишається в грі
	assert.False(t, result.NewState.Players["p2"].IsOut, "p2 має лишитися в грі після неправильного вгадування")

	// Перевірка: подія CARD_PLAYED є
	require.GreaterOrEqual(t, len(result.DomainEvents), 1, "Має бути щонайменше 1 подія")
	assert.Equal(t, engine.EventCardPlayed, result.DomainEvents[0].Type)
}
