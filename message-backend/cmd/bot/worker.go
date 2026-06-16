package bot

import (
	"context"
	"time"

	"secret-message/cmd/engine"

	"github.com/sirupsen/logrus"
)

// Робимо місток до нашої Room з пакету network
type RoomGateway interface {
	PostBotAction(botID string, action engine.Action)
	PostBotChancellorAction(botID string, action engine.ChancellorResolveAction)
}

type BotManager struct {
	RoomID   string
	Trackers map[string]*DeckTracker
	Brain    map[string]*BotBrain
	log      *logrus.Logger
}

func NewBotManager(roomID string, botIDs []string, logger *logrus.Logger) *BotManager {
	trackers := make(map[string]*DeckTracker)
	brains := make(map[string]*BotBrain)

	for _, id := range botIDs {
		trackers[id] = NewDeckTracker(id)
		brains[id] = NewBotBrain(id, logger)
	}

	return &BotManager{
		RoomID:   roomID,
		Trackers: trackers,
		Brain:    brains,
		log:      logger,
	}
}

func (bm *BotManager) HandleGameUpdate(state *engine.GameState, events []engine.DomainEvent) {
	for _, tracker := range bm.Trackers {
		for _, event := range events {
			tracker.HandleDomainEvent(event)
			if event.Type == engine.EventChancellorResolved {
				// Якщо твій payload приходить як вказівник чи структура, кастимо її тут:
				if payload, ok := event.Payload.(engine.ChancellorResolvedPayload); ok {
					if payload.PlayerID == tracker.BotID {
						tracker.RecordChancellorAction(engine.ChancellorResolveAction{
							PlayerID:    payload.PlayerID,
							BottomOrder: payload.BottomOrder,
						})
					}
				}
			}
		}
		tracker.TrackGameState(state)
	}
}

func (bm *BotManager) Reset() {
	for _, tracker := range bm.Trackers {
		tracker.Reset()
	}
}

func (bm *BotManager) RunCheck(ctx context.Context, state *engine.GameState, gateway RoomGateway) {
	if state.IsGameOver || state.Phase == engine.PhaseRoundEnd || state.Phase == "FINISHED" {
		return
	}

	var activePlayerID string
	switch state.Phase {
	case engine.PhaseMainAction:
		if state.CurrentTurn < 0 || state.CurrentTurn >= len(state.TurnOrder) {
			return
		}
		activePlayerID = state.TurnOrder[state.CurrentTurn]
	case engine.PhaseResolveChancellor:
		activePlayerID = state.PendingAction.PlayerID
	}

	brain, isBot := bm.Brain[activePlayerID]
	tracker := bm.Trackers[activePlayerID]
	if !isBot || tracker == nil {
		return
	}

	go func(targetBotID string, targetBrain *BotBrain, targetTracker *DeckTracker) {
		select {
		case <-time.After(time.Duration(1000+time.Now().UnixNano()%1500) * time.Millisecond):
		case <-ctx.Done():
			return
		}

		decision, err := targetBrain.Think(state, targetTracker)
		if err != nil {
			bm.log.WithFields(logrus.Fields{
				"room_id": bm.RoomID,
				"bot_id":  targetBotID,
				"error":   err.Error(),
			}).Error("Бот не зміг прийняти рішення")
			return
		}

		switch act := decision.(type) {
		case engine.Action:
			act.PlayerID = targetBotID
			gateway.PostBotAction(targetBotID, act)
		case engine.ChancellorResolveAction:
			act.PlayerID = targetBotID
			gateway.PostBotChancellorAction(targetBotID, act)
		}
	}(activePlayerID, brain, tracker)
}
