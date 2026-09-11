package bot

import (
	"errors"
	"math"
	"math/rand"
	"slices"

	"secret-message/cmd/engine"

	"github.com/sirupsen/logrus"
)

type BotBrain struct {
	BotID string
	log   *logrus.Logger
}

func NewBotBrain(botID string, logger *logrus.Logger) *BotBrain {
	return &BotBrain{
		BotID: botID,
		log:   logger,
	}
}

func (b *BotBrain) Think(state *engine.GameState, tracker *DeckTracker) (interface{}, error) {
	me, exists := state.Players[b.BotID]
	if !exists || me.IsOut {
		return nil, errors.New("bot is out of game or not found")
	}
	b.log.WithFields(logrus.Fields{
		"bot_id": b.BotID,
		"phase":  state.Phase,
		"hand":   me.Hand,
	}).Debug("Бот починає аналіз ходу")

	switch state.Phase {
	case engine.PhaseMainAction:
		return b.decideMainAction(state, me, tracker)
	case engine.PhaseResolveChancellor:
		return b.decideChancellorResolve(state, me)
	}

	return nil, errors.New("unsupported game phase for bot thinking")
}

func (b *BotBrain) decideMainAction(state *engine.GameState, me engine.Player, tracker *DeckTracker) (engine.Action, error) {
	memory := NewBotMemory(b.BotID, state, tracker)
	var bestAction engine.Action
	bestScore := -math.MaxFloat64

	// Обираємо найкращу ціль на основі наявної інформації
	targetID := b.findBestTarget(state, memory)
	memory.TargetID = targetID

	// ПЕРЕВІРКА: Чи знає хтось із живих суперників нашу карту?
	isMyCardDisclosed := false
	var disclosedCard engine.CardType
	for enemyID, card := range tracker.AmIDisclosedTo {
		if enemy, ok := state.Players[enemyID]; ok && !enemy.IsOut {
			isMyCardDisclosed = true
			disclosedCard = card
			break
		}
	}

	b.log.WithFields(logrus.Fields{
		"bot_id":         b.BotID,
		"target_chosen":  targetID,
		"is_disclosed":   isMyCardDisclosed,
		"disclosed_card": disclosedCard,
		"unknown_cards":  memory.TotalUnknown,
	}).Debug("Аналіз можливих ходів")

	for idx, card := range me.Hand {
		// --- HARD RULES ---
		if card == engine.CardPrincess {
			continue // Пропускаємо Princess (не можна скидати)
		}
		if err := engine.ValidateCountessRule(me, engine.Action{HandIndex: idx}); err != nil {
			continue // Примусове скидання Countess за правилом
		}
		if card == engine.CardKing && b.hasCardInHand(me, engine.CardGuard) {
			continue // Уникаємо розіграшу King, якщо друга карта Guard
		}

		currentScore := 0.0
		var guessCard engine.CardType
		otherCard := me.Hand[(idx+1)%2]

		// СТРАТЕГІЯ ЕНДШПІЛЮ (Останній хід, в колоді 0 карт)
		if len(state.Deck) == 0 {
			// Цінність ходу пропорційна силі карти, яка ЗАЛИШИТЬСЯ в руці
			currentScore = float64(otherCard) * 20.0

			// Виняток: якщо ми можемо вбити Вартовим прямо зараз — це абсолютний пріоритет
			if card == engine.CardGuard && targetID != "" {
				if knownCard, ok := memory.KnownOpponentCards[targetID]; ok && knownCard != engine.CardGuard {
					guessCard = knownCard
					currentScore = 300.0
				}
			}
		} else {
			// БАЗОВА СТРАТЕГІЯ (Колода не порожня)
			switch card {
			case engine.CardHandmaid:
				currentScore = 85.0
				if isMyCardDisclosed {
					currentScore = 200.0 // Ідеальний порятунок
				}

			case engine.CardSpy:
				currentScore = 30.0
				if !memory.IsSpyBonusContested {
					// Динамічна цінність: чим ближче кінець гри, тим цінніший Шпигун
					deckFactor := 15.0 - float64(len(state.Deck))
					currentScore = 50.0 + (deckFactor * 4.0)
				}

			case engine.CardCountess:
				currentScore = 60.0
				// Блеф розіграшу Графині
				if otherCard != engine.CardKing && otherCard != engine.CardPrince {
					if rand.Float64() < 0.15 {
						currentScore = 145.0
					}
				}

			case engine.CardGuard:
				if targetID != "" {
					if knownCard, ok := memory.KnownOpponentCards[targetID]; ok && knownCard != engine.CardGuard {
						guessCard = knownCard
						currentScore = 160.0 // 100% вбивство
					} else {
						guessCard, currentScore = b.evaluateBestGuardGuess(memory)
						currentScore += 35.0
					}
				} else {
					currentScore = 15.0
				}

			case engine.CardPriest:
				if targetID != "" {
					// Якщо ми вже знаємо карту цієї цілі, цінність Священника падає
					if _, ok := memory.KnownOpponentCards[targetID]; ok {
						currentScore = 10.0
					} else {
						currentScore = 55.0
					}
				}

			case engine.CardBaron:
				if targetID != "" {
					currentScore = b.evaluateBaronRisk(otherCard, memory)
				} else {
					currentScore = 5.0
				}

			case engine.CardPrince:
				if targetID != "" {
					currentScore = 55.0
				} else {
					// Захист від самогубства Принца (всі суперники під ефектом Служниці)
					if otherCard == engine.CardPrincess {
						currentScore = -500.0
					} else if otherCard < engine.CardChancellor {
						currentScore = 90.0 // Вигідно спалити слабку карту під себе
					} else {
						currentScore = 10.0 // Шкода скидати Короля/Графиню
					}
				}

			case engine.CardChancellor:
				currentScore = 65.0

			case engine.CardKing:
				if targetID != "" {
					currentScore = 45.0
					if isMyCardDisclosed && card == disclosedCard {
						currentScore = 160.0 // Здихаємося скомпрометованого Короля
					}
				}
			}
		}

		// СТРАТЕГІЯ СКИДАННЯ КОМПРОМАТУ
		if len(state.Deck) > 0 && isMyCardDisclosed && card == disclosedCard && card != engine.CardGuard {
			currentScore += 95.0
		}

		b.log.WithFields(logrus.Fields{
			"bot_id": b.BotID,
			"card":   card,
			"score":  currentScore,
			"guess":  guessCard,
		}).Debug("Оцінено вагу карти")

		if currentScore > bestScore {
			bestScore = currentScore
			bestAction = engine.Action{
				PlayerID:  b.BotID,
				HandIndex: idx,
				TargetID:  targetID,
				Guess:     guessCard,
			}
		}
	}

	if bestScore == -math.MaxFloat64 {
		bestAction = engine.Action{PlayerID: b.BotID, HandIndex: 0, TargetID: targetID}
	}

	b.log.WithFields(logrus.Fields{
		"bot_id":     b.BotID,
		"play_card":  me.Hand[bestAction.HandIndex],
		"best_score": bestScore,
	}).Info("Бот обрав найкращу дію")

	return bestAction, nil
}

