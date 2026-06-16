package network

import (
	"context"
	"fmt"
	"time"

	"secret-message/cmd/engine"

	"github.com/sirupsen/logrus"
)

func (r *Room) PostBotAction(botID string, act engine.Action) {
	r.actions <- inboundAction{
		PlayerID:         botID,
		RequestID:        fmt.Sprintf("bot-action-%d", time.Now().UnixNano()),
		IsChancellorType: false,
		Action:           act,
	}
}

func (r *Room) PostBotChancellorAction(botID string, act engine.ChancellorResolveAction) {
	r.actions <- inboundAction{
		PlayerID:         botID,
		RequestID:        fmt.Sprintf("bot-chancellor-%d", time.Now().UnixNano()),
		IsChancellorType: true,
		ChancellorAction: act,
	}
}

func (r *Room) AddBotPlayer(botName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	log := r.log.WithFields(logrus.Fields{"room_id": r.id, "requested_bot_name": botName})

	if len(r.state.TurnOrder) > 0 {
		return fmt.Errorf("cannot join bot: game has already started")
	}
	if len(r.state.Players) >= MAX_PLAYERS {
		return fmt.Errorf("room is full, cannot add bot")
	}

	// 1. Отримуємо список усіх легітимних ботів із нашого сховища
	staticBots := r.store.GetBotNames()
	var selectedBotPointer *string // Тимчасовий покажчик на обраного бота

	// Якщо ім'я не передано, або передано конкретне — шукаємо підходящого вільного бота
	for _, sb := range staticBots {
		// Якщо користувач просив конкретного бота, перевіряємо ім'я
		if botName != "" && sb.Username != botName {
			continue
		}

		// Перевіряємо, чи цей бот вже є в кімнаті (за його константним ID)
		botIDStr := sb.ID.String()
		if _, exists := r.state.Players[botIDStr]; !exists {
			// Бот вільний для цієї кімнати!
			selectedBot := sb
			selectedBotPointer = &botIDStr
			botName = selectedBot.Username // гарантуємо правильне ім'я
			break
		}
	}

	if selectedBotPointer == nil {
		if botName != "" {
			return fmt.Errorf("bot with name %s is already in this room or does not exist", botName)
		}
		return fmt.Errorf("all available system bots are already in this room")
	}

	botID := *selectedBotPointer

	// 2. Додаємо бота у GameState, використовуючи його константний ID та роль
	r.state.Players[botID] = engine.Player{
		ID:               botID,
		Username:         botName,
		UserRole:         "bot", // Маркер ролі для логіки кімнати та ШІ
		AvatarSeed:       botName,
		Hand:             []engine.CardType{},
		DiscardPile:      []engine.CardType{},
		Score:            0,
		SpyPointsAwarded: false,
	}

	playersCount := len(r.state.Players)
	if err := r.saveToDB(context.Background()); err != nil {
		return fmt.Errorf("failed to save state after bot join: %w", err)
	}

	if r.hub != nil {
		go r.hub.NotifyLobbyUpdate()
	}
	go r.BroadcastState(UpdateTypeRoomUpdated, nil)

	log.Infof("Бот %s [ID: %s] успішно приєднався. Усього гравців: %d", botName, botID, playersCount)
	return nil
}
