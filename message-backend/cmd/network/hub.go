// hub.go
package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"secret-message/cmd/engine"
	"secret-message/cmd/storage"

	"github.com/sirupsen/logrus"
)

// Hub керує життєвим циклом усіх активних ігрових кімнат у пам'яті
type Hub struct {
	store  *storage.Storage
	log    *logrus.Logger
	rooms  map[string]*Room
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Посилання на шлюз подій (можна передати через інтерфейс або вказати напряму)
	OnLobbyChanged func(lobbyRooms []RoomLobbyInfo)
}

// NewHub створює новий екземпляр диспетчера кімнат
func NewHub(store *storage.Storage, logger *logrus.Logger) *Hub {
	return &Hub{
		store: store,
		log:   logger,
		rooms: make(map[string]*Room),
	}
}

// Start запускає Hub та визначає його контекст життєвого циклу
func (h *Hub) Start(ctx context.Context) {
	h.mu.Lock()
	h.ctx, h.cancel = context.WithCancel(ctx)
	h.mu.Unlock()
	h.log.Info("Центральний менеджер кімнат (Hub) успішно запущено")
}

// Stop зупиняє Hub та всі дочірні кімнати
func (h *Hub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cancel != nil {
		h.cancel()
	}
	h.log.Info("Центральний менеджер кімнат (Hub) зупинено")
}

// GetOrCreateRoom знаходить існуючу кімнату в пам'яті або ініціалізує її з БД/нуля
func (h *Hub) GetOrCreateRoom(roomID string) (*Room, error) {
	// 1. Швидка перевірка під RLock (читання), чи є кімната в пам'яті
	h.mu.RLock()
	r, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if exists {
		return r, nil
	}

	// 2. Якщо немає, блокуємо на запис для створення/завантаження
	h.mu.Lock()
	if h.ctx == nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("hub is not started yet")
	}
	// Double-checking
	if r, exists = h.rooms[roomID]; exists {
		h.mu.Unlock()
		return r, nil
	}
	hubCtx := h.ctx
	h.mu.Unlock()

	h.log.WithField("room_id", roomID).Debug("Кімнату не знайдено в пам'яті, завантаження з БД...")

	// Викликаємо NewRoom (повільна операція, I/O з БД виконується без блокування всього хабу)
	newRoom, err := NewRoom(hubCtx, roomID, "", h.store, h.log, h)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize room %s: %w", roomID, err)
	}

	// 3. Знову захоплюємо лок хабу, щоб безпечно зберегти кімнату в мапу
	h.mu.Lock()
	defer h.mu.Unlock()

	// Ще одна перевірка на випадок, якщо інша горутина встигла створити кімнату, поки ми ходили в БД
	if existingRoom, exists := h.rooms[roomID]; exists {
		return existingRoom, nil
	}

	// Запускаємо фоновий цикл обробки подій кімнати
	go newRoom.Start(hubCtx)

	h.rooms[roomID] = newRoom
	h.log.WithField("room_id", roomID).Info("Нову кімнату успішно додано в пам'ять хабу та запущено")

	return newRoom, nil
}

// CreateNewRoom явно створює абсолютно нову кімнату в пам'яті хабу з призначеним хостом
func (h *Hub) CreateNewRoom(roomID string, hostID string) (*Room, error) {
	h.mu.Lock()
	if h.ctx == nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("hub is not started yet")
	}

	// Перевіряємо, чи немає колізії ID кімнат у пам'яті
	if _, exists := h.rooms[roomID]; exists {
		h.mu.Unlock()
		return nil, fmt.Errorf("room with ID %s already exists in memory", roomID)
	}
	hubCtx := h.ctx
	h.mu.Unlock() // Відпускаємо лок хабу перед ініціалізацією кімнати та записом в БД всередині NewRoom

	h.log.WithFields(logrus.Fields{
		"room_id": roomID,
		"host_id": hostID,
	}).Info("Створення нової ігрової кімнати хостом через API...")

	// Створюємо екземпляр кімнати (завдяки порожньому стану в БД, NewRoom створить чистий стейт)
	newRoom, err := NewRoom(hubCtx, roomID, hostID, h.store, h.log, h)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate room: %w", err)
	}

	// Автоматично додаємо хоста як першого гравця в ігровий стейт кімнати
	if err := newRoom.AddPlayer(hostID); err != nil {
		return nil, fmt.Errorf("failed to add host to their own room: %w", err)
	}

	go newRoom.Start(hubCtx)

	// Повертаємо лок хабу для реєстрації кімнати в мапі
	h.mu.Lock()
	defer h.mu.Unlock()

	// Перевірка на випадок паралельного запиту з таким же ID
	if existingRoom, exists := h.rooms[roomID]; exists {
		return existingRoom, nil
	}

	h.rooms[roomID] = newRoom

	fmt.Println(newRoom)

	go h.NotifyLobbyUpdate()

	return newRoom, nil
}

