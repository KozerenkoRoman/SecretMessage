// apply_countess.go
package engine

// ValidateCountessRule перевіряє, чи не порушує гравець обов'язкове правило скидання Графині.
// Ця функція викликається на початку головного Apply для БУДЬ-ЯКОЇ карти.
//
// При порушенні повертає типізовану *GameError з кодом ErrMustPlayCountess —
// фронтенд покаже користувачу локалізоване повідомлення.
func ValidateCountessRule(player Player, action Action) error {
	var hasCountess bool
	var hasKingOrPrince bool

	// Перевіряємо поточний склад руки гравця
	for _, card := range player.Hand {
		if card == CardCountess {
			hasCountess = true
		}
		if card == CardKing || card == CardPrince {
			hasKingOrPrince = true
		}
	}

	// Якщо у гравця комбінація Графиня + (Король або Принц)
	if hasCountess && hasKingOrPrince {
		// Дивимось, яку карту гравець намагається розіграти за переданим індексом
		attemptedCard := player.Hand[action.HandIndex]

		// Якщо він намагається зіграти щось інше, крім Графині — блокуємо хід
		if attemptedCard != CardCountess {
			return NewError(ErrMustPlayCountess,
				"player has Countess + King/Prince but tried to play card=%d", attemptedCard)
		}
	}

	return nil
}

// ApplyCountess реалізує безпосередній ефект розіграшу карти Countess
func ApplyCountess(state GameState, action Action, clock Clock, startEventID uint64) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	// Скидаємо Countess з руки
	var newHand []CardType
	for _, c := range player.Hand {
		if c != CardCountess {
			newHand = append(newHand, c)
		} else {
			player.DiscardPile = append(player.DiscardPile, c)
		}
	}
	player.Hand = newHand
	state.Players[action.PlayerID] = player

	events := []DomainEvent{{
		EventID:   startEventID,
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardCountess},
		Timestamp: clock.Now(),
	}}

	return ApplyResult{
		NewState:     state,
		DomainEvents: events,
	}, nil
}
