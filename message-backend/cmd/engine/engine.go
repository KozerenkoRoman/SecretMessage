package engine

func Apply(state GameState, action Action, rng RNG, clock Clock) (ApplyResult, error) {
	if len(state.TurnOrder) == 0 {
		return ApplyResult{}, NewError(ErrEmptyTurnOrder, "empty TurnOrder")
	}
	if state.CurrentTurn < 0 || state.CurrentTurn >= len(state.TurnOrder) {
		return ApplyResult{}, NewError(ErrCurrentTurnOutOfRange, "CurrentTurn=%d,TurnOrder len=%d", state.CurrentTurn, len(state.TurnOrder))
	}
	if state.Phase != PhaseMainAction {
		return ApplyResult{}, NewError(ErrInvalidPhase, "expected %s,got %s", PhaseMainAction, state.Phase)
	}
	if len(state.Players) == 0 {
		return ApplyResult{}, NewError(ErrNoPlayers, "no players in state")
	}

	activePlayerID := state.TurnOrder[state.CurrentTurn]
	player, exists := state.Players[action.PlayerID]
	if !exists {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if activePlayerID != action.PlayerID {
		return ApplyResult{}, NewError(ErrOutOfTurn, "active=%s,got=%s", activePlayerID, action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}
	if action.HandIndex < 0 || action.HandIndex >= len(player.Hand) {
		return ApplyResult{}, NewError(ErrInvalidHandIndex, "index=%d,hand_size=%d", action.HandIndex, len(player.Hand))
	}
	if len(player.Hand) == 0 {
		return ApplyResult{}, NewError(ErrPlayerHasNoCards, "player_id=%s", action.PlayerID)
	}

	if err := ValidateCountessRule(player, action); err != nil {
		return ApplyResult{}, err
	}

	playedCard := player.Hand[action.HandIndex]
	nextState := state.Clone()
	nextState.Sequence++

	pMod := nextState.Players[action.PlayerID]
	pMod.DiscardPile = append(pMod.DiscardPile, playedCard)
	var newHand []CardType
	if len(pMod.Hand) > 0 {
		newHand = append(newHand, pMod.Hand[:action.HandIndex]...)
		newHand = append(newHand, pMod.Hand[action.HandIndex+1:]...)
	}
	pMod.Hand = newHand
	nextState.Players[action.PlayerID] = pMod

	var effectResult ApplyResult
	var err error

	switch playedCard {
	case CardSpy:
		effectResult, err = ApplySpy(nextState, action, clock)
	case CardGuard:
		effectResult, err = ApplyGuard(nextState, action, clock)
	case CardPriest:
		effectResult, err = ApplyPriest(nextState, action, clock)
	case CardHandmaid:
		effectResult, err = ApplyHandmaid(nextState, action, clock)
	case CardBaron:
		effectResult, err = ApplyBaron(nextState, action, clock)
	case CardPrince:
		effectResult, err = ApplyPrince(nextState, action, rng, clock)
	case CardChancellor:
		effectResult, err = ApplyChancellor(nextState, action, rng, clock)
	case CardCountess:
		effectResult, err = ApplyCountess(nextState, action, clock)
	case CardKing:
		effectResult, err = ApplyKing(nextState, action, clock)
	case CardPrincess:
		effectResult, err = ApplyPrincess(nextState, action, clock)
	default:
		return ApplyResult{}, NewError(ErrCardEffectNotImplemented, "card=%d", playedCard)
	}

	if err != nil {
		return ApplyResult{}, err
	}

	effectState := effectResult.NewState

	// Перевіряємо кінець раунду
	aliveCount := 0
	for _, p := range effectState.Players {
		if !p.IsOut {
			aliveCount++
		}
	}
	if aliveCount <= 1 || len(effectState.Deck) == 0 {
		roundResult := ResolveRoundEnd(effectState, clock)
		effectState = roundResult.NewState
		effectResult.DomainEvents = append(effectResult.DomainEvents, roundResult.DomainEvents...)
		return ApplyResult{NewState: effectState, DomainEvents: effectResult.DomainEvents}, nil
	}

	// ЯКЩО ФАЗА КАНЦЛЕРА — ПЕРЕРИВАЄМО ПЕРЕХІД ХОДУ, ЧЕКАЄМО НА RESOLVE
	if effectState.Phase == PhaseResolveChancellor {
		return ApplyResult{NewState: effectState, DomainEvents: effectResult.DomainEvents}, nil
	}

	// Викликаємо єдину логіку переходу ходу
	effectState, advanceEvents := AdvanceTurn(effectState, clock)
	effectResult.DomainEvents = append(effectResult.DomainEvents, advanceEvents...)

	return ApplyResult{NewState: effectState, DomainEvents: effectResult.DomainEvents}, nil
}

// AdvanceTurn — єдине місце в системі, яке перемикає хід та видає карту наступному гравцю!
func AdvanceTurn(state GameState, clock Clock) (GameState, []DomainEvent) {
	events := []DomainEvent{}

	state.CurrentTurn = (state.CurrentTurn + 1) % len(state.TurnOrder)
	for state.Players[state.TurnOrder[state.CurrentTurn]].IsOut {
		state.CurrentTurn = (state.CurrentTurn + 1) % len(state.TurnOrder)
	}

	newActivePlayerID := state.TurnOrder[state.CurrentTurn]
	newActivePlayer := state.Players[newActivePlayerID]
	newActivePlayer.IsProtected = false

	if len(newActivePlayer.Hand) < 2 && len(state.Deck) > 0 {
		drawnCard := state.Deck[0]
		state.Deck = state.Deck[1:]
		newActivePlayer.Hand = append(newActivePlayer.Hand, drawnCard)

		events = append(events, DomainEvent{
			Type:      EventCardDrawn,
			Payload:   CardDrawnPayload{PlayerID: newActivePlayerID, Card: drawnCard},
			Timestamp: clock.Now(),
		})
	}

	state.Players[newActivePlayerID] = newActivePlayer
	return state, events
}

func ResolveRoundEnd(state GameState, clock Clock) ApplyResult {
	events := []DomainEvent{}
	state = ResolveSpyBonus(state)
	for id, p := range state.Players {
		if p.SpyPointsAwarded {
			events = append(events, DomainEvent{
				Type:      EventSpyBonus,
				Payload:   SpyBonusPayload{PlayerID: id, Points: 1},
				Timestamp: clock.Now(),
			})
		}
	}

	alive := []Player{}
	for _, p := range state.Players {
		if !p.IsOut {
			alive = append(alive, p)
		}
	}

	var winnerID string
	if len(alive) == 1 {
		winnerID = alive[0].ID
	} else if len(alive) > 1 {
		best := alive[0]
		for _, p := range alive[1:] {
			if len(p.Hand) > 0 && len(best.Hand) > 0 {
				if p.Hand[0] > best.Hand[0] {
					best = p
				}
			}
		}
		winnerID = best.ID
	}

	if winnerID != "" {
		w := state.Players[winnerID]
		w.Score++
		state.Players[winnerID] = w
		state.WinnerID = winnerID
		reason := ReasonRoundDeckEmpty
		if len(alive) == 1 {
			reason = ReasonRoundLastStanding
		}
		events = append(events, DomainEvent{
			// 🌟 Тут так само поле EventID більше не потрібне
			Type:      EventRoundEnd,
			Payload:   RoundEndPayload{WinnerID: winnerID, Reason: reason},
			Timestamp: clock.Now(),
		})
	}

	for _, p := range state.Players {
		if p.Score >= 7 {
			state.IsGameOver = true
		}
	}

	state.Phase = PhaseRoundEnd
	state.PendingAction = PendingAction{}
	return ApplyResult{NewState: state, DomainEvents: events}
}

func CanTargetPlayer(state GameState, playerID string, targetID string) (bool, error) {
	hasValidTarget := false
	for id, p := range state.Players {
		if id != playerID && !p.IsOut && !p.IsProtected {
			hasValidTarget = true
			break
		}
	}
	if !hasValidTarget {
		return false, nil
	}
	if targetID == "" || targetID == "null" || targetID == "undefined" {
		return false, NewError(ErrTargetRequired, "target_id is empty")
	}
	target, exists := state.Players[targetID]
	if !exists {
		return false, NewError(ErrTargetNotFound, "target_id=%s", targetID)
	}
	if target.IsOut {
		return false, NewError(ErrTargetAlreadyOut, "target_id=%s", targetID)
	}
	if target.IsProtected {
		return false, NewError(ErrTargetProtected, "target_id=%s", targetID)
	}
	return true, nil
}
