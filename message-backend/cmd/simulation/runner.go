// =============================================================================
// simulation/runner.go
//
// Headless Bot-vs-Bot рушій.
//
// Runner проганяє повну гру (кілька раундів до 7 очок або поки не лишиться
// переможець) без мережі, WebSocket, таймерів та БД. Він відтворює той самий
// життєвий цикл раунду, що й room.StartGame / room.NextRound (див. room.go):
//   1. PrepareNewDeck → burn-карта + колода.
//   2. Роздати по 1 карті кожному живому слоту.
//   3. Активний гравець добирає 2-гу карту.
//   4. Цикл ходів через engine.Apply / engine.ResolveChancellor доки раунд не
//      завершиться (ROUND_END).
//   5. Якщо гра не закінчена — новий раунд; переможець раунду починає наступний.
//
// First-move override: якщо задано, вказаний слот стає індексом 0 у TurnOrder
// кожного НОВОГО раунду (і, відповідно, ходить першим).
//
// Жодних time.Sleep — виконання максимально швидке.
// =============================================================================

package simulation

import (
	"context"
	"math/rand"
	"time"

	"secret-message/cmd/engine"

	"github.com/sirupsen/logrus"
)

// simClock — детермінований Clock для engine (реальний час не потрібен, але
// engine.DomainEvent має Timestamp). Використовуємо фіксований момент.
type simClock struct{ t time.Time }

func (c simClock) Now() time.Time { return c.t }

// goRNG адаптує *rand.Rand під engine.RNG.
type goRNG struct{ r *rand.Rand }

func (g goRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return g.r.Intn(n)
}

// scoreLimit — скільки очок треба для перемоги у грі (як у engine.ResolveRoundEnd).
const scoreLimit = 7

// Runner виконує одну гру між сконфігурованими слотами.
type Runner struct {
	cfg    Config
	log    *logrus.Logger
	strats map[string]Strategy // botID -> стратегія
	order  []string            // базовий порядок слотів (botID)
}

// NewRunner будує рушій під конкретну гру з детермінованим rng (від seed гри).
func NewRunner(cfg Config, gameSeed int64, logger *logrus.Logger) *Runner {
	rng := rand.New(rand.NewSource(gameSeed))
	strats := make(map[string]Strategy, len(cfg.Slots))
	order := make([]string, 0, len(cfg.Slots))
	for _, slot := range cfg.Slots {
		strats[slot.BotID] = buildStrategy(slot, rng, logger)
		order = append(order, slot.BotID)
	}
	return &Runner{cfg: cfg, log: logger, strats: strats, order: order}
}

// slotStrategyMap повертає botID -> назва стратегії (для метрик).
func (r *Runner) slotStrategyMap() map[string]string {
	out := make(map[string]string, len(r.cfg.Slots))
	for _, slot := range r.cfg.Slots {
		out[slot.BotID] = slot.Strategy
	}
	return out
}

// RunGame проганяє одну повну гру та повертає її телеметрію.
func (r *Runner) RunGame(ctx context.Context, gameIndex int, gameSeed int64) (SimGameRecord, error) {
	start := time.Now()
	clock := simClock{t: time.Unix(0, 0)}
	rng := goRNG{r: rand.New(rand.NewSource(gameSeed))}

	rec := SimGameRecord{
		GameIndex: gameIndex,
		Seed:      gameSeed,
		Moves:     make([]SimMoveRecord, 0, 64),
	}

	// Стартовий стан гри (лобі з усіма слотами, рахунок 0).
	state := r.initialState()

	globalTurn := 0
	roundIdx := 0

	// Порядок слотів у першому раунді (з урахуванням first-move override).
	startingBot := r.cfg.FirstMoveBotID

	for !state.IsGameOver {
		select {
		case <-ctx.Done():
			return rec, ctx.Err()
		default:
		}

		// --- Ініціалізація раунду (дзеркало room.StartGame/NextRound) ---
		r.setupRound(&state, rng, clock, startingBot)
		rec.TotalRounds++
		roundIdx++
		if rec.FirstMoveBot == "" {
			rec.FirstMoveBot = state.TurnOrder[state.CurrentTurn]
		}

		// Скидаємо перраундову пам'ять стратегій.
		for _, s := range r.strats {
			s.ResetForNewRound()
		}
		// Даємо стратегіям побачити стартовий стан раунду.
		for _, s := range r.strats {
			s.ObserveEvents(&state, nil)
		}

		// --- Цикл ходів раунду ---
		for state.Phase != engine.PhaseRoundEnd && !state.IsGameOver {
			if globalTurn > hardTurnLimit {
				r.log.Warn("[sim] hard turn limit reached, aborting game")
				break
			}
			select {
			case <-ctx.Done():
				return rec, ctx.Err()
			default:
			}

			events, move, err := r.playOneTurn(&state, rng, clock, globalTurn, roundIdx)
			if err != nil {
				// Не мали б доходити сюди: playOneTurn вже застосовує fallback.
				r.log.WithError(err).Warn("[sim] unrecoverable turn error")
				break
			}
			rec.Moves = append(rec.Moves, move)
			globalTurn++
			rec.TotalTurns++

			// Оновлюємо пам'ять усіх стратегій подіями цього переходу.
			for _, s := range r.strats {
				s.ObserveEvents(&state, events)
			}
		}

		// Наступний раунд починає переможець попереднього (як у room.NextRound),
		// АЛЕ first-move override має пріоритет, якщо заданий.
		if r.cfg.FirstMoveBotID != "" {
			startingBot = r.cfg.FirstMoveBotID
		} else {
			startingBot = state.WinnerID
		}

		// Готуємо стан до наступного раунду: знімаємо RoundEnd, лишаємо рахунок.
		if !state.IsGameOver {
			state.Phase = engine.PhaseMainAction
			state.WinnerID = ""
		}
	}

	// Фінальний переможець гри = слот із найбільшим рахунком (>=7).
	r.finalizeWinner(&state, &rec)
	rec.DurationMs = time.Since(start).Milliseconds()
	return rec, nil
}

