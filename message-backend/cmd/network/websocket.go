package network

import (
	"encoding/json"
	"net/http"

	"secret-message/cmd/auth"
	"secret-message/cmd/engine"
	"secret-message/cmd/storage"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WSMessage struct {
	Type             string                          `json:"type"`
	RoomID           string                          `json:"room_id"`
	RequestID        string                          `json:"request_id"`
	Action           *engine.Action                  `json:"action,omitempty"`
	ChancellorAction *engine.ChancellorResolveAction `json:"chancellor_action,omitempty"`
}

type wsAckResponse struct {
	Status    string `json:"status"`
	RequestID string `json:"request_id"`
	Type      string `json:"type"`
}

type wsPongResponse struct {
	Type string `json:"type"`
}

type Server struct {
	hub      *Hub
	gateway  *Gateway
	log      *logrus.Logger
	upgrader websocket.Upgrader
	store    *storage.Storage
}

func NewServer(hub *Hub, gateway *Gateway, logger *logrus.Logger, store *storage.Storage) *Server {
	return &Server{
		hub:     hub,
		gateway: gateway,
		log:     logger,
		store:   store,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	// 1. Валідуємо користувача з контексту Middleware
	claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
	if !ok {
		s.log.Error("WebSocket з'єднання відхилено: користувач не авторизований Middleware")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	currentPlayerID := claims.UserID

	// 1a. Жорстка валідація UUID на найвищому шарі.
	// Якщо токен містить нестандартний UserID, відхиляємо з'єднання
	// ще ДО апгрейду WebSocket — нижні шари (Hub/Room/AddPlayer)
	// можуть покладатись на те, що ID гарантовано є валідним UUID,
	// і вже не зобов'язані повторно це перевіряти і логувати warning'и.
	if _, parseErr := uuid.Parse(currentPlayerID); parseErr != nil {
		s.log.WithFields(logrus.Fields{
			"player_id": currentPlayerID,
			"error":     parseErr.Error(),
		}).Error("WebSocket з'єднання відхилено: UserID не є валідним UUID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. Робимо апгрейд з'єднання до WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.WithField("error", err.Error()).Error("Помилка апгрейду до WebSocket")
		return
	}

	s.log.Infof("Користувач %s `%s` успішно підключився (IP: %s)", claims.Username, currentPlayerID, r.RemoteAddr)

	// ЗМІНЕНО: Реєструємо клієнта в нашому глобальному Gateway десктопу
	s.gateway.AddClient(currentPlayerID, conn)
	s.gateway.mu.RLock()
	client := s.gateway.clients[currentPlayerID]
	s.gateway.mu.RUnlock()

	sendToClient := func(msg any) {
		data, _ := json.Marshal(msg)
		select {
		case client.SendChan <- data:
		default:
			s.log.Warn("Канал клієнта переповнений")
		}
	}

	if s.hub != nil {
		// Припускаємо, що у вашому Hub є метод отримання кімнат,
		// або ми можемо примусово викликати Broadcast для оновлення лобі
		s.gateway.BroadcastToDesktop(s.hub.GetLobbyRooms())
	}

	var currentRoomID string

	// Гарантоване очищення клієнта при виході з методу сокету
	defer func() {
		//Видаляємо клієнта з глобального Gateway
		s.gateway.RemoveClient(currentPlayerID)
		conn.Close()

		if currentRoomID != "" {
			if room, _ := s.hub.GetOrCreateRoom(currentRoomID); room != nil {
				room.UnregisterClient(currentPlayerID)
				s.log.Infof("Користувач %s відключився, сесію оброблено в кімнаті %s", currentPlayerID, currentRoomID)
			}
		}
	}()

	// 3. Нескінченний цикл обробки повідомлень від конкретного гравця
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.log.WithFields(logrus.Fields{
					"player_id": currentPlayerID,
					"error":     err.Error(),
				}).Warn("Немотивоване закриття сокету клієнтом")
			} else {
				s.log.Infof("Користувач %s планово розірвав з'єднання", currentPlayerID)
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			s.log.WithFields(logrus.Fields{
				"player_id": currentPlayerID,
				"raw_msg":   string(msgBytes),
				"error":     err.Error(),
			}).Warn("Отримано некоректний JSON від клієнта")
			s.sendError(client, "invalid_json", "Помилка парсингу JSON")
			continue
		}

		// Якщо клієнт на /desktop шле сервісні повідомлення (наприклад, PING чи оновлення статусу),
		// йому не обов'язково передавати room_id.
		if msg.Type == "" {
			s.sendError(client, "missing_fields", "Поле type є обов'язковим")
			continue
		}

		// Якщо повідомлення суто лобі/сервісне і не потребує кімнати — обробляємо його окремо
		if msg.RoomID == "" {
			if msg.Type == "PING" {
				sendToClient(wsPongResponse{Type: "PONG"})
			}
			continue
		}

		room, err := s.hub.GetOrCreateRoom(msg.RoomID)
		if err != nil {
			s.sendError(client, "room_error", err.Error())
			continue
		}

		if currentRoomID != msg.RoomID {
			if currentRoomID != "" {
				if oldRoom, _ := s.hub.GetOrCreateRoom(currentRoomID); oldRoom != nil {
					oldRoom.UnregisterClient(currentPlayerID)
				}
			}

			currentRoomID = msg.RoomID
			room.RegisterClient(currentPlayerID, client)
		}

		// 4. Обробка бізнес-повідомлень гри
		switch msg.Type {
		case MsgJoin:
			err = room.AddPlayer(currentPlayerID)
			if err != nil {
				s.sendError(client, "join_failed", err.Error())
				continue
			}

		case MsgStartGame:
			err = room.StartGame()
			if err != nil {
				s.sendError(client, "start_failed", err.Error())
				continue
			}

		case MsgNextRound:
			if err := room.NextRound(); err != nil {
				s.sendError(client, "failed_to_next_round", err.Error())
			}

		case MsgAction:
			if msg.Action == nil {
				s.sendError(client, "missing_payload", "Поле action відсутнє")
				continue
			}
			err = room.SubmitAction(s.hub.ctx, currentPlayerID, msg.RequestID, *msg.Action)
			if err != nil {
				// Доменна помилка від engine.Apply: витягуємо стабільний
				// error_code (наприклад "ERR_PLAYER_ALREADY_OUT") і шлемо
				// його у пакеті — фронт сам перекладе.
				s.sendEngineError(client, msg.RequestID, "action_rejected", err)
				continue
			}

		case MsgChancellorResolve:
			if msg.ChancellorAction == nil {
				s.sendError(client, "missing_payload", "Поле chancellor_action відсутнє")
				continue
			}
			err = room.SubmitChancellorAction(s.hub.ctx, currentPlayerID, msg.RequestID, *msg.ChancellorAction)
			if err != nil {
				s.sendEngineError(client, msg.RequestID, "action_rejected", err)
				continue
			}

		case MsgLeave:
			if err := room.HandlePlayerLeave(s.hub.ctx, currentPlayerID); err != nil {
				s.sendError(client, "leave_failed", err.Error())
				continue
			}
			// Після виходу з гри відписуємо клієнта від кімнати, але сокет залишається активним
			room.UnregisterClient(currentPlayerID)
			currentRoomID = ""

		default:
			s.sendError(client, "unknown_type", "Невідомий тип повідомлення")
			continue
		}

		s.log.WithFields(logrus.Fields{
			"type":      msg.Type,
			"room_id":   msg.RoomID,
			"player_id": currentPlayerID,
			"req_id":    msg.RequestID,
		}).Debug("Запит успішно оброблено двигуном")

		sendToClient(wsAckResponse{
			Status:    "success",
			RequestID: msg.RequestID,
			Type:      msg.Type + "_ACK",
		})
	}
}

// sendError шле клієнту помилку транспортного рівня (некоректний JSON,
// відсутні обов'язкові поля, помилка авторизації тощо).
//
// Поле `code` тут — це КАТЕГОРІЯ помилки на рівні WS-протоколу
// ("invalid_json", "missing_payload", "join_failed" і т.д.).
// Поле `message` — англомовний/змішаний dev-message, який МОЖЕ бути
// показаний у девтулзах, але НЕ має використовуватись як готовий текст
// для UI без локалізації.
//
// Для помилок ігрового рушія використовуйте sendEngineError — він додає
// поле "error_code" зі стабільним кодом (engine.ErrorCode), яке фронт
// мапить у локалізовані рядки.
func (s *Server) sendError(client *GlobalClient, code, message string) {
	data, _ := json.Marshal(map[string]any{
		"status":  "error",
		"code":    code,
		"message": message,
	})
	select {
	case client.SendChan <- data:
	default:
		s.log.Warn("Не вдалося відправити помилку: канал переповнений")
	}
}

// sendEngineError шле клієнту доменну помилку від engine.Apply / room.processAction.
//
// Вихідний JSON-пакет:
//
//	{
//	  "status":     "error",
//	  "code":       "action_rejected",        // транспортна категорія (legacy)
//	  "error_code": "ERR_PLAYER_ALREADY_OUT", // СТАБІЛЬНИЙ код для i18n на фронті
//	  "message":    "ERR_PLAYER_ALREADY_OUT: player_id=p2", // dev-only
//	  "request_id": "<echo>",
//	  "details":    { ... }                    // опціонально, з GameError.Details
//	}
//
// Фронтенд має зчитувати ВИКЛЮЧНО error_code і не парсити message для UI.
func (s *Server) sendEngineError(client *GlobalClient, requestID, transportCode string, err error) {
	if err == nil {
		return
	}

	// Витягаємо стабільний код. Якщо err не є *engine.GameError —
	// CodeOf поверне ErrInternal, що теж буде нормально оброблено фронтом.
	errorCode := engine.CodeOf(err)

	packet := map[string]any{
		"status":     "error",
		"code":       transportCode,
		"error_code": string(errorCode),
		"message":    err.Error(), // тільки для логів/девтулзів
		"request_id": requestID,
	}

	// Якщо це наша типізована помилка з контекстом — приклеюємо details.
	if ge, ok := engine.AsGameError(err); ok && len(ge.Details) > 0 {
		packet["details"] = ge.Details
	}

	data, marshalErr := json.Marshal(packet)
	if marshalErr != nil {
		s.log.WithField("error", marshalErr.Error()).
			Error("Не вдалося серіалізувати engine-error пакет")
		return
	}

	select {
	case client.SendChan <- data:
	default:
		s.log.Warn("Не вдалося відправити engine-error: канал переповнений")
	}
}