// GetLobbyRooms повертає список інформаційних об'єктів для всіх кімнат,
// які очікують гравців (ігри в яких ще НЕ розпочалися)
func (h *Hub) GetLobbyRooms() []RoomLobbyInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var lobbyList []RoomLobbyInfo

	for _, room := range h.rooms {
		info := room.GetLobbyInfo()
		// Нам потрібні лише кімнати, які ще збирають лобі (гра не стартувала)
		// та де є вільні місця
		if !info.IsStarted && info.PlayerCount < info.MaxPlayers {
			lobbyList = append(lobbyList, info)
		}
	}

	return lobbyList
}

func (h *Hub) NotifyLobbyUpdate() {
	if h.OnLobbyChanged != nil {
		rooms := h.GetLobbyRooms()
		h.OnLobbyChanged(rooms)
	}
}

// ForwardAction маршрутизує ігрову дію гравця у відповідну кімнату
func (h *Hub) ForwardAction(roomID string, playerID string, reqID string, action engine.Action) error {
	room, err := h.GetOrCreateRoom(roomID)
	if err != nil {
		return err
	}

	// Передаємо екшен у чергу повідомлень кімнати (безпечно для конкурентного середовища)
	return room.SubmitAction(h.ctx, playerID, reqID, action)
}

// CloseRoom видаляє кімнату з пам'яті хабу (наприклад, коли гра завершилась)
func (h *Hub) CloseRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.rooms[roomID]; exists {
		// Тут у майбутньому варто викликати локальний room.Stop(),
		// якщо ви додасте Context / CancelFunc всередину самої структури Room,
		// щоб примусово гасити горутину room.Start().
		delete(h.rooms, roomID)
		h.log.WithField("room_id", roomID).Info("Кімнату вивантажено з оперативної пам'яті хабу")
	}
}

// GetActiveRooms повертає список ідентифікаторів усіх кімнат, які зараз активні в пам'яті
func (h *Hub) GetActiveRooms() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]string, 0, len(h.rooms))
	for id := range h.rooms {
		rooms = append(rooms, id)
	}
	return rooms
}

// DisconnectUserGlobally знаходить користувача в усіх активних кімнатах,
// закриває його WebSocket з'єднання та видаляє його з ігрових сесій.
func (h *Hub) DisconnectUserGlobally(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.log.Infof("Запуск глобального відключення для заблокованого користувача: %s", userID)

	for roomID, room := range h.rooms {
		room.mu.Lock()

		// Перевіряємо, чи є в цієї кімнати активне сокет-з'єднання з цим користувачем
		if client, exists := room.conns[userID]; exists {
			// Надсилаємо фінальне повідомлення через SendChan,
			// щоб лише StartWriter писав у *websocket.Conn (уникаємо concurrent write panic)
			notice, _ := json.Marshal(map[string]string{
				"status":  "error",
				"code":    "account_suspended",
				"message": "Your account has been suspended.",
			})
			select {
			case client.SendChan <- notice:
			default:
			}

			// Закриваємо з'єднання — StartWriter завершиться при наступній спробі запису.
			// Власне очищення SendChan/мапи Gateway виконає RemoveClient через defer у HandleWS.
			_ = client.Conn.Close()

			// Видаляємо з'єднання з мапи кімнати (робимо інлайново замість UnregisterClient, щоб уникнути self-deadlock)
			delete(room.conns, userID)
			room.log.Infof("Користувача %s примусово видалено з кімнати %s", userID, roomID)

			// Якщо в кімнаті більше немає гравців, її можна закрити
			if len(room.conns) == 0 {
				room.log.Infof("Кімната %s спорожніла. Видалення кімнати з Hub.", roomID)
				delete(h.rooms, roomID)
			}
		}

		room.mu.Unlock()
	}
}
