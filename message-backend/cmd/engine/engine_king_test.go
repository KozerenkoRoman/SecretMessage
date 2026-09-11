package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupKingSuite створює початковий ігровий стан для тесту карти Король.
func setupKingSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2032,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard},
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			// p1 тримає єдину карту — Короля
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardKing}, IsOut: false, Score: 0},
			// p2 тримає Принцесу
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false, Score: 0},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestApply_KingSwapsHands перевіряє правильність логіки обміну картами між гравцями.
func TestApply_KingSwapsHands(t *testing.T) {
	state, clock, rng := setupKingSuite()

	// Дія: Гравець p1 розігрує Короля (індекс 0) проти гравця p2
	action := engine.Action{PlayerID: "p1", HandIndex: 0, TargetID: "p2"}
	result, err := engine.Apply(state, action, rng, clock)
	require.NoError(t, err)

	// Отримуємо фінальний стан рук після виконання дії
	p1Hand := result.NewState.Players["p1"].Hand
	p2Hand := result.NewState.Players["p2"].Hand

	// Перевірка обміну: p1 повинен отримати карту, яка була у p2 (Принцесу)
	assert.Equal(t, []engine.CardType{engine.CardPrincess}, p1Hand, "p1 має отримати карту від p2")

	// Після обміну p2 отримав порожню руку p1. Далі хід переходить до p2
	// (AdvanceTurn), і оскільки в p2 менше 2 карт, він добирає карту з колоди
	// (Guard) — це початок ЙОГО ходу за правилами Love Letter.
	assert.Equal(t, []engine.CardType{engine.CardGuard}, p2Hand,
		"p2 добирає карту на початку свого ходу після обміну")

	// Події: CARD_PLAYED (King), HANDS_SWAPPED (обмін), CARD_DRAWN (добір p2 на його ході).
	require.Len(t, result.DomainEvents, 3, "Має бути 3 події: CARD_PLAYED, HANDS_SWAPPED, CARD_DRAWN")
	assert.Equal(t, engine.EventCardPlayed, result.DomainEvents[0].Type)
	assert.Equal(t, engine.EventHandsSwapped, result.DomainEvents[1].Type)
	assert.Equal(t, engine.EventCardDrawn, result.DomainEvents[2].Type)
}
