/* ===== FILE: engine\apply_princess.go ===== */
package engine

func ApplyPrincess(state GameState, action Action, clock Clock) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	events := []DomainEvent{{
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardPrincess},
		Timestamp: clock.Now(),
	}}

	player.IsOut = true
	// Сама Принцеса вже потрапила у DiscardPile в engine.go. Тепер, при вибутті
	// гравця, решта його руки ТЕЖ має піти у відбій — так само, як це роблять
	// ApplyGuard/ApplyBaron/room.eliminatePlayer. Це критично для правила
	// Шпигуна: історія розіграних/скинутих карт тримається у DiscardPile, тож
	// якщо в руці лишався Шпигун, він має зберегтися для підрахунку бонусу
	// навіть у вибулого гравця.
	player.DiscardPile = append(player.DiscardPile, player.Hand...)
	player.Hand = []CardType{}
	state.Players[action.PlayerID] = player

	events = append(events, DomainEvent{
		Type:      EventPlayerEliminated,
		Payload:   PlayerEliminatedPayload{PlayerID: action.PlayerID, Reason: ReasonPrincessPlayed},
		Timestamp: clock.Now(),
	})

	// ФІКС: Обов'язкова перевірка кінця раунду
	aliveCount := 0
	for _, p := range state.Players {
		if !p.IsOut {
			aliveCount++
		}
	}
	if aliveCount <= 1 {
		roundResult := ResolveRoundEnd(state, clock)
		state = roundResult.NewState
		events = append(events, roundResult.DomainEvents...)
	}

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}
