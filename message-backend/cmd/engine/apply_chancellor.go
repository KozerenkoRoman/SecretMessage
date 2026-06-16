/* ===== ФІКС FILE: engine\apply_chancellor.go ===== */
package engine

func ApplyChancellor(state GameState, action Action, rng RNG, clock Clock) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	events := []DomainEvent{{
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardChancellor},
		Timestamp: clock.Now(),
	}}

	drawCount := min(len(state.Deck), 2)
	drawn := make([]CardType, drawCount)
	copy(drawn, state.Deck[:drawCount])
	state.Deck = state.Deck[drawCount:]

	player.Hand = append(player.Hand, drawn...)
	state.Players[action.PlayerID] = player

	events = append(events, DomainEvent{
		Type:      EventChancellorDrawn,
		Payload:   ChancellorDrawnPayload{PlayerID: action.PlayerID, Cards: drawn},
		Timestamp: clock.Now(),
	})

	state.Phase = PhaseResolveChancellor
	state.PendingAction = PendingAction{
		Type:     PhaseResolveChancellor,
		PlayerID: action.PlayerID,
	}

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}

func ResolveChancellor(state GameState, action ChancellorResolveAction, clock Clock) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}
	if state.Phase != PhaseResolveChancellor || state.PendingAction.PlayerID != action.PlayerID {
		return ApplyResult{}, NewError(ErrInvalidPhase, "not in chancellor resolve phase for this player")
	}
	if action.KeepHandIndex < 0 || action.KeepHandIndex >= len(player.Hand) {
		return ApplyResult{}, NewError(ErrInvalidHandIndex, "index=%d,hand_size=%d", action.KeepHandIndex, len(player.Hand))
	}

	kept := player.Hand[action.KeepHandIndex]

	expectedBottom := make(map[CardType]int)
	for i, card := range player.Hand {
		if i != action.KeepHandIndex {
			expectedBottom[card]++
		}
	}

	expectedCount := len(player.Hand) - 1
	if len(action.BottomOrder) != expectedCount {
		return ApplyResult{}, NewError(ErrInvalidAction, "invalid bottom cards count: expected %d, got %d", expectedCount, len(action.BottomOrder))
	}

	bottom := make([]CardType, len(action.BottomOrder))
	for i, card := range action.BottomOrder {
		if expectedBottom[card] <= 0 {
			return ApplyResult{}, NewError(ErrInvalidAction, "card %d cannot be put to bottom (not in hand or duplicate)", card)
		}
		expectedBottom[card]--
		bottom[i] = card
	}

	player.Hand = []CardType{kept}
	state.Players[action.PlayerID] = player

	// Кладемо карти на дно колоди
	state.Deck = append(state.Deck, bottom...)

	events := []DomainEvent{{
		Type:      EventChancellorResolved,
		Payload:   ChancellorResolvedPayload{PlayerID: action.PlayerID, Kept: kept, BottomOrder: bottom},
		Timestamp: clock.Now(),
	}}

	// Повертаємо гру в нормальну фазу
	state.Phase = PhaseMainAction
	state.PendingAction = PendingAction{}

	// Перевіряємо умови завершення раунду перед передачею ходу
	aliveCount := 0
	for _, p := range state.Players {
		if !p.IsOut {
			aliveCount++
		}
	}
	if aliveCount <= 1 || len(state.Deck) == 0 {
		roundResult := ResolveRoundEnd(state, clock)
		state = roundResult.NewState
		events = append(events, roundResult.DomainEvents...)
		return ApplyResult{NewState: state, DomainEvents: events}, nil
	}

	var advanceEvents []DomainEvent
	state, advanceEvents = AdvanceTurn(state, clock)
	events = append(events, advanceEvents...)

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
