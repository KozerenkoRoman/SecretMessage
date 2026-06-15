package engine

import "slices"

func ApplySpy(state GameState, action Action, clock Clock, startEventID uint64) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	events := []DomainEvent{
		{
			EventID:   startEventID,
			Type:      EventCardPlayed,
			Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardSpy},
			Timestamp: clock.Now(),
		},
	}

	return ApplyResult{
		NewState:     state,
		DomainEvents: events,
	}, nil
}

func ResolveSpyBonus(state GameState) GameState {
	activeSpyOwners := make([]string, 0)

	// 1. Спочатку скидаємо прапорець абсолютно всім гравцям перед перевіркою
	for id, p := range state.Players {
		p.SpyPointsAwarded = false
		state.Players[id] = p
	}

	// 2. Шукаємо тих, хто зіграв шпигуна (НЕВАЖЛИВО, чи гравець IsOut чи ні)
	for id, p := range state.Players {
		hasSpy := slices.Contains(p.DiscardPile, CardSpy)
		if hasSpy {
			activeSpyOwners = append(activeSpyOwners, id)
		}
	}

	// 3. Якщо такий гравець рівно один — нараховуємо бал
	if len(activeSpyOwners) == 1 {
		winnerID := activeSpyOwners[0]
		p := state.Players[winnerID]
		p.SpyPointsAwarded = true
		p.Score++
		state.Players[winnerID] = p
	}

	return state
}
