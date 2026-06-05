package engine

// ApplyHandmaid реалізує ефект карти Handmaid
func ApplyHandmaid(state GameState, action Action, clock Clock, startEventID uint64) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	// Гравець зіграв Handmaid
	events := []DomainEvent{{
		EventID:   startEventID,
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardHandmaid},
		Timestamp: clock.Now(),
	}}

	// Ефект: гравець отримує захист до наступного ходу
	player.IsProtected = true
	state.Players[action.PlayerID] = player

	return ApplyResult{
		NewState:     state,
		DomainEvents: events,
	}, nil
}
