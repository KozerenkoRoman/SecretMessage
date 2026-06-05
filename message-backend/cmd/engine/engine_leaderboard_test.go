package engine_test

import (
	"testing"
	"time"

	"secret-message/cmd/engine"

	"github.com/stretchr/testify/assert"
)

func setupLeaderboardSuite() (engine.GameState, *mockClock, *mockRNG) {
	return engine.GameState{
		Seed:        2031,
		Sequence:    1,
		Phase:       engine.PhaseMainAction,
		Deck:        []engine.CardType{engine.CardGuard, engine.CardPriest},
		TurnOrder:   []string{"p1", "p2", "p3"},
		CurrentTurn: 0,
		Players: map[string]engine.Player{
			"p1": {ID: "p1", Hand: []engine.CardType{engine.CardSpy}, IsOut: false, Score: 7},
			"p2": {ID: "p2", Hand: []engine.CardType{engine.CardPrincess}, IsOut: true, Score: 3},
			"p3": {ID: "p3", Hand: []engine.CardType{engine.CardBaron}, IsOut: true, Score: 5},
		},
		IsGameOver: true,
		WinnerID:   "p1",
	}, &mockClock{fixedTime: time.Now()}, &mockRNG{}
}

func TestLeaderboardAfterGameEnd(t *testing.T) {
	state, _, _ := setupLeaderboardSuite()

	// Перевірка: гра завершена
	assert.True(t, state.IsGameOver, "Гра має бути завершена")
	assert.Equal(t, "p1", state.WinnerID, "Переможцем має бути p1")

	// Перевірка очок напряму
	assert.Equal(t, 7, state.Players["p1"].Score, "p1 має 7 очок")
	assert.Equal(t, 3, state.Players["p2"].Score, "p2 має 3 очки")
	assert.Equal(t, 5, state.Players["p3"].Score, "p3 має 5 очок")

	// Перевірка: переможець має найбільший рахунок
	maxScore := 0
	winner := ""
	for id, p := range state.Players {
		if p.Score > maxScore {
			maxScore = p.Score
			winner = id
		}
	}
	assert.Equal(t, state.WinnerID, winner, "Переможець у таблиці має збігатися з WinnerID")
}
