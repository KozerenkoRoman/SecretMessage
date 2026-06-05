package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSpyRoundSuite налаштовує початковий стан для тестів із карткою Шпигуна.
func setupSpyRoundSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2027,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard, engine.CardPriest},
		TurnOrder:   []string{"p1", "p2", "p3"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			// p1: грає Шпигуна, Принцеса залишається в руці (сила 9)
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardSpy, engine.CardPrincess}, IsOut: false, Score: 0},
			// p2: грає Принцесу і вибуває
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false, Score: 0},
			// p3: даємо дві карти. Він грає Барона (індекс 0), а Вартовий (сила 1) залишається в руці для порівняння
			"p3": {ID: "p3", Hand: []engine.CardType{engine.CardBaron, engine.CardGuard}, IsOut: false, Score: 0},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestApply_SpyBonusAwarded перевіряє нарахування бонусу Шпигуна переможцю раунду.
func TestApply_SpyBonusAwarded(t *testing.T) {
	state, clock, rng := setupSpyRoundSuite()

	// Хід 1: p1 грає Spy
	action1 := engine.Action{PlayerID: "p1", HandIndex: 0}
	result1, err := engine.Apply(state, action1, rng, clock, 100)
	require.NoError(t, err)

	// Хід 2: p2 грає Princess і вибуває
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0}
	result2, err := engine.Apply(result1.NewState, action2, rng, clock, 200)
	require.NoError(t, err)

	// Завершуємо раунд. Порівнюються залишки карт: p1 (Princess, 9) проти p3 (Baron, 3).
	// Оскільки в цьому тесті ми примусово викликаємо ResolveRoundEnd без Ходу 3, рушій порівняє перші карти в руках.
	final := engine.ResolveRoundEnd(result2.NewState, clock, 300).NewState

	// Перевірки з урахуванням нової логіки бекенду
	assert.True(t, final.Players["p1"].SpyPointsAwarded, "p1 має отримати бонус Spy")
	assert.Equal(t, "p1", final.WinnerID, "Переможцем має бути p1")

	// ВИПРАВЛЕНО: p1 отримує 1 очко за перемогу за картою + 1 очко за Шпигуна = разом 2 очки
	assert.Equal(t, 2, final.Players["p1"].Score, "p1 має отримати разом 2 очки (перемога + шпигун)")
}

// TestFullRoundWithSpyBonus перевіряє ізоляцію бонусу Шпигуна під час активних ходів інших гравців.
func TestFullRoundWithSpyBonus(t *testing.T) {
	state, clock, rng := setupSpyRoundSuite()

	// Хід 1: p1 грає Spy. В руці залишається [CardPrincess]
	action1 := engine.Action{PlayerID: "p1", HandIndex: 0}
	result1, err := engine.Apply(state, action1, rng, clock, 100)
	require.NoError(t, err)

	// Хід 2: p2 грає Princess і вибуває.
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0}
	result2, err := engine.Apply(result1.NewState, action2, rng, clock, 200)
	require.NoError(t, err)

	// Хід 3: p3 грає Baron (індекс 0) проти p1.
	// В руці у p3 залишається [CardGuard] (сила 1). У p1 в руці [CardPrincess] (сила 9).
	// Карта p3 слабша, тому p3 вибуває з гри.
	action3 := engine.Action{PlayerID: "p3", HandIndex: 0, TargetID: "p1"}
	result3, err := engine.Apply(result2.NewState, action3, rng, clock, 300)
	require.NoError(t, err)

	// Завершення раунду
	final := engine.ResolveRoundEnd(result3.NewState, clock, 400).NewState

	// Перевірка: бонус має отримати лише гравець p1
	assert.True(t, final.Players["p1"].SpyPointsAwarded, "p1 має отримати бонус за Spy")
	assert.False(t, final.Players["p3"].SpyPointsAwarded, "p3 не має отримати бонус")

	// Оскільки p3 вибув, p1 залишився останнім живим граючим і виграв раунд
	assert.Equal(t, 2, final.Players["p1"].Score, "p1 отримує 2 очки (шпигун + останній живий)")
}

