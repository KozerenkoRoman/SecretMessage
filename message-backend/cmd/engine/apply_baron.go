/* ===== FILE: engine\apply_baron.go ===== */
package engine

func ApplyBaron(state GameState, action Action, clock Clock, startEventID uint64) (ApplyResult, error) {
	shouldApply, err := CanTargetPlayer(state, action.PlayerID, action.TargetID)
	if err != nil {
		return ApplyResult{}, err
	}

	events := []DomainEvent{{
		EventID: startEventID,
		Type:    EventCardPlayed,
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
	if len(player.Hand) == 0 || len(target.Hand) == 0 {
		return ApplyResult{}, NewError(ErrBaronNoCardsToCompare, "player_hand=%d,target_hand=%d", len(player.Hand), len(target.Hand))
	}

	playerCard := player.Hand[0]
	targetCard := target.Hand[0]

	// Базова подія порівняння (залишається для внутрішньої логіки/анімацій)
	events = append(events, DomainEvent{
		EventID: startEventID + 1,
		Type:    EventRoundCompared,
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

		// ПЕРЕГЛЯНЬ ТУТ: ДодаємоWinnerCard та LoserCard у Payload події BaronResult,
		// щоб GameLogPanel.vue міг зчитати їх через log.namedArgs.winnerCard/loserCard
		events = append(events, DomainEvent{
			EventID: startEventID + 2,
			Type:    EventBaronResult,
			Payload: BaronResultPayload{
				WinnerID:   action.PlayerID,
				LoserID:    action.TargetID,
				WinnerCard: playerCard, // Передаємо як стрінгу або int (переконайся, що JSON мапиться на назву карти)
				LoserCard:  targetCard,
			},
			Timestamp: clock.Now(),
		}, DomainEvent{
			EventID: startEventID + 3,
			Type:    EventPlayerEliminated,
			Payload: PlayerEliminatedPayload{
				PlayerID: action.TargetID,
				Reason:   ReasonBaronLost,
			},
			Timestamp: clock.Now(),
		})

	} else if targetCard > playerCard {
		player.IsOut = true
		player.DiscardPile = append(player.DiscardPile, player.Hand...)
		player.Hand = []CardType{}
		state.Players[action.PlayerID] = player

		events = append(events, DomainEvent{
			EventID: startEventID + 2,
			Type:    EventBaronResult,
			Payload: BaronResultPayload{
				WinnerID:   action.TargetID,
				LoserID:    action.PlayerID,
				WinnerCard: targetCard,
				LoserCard:  playerCard,
			},
			Timestamp: clock.Now(),
		}, DomainEvent{
			EventID: startEventID + 3,
			Type:    EventPlayerEliminated,
			Payload: PlayerEliminatedPayload{
				PlayerID: action.PlayerID,
				Reason:   ReasonBaronLost,
			},
			Timestamp: clock.Now(),
		})
	}

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
