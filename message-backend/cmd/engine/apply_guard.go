package engine

func ApplyGuard(state GameState, action Action, clock Clock, startEventID uint64) (ApplyResult, error) {
	shouldApply, err := CanTargetPlayer(state, action.PlayerID, action.TargetID)
	if err != nil {
		return ApplyResult{}, err
	}

	events := []DomainEvent{{
		EventID:   startEventID,
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardGuard, TargetID: action.TargetID},
		Timestamp: clock.Now(),
	}}

	// Якщо ефект не застосовується (всі захищені)
	if !shouldApply {
		return ApplyResult{NewState: state, DomainEvents: events}, nil
	}

	if action.TargetID == action.PlayerID {
		return ApplyResult{}, NewError(ErrCannotTargetSelf, "card=Guard")
	}

	target, ok := state.Players[action.TargetID]
	if !ok {
		return ApplyResult{}, NewError(ErrTargetNotFound, "target_id=%s", action.TargetID)
	}
	if target.IsOut {
		return ApplyResult{}, NewError(ErrTargetAlreadyOut, "target_id=%s", action.TargetID)
	}

	if action.Guess == CardGuard {
		return ApplyResult{}, NewError(ErrGuardCannotGuessGuard, "")
	}

	if len(target.Hand) > 0 && target.Hand[0] == action.Guess {
		target.IsOut = true
		state.Players[action.TargetID] = target
		events = append(events,
			DomainEvent{
				EventID: startEventID + 1,
				Type:    EventGuardHit,
				Payload: GuardHitPayload{
					PlayerID: action.PlayerID,
					TargetID: action.TargetID,
					Guess:    action.Guess,
				},
				Timestamp: clock.Now(),
			},
			DomainEvent{
				EventID: startEventID + 2,
				Type:    EventPlayerEliminated,
				Payload: PlayerEliminatedPayload{
					PlayerID: action.TargetID,
					Reason:   ReasonGuardHit,
				},
				Timestamp: clock.Now(),
			},
		)
	} else {
		events = append(events, DomainEvent{
			EventID: startEventID + 1,
			Type:    EventGuardMiss,
			Payload: GuardMissPayload{
				PlayerID: action.PlayerID,
				TargetID: action.TargetID,
				Guess:    action.Guess,
			},
			Timestamp: clock.Now(),
		})
	}

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