// initialState будує стан-лобі з усіма слотами (рахунок 0, ще без роздачі).
func (r *Runner) initialState() engine.GameState {
	players := make(map[string]engine.Player, len(r.cfg.Slots))
	for _, slot := range r.cfg.Slots {
		players[slot.BotID] = engine.Player{
			ID:          slot.BotID,
			Username:    slot.Username,
			UserRole:    "bot",
			Hand:        []engine.CardType{},
			DiscardPile: []engine.CardType{},
			Score:       0,
		}
	}
	return engine.GameState{
		Phase:   engine.PhaseMainAction,
		Players: players,
	}
}

// setupRound роздає нову колоду й формує TurnOrder (з first-move override).
func (r *Runner) setupRound(state *engine.GameState, rng engine.RNG, clock engine.Clock, startingBot string) {
	deck, burn := engine.PrepareNewDeck(rng)
	state.Deck = deck
	state.BurnCard = &burn
	state.Phase = engine.PhaseMainAction
	state.IsGameOver = false
	state.WinnerID = ""
	state.PendingAction = engine.PendingAction{}
	state.Sequence++

	// Формуємо порядок ходу: базовий r.order, але якщо startingBot заданий і
	// валідний — ставимо його на позицію 0 (ротацією).
	order := make([]string, len(r.order))
	copy(order, r.order)
	if startingBot != "" {
		for i, id := range order {
			if id == startingBot {
				order = append(order[i:], order[:i]...)
				break
			}
		}
	}
	state.TurnOrder = order
	state.CurrentTurn = 0

	// Роздаємо по 1 карті кожному слоту + скидаємо перраундовий стан гравця.
	for _, id := range order {
		p := state.Players[id]
		card := state.Deck[0]
		state.Deck = state.Deck[1:]
		p.Hand = []engine.CardType{card}
		p.DiscardPile = []engine.CardType{}
		p.IsOut = false
		p.IsProtected = false
		p.SpyPointsAwarded = false
		state.Players[id] = p
	}

	// Перший гравець добирає 2-гу карту (як у room.StartGame).
	firstID := order[0]
	fp := state.Players[firstID]
	if len(state.Deck) > 0 {
		draw := state.Deck[0]
		state.Deck = state.Deck[1:]
		fp.Hand = append(fp.Hand, draw)
		state.Players[firstID] = fp
	}
}

// playOneTurn запитує рішення активного слота, застосовує його через engine,
// і повертає події переходу + телеметрію ходу.
func (r *Runner) playOneTurn(state *engine.GameState, rng engine.RNG, clock engine.Clock, globalTurn, roundIdx int) ([]engine.DomainEvent, SimMoveRecord, error) {
	activeID := r.activePlayer(state)
	strat := r.strats[activeID]

	move := SimMoveRecord{
		TurnIndex:   globalTurn,
		RoundIndex:  roundIdx,
		BotID:       activeID,
		BotStrategy: r.strategyName(activeID),
	}

	decisionStart := time.Now()

	if state.Phase == engine.PhaseResolveChancellor {
		act, err := strat.DecideChancellor(state, activeID)
		if err != nil {
			// Fallback: лишаємо першу карту, решту на дно.
			act = fallbackChancellor(state, activeID)
			move.UsedFallback = true
		}
		move.DecisionMs = time.Since(decisionStart).Milliseconds()
		move.PlayedCard = int(engine.CardChancellor)

		res, applyErr := engine.ResolveChancellor(*state, act, clock)
		if applyErr != nil {
			move.InvalidAttempt = true
			// Форсуємо валідний fallback і повторюємо.
			act = fallbackChancellor(state, activeID)
			move.UsedFallback = true
			res, applyErr = engine.ResolveChancellor(*state, act, clock)
			if applyErr != nil {
				return nil, move, applyErr
			}
		}
		*state = res.NewState
		annotateEliminations(&move, res.DomainEvents)
		return res.DomainEvents, move, nil
	}

	// MAIN_ACTION.
	act, err := strat.DecideMain(state, activeID)
	if err != nil {
		act = fallbackMain(state, activeID)
		move.UsedFallback = true
	}
	move.DecisionMs = time.Since(decisionStart).Milliseconds()

	me := state.Players[activeID]
	if act.HandIndex >= 0 && act.HandIndex < len(me.Hand) {
		move.PlayedCard = int(me.Hand[act.HandIndex])
	}
	move.TargetID = act.TargetID
	if act.Guess != 0 || move.PlayedCard == int(engine.CardGuard) {
		g := int(act.Guess)
		move.GuessCard = &g
	}

	res, applyErr := engine.Apply(*state, act, rng, clock)
	if applyErr != nil {
		// Невалідний хід — фіксуємо і застосовуємо гарантовано-легальний fallback.
		move.InvalidAttempt = true
		move.UsedFallback = true
		act = fallbackMain(state, activeID)
		if act.HandIndex >= 0 && act.HandIndex < len(me.Hand) {
			move.PlayedCard = int(me.Hand[act.HandIndex])
		}
		move.TargetID = act.TargetID
		res, applyErr = engine.Apply(*state, act, rng, clock)
		if applyErr != nil {
			return nil, move, applyErr
		}
	}

	*state = res.NewState
	annotateEliminations(&move, res.DomainEvents)
	return res.DomainEvents, move, nil
}

