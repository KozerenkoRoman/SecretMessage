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
	result1, err := engine.Apply(state, action1, rng, clock)
	require.NoError(t, err)

	// Хід 2: p2 грає Princess і вибуває
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0}
	result2, err := engine.Apply(result1.NewState, action2, rng, clock)
	require.NoError(t, err)

	// Завершуємо раунд. Порівнюються залишки карт: p1 (Princess, 9) проти p3 (Baron, 3).
	// Оскільки в цьому тесті ми примусово викликаємо ResolveRoundEnd без Ходу 3, рушій порівняє перші карти в руках.
	final := engine.ResolveRoundEnd(result2.NewState, clock).NewState

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
	result1, err := engine.Apply(state, action1, rng, clock)
	require.NoError(t, err)

	// Хід 2: p2 грає Princess і вибуває.
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0}
	result2, err := engine.Apply(result1.NewState, action2, rng, clock)
	require.NoError(t, err)

	// Хід 3: p3 грає Baron (індекс 0) проти p1.
	// В руці у p3 залишається [CardGuard] (сила 1). У p1 в руці [CardPrincess] (сила 9).
	// Карта p3 слабша, тому p3 вибуває з гри.
	action3 := engine.Action{PlayerID: "p3", HandIndex: 0, TargetID: "p1"}
	result3, err := engine.Apply(result2.NewState, action3, rng, clock)
	require.NoError(t, err)

	// Завершення раунду
	final := engine.ResolveRoundEnd(result3.NewState, clock).NewState

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
	result1, err := engine.Apply(state, action1, rng, clock)
	require.NoError(t, err)

	// Хід 2: Навмисно вибиваємо p1 з гри (наприклад, p2 за допомогою Вартового вгадує Принцесу у p1).
	// Для симуляції міняємо руку p2 на Вартового
	modState := result1.NewState
	p2 := modState.Players["p2"]
	p2.Hand = []engine.CardType{engine.CardGuard}
	modState.Players["p2"] = p2

	// p2 вгадує, що у p1 в руці Принцеса (9). p1 вибуває.
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0, TargetID: "p1", Guess: engine.CardPrincess}
	result2, err := engine.Apply(modState, action2, rng, clock)
	require.NoError(t, err)

	// Переконуємось, що p1 дійсно вибув, але його відбій містить Шпигуна
	assert.True(t, result2.NewState.Players["p1"].IsOut)

	// Примусово завершуємо раунд. Тепер живі лише p2 та p3. Припустимо, виграє p2.
	final := engine.ResolveRoundEnd(result2.NewState, clock).NewState

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
	result1, err := engine.Apply(state, action1, rng, clock)
	require.NoError(t, err)

	// Хід 2: p2 скидає Принцесу і вибуває
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0}
	result2, err := engine.Apply(result1.NewState, action2, rng, clock)
	require.NoError(t, err)

	// Хід 3: p3 теж грає свого Spy з руки (індекс 0)
	action3 := engine.Action{PlayerID: "p3", HandIndex: 0}
	result3, err := engine.Apply(result2.NewState, action3, rng, clock)
	require.NoError(t, err)

	// Завершення раунду
	final := engine.ResolveRoundEnd(result3.NewState, clock).NewState

	// ПЕРЕВІРКА: Оскільки Шпигунів на столі двоє (у p1 та p3), бонус Шпигуна
	// НЕ нараховується жодному з них.
	assert.False(t, final.Players["p1"].SpyPointsAwarded, "Бонус Шпигуна анульовано для p1")
	assert.False(t, final.Players["p3"].SpyPointsAwarded, "Бонус Шпигуна анульовано для p3")

	// Важливо відокремити бонус Шпигуна від бала за перемогу в раунді.
	// Після скидання Шпигунів у руках лишились: p1 = [Princess(9)], p3 = [Guard(1)].
	// Колода порожня -> порівняння карт: p1 перемагає й отримує 1 бал ЗА ПЕРЕМОГУ
	// (а не за Шпигуна). p3 програє й лишається з 0.
	assert.Equal(t, "p1", final.WinnerID, "Переможцем раунду за старшою картою має бути p1")
	assert.Equal(t, 1, final.Players["p1"].Score, "p1 отримує лише бал за перемогу, без бонусу Шпигуна")
	assert.Equal(t, 0, final.Players["p3"].Score, "p3 не отримує жодних балів (Шпигун анульовано, раунд програно)")
}

// TestSpyBonus_SinglePlayerPlayedTwice: один гравець розіграв ДВОХ Шпигунів
// за раунд. Він усе одно єдиний власник Шпигуна й отримує рівно 1 бонусний бал
// (рахуємо гравців, а не карти).
func TestSpyBonus_SinglePlayerPlayedTwice(t *testing.T) {
	state := engine.GameState{
		Phase:     engine.PhaseMainAction,
		TurnOrder: []string{"p1", "p2"},
		Players: map[string]engine.Player{
			// p1 скинув двох Шпигунів і ще тримає Гвардійця (сила 1).
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardGuard},
				DiscardPile: []engine.CardType{engine.CardSpy, engine.CardSpy}, IsOut: false, Score: 0},
			// p2 без Шпигунів, у руці Барон (сила 3) -> p2 виграє раунд за картою.
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardBaron},
				DiscardPile: []engine.CardType{}, IsOut: false, Score: 0},
		},
	}
	clock := &mockClock{fixedTime: time.Now()}

	final := engine.ResolveRoundEnd(state, clock).NewState

	// p1 — єдиний власник Шпигуна -> рівно 1 бонусний бал, попри дві карти.
	assert.True(t, final.Players["p1"].SpyPointsAwarded, "p1 має отримати бонус Шпигуна")
	assert.Equal(t, 1, final.Players["p1"].Score, "Два зіграні Шпигуни дають РІВНО 1 бонусний бал")
	// p2 виграє раунд за старшою картою -> 1 бал за перемогу, без Шпигуна.
	assert.Equal(t, "p2", final.WinnerID)
	assert.False(t, final.Players["p2"].SpyPointsAwarded)
	assert.Equal(t, 1, final.Players["p2"].Score, "p2 отримує лише бал за перемогу")
}

