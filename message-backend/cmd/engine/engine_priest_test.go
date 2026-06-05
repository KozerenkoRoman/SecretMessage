package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupPriestSuite створює початковий стан для перевірки карти Священника.
func setupPriestSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2038,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard},
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			// Гравець p1 тримає Священника
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardPriest}, IsOut: false, Score: 0},
			// Гравець p2 тримає Принцесу, яку ми будемо перевіряти
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false, Score: 0},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestApply_PriestRevealsOpponentCard перевіряє генерацію події CARD_REVEALED із правильною структурою даних.
func TestApply_PriestRevealsOpponentCard(t *testing.T) {
	state, clock, rng := setupPriestSuite()

	// Хід: p1 грає Priest проти p2
	action := engine.Action{
		PlayerID:  "p1",
		HandIndex: 0,
		TargetID:  "p2",
	}
	result, err := engine.Apply(state, action, rng, clock, 100)
	require.NoError(t, err)

	// Перевірка кількості згенерованих подій
	require.GreaterOrEqual(t, len(result.DomainEvents), 2, "Має бути щонайменше 2 події")

	// Перевірка типів подій (CARD_PLAYED та PRIEST_EFFECT)
	assert.Equal(t, engine.EventCardPlayed, result.DomainEvents[0].Type)
	assert.Equal(t, engine.EventPriestEffect, result.DomainEvents[1].Type)

	// Робимо точне приведення інтерфейсу до структури з вашого доменного шару
	payload, ok := result.DomainEvents[1].Payload.(engine.PriestEffectPayload)
	require.True(t, ok, "Payload має містити структуру типу engine.PriestEffectPayload")

	// Перевірка полів структури
	assert.Equal(t, "p2", payload.TargetID, "Цільовий гравець має бути p2")
	assert.Equal(t, engine.CardPrincess, payload.Card, "Відкрита карта має бути CardPrincess (Принцеса)")
}
