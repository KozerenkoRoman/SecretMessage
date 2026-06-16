package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMultiRoundSuite налаштовує початковий стан для першого раунду.
func setupMultiRoundSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2028,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard, engine.CardPriest, engine.CardBaron},
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardSpy}, IsOut: false, Score: 0},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false, Score: 0},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestMultiRoundSimulation перевіряє повний цикл роботи гри протягом двох раундів.
func TestMultiRoundSimulation(t *testing.T) {
	state, clock, rng := setupMultiRoundSuite()

	// --- РАУНД 1 ---
	// Крок 1: p1 грає Шпигуна (Spy)
	action1 := engine.Action{PlayerID: "p1", HandIndex: 0}
	result1, err := engine.Apply(state, action1, rng, clock)
	require.NoError(t, err)

	// Крок 2: p2 грає Принцесу (Princess) і вибуває
	action2 := engine.Action{PlayerID: "p2", HandIndex: 0}
	result2, err := engine.Apply(result1.NewState, action2, rng, clock)
	require.NoError(t, err)

	// Завершуємо перший раунд підрахунком балів
	final1 := engine.ResolveRoundEnd(result2.NewState, clock).NewState

	// Перевірка балів після Раунду 1 (Перемога + Шпигунський бонус = 2 очки)
	assert.True(t, final1.Players["p1"].SpyPointsAwarded, "p1 має отримати бонус Spy")
	assert.Equal(t, 2, final1.Players["p1"].Score, "p1 має отримати 2 очки (перемога + шпигунський бонус)")

	// --- РАУНД 2 ---
	// Створюємо стан для початку другого раунду з урахуванням збереження очок.
	newRound := engine.GameState{
		Seed:        2029,
		Sequence:    final1.Sequence + 1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard, engine.CardPriest},
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			// Даємо p1 дві карти Guard. Після розіграшу однієї, друга залишиться для порівняння з Бароном
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardGuard, engine.CardGuard}, IsOut: false, Score: final1.Players["p1"].Score},
			// Даємо p2 дві карти: Барона для ходу та Принца/Священника для порівняння
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardBaron, engine.CardPriest}, IsOut: false, Score: final1.Players["p2"].Score},
		},
	}

	// Крок 3: p1 грає першого Guard (індекс 0) проти p2, але називає неправильну карту (промах)
	action3 := engine.Action{PlayerID: "p1", HandIndex: 0, TargetID: "p2", Guess: engine.CardPrincess}
	result3, err := engine.Apply(newRound, action3, rng, clock)
	require.NoError(t, err)

	// Крок 4: p2 грає Барона (Baron, індекс 0) проти p1.
	// Тепер у p1 в руці залишився Guard (сила 1), а у p2 в руці залишився Priest (сила 2). Порівняння можливе!
	action4 := engine.Action{PlayerID: "p2", HandIndex: 0, TargetID: "p1"}
	result4, err := engine.Apply(result3.NewState, action4, rng, clock)
	require.NoError(t, err)

	// Завершуємо другий раунд
	final2 := engine.ResolveRoundEnd(result4.NewState, clock).NewState

	// Перевірка фіналу: в результаті боїв має залишитися один активний гравець
	activeCount := 0
	for _, p := range final2.Players {
		if !p.IsOut {
			activeCount++
		}
	}
	assert.Equal(t, 1, activeCount, "У кінці раунду має лишитися один активний гравець")

	// Перевірка загального рахунку
	assert.True(t, final2.Players["p1"].Score > 0 || final2.Players["p2"].Score > 0, "Хтось має отримати очки після двох раундів")
}
