package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupBaronSuite налаштовує початковий стан гри для тестування Барона.
func setupBaronSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2033,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard},
		TurnOrder:   []string{"player1", "player2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			"player1": {ID: "player1", Hand: []engine.CardType{engine.CardBaron, engine.CardKing}, IsOut: false},
			"player2": {ID: "player2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false, Score: 0},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestApply_BaronEliminatesWeaker перевіряє, що слабший гравець вибуває,
// а раунд автоматично завершується, якщо залишився один гравець.
func TestApply_BaronEliminatesWeaker(t *testing.T) {
	state, clock, rng := setupBaronSuite()

	// Дія: player1 грає Барона проти player2
	action := engine.Action{
		PlayerID:  "player1",
		TargetID:  "player2",
		HandIndex: 0,
	}

	// Виклик головного рушія гри
	result, err := engine.Apply(state, action, rng, clock)
	require.NoError(t, err)

	// Перевірка змін у стані гравців
	assert.True(t, result.NewState.Players["player1"].IsOut, "player1 має вибути після порівняння")
	assert.False(t, result.NewState.Players["player2"].IsOut, "player2 має залишитися в грі")

	// Перевірка доменних подій. Тепер ми очікуємо 5 подій через завершення раунду.
	require.Len(t, result.DomainEvents, 5, "Має бути згенеровано рівно 5 подій")

	// Подія 1: Карта розіграна
	assert.Equal(t, engine.EventCardPlayed, result.DomainEvents[0].Type)

	// Подія 2: Карти порівняно всередині рушія
	assert.Equal(t, engine.EventRoundCompared, result.DomainEvents[1].Type)

	// Подія 3: Результат Барона (визначення переможця ефекту)
	assert.Equal(t, engine.EventBaronResult, result.DomainEvents[2].Type)

	// Подія 4: Офіційне вибуття гравця player1
	assert.Equal(t, engine.EventPlayerEliminated, result.DomainEvents[3].Type)

	// Подія 5: Автоматичне завершення раунду, бо залишився 1 гравець
	assert.Equal(t, engine.EventRoundEnd, result.DomainEvents[4].Type)
}

func TestApply_BaronTieKeepsBothAlive(t *testing.T) {
	// Створюємо стан для нічиєї: у player1 та player2 однакові карти в руках
	state := engine.GameState{
		Seed:        2033,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard},
		TurnOrder:   []string{"player1", "player2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			// player1 скидає Барона (індекс 0), в руці залишається Guard (сила 1)
			"player1": {ID: "player1", Hand: []engine.CardType{engine.CardBaron, engine.CardGuard}, IsOut: false},
			// player2 тримає в руці Guard (сила 1)
			"player2": {ID: "player2", Hand: []engine.CardType{engine.CardGuard}, IsOut: false, Score: 0},
		},
	}
	clock := &mockClock{fixedTime: time.Now()}
	rng := &mockRNG{}

	// Дія: player1 грає Барона проти player2
	action := engine.Action{
		PlayerID:  "player1",
		TargetID:  "player2",
		HandIndex: 0,
	}

	// Виклик рушія гри
	result, err := engine.Apply(state, action, rng, clock)
	require.NoError(t, err)

	// Перевірка стану: Обидва гравці повинні залишитися в грі!
	assert.False(t, result.NewState.Players["player1"].IsOut, "player1 НЕ має вибути при нічиї")
	assert.False(t, result.NewState.Players["player2"].IsOut, "player2 НЕ має вибути при нічиї")

	// Перевірка доменних подій: при нічиї генеруються тільки 2 події
	require.Len(t, result.DomainEvents, 2, "При нічиї має бути згенеровано рівно 2 події")
	assert.Equal(t, engine.EventCardPlayed, result.DomainEvents[0].Type)
	assert.Equal(t, engine.EventRoundCompared, result.DomainEvents[1].Type)
}
