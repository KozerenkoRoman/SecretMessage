package engine

func ApplyPrince(state GameState, action Action, rng RNG, clock Clock) (ApplyResult, error) {
	shouldApply, err := CanTargetPlayer(state, action.PlayerID, action.TargetID)
	if err != nil {
		return ApplyResult{}, err
	}
	if !shouldApply && (action.TargetID == "" || action.TargetID == "null" || action.TargetID == "undefined") {
		action.TargetID = action.PlayerID
	}

	target, ok := state.Players[action.TargetID]
	if !ok {
		return ApplyResult{}, NewError(ErrTargetNotFound, "target_id=%s", action.TargetID)
	}
	if target.IsOut {
		return ApplyResult{}, NewError(ErrTargetAlreadyOut, "target_id=%s", action.TargetID)
	}
	if target.IsProtected {
		events := []DomainEvent{{
			Type:      EventCardPlayed,
			Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardPrince, TargetID: action.TargetID},
			Timestamp: clock.Now(),
		}}
		return ApplyResult{NewState: state, DomainEvents: events}, nil
	}

	var discardedCard CardType
	hasCardToDiscard := len(target.Hand) > 0
	if hasCardToDiscard {
		discardedCard = target.Hand[0]
	}

	events := []DomainEvent{{
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, TargetID: action.TargetID, Card: CardPrince, DiscardedCard: discardedCard},
		Timestamp: clock.Now(),
	}}

	if hasCardToDiscard {
		target.DiscardPile = append(target.DiscardPile, discardedCard)
		target.Hand = []CardType{}

		if discardedCard == CardPrincess {
			target.IsOut = true
			state.Players[action.TargetID] = target
			events = append(events, DomainEvent{
				Type:      EventPlayerEliminated,
				Payload:   PlayerEliminatedPayload{PlayerID: action.TargetID, Reason: ReasonPrincessPlayed, Card: CardPrincess},
				Timestamp: clock.Now(),
			})
			return ApplyResult{NewState: state, DomainEvents: events}, nil
		}
	}

	if len(state.Deck) > 0 {
		newCard := state.Deck[0]
		state.Deck = state.Deck[1:]
		target.Hand = append(target.Hand, newCard)
		state.Players[action.TargetID] = target
		events = append(events, DomainEvent{
			Type:      EventCardDrawn,
			Payload:   CardDrawnPayload{PlayerID: action.TargetID, Card: newCard},
			Timestamp: clock.Now(),
		})
	} else if state.BurnCard != nil {
		target.Hand = append(target.Hand, *state.BurnCard)
		state.Players[action.TargetID] = target
		events = append(events, DomainEvent{
			Type:      EventCardDrawn,
			Payload:   CardDrawnPayload{PlayerID: action.TargetID, Card: *state.BurnCard},
			Timestamp: clock.Now(),
		})
		state.BurnCard = nil
	}

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
