// =============================================================================
// simulation/strategy.go
//
// Стратегії прийняття рішень для headless Bot-vs-Bot симуляції.
//
// Стратегія — це чисте, детерміноване (за наявності seed) джерело рішень для
// одного «слота» гравця. Вона НЕ залежить від WebSocket, БД чи room.go — лише
// від engine.GameState. Це дозволяє ганяти тисячі матчів у пам'яті.
//
// Дві вбудовані реалізації:
//   - SmartStrategy  — обгортка навколо продакшн-логіки bot.BotBrain (+ DeckTracker),
//                      тобто «розумний» бот, яким грають живі кімнати.
//   - RandomStrategy — базова лінія: легальний випадковий хід. Потрібна для
//                      A/B-порівняння: наскільки SmartStrategy кращий за випадок.
//
// Додати нову «особистість» = реалізувати Strategy і зареєструвати у registry.
// =============================================================================

package simulation

import (
	"math/rand"

	"secret-message/cmd/bot"
	"secret-message/cmd/engine"

	"github.com/sirupsen/logrus"
)

// Strategy — рішення для одного слота гравця у headless-матчі.
type Strategy interface {
	// Name — стабільний ідентифікатор алгоритму (для телеметрії).
	Name() string

	// DecideMain повертає дію у фазі MAIN_ACTION.
	DecideMain(state *engine.GameState, botID string) (engine.Action, error)

	// DecideChancellor повертає резолв у фазі RESOLVE_CHANCELLOR.
	DecideChancellor(state *engine.GameState, botID string) (engine.ChancellorResolveAction, error)

	// ObserveEvents отримує події останнього переходу стану, щоб стратегія
	// могла оновити внутрішню пам'ять (напр., DeckTracker). Може бути no-op.
	ObserveEvents(state *engine.GameState, events []engine.DomainEvent)

	// ResetForNewRound скидає перраундову пам'ять (нова колода/роздача).
	ResetForNewRound()
}

// =============================================================================
// SmartStrategy — обгортка над продакшн-мозком бота.
// =============================================================================

type SmartStrategy struct {
	botID   string
	brain   *bot.BotBrain
	tracker *bot.DeckTracker
}

func NewSmartStrategy(botID string, logger *logrus.Logger) *SmartStrategy {
	return &SmartStrategy{
		botID:   botID,
		brain:   bot.NewBotBrain(botID, logger),
		tracker: bot.NewDeckTracker(botID),
	}
}

func (s *SmartStrategy) Name() string { return "smart" }

func (s *SmartStrategy) DecideMain(state *engine.GameState, _ string) (engine.Action, error) {
	decision, err := s.brain.Think(state, s.tracker)
	if err != nil {
		return engine.Action{}, err
	}
	act, ok := decision.(engine.Action)
	if !ok {
		return engine.Action{}, errUnexpectedDecision
	}
	act.PlayerID = s.botID
	return act, nil
}

func (s *SmartStrategy) DecideChancellor(state *engine.GameState, _ string) (engine.ChancellorResolveAction, error) {
	decision, err := s.brain.Think(state, s.tracker)
	if err != nil {
		return engine.ChancellorResolveAction{}, err
	}
	act, ok := decision.(engine.ChancellorResolveAction)
	if !ok {
		return engine.ChancellorResolveAction{}, errUnexpectedDecision
	}
	act.PlayerID = s.botID
	return act, nil
}

func (s *SmartStrategy) ObserveEvents(state *engine.GameState, events []engine.DomainEvent) {
	// Дзеркалимо логіку BotManager.HandleGameUpdate для одного трекера.
	for _, ev := range events {
		s.tracker.HandleDomainEvent(ev)
		if ev.Type == engine.EventChancellorResolved {
			if payload, ok := ev.Payload.(engine.ChancellorResolvedPayload); ok && payload.PlayerID == s.botID {
				s.tracker.RecordChancellorAction(engine.ChancellorResolveAction{
					PlayerID:    payload.PlayerID,
					BottomOrder: payload.BottomOrder,
				})
			}
		}
	}
	s.tracker.TrackGameState(state)
}

func (s *SmartStrategy) ResetForNewRound() {
	s.tracker.Reset()
}

// =============================================================================
// RandomStrategy — базова лінія: випадковий легальний хід.
// =============================================================================

type RandomStrategy struct {
	botID string
	rng   *rand.Rand
}

func NewRandomStrategy(botID string, rng *rand.Rand) *RandomStrategy {
	return &RandomStrategy{botID: botID, rng: rng}
}

func (s *RandomStrategy) Name() string { return "random" }

func (s *RandomStrategy) DecideMain(state *engine.GameState, _ string) (engine.Action, error) {
	me, ok := state.Players[s.botID]
	if !ok || me.IsOut || len(me.Hand) == 0 {
		return engine.Action{}, errNoLegalMove
	}

	// Збираємо легальні індекси руки з урахуванням правила Графині та Принцеси.
	legalIdx := make([]int, 0, len(me.Hand))
	for idx, card := range me.Hand {
		if card == engine.CardPrincess && len(me.Hand) > 1 {
			// Не скидаємо Принцесу добровільно, якщо є вибір.
			continue
		}
		if err := engine.ValidateCountessRule(me, engine.Action{HandIndex: idx}); err != nil {
			continue
		}
		legalIdx = append(legalIdx, idx)
	}
	if len(legalIdx) == 0 {
		legalIdx = []int{0}
	}
	chosenIdx := legalIdx[s.rng.Intn(len(legalIdx))]
	chosenCard := me.Hand[chosenIdx]

	// Обираємо випадкову легальну ціль (для карт, що потребують ціль).
	targetID := s.randomTarget(state)
	var guess engine.CardType
	if chosenCard == engine.CardGuard {
		// Вгадуємо будь-яку карту, окрім Вартового (заборонено правилами).
		guess = engine.CardType(engine.CardPriest + engine.CardType(s.rng.Intn(int(engine.CardPrincess-engine.CardPriest+1))))
		if guess == engine.CardGuard {
			guess = engine.CardPriest
		}
	}

	return engine.Action{
		PlayerID:  s.botID,
		HandIndex: chosenIdx,
		TargetID:  targetID,
		Guess:     guess,
	}, nil
}

func (s *RandomStrategy) randomTarget(state *engine.GameState) string {
	candidates := make([]string, 0)
	for id, p := range state.Players {
		if id == s.botID || p.IsOut || p.IsProtected {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return ""
	}
	return candidates[s.rng.Intn(len(candidates))]
}

func (s *RandomStrategy) DecideChancellor(state *engine.GameState, _ string) (engine.ChancellorResolveAction, error) {
	me, ok := state.Players[s.botID]
	if !ok || len(me.Hand) == 0 {
		return engine.ChancellorResolveAction{}, errNoLegalMove
	}
	keep := s.rng.Intn(len(me.Hand))
	bottom := make([]engine.CardType, 0, len(me.Hand)-1)
	for idx, card := range me.Hand {
		if idx == keep {
			continue
		}
		bottom = append(bottom, card)
	}
	return engine.ChancellorResolveAction{
		PlayerID:      s.botID,
		KeepHandIndex: keep,
		BottomOrder:   bottom,
	}, nil
}

func (s *RandomStrategy) ObserveEvents(_ *engine.GameState, _ []engine.DomainEvent) {}
func (s *RandomStrategy) ResetForNewRound()                                         {}
