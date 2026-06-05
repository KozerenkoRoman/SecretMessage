package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupHandmaidSuite налаштовує початковий стан для тесту Покоївки.
func setupHandmaidSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2039,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard, engine.CardGuard},
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardHandmaid, engine.CardPrince}, IsOut: false, Score: 0, IsProtected: false},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardGuard, engine.CardBaron}, IsOut: false, Score: 0, IsProtected: false},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestApply_HandmaidProtectsFromTargetedAttack перевіряє, що захист Покоївки
// рятує гравця від атаки Вартового.
func TestApply_HandmaidProtectsFromTargetedAttack(t *testing.T) {
	state, clock, rng := setupHandmaidSuite()

	// 1. Хід: p1 розігрує Поковку (Handmaid) під індексом 0
	action1 := engine.Action{PlayerID: "p1", HandIndex: 0}
	result1, err := engine.Apply(state, action1, rng, clock, 100)
	require.NoError(t, err)

	// Перевірка: статус захисту змінився на true
	assert.True(t, result1.NewState.Players["p1"].IsProtected, "p1 має бути захищений після розіграшу Handmaid")

	// 2. Хід: p2 намагається атакувати гравця p1 за допомогою Вартового (Guard)
	action2 := engine.Action{
		PlayerID:  "p2",
		HandIndex: 0,
		TargetID:  "p1",
		Guess:     engine.CardPrincess,
	}
	result2, err := engine.Apply(result1.NewState, action2, rng, clock, 200)
	require.NoError(t, err)

	// Головна перевірка безпеки: p1 живий, бо ефект карти Guard змазався об захист
	assert.False(t, result2.NewState.Players["p1"].IsOut, "p1 не має вибути, бо його захищає Покоївка")

	// Перевірка доменних подій, які згенерував рушій
	require.GreaterOrEqual(t, len(result2.DomainEvents), 2, "Має бути щонайменше 2 події")

	// Перша подія: карта Вартового була розіграна
	assert.Equal(t, engine.EventCardPlayed, result2.DomainEvents[0].Type)

	// Друга подія: оскільки єдина ціль (p1) захищена, Guard не має ефекту;
	// рушій просто завершує хід p2 і добирає карту наступному гравцеві.
	// Тому другою подією є EventCardDrawn.
	assert.Equal(t, engine.EventCardDrawn, result2.DomainEvents[1].Type)
}

func TestHandmaidProtectionLifecycle(t *testing.T) {
	p1ID := "player_1"
	p2ID := "player_2"

	// 1. Ініціалізуємо початковий стан гри для 2 гравців
	state := engine.GameState{
		Phase:       engine.PhaseMainAction,
		CurrentTurn: 0, // Хід Гравця 1
		TurnOrder:   []string{p1ID, p2ID},
		Deck:        []engine.CardType{engine.CardGuard, engine.CardPriest}, // Карти для добору
		Players: map[string]engine.Player{
			p1ID: {
				ID:          p1ID,
				Hand:        []engine.CardType{engine.CardHandmaid, engine.CardGuard}, // Гравець 1 має Служницю
				DiscardPile: []engine.CardType{},
				IsProtected: false,
				IsOut:       false,
			},
			p2ID: {
				ID:          p2ID,
				Hand:        []engine.CardType{engine.CardHandmaid}, // ВИПРАВЛЕННЯ: Даємо Служницю, вона НЕ вимагає TargetID
				DiscardPile: []engine.CardType{},
				IsProtected: false,
				IsOut:       false,
			},
		},
	}

	rng := &mockRNG{}
	clock := &mockClock{fixedTime: time.Now()}

	// 2. Крок 1: Гравець 1 грає Служницю (індекс 0 у його руці)
	action1 := engine.Action{
		PlayerID:  p1ID,
		HandIndex: 0, // CardHandmaid
	}

	result1, err := engine.Apply(state, action1, rng, clock, 1000)
	require.NoError(t, err)

	// Перевіряємо, що після розіграшу Служниці Гравець 1 став ЗАХИЩЕНИМ
	p1AfterHandmaid := result1.NewState.Players[p1ID]
	assert.True(t, p1AfterHandmaid.IsProtected, "Expected Player 1 to be protected right after playing Handmaid")

	// Перевіряємо, що хід перейшов до Гравця 2
	assert.Equal(t, 1, result1.NewState.CurrentTurn, "Expected CurrentTurn to be 1 (Player 2)")

	// 3. Крок 2: Гравець 2 робить свій хід. Він теж грає Служницю (вона безпечна і не вимагає цілей)
	state2 := result1.NewState

	action2 := engine.Action{
		PlayerID:  p2ID,
		HandIndex: 0, // Скидає свою Служницю
	}

	result2, err := engine.Apply(state2, action2, rng, clock, 2000)
	require.NoError(t, err)

	// Тепер хід знову має повернутися до Гравця 1 (індекс 0 в TurnOrder)
	assert.Equal(t, 0, result2.NewState.CurrentTurn, "Expected CurrentTurn to return to 0 (Player 1)")

	// КРИТИЧНА ПЕРЕВІРКА: Оскільки хід повернувся до Гравця 1, його статус IsProtected має автоматично скинутися в false!
	p1AtNewTurn := result2.NewState.Players[p1ID]
	assert.False(t, p1AtNewTurn.IsProtected, "Expected Player 1 protection to be reset to false at the start of their new turn")
}
