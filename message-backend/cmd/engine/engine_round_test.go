package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRoundSuite готує чистий стан гри для симуляції одного раунду.
// У нас є два гравці: p1 (зараз його хід) та p2.
func setupRoundSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        777,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard, engine.CardPriest, engine.CardBaron, engine.CardPrincess},
		TurnOrder:   []string{"p1", "p2"},
		CurrentTurn: 0, // Ходить p1
		Players: map[string]engine.Player{
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardGuard}, IsOut: false},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: false},
		},
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

// TestFullRoundSimulation перевіряє повний цикл ходу, автоматичний пропуск
// гравців, які вибули, та успішне завершення раунду.
func TestFullRoundSimulation(t *testing.T) {
	state, clock, rng := setupRoundSuite()

	// Хід 1: p1 грає картку Guard (Вартовий) проти p2 і вгадує, що у p2 картка Princess.
	action1 := engine.Action{
		PlayerID:  "p1",
		HandIndex: 0,
		TargetID:  "p2",
		Guess:     engine.CardPrincess, // Правильне вгадування!
	}

	// Застосовуємо дію до нашого стану
	result1, err := engine.Apply(state, action1, rng, clock)
	require.NoError(t, err, "Дія має виконатися без помилок")

	finalState := result1.NewState

	// 1. Перевірка: p2 мав вибути, бо його карту вгадали
	assert.True(t, finalState.Players["p2"].IsOut, "Гравець p2 має вибути після того, як вгадали його Принцесу")

	// 2. Перевірка черговості ходу:
	// Рушій гри автоматично побачив, що p2 вибув, і перевів хід далі по колу — знову на p1.
	// Тому CurrentTurn має дорівнювати 0 (індекс гравця p1 у масиві TurnOrder).
	assert.Equal(t, 0, finalState.CurrentTurn, "Хід має автоматично повернутися до p1 (індекс 0), оскільки p2 вибув")
	assert.Equal(t, "p1", finalState.TurnOrder[finalState.CurrentTurn], "Поточний активний гравець у TurnOrder — p1")

	// 3. Перевірка завершення раунду:
	// Оскільки живим залишився лише 1 гравець, рушій наприкінці функції Apply автоматично
	// викликає функцію завершення раунду (ResolveRoundEnd).
	// Перевіримо, чи змінилася фаза гри на ROUND_END.
	assert.Equal(t, engine.PhaseRoundEnd, finalState.Phase, "Раунд має автоматично завершитися")

	// 4. Рахуємо кількість активних гравців
	activeCount := 0
	for _, p := range finalState.Players {
		if !p.IsOut {
			activeCount++
		}
	}
	assert.Equal(t, 1, activeCount, "У кінці раунду має лишитися рівно один активний гравець (переможець)")

	// 5. Перевіряємо, чи визначено переможця раунду
	assert.Equal(t, "p1", finalState.WinnerID, "Переможцем раунду має стати гравець p1")
}