// НОВИЙ ТЕСТ: Перевіряє випадок, коли гравець вибув, але все одно бере участь у підрахунку Шпигунів
func TestSpyBonus_WhenPlayerIsOut(t *testing.T) {
	state, clock, rng := setupSpyRoundSuite()

	// Хід 1: p1 грає Spy. В руці залишається [CardPrincess]
	action1 := engine.Action{PlayerID: "p1", HandIndex: 0}
	result1, err := engine.Apply(state, action1, rng, clock, 100)
	require.NoError(t, err)

	// Хід 2: Навмисно вибиваємо p1 з гри (наприклад, p2 за допомогою Вартового вгадує Принцесу у p1).
	// Для симуляції міняємо руку p2 на Вартового
	modState := result1.NewState
	p2 := modState.Players["p2"]
	p2.Hand = []engine.CardType{engine.CardGuard}
	modState.Players["p2"] = p2

	// p2 вгадує, що у p1 в руці Принцеса (9). p1 вибуває.
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0, TargetID: "p1", Guess: engine.CardPrincess}
	result2, err := engine.Apply(modState, action2, rng, clock, 200)
	require.NoError(t, err)

	// Переконуємось, що p1 дійсно вибув, але його відбій містить Шпигуна
	assert.True(t, result2.NewState.Players["p1"].IsOut)

	// Примусово завершуємо раунд. Тепер живі лише p2 та p3. Припустимо, виграє p2.
	final := engine.ResolveRoundEnd(result2.NewState, clock, 300).NewState

	// КРИТИЧНА ПЕРЕВІРКА: p1 вибув, але він єдиний, хто розіграв Шпигуна за раунд.
	// Він ПОВИНЕН отримати бонусний бал в Score!
	assert.True(t, final.Players["p1"].SpyPointsAwarded, "Вибулий p1 все одно фіксується як володар бонусу Шпигуна")
	assert.Equal(t, 1, final.Players["p1"].Score, "Вибулий p1 має отримати 1 очко за успішне шпигунство")
}

// НОВИЙ ТЕСТ: Перевіряє взаємне знищення ефекту, якщо двоє різних гравців скинули по Шпигуну
func TestSpyBonus_CanceledWhenMultipleOwners(t *testing.T) {
	state, clock, rng := setupSpyRoundSuite()

	// Даємо Шпигуна також і гравцю p3 в руку
	p3 := state.Players["p3"]
	p3.Hand = []engine.CardType{engine.CardSpy, engine.CardGuard}
	state.Players["p3"] = p3

	// Хід 1: p1 грає Spy
	action1 := engine.Action{PlayerID: "p1", HandIndex: 0}
	result1, err := engine.Apply(state, action1, rng, clock, 100)
	require.NoError(t, err)

	// Хід 2: p2 скидає Принцесу і вибуває
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0}
	result2, err := engine.Apply(result1.NewState, action2, rng, clock, 200)
	require.NoError(t, err)

	// Хід 3: p3 теж грає свого Spy з руки (індекс 0)
	action3 := engine.Action{PlayerID: "p3", HandIndex: 0}
	result3, err := engine.Apply(result2.NewState, action3, rng, clock, 300)
	require.NoError(t, err)

	// Завершення раунду
	final := engine.ResolveRoundEnd(result3.NewState, clock, 400).NewState

	// ПЕРЕВІРКА: Оскільки Шпигунів на столі двоє (у p1 та p3), ніхто не отримує бонусних балів
	assert.False(t, final.Players["p1"].SpyPointsAwarded)
	assert.False(t, final.Players["p3"].SpyPointsAwarded)
	assert.Equal(t, 0, final.Players["p1"].Score, "Бонус Шпигуна анульовано для p1")
	assert.Equal(t, 0, final.Players["p3"].Score, "Бонус Шпигуна анульовано для p3")
}
