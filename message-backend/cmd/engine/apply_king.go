package engine

func ApplyKing(state GameState, action Action, clock Clock, startEventID uint64) (ApplyResult, error) {
	shouldApply, err := CanTargetPlayer(state, action.PlayerID, action.TargetID)
	if err != nil {
		return ApplyResult{}, err
	}

	events := []DomainEvent{{
		EventID:   startEventID,
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardKing, TargetID: action.TargetID},
		Timestamp: clock.Now(),
	}}

	// Якщо всі захищені — просто скидаємо карту
	if !shouldApply {
		return ApplyResult{NewState: state, DomainEvents: events}, nil
	}

	if action.TargetID == action.PlayerID {
		return ApplyResult{}, NewError(ErrCannotTargetSelf, "card=King")
	}

	target, ok := state.Players[action.TargetID]
	if !ok {
		return ApplyResult{}, NewError(ErrTargetNotFound, "target_id=%s", action.TargetID)
	}
	if target.IsOut {
		return ApplyResult{}, NewError(ErrTargetAlreadyOut, "target_id=%s", action.TargetID)
	}

	player := state.Players[action.PlayerID]
	playerHand := player.Hand
	targetHand := target.Hand

	player.Hand = targetHand
	target.Hand = playerHand

	state.Players[action.PlayerID] = player
	state.Players[action.TargetID] = target

	events = append(events, DomainEvent{
		EventID:   startEventID + 1,
		Type:      EventHandsSwapped,
		Payload:   HandsSwappedPayload{PlayerID: action.PlayerID, TargetID: action.TargetID},
		Timestamp: clock.Now(),
	})

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
