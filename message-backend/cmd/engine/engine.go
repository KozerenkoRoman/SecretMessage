package engine

func Apply(state GameState, action Action, rng RNG, clock Clock, startEventID uint64) (ApplyResult, error) {
	// Базові перевірки стану
	if len(state.TurnOrder) == 0 {
		return ApplyResult{}, NewError(ErrEmptyTurnOrder, "empty TurnOrder")
	}
	if state.CurrentTurn < 0 || state.CurrentTurn >= len(state.TurnOrder) {
		return ApplyResult{}, NewError(ErrCurrentTurnOutOfRange,
			"CurrentTurn=%d, TurnOrder len=%d", state.CurrentTurn, len(state.TurnOrder))
	}
	if state.Phase != PhaseMainAction {
		return ApplyResult{}, NewError(ErrInvalidPhase,
			"expected %s, got %s", PhaseMainAction, state.Phase)
	}
	if len(state.Players) == 0 {
		return ApplyResult{}, NewError(ErrNoPlayers, "no players in state")
	}

	// 1. Перевірка черговості ходу
	activePlayerID := state.TurnOrder[state.CurrentTurn]
	player, exists := state.Players[action.PlayerID]
	if !exists {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if activePlayerID != action.PlayerID {
		return ApplyResult{}, NewError(ErrOutOfTurn,
			"active=%s, got=%s", activePlayerID, action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}
	if action.HandIndex < 0 || action.HandIndex >= len(player.Hand) {
		return ApplyResult{}, NewError(ErrInvalidHandIndex,
			"index=%d, hand_size=%d", action.HandIndex, len(player.Hand))
	}

	// 2. Яку карту граємо
	if len(player.Hand) == 0 {
		return ApplyResult{}, NewError(ErrPlayerHasNoCards, "player_id=%s", action.PlayerID)
	}

	// ВИКЛИК ВАЛІДАТОРУ З ФАЙЛУ apply_countess.go
	if err := ValidateCountessRule(player, action); err != nil {
		return ApplyResult{}, err // Негайно повертаємо помилку валідації
	}

	playedCard := player.Hand[action.HandIndex]

	// 3. Клонування стану
	nextState := state.Clone()
	nextState.Sequence++

	// 4. БЕЗПЕЧНЕ скидання карти в дискард та видалення з руки
	pMod := nextState.Players[action.PlayerID]

	// Додаємо карту до стопки скидання
	pMod.DiscardPile = append(pMod.DiscardPile, playedCard)

	// Створюємо новий слайс методом фільтрації
	var newHand []CardType
	if len(pMod.Hand) > 0 {
		newHand = append(newHand, pMod.Hand[:action.HandIndex]...)
		newHand = append(newHand, pMod.Hand[action.HandIndex+1:]...)
	}
	pMod.Hand = newHand
	nextState.Players[action.PlayerID] = pMod

	// 5. Делегація ефекту карти
	var effectResult ApplyResult
	var err error

	switch playedCard {
	case CardSpy:
		effectResult, err = ApplySpy(nextState, action, clock, startEventID)
	case CardGuard:
		effectResult, err = ApplyGuard(nextState, action, clock, startEventID)
	case CardPriest:
		effectResult, err = ApplyPriest(nextState, action, clock, startEventID)
	case CardHandmaid:
		effectResult, err = ApplyHandmaid(nextState, action, clock, startEventID)
	case CardBaron:
		effectResult, err = ApplyBaron(nextState, action, clock, startEventID)
	case CardPrince:
		effectResult, err = ApplyPrince(nextState, action, rng, clock, startEventID)
	case CardChancellor:
		effectResult, err = ApplyChancellor(nextState, action, rng, clock, startEventID)
	case CardCountess:
		effectResult, err = ApplyCountess(nextState, action, clock, startEventID)
	case CardKing:
		effectResult, err = ApplyKing(nextState, action, clock, startEventID)
	case CardPrincess:
		effectResult, err = ApplyPrincess(nextState, action, clock, startEventID)
	default:
		return ApplyResult{}, NewError(ErrCardEffectNotImplemented,
			"card=%d", playedCard)
	}
	if err != nil {
		return ApplyResult{}, err
	}

	// 6. Обробка завершення ходу та перехід до наступного гравця
	effectState := effectResult.NewState

	// Рахуємо кількість живих гравців ПІСЛЯ застосування ефекту карти
	aliveCount := 0
	for _, p := range effectState.Players {
		if !p.IsOut {
			aliveCount++
		}
	}

	// ПЕРЕВІРКА КІНЦЯ РАУНДУ: якщо залишився 1 живий АБО колода порожня НА МОМЕНТ завершення поточного ходу
	if aliveCount <= 1 || len(effectState.Deck) == 0 {
		roundResult := ResolveRoundEnd(effectState, clock, startEventID+uint64(len(effectResult.DomainEvents)))
		effectState = roundResult.NewState
		effectResult.DomainEvents = append(effectResult.DomainEvents, roundResult.DomainEvents...)

		return ApplyResult{
			NewState:     effectState,
			DomainEvents: effectResult.DomainEvents,
		}, nil
	}

	if effectState.Phase == PhaseResolveChancellor {
		return ApplyResult{
			NewState:     effectState,
			DomainEvents: effectResult.DomainEvents,
		}, nil
	}

	// Якщо раунд триває, визначаємо наступного активного гравця (пропускаємо тих, хто вибув)
	effectState.CurrentTurn = (effectState.CurrentTurn + 1) % len(effectState.TurnOrder)
	for effectState.Players[effectState.TurnOrder[effectState.CurrentTurn]].IsOut {
		effectState.CurrentTurn = (effectState.CurrentTurn + 1) % len(effectState.TurnOrder)
	}

	newActivePlayerID := effectState.TurnOrder[effectState.CurrentTurn]
	newActivePlayer := effectState.Players[newActivePlayerID]

	// Знімаємо захист Служниці (IsProtected) на початку ВЛАСНОГО ходу гравця
	newActivePlayer.IsProtected = false

	// ДОБІР КАРТИ: Оскільки ми вище перевірили, що len(Deck) > 0, тут гарантовано є карта
	drawnCard := effectState.Deck[0]
	effectState.Deck = effectState.Deck[1:] // Видаляємо карту з колоди

	// Додаємо карту в руку новому активному гравцю
	newActivePlayer.Hand = append(newActivePlayer.Hand, drawnCard)

	// Створюємо доменну подію про добір карти
	effectResult.DomainEvents = append(effectResult.DomainEvents, DomainEvent{
		EventID:   startEventID + uint64(len(effectResult.DomainEvents)) + 1,
		Type:      EventCardDrawn,
		Payload:   CardDrawnPayload{PlayerID: newActivePlayerID, Card: drawnCard},
		Timestamp: clock.Now(),
	})

	// Зберігаємо оновленого гравця в стан
	effectState.Players[newActivePlayerID] = newActivePlayer

	return ApplyResult{
		NewState:     effectState,
		DomainEvents: effectResult.DomainEvents,
	}, nil
}

// ResolveRoundEnd завершує раунд, визначає переможця та нараховує бонуси
func ResolveRoundEnd(state GameState, clock Clock, startEventID uint64) ApplyResult {
	events := []DomainEvent{}
	// Оновлюємо стейт (там підрахує прапорці та додасть p.Score++)
	state = ResolveSpyBonus(state)
	// Збираємо івенти для фронтенду на основі оновленого стейту
	for id, p := range state.Players {
		if p.SpyPointsAwarded {
			events = append(events, DomainEvent{
				EventID:   startEventID,
				Type:      EventSpyBonus,
				Payload:   SpyBonusPayload{PlayerID: id, Points: 1},
				Timestamp: clock.Now(),
			})
			startEventID++ // Інкрементуємо ID для наступного івенту
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
			EventID:   startEventID,
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

	return true, nil // Ціль валідна, ефект має виконатися
}
