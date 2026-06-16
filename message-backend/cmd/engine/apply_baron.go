package engine

func ApplyBaron(state GameState, action Action, clock Clock) (ApplyResult, error) {
	shouldApply, err := CanTargetPlayer(state, action.PlayerID, action.TargetID)
	if err != nil {
		return ApplyResult{}, err
	}

	events := []DomainEvent{{
		Type: EventCardPlayed,
		Payload: CardPlayedPayload{
			PlayerID: action.PlayerID,
			Card:     CardBaron,
			TargetID: action.TargetID,
		},
		Timestamp: clock.Now(),
	}}

	if !shouldApply {
		return ApplyResult{NewState: state, DomainEvents: events}, nil
	}

	if action.TargetID == action.PlayerID {
		return ApplyResult{}, NewError(ErrCannotTargetSelf, "card=Baron")
	}

	target, ok := state.Players[action.TargetID]
	if !ok {
		return ApplyResult{}, NewError(ErrTargetNotFound, "target_id=%s", action.TargetID)
	}
	if target.IsOut {
		return ApplyResult{}, NewError(ErrTargetAlreadyOut, "target_id=%s", action.TargetID)
	}

	player := state.Players[action.PlayerID]
	var playerCard CardType

	// Розумне визначення карти з урахуванням дублікатів Баронів
	if len(player.Hand) == 2 {
		if player.Hand[0] == CardBaron {
			playerCard = player.Hand[1]
		} else {
			playerCard = player.Hand[0]
		}
	} else {
		if len(player.Hand) == 0 || len(target.Hand) == 0 {
			return ApplyResult{}, NewError(ErrBaronNoCardsToCompare, "player_hand=%d,target_hand=%d", len(player.Hand), len(target.Hand))
		}
		playerCard = player.Hand[0]
	}

	targetCard := target.Hand[0]

	events = append(events, DomainEvent{
		Type: EventRoundCompared,
		Payload: RoundComparedPayload{
			PlayerID:   action.PlayerID,
			TargetID:   action.TargetID,
			PlayerCard: playerCard,
			TargetCard: targetCard,
		},
		Timestamp: clock.Now(),
	})

	if playerCard > targetCard {
		target.IsOut = true
		target.DiscardPile = append(target.DiscardPile, target.Hand...)
		target.Hand = []CardType{}
		state.Players[action.TargetID] = target
		events = append(events, DomainEvent{
			Type: EventBaronResult,
			Payload: BaronResultPayload{
				WinnerID: action.PlayerID, LoserID: action.TargetID, WinnerCard: playerCard, LoserCard: targetCard,
			},
			Timestamp: clock.Now(),
		}, DomainEvent{
			Type: EventPlayerEliminated,
			Payload: PlayerEliminatedPayload{
				PlayerID: action.TargetID, Reason: ReasonBaronLost,
			},
			Timestamp: clock.Now(),
		})
	} else if targetCard > playerCard {
		player.IsOut = true
		player.DiscardPile = append(player.DiscardPile, player.Hand...)
		player.Hand = []CardType{}
		state.Players[action.PlayerID] = player
		events = append(events, DomainEvent{
			Type: EventBaronResult,
			Payload: BaronResultPayload{
				WinnerID: action.TargetID, LoserID: action.PlayerID, WinnerCard: targetCard, LoserCard: playerCard,
			},
			Timestamp: clock.Now(),
		}, DomainEvent{
			Type:      EventPlayerEliminated,
			Payload:   PlayerEliminatedPayload{PlayerID: action.PlayerID, Reason: ReasonBaronLost},
			Timestamp: clock.Now(),
		})
	}
	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
