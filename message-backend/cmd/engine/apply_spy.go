package engine

import "slices"

func ApplySpy(state GameState, action Action, clock Clock) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	events := []DomainEvent{{
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardSpy},
		Timestamp: clock.Now(),
	}}

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}

func ResolveSpyBonus(state GameState) GameState {
	activeSpyOwners := make([]string, 0)
	for id, p := range state.Players {
		p.SpyPointsAwarded = false
		state.Players[id] = p
	}
	for id, p := range state.Players {
		if slices.Contains(p.DiscardPile, CardSpy) && !p.IsOut {
			activeSpyOwners = append(activeSpyOwners, id)
		}
	}
	if len(activeSpyOwners) == 1 {
		winnerID := activeSpyOwners[0]
		p := state.Players[winnerID]
		p.SpyPointsAwarded = true
		p.Score++
		state.Players[winnerID] = p
	}
	return state
}