func (b *BotBrain) decideChancellorResolve(state *engine.GameState, me engine.Player) (engine.ChancellorResolveAction, error) {
	if len(me.Hand) == 0 {
		return engine.ChancellorResolveAction{}, errors.New("no cards to resolve for chancellor")
	}

	bestKeepIndex := 0
	highestWeight := -1.0

	// Використовуємо зважену оцінку карт для Канцлера
	for idx, card := range me.Hand {
		weight := getChancellorCardWeight(card, len(state.Deck))
		if weight > highestWeight {
			highestWeight = weight
			bestKeepIndex = idx
		}
	}

	bottomOrder := make([]engine.CardType, 0, 2)
	for idx, card := range me.Hand {
		if idx == bestKeepIndex {
			continue
		}
		bottomOrder = append(bottomOrder, card)
	}

	b.log.WithFields(logrus.Fields{
		"bot_id": b.BotID,
		"keep":   me.Hand[bestKeepIndex],
		"bottom": bottomOrder,
	}).Info("Chancellor: Бот вибрав яку карту залишити")

	return engine.ChancellorResolveAction{
		PlayerID:      b.BotID,
		KeepHandIndex: bestKeepIndex,
		BottomOrder:   bottomOrder,
	}, nil
}

func (b *BotBrain) findBestTarget(state *engine.GameState, memory *BotMemory) string {
	var backupTarget string

	// Пріоритет 1: Шукаємо гравця, чию карту ми знаємо на 100% (і це не Вартовий)
	for id, player := range state.Players {
		if id == b.BotID || player.IsOut || player.IsProtected {
			continue
		}
		if backupTarget == "" {
			backupTarget = id // Запам'ятовуємо хоч когось на випадок відсутності інфи
		}

		if knownCard, ok := memory.KnownOpponentCards[id]; ok && knownCard != engine.CardGuard {
			return id // Знайшли ідеальну мішень для атак
		}
	}

	return backupTarget
}