func (r *Runner) activePlayer(state *engine.GameState) string {
	if state.Phase == engine.PhaseResolveChancellor {
		return state.PendingAction.PlayerID
	}
	if state.CurrentTurn >= 0 && state.CurrentTurn < len(state.TurnOrder) {
		return state.TurnOrder[state.CurrentTurn]
	}
	return ""
}

func (r *Runner) strategyName(botID string) string {
	for _, slot := range r.cfg.Slots {
		if slot.BotID == botID {
			return slot.Strategy
		}
	}
	return "unknown"
}

// finalizeWinner визначає переможця гри та фіксує його у записі.
func (r *Runner) finalizeWinner(state *engine.GameState, rec *SimGameRecord) {
	bestScore := -1
	bestID := ""
	bestSlot := -1
	for slotIdx, id := range r.order {
		p := state.Players[id]
		if p.Score > bestScore {
			bestScore = p.Score
			bestID = id
			bestSlot = slotIdx
		}
	}
	if bestID != "" && bestScore >= scoreLimit {
		rec.WinnerBot = bestID
		s := bestSlot
		rec.WinnerSlot = &s
	} else if bestID != "" {
		// Гра могла завершитись через hardTurnLimit — беремо лідера як переможця.
		rec.WinnerBot = bestID
		s := bestSlot
		rec.WinnerSlot = &s
	}
	// Хто взяв бонус Шпигуна в ОСТАННЬОМУ раунді (інформативно).
	for _, id := range r.order {
		if state.Players[id].SpyPointsAwarded {
			rec.SpyWinnerBot = id
			break
		}
	}
}

// annotateEliminations витягує причину вибуття з подій переходу (якщо була).
func annotateEliminations(move *SimMoveRecord, events []engine.DomainEvent) {
	for _, ev := range events {
		if ev.Type == engine.EventPlayerEliminated {
			if p, ok := ev.Payload.(engine.PlayerEliminatedPayload); ok {
				move.EliminatedReason = p.Reason
			}
		}
	}
}

// fallbackMain — гарантовано легальна дія: перша некритична карта без цілі
// (або з першою легальною ціллю). Використовується, коли стратегія помилилась.
func fallbackMain(state *engine.GameState, botID string) engine.Action {
	me := state.Players[botID]
	// Обираємо індекс, що не порушує правило Графині; уникаємо Принцеси.
	chosen := 0
	for idx, card := range me.Hand {
		if card == engine.CardPrincess && len(me.Hand) > 1 {
			continue
		}
		if err := engine.ValidateCountessRule(me, engine.Action{HandIndex: idx}); err != nil {
			continue
		}
		chosen = idx
		break
	}

	act := engine.Action{PlayerID: botID, HandIndex: chosen}

	// Якщо карта потребує ціль — беремо першу валідну.
	card := me.Hand[chosen]
	if needsTarget(card) {
		for id, p := range state.Players {
			if id == botID || p.IsOut || p.IsProtected {
				continue
			}
			act.TargetID = id
			break
		}
		if card == engine.CardGuard {
			act.Guess = engine.CardPriest // будь-що, окрім Guard
		}
	}
	return act
}

func fallbackChancellor(state *engine.GameState, botID string) engine.ChancellorResolveAction {
	me := state.Players[botID]
	bottom := make([]engine.CardType, 0, len(me.Hand)-1)
	for i, card := range me.Hand {
		if i == 0 {
			continue
		}
		bottom = append(bottom, card)
	}
	return engine.ChancellorResolveAction{
		PlayerID:      botID,
		KeepHandIndex: 0,
		BottomOrder:   bottom,
	}
}

func needsTarget(card engine.CardType) bool {
	switch card {
	case engine.CardGuard, engine.CardPriest, engine.CardBaron, engine.CardPrince, engine.CardKing:
		return true
	default:
		return false
	}
}
