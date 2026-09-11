package engine

import "slices"

func ApplySpy(state GameState, action Action, clock Clock) (ApplyResult, error) {
	player, ok := state.Players[action.PlayerID]
	if !ok {
		return ApplyResult{}, NewError(ErrPlayerNotFound, "player_id=%s", action.PlayerID)
	}
	if player.IsOut {
		return ApplyResult{}, NewError(ErrPlayerAlreadyOut, "player_id=%s", action.PlayerID)
	}

	events := []DomainEvent{{
		Type:      EventCardPlayed,
		Payload:   CardPlayedPayload{PlayerID: action.PlayerID, Card: CardSpy},
		Timestamp: clock.Now(),
	}}

	return ApplyResult{NewState: state, DomainEvents: events}, nil
}

// ResolveSpyBonus нараховує бонусний бал Шпигуна за офіційним правилом
// Love Letter:
//
//   - У кінці раунду 1 бонусний бал отримує гравець, ЯКЩО він єдиний, хто
//     зіграв (скинув) хоча б одного Шпигуна протягом цього раунду.
//   - Якщо один гравець зіграв кількох Шпигунів — він все одно єдиний
//     "власник" і отримує рівно 1 бал (рахуємо гравців, а не карти).
//   - Якщо двоє чи більше різних гравців зіграли Шпигуна — бонус скасовується,
//     ніхто його не отримує.
//
// ВАЖЛИВО: бал нараховується НЕЗАЛЕЖНО від того, чи гравець вибув, вийшов або
// відключився до завершення раунду. Історія розіграних карт зберігається у
// DiscardPile гравця (при виключенні/дисконекті рука також переноситься у
// DiscardPile — див. room.eliminatePlayer), тож IsOut НЕ впливає на право
// на бонус.
func ResolveSpyBonus(state GameState) GameState {
	// Скидаємо прапорець з попереднього раунду для всіх гравців.
	for id, p := range state.Players {
		p.SpyPointsAwarded = false
		state.Players[id] = p
	}

	// Збираємо УНІКАЛЬНИХ гравців, у чиїй історії розіграних карт (DiscardPile)
	// є хоча б один Шпигун. Вибулі/відключені гравці теж враховуються.
	spyOwners := make([]string, 0)
	for id, p := range state.Players {
		if slices.Contains(p.DiscardPile, CardSpy) {
			spyOwners = append(spyOwners, id)
		}
	}

	// Бонус нараховується лише якщо власник Шпигуна РІВНО один.
	if len(spyOwners) == 1 {
		winnerID := spyOwners[0]
		p := state.Players[winnerID]
		p.SpyPointsAwarded = true
		p.Score++
		state.Players[winnerID] = p
	}
	return state
}
