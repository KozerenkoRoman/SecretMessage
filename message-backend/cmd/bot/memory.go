package bot

import "secret-message/cmd/engine"

var CardRegistry = map[engine.CardType]int{
	engine.CardSpy:        2,
	engine.CardGuard:      6,
	engine.CardPriest:     2,
	engine.CardBaron:      2,
	engine.CardHandmaid:   2,
	engine.CardPrince:     2,
	engine.CardChancellor: 2,
	engine.CardKing:       1,
	engine.CardCountess:   1,
	engine.CardPrincess:   1,
}

type BotMemory struct {
	BotID               string
	VisibleCounts       map[engine.CardType]int
	TotalUnknown        int
	KnownOpponentCards  map[string]engine.CardType
	LastPlayedCard      map[string]engine.CardType // остання зіграна карта
	TargetID            string                     // поточна ціль бота
	IsSpyBonusContested bool
}

func NewBotMemory(botID string, state *engine.GameState, tracker *DeckTracker) *BotMemory {
	mem := &BotMemory{
		BotID:              botID,
		VisibleCounts:      make(map[engine.CardType]int),
		KnownOpponentCards: tracker.KnownOpponentCards,
		LastPlayedCard:     tracker.LastPlayedCard, // Лінк на трекер
	}

	spyDiscardedCount := 0
	for _, player := range state.Players {
		for _, card := range player.DiscardPile {
			mem.VisibleCounts[card]++
			if card == engine.CardSpy {
				spyDiscardedCount++
			}
		}
		if player.IsOut && len(player.Hand) > 0 {
			mem.VisibleCounts[player.Hand[0]]++
		}
	}

	if spyDiscardedCount > 0 {
		mem.IsSpyBonusContested = true
	}

	visibleSum := 0
	for _, count := range mem.VisibleCounts {
		visibleSum += count
	}

	myHandSize := 0
	if me, exists := state.Players[botID]; exists {
		myHandSize = len(me.Hand)
		for _, card := range me.Hand {
			mem.VisibleCounts[card]++
		}
	}

	mem.TotalUnknown = max(21-visibleSum-myHandSize, 0)
	return mem
}

func (m *BotMemory) GetCardProbability(card engine.CardType) float64 {
	if m.TotalUnknown <= 0 {
		return 0.0
	}
	maxInDeck := CardRegistry[card]
	visible := m.VisibleCounts[card]
	left := maxInDeck - visible
	if left <= 0 {
		return 0.0
	}
	return float64(left) / float64(m.TotalUnknown)
}
