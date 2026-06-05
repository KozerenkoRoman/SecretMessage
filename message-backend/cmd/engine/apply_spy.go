package engine

func ApplySpy(state GameState, action Action, clock Clock, startEventID uint64) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	var newHand []CardType
	for _, c := range player.Hand {
		if c != CardSpy {
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
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardSpy},
		Timestamp: clock.Now(),
	}}

	return ApplyResult{
		NewState:     state,
		DomainEvents: events,
	}, nil
}

func ResolveSpyBonus(state GameState) GameState {
	spyOwners := []string{}

	for id, p := range state.Players {
		for _, c := range p.DiscardPile {
			if c == CardSpy {
				spyOwners = append(spyOwners, id)
				break
			}
		}
	}

	// Бонус нараховується ЛІШЕ якщо Шпигун у відбої є рівно в ОДНОГО гравця
	if len(spyOwners) == 1 {
		winnerID := spyOwners[0]
		p := state.Players[winnerID]
		p.SpyPointsAwarded = true
		p.Score++ // Обов'язково додаємо очко в структуру даних сервера!
		state.Players[winnerID] = p
	}

	return state
}
