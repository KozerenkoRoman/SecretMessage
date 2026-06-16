package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApply_Success(t *testing.T) {
	state := engine.GameState{
		Seed:           2040,
		Sequence:       42,
		Phase:          engine.PhaseMainAction,
		Deck:           []engine.CardType{engine.CardGuard, engine.CardPriest}, // додали карти
		TurnOrder:      []string{"player1", "player2"},
		CurrentTurn:    0,
		TransitionHash: "initial_hash",
		Players: map[string]engine.Player{
			"player1": {ID: "player1", Hand: []engine.CardType{engine.CardHandmaid}, IsOut: false},
			"player2": {ID: "player2", Hand: []engine.CardType{engine.CardGuard}, IsOut: false},
		},
	}

	action := engine.Action{PlayerID: "player1", HandIndex: 0}

	result, err := engine.Apply(state, action, &mockRNG{}, &mockClock{fixedTime: time.Now()})
	require.NoError(t, err)

	// Перевірка: Sequence збільшився
	assert.Equal(t, state.Sequence+1, result.NewState.Sequence)

	// Перевірка: є події
	require.NotEmpty(t, result.DomainEvents, "Має бути хоча б одна подія")
	assert.Equal(t, engine.EventCardPlayed, result.DomainEvents[0].Type)
}

func TestApply_OutOfTurnRejected(t *testing.T) {
	// Arrange: Черга гравця "player1"
	state := engine.GameState{
		TurnOrder:      []string{"player1", "player2"},
		CurrentTurn:    0,
		Sequence:       1,
		TransitionHash: "some_hash",
	}

	// Act: Намагається виконати дію "player2"
	action := engine.Action{
		PlayerID:  "player2",
		HandIndex: 0,
	}

	_, err := engine.Apply(state, action, &mockRNG{}, &mockClock{fixedTime: time.Now()})

	// Assert: Запит має бути відхилений
	if err == nil {
		t.Fatal("expected error for out of turn action, got nil")
	}
}

func TestApply_EmptyTurnOrderRejected(t *testing.T) {
	state := engine.GameState{
		TurnOrder: []string{}, // порожній список
		Phase:     engine.PhaseMainAction,
		Players:   map[string]engine.Player{},
	}

	action := engine.Action{PlayerID: "p1", HandIndex: 0}
	_, err := engine.Apply(state, action, &mockRNG{}, &mockClock{fixedTime: time.Now()})

	require.Error(t, err)
	require.Equal(t, engine.ErrEmptyTurnOrder, engine.CodeOf(err))
}

func TestApply_InvalidPhaseRejected(t *testing.T) {
	state := engine.GameState{
		TurnOrder: []string{"p1"},
		Phase:     engine.PhaseRoundEnd, // неправильна фаза
		Players: map[string]engine.Player{
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardGuard}},
		},
	}

	action := engine.Action{PlayerID: "p1", HandIndex: 0}
	_, err := engine.Apply(state, action, &mockRNG{}, &mockClock{fixedTime: time.Now()})

	require.Error(t, err)
	// Перевіряємо стабільний код помилки, а не англійський текст —
	// текст може змінюватись, код є публічним контрактом.
	require.Equal(t, engine.ErrInvalidPhase, engine.CodeOf(err))
}

func TestApply_NoPlayersRejected(t *testing.T) {
	state := engine.GameState{
		TurnOrder: []string{"p1"},
		Phase:     engine.PhaseMainAction,
		Players:   map[string]engine.Player{}, // немає гравців
	}

	action := engine.Action{PlayerID: "p1", HandIndex: 0}
	_, err := engine.Apply(state, action, &mockRNG{}, &mockClock{fixedTime: time.Now()})

	require.Error(t, err)
	require.Equal(t, engine.ErrNoPlayers, engine.CodeOf(err))
}