func (b *BotBrain) evaluateBestGuardGuess(mem *BotMemory) (engine.CardType, float64) {
	if mem.TargetID != "" {
		if lastCard, ok := mem.LastPlayedCard[mem.TargetID]; ok && lastCard == engine.CardCountess {

			// Формуємо пул карт, які зазвичай супроводжують Графиню
			countessTriggers := []engine.CardType{engine.CardPrince, engine.CardKing, engine.CardPrincess}

			var bestCountessGuess engine.CardType = engine.CardPrince
			maxCountessProb := -1.0

			for _, card := range countessTriggers {
				prob := mem.GetCardProbability(card)
				if prob > maxCountessProb {
					maxCountessProb = prob
					bestCountessGuess = card
				}
			}

			// Якщо в колоді ще залишилися ці карти, повертаємо найімовірнішу з них із вагомим бонусом
			if maxCountessProb > 0 {
				b.log.WithFields(logrus.Fields{
					"bot_id":      b.BotID,
					"target_id":   mem.TargetID,
					"reason":      "target_played_countess_before",
					"best_guess":  bestCountessGuess,
					"probability": maxCountessProb,
				}).Debug("Бот помітив скинуту Графиню і фокусується на [5, 7, 9]")
				return bestCountessGuess, maxCountessProb * 150.0 // Коефіцієнт впевненості
			}
		}
	}

	// 2. ДЕФОЛТНА ЛОГІКА (якщо Графині не було):
	var bestGuess engine.CardType = engine.CardPriest
	maxProb := -1.0

	for card := engine.CardSpy; card <= engine.CardPrincess; card++ {
		if card == engine.CardGuard {
			continue
		}
		prob := mem.GetCardProbability(card)
		if prob > maxProb {
			maxProb = prob
			bestGuess = card
		} else if math.Abs(prob-maxProb) < 1e-9 && card == engine.CardPrince {
			bestGuess = engine.CardPrince
		}
	}
	return bestGuess, maxProb * 100.0
}

func (b *BotBrain) evaluateBaronRisk(myOtherCard engine.CardType, mem *BotMemory) float64 {
	baseScore := float64(myOtherCard) * 12.0
	winScore := 0.0
	loseScore := 0.0

	for card := engine.CardSpy; card <= engine.CardPrincess; card++ {
		prob := mem.GetCardProbability(card)

		if myOtherCard > card {
			winScore += prob * 90.0
		} else if myOtherCard < card {
			loseScore += prob * 120.0
		}
	}

	return baseScore + winScore - loseScore
}

func (b *BotBrain) hasCardInHand(p engine.Player, card engine.CardType) bool {
	return slices.Contains(p.Hand, card)
}

func getChancellorCardWeight(card engine.CardType, deckSize int) float64 {
	switch card {
	case engine.CardPrincess:
		if deckSize > 6 {
			return 2.0 // Початок гри — ховаємо Принцесу глибше
		}
		return 9.5 // Ендшпіль — Принцеса принесе нам перемогу за очками
	case engine.CardCountess:
		return 8.5
	case engine.CardHandmaid:
		return 8.0
	case engine.CardChancellor:
		return 4.0
	default:
		return float64(card)
	}
}
