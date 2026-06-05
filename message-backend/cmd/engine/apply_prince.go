package engine

// ApplyPrince реалізує ефект карти Prince
func ApplyPrince(state GameState, action Action, rng RNG, clock Clock, startEventID uint64) (ApplyResult, error) {
	shouldApply, err := CanTargetPlayer(state, action.PlayerID, action.TargetID)
	if err != nil {
		return ApplyResult{}, err
	}

	// Якщо всі інші захищені, а фронтенд прислав порожню ціль, Принц автоматично б'є по самому собі
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
		// Якщо таргет під Handmaid — ефект ігнорується
		events := []DomainEvent{{
			EventID:   startEventID,
			Type:      EventCardPlayed,
			Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardPrince, TargetID: action.TargetID},
			Timestamp: clock.Now(),
		}}
		return ApplyResult{NewState: state, DomainEvents: events}, nil
	}

	events := []DomainEvent{{
		EventID:   startEventID,
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardPrince, TargetID: action.TargetID},
		Timestamp: clock.Now(),
	}}

	// Скидання карти з руки таргета
	if len(target.Hand) > 0 {
		discarded := target.Hand[0]
		target.DiscardPile = append(target.DiscardPile, discarded)
		target.Hand = []CardType{}

		// Якщо скинута Princess → миттєвий програш
		if discarded == CardPrincess {
			target.IsOut = true
			state.Players[action.TargetID] = target

			events = append(events, DomainEvent{
				EventID:   startEventID + 1,
				Type:      EventPlayerEliminated,
				Payload:   PlayerEliminatedPayload{PlayerID: action.TargetID, Reason: ReasonPrincessPlayed},
				Timestamp: clock.Now(),
			})

			return ApplyResult{NewState: state, DomainEvents: events}, nil
		}
	}

	// Добір нової карти
	if len(state.Deck) > 0 {
		drawIndex := rng.Intn(len(state.Deck))
		newCard := state.Deck[drawIndex]

		// Видаляємо карту з колоди
		state.Deck = append(state.Deck[:drawIndex], state.Deck[drawIndex+1:]...)
		target.Hand = append(target.Hand, newCard)
		state.Players[action.TargetID] = target

		events = append(events, DomainEvent{
			EventID:   startEventID + 2,
			Type:      EventCardDrawn,
			Payload:   CardDrawnPayload{PlayerID: action.TargetID, Card: newCard},
			Timestamp: clock.Now(),
		})
	} else {
		// Якщо колода порожня → таргет бере BurnCard
		if state.BurnCard != nil {
			target.Hand = append(target.Hand, *state.BurnCard)
			state.Players[action.TargetID] = target

			events = append(events, DomainEvent{
				EventID:   startEventID + 2,
				Type:      EventCardDrawn,
				Payload:   CardDrawnPayload{PlayerID: action.TargetID, Card: *state.BurnCard},
				Timestamp: clock.Now(),
			})
			// BurnCard використана
			state.BurnCard = nil
		}
	}

	return ApplyResult{
		NewState:     state,
		DomainEvents: events,
	}, nil
}
