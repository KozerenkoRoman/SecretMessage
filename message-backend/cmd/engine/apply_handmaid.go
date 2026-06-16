package engine

func ApplyHandmaid(state GameState, action Action, clock Clock) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	events := []DomainEvent{{
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardHandmaid},
		Timestamp: clock.Now(),
	}}

	player.IsProtected = true
	state.Players[action.PlayerID] = player

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
