package engine

func ValidateCountessRule(player Player, action Action) error {
	var hasCountess bool
	var hasKingOrPrince bool
	for _, card := range player.Hand {
		if card == CardCountess {
			hasCountess = true
		}
		if card == CardKing || card == CardPrince {
			hasKingOrPrince = true
		}
	}
	if hasCountess && hasKingOrPrince {
		attemptedCard := player.Hand[action.HandIndex]
		if attemptedCard != CardCountess {
			return NewError(ErrMustPlayCountess, "player has Countess + King/Prince but tried to play card=%d", attemptedCard)
		}
	}
	return nil
}

func ApplyCountess(state GameState, action Action, clock Clock) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	// Оскільки engine.go вже прибрав карту з руки та поклав у DiscardPile,
	// тут нам потрібно лише сгенерувати івент.
	events := []DomainEvent{{
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardCountess},
		Timestamp: clock.Now(),
	}}
	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
