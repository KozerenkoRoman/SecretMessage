package engine

func ApplyPrincess(state GameState, action Action, clock Clock, startEventID uint64) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	events := []DomainEvent{{
		EventID:   startEventID,
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardPrincess},
		Timestamp: clock.Now(),
	}}

	// Princess → миттєве вибуття
	player.IsOut = true
	player.DiscardPile = append(player.DiscardPile, CardPrincess)
	player.Hand = []CardType{}
	state.Players[action.PlayerID] = player

	events = append(events, DomainEvent{
		EventID:   startEventID + 1,
		Type:      EventPlayerEliminated,
		Payload:   PlayerEliminatedPayload{PlayerID: action.PlayerID, Reason: ReasonPrincessPlayed},
		Timestamp: clock.Now(),
	})

	return ApplyResult{
		NewState:     state,
		DomainEvents: events,
	}, nil
}