// TestSpyBonus_EliminatedPlayerWinsSpyWhileOtherWinsRound: класичний edge-case
// з вимог — вибулий/відключений гравець отримує бонус Шпигуна, тоді як раунд
// виграє ЗОВСІМ ІНШИЙ (живий) гравець.
func TestSpyBonus_EliminatedPlayerWinsSpyWhileOtherWinsRound(t *testing.T) {
	state := engine.GameState{
		Phase:     engine.PhaseMainAction,
		TurnOrder: []string{"p1", "p2", "p3"},
		Players: map[string]engine.Player{
			// p1 вибув, але встиг скинути Шпигуна цього раунду (рука перенесена у DiscardPile).
			"p1": {ID: "p1", Hand: nil,
				DiscardPile: []engine.CardType{engine.CardSpy, engine.CardPrincess}, IsOut: true, Score: 3},
			// p2 живий, найстарша карта (King=8) -> виграє раунд.
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardKing},
				DiscardPile: []engine.CardType{}, IsOut: false, Score: 2},
			// p3 живий, слабша карта (Priest=2).
			"p3": {ID: "p3", Hand: []engine.CardType{engine.CardPriest},
				DiscardPile: []engine.CardType{}, IsOut: false, Score: 2},
		},
	}
	clock := &mockClock{fixedTime: time.Now()}

	final := engine.ResolveRoundEnd(state, clock).NewState

	// КРИТИЧНО: вибулий p1 — єдиний власник Шпигуна -> отримує бонус, попри IsOut.
	assert.True(t, final.Players["p1"].SpyPointsAwarded, "Вибулий p1 має отримати бонус Шпигуна")
	assert.Equal(t, 4, final.Players["p1"].Score, "p1: 3 + 1 бонус Шпигуна = 4")

	// Раунд виграє інший (живий) гравець p2 за старшою картою.
	assert.Equal(t, "p2", final.WinnerID)
	assert.False(t, final.Players["p2"].SpyPointsAwarded)
	assert.Equal(t, 3, final.Players["p2"].Score, "p2: 2 + 1 за перемогу = 3")
	assert.Equal(t, 2, final.Players["p3"].Score, "p3 без змін")
}

// TestSpyBonus_NobodyPlayedSpy: якщо ніхто не грав Шпигуна, бонус не видається.
func TestSpyBonus_NobodyPlayedSpy(t *testing.T) {
	state := engine.GameState{
		Phase:     engine.PhaseMainAction,
		TurnOrder: []string{"p1", "p2"},
		Players: map[string]engine.Player{
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardKing},
				DiscardPile: []engine.CardType{engine.CardGuard}, Score: 0},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPriest},
				DiscardPile: []engine.CardType{engine.CardBaron}, Score: 0},
		},
	}
	clock := &mockClock{fixedTime: time.Now()}

	final := engine.ResolveRoundEnd(state, clock).NewState

	assert.False(t, final.Players["p1"].SpyPointsAwarded)
	assert.False(t, final.Players["p2"].SpyPointsAwarded)
	// Лише переможець раунду (p1 з King=8) отримує бал.
	assert.Equal(t, "p1", final.WinnerID)
	assert.Equal(t, 1, final.Players["p1"].Score)
	assert.Equal(t, 0, final.Players["p2"].Score)
}

// TestResolveRoundEnd_Idempotent: повторний виклик ResolveRoundEnd на вже
// завершеному раунді НЕ дублює ані бонус Шпигуна, ані бал переможця.
func TestResolveRoundEnd_Idempotent(t *testing.T) {
	state := engine.GameState{
		Phase:     engine.PhaseMainAction,
		TurnOrder: []string{"p1", "p2"},
		Players: map[string]engine.Player{
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardPrincess},
				DiscardPile: []engine.CardType{engine.CardSpy}, Score: 0},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardGuard},
				DiscardPile: []engine.CardType{}, Score: 0},
		},
	}
	clock := &mockClock{fixedTime: time.Now()}

	first := engine.ResolveRoundEnd(state, clock)
	afterFirst := first.NewState
	// p1: 1 (перемога) + 1 (Шпигун) = 2.
	assert.Equal(t, 2, afterFirst.Players["p1"].Score)
	assert.Equal(t, engine.PhaseRoundEnd, afterFirst.Phase)

	// Повторний виклик — жодних змін і жодних нових подій.
	second := engine.ResolveRoundEnd(afterFirst, clock)
	assert.Equal(t, 2, second.NewState.Players["p1"].Score, "Повторний резолв НЕ має подвоювати бали")
	assert.Empty(t, second.DomainEvents, "Повторний резолв НЕ має генерувати нові події")
}
