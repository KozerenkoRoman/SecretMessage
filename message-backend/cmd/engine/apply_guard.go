package engine

func ApplyGuard(state GameState, action Action, clock Clock) (ApplyResult, error) {
	shouldApply, err := CanTargetPlayer(state, action.PlayerID, action.TargetID)
	if err != nil {
		return ApplyResult{}, err
	}

	events := []DomainEvent{{
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardGuard, TargetID: action.TargetID},
		Timestamp: clock.Now(),
	}}

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
		target.DiscardPile = append(target.DiscardPile, target.Hand...)
		target.Hand = []CardType{}
		state.Players[action.TargetID] = target

		events = append(events, DomainEvent{
			Type:      EventGuardHit,
			Payload:   GuardHitPayload{PlayerID: action.PlayerID, TargetID: action.TargetID, Guess: action.Guess},
			Timestamp: clock.Now(),
		}, DomainEvent{
			Type:      EventPlayerEliminated,
			Payload:   PlayerEliminatedPayload{PlayerID: action.TargetID, Reason: ReasonGuardHit},
			Timestamp: clock.Now(),
		})
	} else {
		events = append(events, DomainEvent{
			Type:      EventGuardMiss,
			Payload:   GuardMissPayload{PlayerID: action.PlayerID, TargetID: action.TargetID, Guess: action.Guess},
			Timestamp: clock.Now(),
		})
	}

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
