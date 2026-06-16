package bot

import "secret-message/cmd/engine"

type DeckTracker struct {
	BotID              string
	KnownBottomCards   []engine.CardType
	KnownOpponentCards map[string]engine.CardType
	// Нове поле: хто з гравців знає нашу конкретну карту
	AmIDisclosedTo map[string]engine.CardType
}

func NewDeckTracker(botID string) *DeckTracker {
	return &DeckTracker{
		BotID:              botID,
		KnownBottomCards:   make([]engine.CardType, 0),
		KnownOpponentCards: make(map[string]engine.CardType),
		AmIDisclosedTo:     make(map[string]engine.CardType),
	}
}

func (t *DeckTracker) Reset() {
	t.KnownBottomCards = make([]engine.CardType, 0)
	t.KnownOpponentCards = make(map[string]engine.CardType)
	t.AmIDisclosedTo = make(map[string]engine.CardType)
}

// RecordChancellorAction фіксує порядок карт, скинутих Канцлером
func (t *DeckTracker) RecordChancellorAction(action engine.ChancellorResolveAction) {
	if action.PlayerID != t.BotID {
		return
	}
	// Карти, які лягли на дно колоди (остання в масиві лежить найглибше)
	for i := len(action.BottomOrder) - 1; i >= 0; i-- {
		t.KnownBottomCards = append(t.KnownBottomCards, action.BottomOrder[i])
	}
}

// TrackGameState коригує трекер при вичерпанні колоди
func (t *DeckTracker) TrackGameState(state *engine.GameState) {
	deckSize := len(state.Deck)

	// Якщо гравець вилетів, видаляємо його з пулу відомих рук
	for id, p := range state.Players {
		if p.IsOut {
			delete(t.KnownOpponentCards, id)
			delete(t.AmIDisclosedTo, id) // Якщо ворог вибув, він більше не загроза
		}
	}

	// Слідкуємо за залишком відомого нам дна колоди
	if deckSize < len(t.KnownBottomCards) && deckSize > 0 {
		t.KnownBottomCards = t.KnownBottomCards[:deckSize]
	} else if deckSize == 0 {
		t.KnownBottomCards = make([]engine.CardType, 0)
	}

	// Обчислюємо, чи взяв суперник карту з нашого прорахованого дна колоди
	activePlayerID := state.TurnOrder[state.CurrentTurn]

	// Якщо настав хід нашого бота, ворог тепер знає лише 50% (бо ми взяли другу карту).
	// Для агресивного захисту ми можемо зберігати статус «розкритий», але очистимо його,
	// якщо події показують, що загроза минула.
	if state.Phase == engine.PhaseMainAction && activePlayerID == t.BotID {
		// Залишаємо AmIDisclosedTo для аналізу всередині decideMainAction.
		// Він очиститься в HandleDomainEvent, коли карту буде зіграно.
	}

	if state.Phase == engine.PhaseMainAction && activePlayerID != t.BotID {
		if len(t.KnownBottomCards) > 0 && len(state.Deck) == len(t.KnownBottomCards)-1 {
			drawnFromBottom := t.KnownBottomCards[len(t.KnownBottomCards)-1]
			t.KnownOpponentCards[activePlayerID] = drawnFromBottom
		}
	}
}

// HandleDomainEvent тепер використовує твої рідні константи EventType
func (t *DeckTracker) HandleDomainEvent(event engine.DomainEvent) {
	switch event.Type {
	case engine.EventCardPlayed:
		if payload, ok := event.Payload.(engine.CardPlayedPayload); ok {
			// Сценарій 1: Гравець зіграв карту зі своєї руки.
			// Якщо це не наш бот — очищуємо пам'ять про нього, бо його рука змінилася.
			if payload.PlayerID != t.BotID {
				delete(t.KnownOpponentCards, payload.PlayerID)
			}

			// Сценарій 2: Був зіграний Принц, який змусив жертву скинути карту.
			// Перевіряємо, чи є TargetID, чи це не сам бот, і чи дійсно карта була скинута.
			if payload.Card == engine.CardPrince && payload.TargetID != "" && payload.TargetID != t.BotID {
				delete(t.KnownOpponentCards, payload.TargetID)
			}

			// Логіка для самого бота: якщо бот скинув карту, яку знали вороги,
			// видаляємо її зі списку "я розкритий перед ворогами".
			if payload.PlayerID == t.BotID {
				for enemyID, knownCard := range t.AmIDisclosedTo {
					if knownCard == payload.Card {
						delete(t.AmIDisclosedTo, enemyID)
					}
				}
			}
		}

	case engine.EventHandsSwapped:
		if payload, ok := event.Payload.(engine.HandsSwappedPayload); ok {
			// Міняємо місцями наші знання про карти гравців
			card1, ok1 := t.KnownOpponentCards[payload.PlayerID]
			card2, ok2 := t.KnownOpponentCards[payload.TargetID]

			if ok1 {
				t.KnownOpponentCards[payload.TargetID] = card1
			} else {
				delete(t.KnownOpponentCards, payload.TargetID)
			}

			if ok2 {
				t.KnownOpponentCards[payload.PlayerID] = card2
			} else {
				delete(t.KnownOpponentCards, payload.PlayerID)
			}

			// Логіка для AmIDisclosedTo: при обміні Королем карти міняються!
			if payload.PlayerID == t.BotID {
				// Ми віддали карту TargetID. Тепер TargetID точно знає, що у нього наша стара карта,
				// а інші гравці, які знали нашу карту, тепер мають хибну інформацію.
				t.AmIDisclosedTo = make(map[string]engine.CardType)
				// Але тепер TargetID знає нашу нову карту? Ні, бо він дав нам свою наосліп.
				// Проте TargetID знає, яку карту отримав ВІД нас.
			} else if payload.TargetID == t.BotID {
				t.AmIDisclosedTo = make(map[string]engine.CardType)
			}
		}

	case engine.EventPriestEffect:
		if payload, ok := event.Payload.(engine.PriestEffectPayload); ok {
			// Якщо підглядали МИ, записуємо результат у пам'ять
			if payload.ViewerID == t.BotID {
				t.KnownOpponentCards[payload.TargetID] = payload.Card
			}
			// ОПАНА! Хтось подивився карту нашого бота за допомогою Священника!
			if payload.TargetID == t.BotID && payload.ViewerID != t.BotID {
				t.AmIDisclosedTo[payload.ViewerID] = payload.Card
			}
		}
	}
}
