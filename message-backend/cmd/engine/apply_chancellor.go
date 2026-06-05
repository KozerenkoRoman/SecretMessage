package engine

// ApplyChancellor реалізує ефект карти Chancellor
func ApplyChancellor(state GameState, action Action, rng RNG, clock Clock, startEventID uint64) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	// Подія розіграшу
	events := []DomainEvent{{
		EventID:   startEventID,
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardChancellor},
		Timestamp: clock.Now(),
	}}

	// Добір двох карт (або менше, якщо колода мала)
	drawCount := 2
	if len(state.Deck) < drawCount {
		drawCount = len(state.Deck)
	}

	drawn := []CardType{}
	for i := 0; i < drawCount; i++ {
		drawIndex := rng.Intn(len(state.Deck))
		card := state.Deck[drawIndex]
		state.Deck = append(state.Deck[:drawIndex], state.Deck[drawIndex+1:]...)
		drawn = append(drawn, card)
	}

	// Додаємо карти до руки гравця
	player.Hand = append(player.Hand, drawn...)
	state.Players[action.PlayerID] = player

	// Подія добору
	events = append(events, DomainEvent{
		EventID:   startEventID + 1,
		Type:      EventChancellorDrawn,
		Payload:   ChancellorDrawnPayload{PlayerID: action.PlayerID, Cards: drawn},
		Timestamp: clock.Now(),
	})

	// Перехід у фазу RESOLVE_CHANCELLOR
	state.Phase = PhaseResolveChancellor

	return ApplyResult{
		NewState:     state,
		DomainEvents: events,
	}, nil
}

// ResolveChancellor застосовує вибір гравця: яку карту залишити, а які покласти вниз колоди
func ResolveChancellor(state GameState, action ChancellorResolveAction, clock Clock, startEventID uint64) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}
	if action.KeepHandIndex < 0 || action.KeepHandIndex >= len(player.Hand) {
		return ApplyResult{}, NewError(ErrInvalidHandIndex,
			"index=%d, hand_size=%d", action.KeepHandIndex, len(player.Hand))
	}

	// Вибрана карта
	kept := player.Hand[action.KeepHandIndex]

	// Решта карт і порядок для низу колоди
	bottom := []CardType{}
	for _, c := range action.BottomOrder {
		bottom = append(bottom, c)
	}

	// Оновлюємо руку поточного гравця (залишається суворо ОДНА карта)
	player.Hand = []CardType{kept}
	state.Players[action.PlayerID] = player

	// Кладемо решту вниз колоди у заданому порядку
	state.Deck = append(state.Deck, bottom...)

	events := []DomainEvent{{
		EventID:   startEventID,
		Type:      EventChancellorResolved,
		Payload:   ChancellorResolvedPayload{PlayerID: action.PlayerID, Kept: kept, BottomOrder: bottom},
		Timestamp: clock.Now(),
	}}

	// Повертаємо фазу у MAIN_ACTION
	state.Phase = PhaseMainAction

	// Перевіряємо, чи не закінчився раунд (наприклад, якщо карт у колоді більше немає)
	aliveCount := 0
	for _, p := range state.Players {
		if !p.IsOut {
			aliveCount++
		}
	}

	if aliveCount <= 1 || len(state.Deck) == 0 {
		roundResult := ResolveRoundEnd(state, clock, startEventID+uint64(len(events)))
		state = roundResult.NewState
		events = append(events, roundResult.DomainEvents...)

		return ApplyResult{
			NewState:     state,
			DomainEvents: events,
		}, nil
	}

	// Перемикаємо хід на наступного живого гравця
	state.CurrentTurn = (state.CurrentTurn + 1) % len(state.TurnOrder)
	for state.Players[state.TurnOrder[state.CurrentTurn]].IsOut {
		state.CurrentTurn = (state.CurrentTurn + 1) % len(state.TurnOrder)
	}

	newActivePlayerID := state.TurnOrder[state.CurrentTurn]
	newActivePlayer := state.Players[newActivePlayerID]

	// Знімаємо захист Служниці
	newActivePlayer.IsProtected = false

	// Новий гравець бере карту з колоди
	drawnCard := state.Deck[0]
	state.Deck = state.Deck[1:]
	newActivePlayer.Hand = append(newActivePlayer.Hand, drawnCard)

	// Додаємо подію взяття карти для нового гравця
	events = append(events, DomainEvent{
		EventID:   startEventID + uint64(len(events)),
		Type:      EventCardDrawn,
		Payload:   CardDrawnPayload{PlayerID: newActivePlayerID, Card: drawnCard},
		Timestamp: clock.Now(),
	})

	// Зберігаємо нового активного гравця в стан
	state.Players[newActivePlayerID] = newActivePlayer

	return ApplyResult{
		NewState:     state,
		DomainEvents: events,
	}, nil
}
