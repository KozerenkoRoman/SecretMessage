// =============================================================================
// network/room.go
//
// Room — центральний об'єкт життєвого циклу однієї партії.
//
// Концепція concurrency:
//   Room — це shared mutable state, до якого мають доступ:
//     • HTTP/WS-обробники (виклики AddPlayer / StartGame / NextRound /
//       SubmitAction / SubmitChancellorAction / HandlePlayerLeave / ...)
//     • Власна горутина Start(ctx), що споживає чергу r.actions і
//       обробляє AFK-таймер.
//     • BroadcastState — викликається з горутин після кожного апдейту.
//
// Правила синхронізації:
//   1. УСЯ робота зі станом гри (r.state, r.conns, r.recentRequests,
//      r.hostID) виконується ВИКЛЮЧНО під r.mu.
//   2. Жоден метод, що тримає r.mu.Lock(), НЕ викликає інші методи Room,
//      що знову запитують r.mu (RLock або Lock) — це призводить до
//      self-deadlock при апгрейді RWMutex.
//      Замість r.GetState() використовуємо безпосередній доступ r.state.*
//      доки тримаємо лок.
//   3. ВСІ дії з зовнішнім світом (saveToDB, BroadcastState, hub.NotifyLobbyUpdate)
//      виконуються або через `go ...` (поза локом) або у виокремлених
//      областях коду ПІСЛЯ r.mu.Unlock(). saveToDB використовує внутрішній
//      lock БД; ми викликаємо її під r.mu (де читаємо r.state) виключно
//      синхронно й коротко.
//   4. Методи з суфіксом *Locked припускають, що викликач уже тримає r.mu.
//
// Дизайн анти-deadlock:
//   • r.mu — sync.RWMutex. RLock допустимий лише на швидкі snapshot-операції
//     (GetLobbyInfo, GetState, BroadcastState — копіювання conns/state).
//   • Усі мутаційні шляхи (Lock) йдуть через короткий критичний секцію,
//     потім — звільняють лок і виконують side-effects.
//
// Дизайн анти-leak:
//   • turnTimer — використовуємо ідіоматичний паттерн drain-then-Reset.
//   • recentRequests — кліренимо при виході гравця у UnregisterClient та
//     HandlePlayerLeave (lobby-branch), щоб мапа не росла нескінченно.
// =============================================================================

package network

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"secret-message/cmd/engine"
	"secret-message/cmd/storage"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// =============================================================================
// Constants & helper types
// =============================================================================

const (
	// MAX_PLAYERS — максимум учасників у одній кімнаті.
	MAX_PLAYERS = 4
	// recentRequestsCap — м'який ліміт історії request_id на одного гравця
	// у механізмі дедуплікації. При перевищенні видаляється найстаріший запис.
	recentRequestsCap = 20
	// systemPlayerID — псевдо-ID для системних дій (наприклад, маркер
	// перезапуску таймера у внутрішній черзі r.actions).
	systemPlayerID = "SYSTEM"
)

// realRNG — стандартна реалізація engine.RNG поверх math/rand.
type realRNG struct{}

func (realRNG) Intn(n int) int { return rand.Intn(n) }

// realClock — стандартна реалізація engine.Clock.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// inboundAction — внутрішня команда, яку зовнішні API-ручки кладуть у
// канал r.actions, щоб її виконав єдиний "ігровий" goroutine у Start().
// errChan дозволяє синхронно повернути результат викликачу.
type inboundAction struct {
	PlayerID         string                         `json:"player_id"`
	RequestID        string                         `json:"request_id"`
	IsChancellorType bool                           `json:"is_chancellor_type"`
	Action           engine.Action                  `json:"action"`
	ChancellorAction engine.ChancellorResolveAction `json:"chancellor_action"`
	errChan          chan error
}

// =============================================================================
// Room
// =============================================================================

// Room керує одним матчем: учасники, стан гри, черга дій, таймери.
type Room struct {
	id        string
	hub       *Hub
	hostID    string
	createdAt time.Time
	store     *storage.Storage

	// actions — черга команд. Ємність 100 — захист від короткочасних сплесків;
	// при переповненні викликач отримає помилку via select default.
	actions chan inboundAction

	// recentRequests: playerID → reqID → timestamp.
	// Захищено r.mu. Очищується при виході гравця, щоб уникнути витоку пам'яті.
	recentRequests map[string]map[string]time.Time

	mu    sync.RWMutex
	rng   engine.RNG
	clock engine.Clock

	conns map[string]*GlobalClient
	log   *logrus.Logger

	state         engine.GameState
	turnStartedAt time.Time
}

// RoomLobbyInfo — DTO для списку лобі.
type RoomLobbyInfo struct {
	ID             string   `json:"room_id"`
	HostID         string   `json:"host_id"`
	PlayerCount    int      `json:"player_count"`
	MaxPlayers     int      `json:"max_players"`
	IsStarted      bool     `json:"is_started"`
	Players        []string `json:"players"`
	PlayerNames    []string `json:"player_names"`
	CreatedAt      int64    `json:"created_at"`
	HostAvatarSeed string   `json:"host_avatar_seed"`
}

// =============================================================================
// Construction
// =============================================================================

// NewRoom створює нову кімнату або відновлює стан з БД.
func NewRoom(ctx context.Context, id string, hostID string, store *storage.Storage, logger *logrus.Logger, hub *Hub) (*Room, error) {
	r := &Room{
		id:             id,
		hostID:         hostID,
		createdAt:      time.Now(),
		store:          store,
		hub:            hub,
		actions:        make(chan inboundAction, 100),
		recentRequests: make(map[string]map[string]time.Time),
		rng:            realRNG{},
		clock:          realClock{},
		log:            logger,
		conns:          make(map[string]*GlobalClient),
		turnStartedAt:  time.Now(),
	}

	if r.hostID == "" {
		r.hostID = systemPlayerID
	}

	log := r.log.WithField("room_id", r.id)
	dbRoom, err := store.GetRoomByID(ctx, id)
	if err == nil {
		var state engine.GameState
		if err := json.Unmarshal(dbRoom.State, &state); err != nil {
			return nil, fmt.Errorf("failed to unmarshal game state: %w", err)
		}
		r.state = state
		if len(state.TurnOrder) > 0 {
			r.hostID = state.TurnOrder[0]
		}
		log.WithField("seq", state.Sequence).Info("Успішно відновлено стан кімнати із бази даних")
	} else {
		log.Info("Кімнату не знайдено в БД, ініціалізуємо новий стан")
		r.state = engine.GameState{
			Seed:      time.Now().UnixNano(),
			Sequence:  0,
			Phase:     engine.PhaseMainAction,
			Players:   make(map[string]engine.Player),
			TurnOrder: []string{},
		}
		if err := r.saveToDB(ctx); err != nil {
			return nil, fmt.Errorf("failed to save initial room state: %w", err)
		}
	}

	return r, nil
}

// =============================================================================
// Read-only API (під RLock)
// =============================================================================

// GetLobbyInfo — швидкий snapshot для лобі-списку.
func (r *Room) GetLobbyInfo() RoomLobbyInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]string, 0, len(r.state.Players))
	names := make([]string, 0, len(r.state.Players))
	for pID, p := range r.state.Players {
		players = append(players, pID)
		names = append(names, p.Username)
	}

	return RoomLobbyInfo{
		ID:             r.id,
		HostID:         r.hostID,
		PlayerCount:    len(r.state.Players),
		MaxPlayers:     MAX_PLAYERS,
		IsStarted:      len(r.state.TurnOrder) > 0,
		Players:        players,
		PlayerNames:    names,
		CreatedAt:      r.createdAt.Unix(),
		HostAvatarSeed: r.state.Players[r.hostID].AvatarSeed,
	}
}

// GetState повертає глибоку копію стану (deep clone), яку викликач може
// модифікувати безпечно. НЕ використовуйте з тримання r.mu.Lock() — буде
// self-deadlock на RLock.
func (r *Room) GetState() engine.GameState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.Clone()
}

// =============================================================================
// Connection lifecycle
// =============================================================================

func (r *Room) RegisterClient(playerID string, client *GlobalClient) {
	r.mu.Lock()
	r.conns[playerID] = client
	r.log.WithFields(logrus.Fields{"room_id": r.id, "player_id": playerID}).
		Debug("Сокет гравця успішно зареєстровано в кімнаті")
	r.mu.Unlock()
	// ПРИБРАНО: go r.BroadcastState(UpdateTypeRoomUpdated, nil)
}

// UnregisterClient — обробка повного дисконекту.
//
// Поведінка:
//   - Завжди прибирає сокет із r.conns.
//   - Завжди прибирає історію request_id у r.recentRequests (запобігає memory leak).
//   - Якщо гра ще не стартувала — гравець вилучається з лобі.
//   - Якщо лобі спорожніло — кімната видаляється з БД та з Hub.
func (r *Room) UnregisterClient(playerID string) {
	r.mu.Lock()
	// Видаляємо сокет
	delete(r.conns, playerID)
	delete(r.recentRequests, playerID)

	isGameStarted := len(r.state.TurnOrder) > 0

	if isGameStarted {
		// ВИКЛИКАЄМО ЛОГІКУ ВИКЛЮЧЕННЯ
		err := r.eliminatePlayer(playerID)
		if err != nil {
			r.log.Errorf("Помилка при спробі виключити гравця: %v", err)
		}

		// Зберігаємо стан
		r.saveToDB(context.Background())
	} else {
		// Логіка для лобі (якщо гра не почалася)
		delete(r.state.Players, playerID)
	}

	// Перевіряємо чи кімната порожня для видалення
	roomIsEmpty := len(r.state.Players) == 0
	r.mu.Unlock()

	if roomIsEmpty {
		go func(store *storage.Storage, id string) {
			_ = store.DeleteRoomByID(context.Background(), id)
		}(r.store, r.id)
		if r.hub != nil {
			r.hub.CloseRoom(r.id)
		}
	} else {
		// Оновлюємо стан для інших гравців
		go r.BroadcastState(UpdateTypeRoomUpdated, nil)
	}
}

// HandlePlayerLeave — гравець натиснув "Вийти" (MsgLeave) у живій грі.
func (r *Room) HandlePlayerLeave(ctx context.Context, playerID string) error {
	r.mu.Lock()
	log := r.log.WithFields(logrus.Fields{"room_id": r.id, "player_id": playerID})

	// --- Гілка 1: Вихід з лобі ДО початку гри ---
	if len(r.state.TurnOrder) == 0 {
		if _, exists := r.state.Players[playerID]; !exists {
			r.mu.Unlock()
			return fmt.Errorf("player %s is not in this room", playerID)
		}
		delete(r.state.Players, playerID)
		// Гравець більше не активний у кімнаті — звільняємо його request-історію.
		delete(r.recentRequests, playerID)
		log.Info("Гравця видалено з лобі за його запитом MsgLeave")

		if len(r.state.Players) == 0 {
			r.mu.Unlock()
			go func(store *storage.Storage, id string) {
				_ = store.DeleteRoomByID(context.Background(), id)
			}(r.store, r.id)
			if r.hub != nil {
				r.hub.CloseRoom(r.id)
				go r.hub.NotifyLobbyUpdate()
			}
			return nil
		}

		if err := r.saveToDB(ctx); err != nil {
			r.mu.Unlock()
			return fmt.Errorf("failed to save state after leave: %w", err)
		}
		r.mu.Unlock()
		go r.BroadcastState(UpdateTypeRoomUpdated, nil)
		if r.hub != nil {
			go r.hub.NotifyLobbyUpdate()
		}
		return nil
	}

	// --- Гілка 2: Капітуляція ПІД ЧАС активної партії ---
	player, exists := r.state.Players[playerID]
	if !exists {
		r.mu.Unlock()
		return fmt.Errorf("player %s is not in this room", playerID)
	}
	if r.state.Phase == engine.PhaseFinished {
		r.mu.Unlock()
		return fmt.Errorf("game has already finished")
	}
	if player.IsOut {
		r.mu.Unlock()
		return fmt.Errorf("player has already been eliminated")
	}

	log.Info("Гравець вийшов з активної партії — фіксуємо поразку")

	for _, card := range player.Hand {
		if card != engine.CardSpy {
			player.DiscardPile = append(player.DiscardPile, card)
		}
	}
	player.Hand = nil
	player.IsOut = true
	player.IsProtected = false
	r.state.Players[playerID] = player
	r.state.Sequence++

	if r.state.Phase == engine.PhaseResolveChancellor {
		currActivePlayerID := r.state.TurnOrder[r.state.CurrentTurn]
		if currActivePlayerID == playerID {
			r.state.Phase = engine.PhaseMainAction
		}
	}

	leaveEvent := engine.DomainEvent{
		EventID:   r.state.Sequence * 1000,
		Type:      engine.EventPlayerLeft,
		Payload:   engine.PlayerLeftPayload{PlayerID: playerID},
		Timestamp: r.clock.Now(),
	}
	events := []engine.DomainEvent{leaveEvent}

	// Скільки лишилось у живих?
	aliveIDs := make([]string, 0, len(r.state.Players))
	for _, p := range r.state.Players {
		if !p.IsOut {
			aliveIDs = append(aliveIDs, p.ID)
		}
	}

	// --- Кінець партії: лишився ≤ 1 живий ---
	if len(aliveIDs) <= 1 {
		if len(aliveIDs) == 1 {
			winnerID := aliveIDs[0]
			w := r.state.Players[winnerID]
			w.Score++
			r.state.Players[winnerID] = w
			r.state.WinnerID = winnerID
			r.state.Phase = engine.PhaseFinished

			events = append(events, engine.DomainEvent{
				EventID:   r.state.Sequence*1000 + 1,
				Type:      engine.EventRoundEnd,
				Payload:   engine.RoundEndPayload{WinnerID: winnerID, Reason: engine.ReasonRoundOpponentLeft},
				Timestamp: r.clock.Now(),
			})
			log.WithField("winner_id", winnerID).Info("Партію завершено — переможець визначений після виходу опонента")
		} else {
			log.Warn("Партію завершено без переможця: усі гравці вийшли")
		}

		// Якщо гра завершилась через вихід опонентів, також підраховуємо Шпигуна
		startEventID := r.state.Sequence*1000 + 2
		roundResult := engine.ResolveRoundEnd(r.state, r.clock, startEventID)
		r.state = roundResult.NewState
		events = append(events, roundResult.DomainEvents...)

		if err := r.saveToDB(ctx); err != nil {
			r.mu.Unlock()
			return fmt.Errorf("failed to save final state after leave: %w", err)
		}

		stateForArchive := r.state.Clone()
		r.mu.Unlock()

		r.archiveFinishedGame(stateForArchive)
		go r.BroadcastState(UpdateTypeRoomUpdated, events)
		if r.hub != nil {
			go r.hub.NotifyLobbyUpdate()
		}
		return nil
	}

	// --- Партія триває: передаємо хід наступному ---
	if len(r.state.TurnOrder) > 0 {
		currIdx := r.state.CurrentTurn
		if currIdx >= 0 && currIdx < len(r.state.TurnOrder) && r.state.TurnOrder[currIdx] == playerID {
			r.switchToNextActivePlayerAndDraw(&events)
		}
	}

	// Якщо колода випорожнилась ПІСЛЯ передачі ходу — резолвимо раунд безпечно
	if len(r.state.Deck) == 0 {
		startEventID := r.state.Sequence * 1000
		roundResult := engine.ResolveRoundEnd(r.state, r.clock, startEventID+50)
		r.state = roundResult.NewState
		events = append(events, roundResult.DomainEvents...)
	}

	if err := r.saveToDB(ctx); err != nil {
		r.mu.Unlock()
		return fmt.Errorf("failed to save state after leave: %w", err)
	}
	stateForArchive := r.state.Clone()
	r.mu.Unlock()

	r.archiveFinishedGame(stateForArchive)
	go r.BroadcastState(UpdateTypeRoomUpdated, events)
	return nil
}

// =============================================================================
// Broadcast
// =============================================================================

// outgoingPacket — типізована форма WS-пакета. Ми використовуємо struct замість
// map[string]any, щоб JSON-енкодер міг закешувати reflection-дескриптор поля
// та виконати рівно одну аллокацію JSON-байтів на гравця, без зайвих
// проміжних map[string]interface{} під капотом.
type outgoingPacket struct {
	Status    string               `json:"status"`
	Type      string               `json:"type"`
	Timestamp int64                `json:"timestamp"`
	State     engine.GameState     `json:"state"`
	Events    []engine.DomainEvent `json:"events"`
}

// BroadcastState надсилає поточний стан кожному підключеному клієнту з
// маскуванням чужих карт і чутливих payload-ів.
//
// Оптимізації:
//   - Один Clone() лише для того, щоб зрізати загальні поля; per-viewer
//     робиться лише ДРІБНА копія мапи players (це ~N маленьких структур,
//     а не повний стан з декою/discard-пілами).
//   - Слайси Deck/TurnOrder/BurnCard у viewer-state переюзаються (read-only),
//     бо JSON-маршалінг лише читає їх.
//   - filteredEvents алокується точно з потрібним капасіті.
//   - Подія, чий Mask(viewerID) повернув той самий obj без копіювання,
//     не призводить до зайвих аллокацій.
func (r *Room) BroadcastState(updateType string, events []engine.DomainEvent) {
	r.mu.RLock()
	baseState := r.state.Clone()

	secondsLeft := int(engine.TurnDuration.Seconds())
	if len(baseState.TurnOrder) > 0 && baseState.Phase != engine.PhaseFinished && baseState.Phase != engine.PhaseRoundEnd {
		elapsed := time.Since(r.turnStartedAt)
		timeLeft := engine.TurnDuration - elapsed

		secondsLeft = int(timeLeft.Seconds())
		if secondsLeft < 0 {
			secondsLeft = 0
		}
	}

	clientsCopy := make(map[string]*GlobalClient, len(r.conns))
	for pID, cl := range r.conns {
		clientsCopy[pID] = cl
	}
	r.mu.RUnlock()

	if len(clientsCopy) == 0 {
		return
	}

	for playerID, client := range clientsCopy {
		viewerPlayers := make(map[string]engine.Player, len(baseState.Players))
		for id, p := range baseState.Players {
			maskedPlayer := engine.Player{
				ID:               p.ID,
				Username:         p.Username,
				AvatarSeed:       p.AvatarSeed,
				DiscardPile:      p.DiscardPile,
				Score:            p.Score,
				IsOut:            p.IsOut,
				IsProtected:      p.IsProtected,
				SpyPointsAwarded: p.SpyPointsAwarded,
			}

			// Якщо це опонент і він має карти в руках — ховаємо їхній тип (замінюємо на 0 / невідомо)
			if id != playerID && len(p.Hand) > 0 {
				maskedPlayer.Hand = make([]engine.CardType, len(p.Hand)) // створює масив «порожніх» карт тієї ж кількості
			} else {
				// Якщо це сам гравець — віддаємо його руку як є
				maskedPlayer.Hand = p.Hand
			}

			viewerPlayers[id] = maskedPlayer
		}

		viewerState := baseState
		viewerState.Players = viewerPlayers
		viewerState.SecondsLeft = secondsLeft

		var filteredEvents []engine.DomainEvent
		if len(events) > 0 {
			filteredEvents = make([]engine.DomainEvent, len(events))
			for i, ev := range events {
				if ev.Payload != nil {
					ev.Payload = ev.Payload.Mask(playerID)
				}
				filteredEvents[i] = ev
			}
		}

		packet := outgoingPacket{
			Status:    "success",
			Type:      updateType,
			Timestamp: time.Now().Unix(),
			State:     viewerState,
			Events:    filteredEvents,
		}

		data, err := json.Marshal(packet)
		if err != nil {
			r.log.WithField("error", err.Error()).Warn("Не вдалося серіалізувати пакет стану")
			continue
		}

		select {
		case client.SendChan <- data:
		default:
			r.log.WithField("player_id", playerID).
				Warn("Канал клієнта переповнений, пропускаємо оновлення")
		}
	}
}

// =============================================================================
// Main loop (ігровий goroutine)
// =============================================================================

// Start — головний споживач r.actions та AFK-таймер.
// Запускається один раз на кімнату (h.go викликає це при створенні).
func (r *Room) Start(ctx context.Context) {
	r.log.WithField("room_id", r.id).Info("Головний цикл обробки ходів запущено")

	// Створюємо таймер у "роззбройованому" стані: time.NewTimer одразу
	// запускає відлік, тому ми його зупиняємо і дренимо канал, якщо
	// він уже встиг вистрілити (race між NewTimer і Stop).
	turnTimer := time.NewTimer(engine.TurnDuration)
	if !turnTimer.Stop() {
		select {
		case <-turnTimer.C:
		default:
		}
	}
	defer turnTimer.Stop()

	// resetTurnTimer — БЕЗПЕЧНИЙ ідіоматичний reset.
	//
	// Чому саме так:
	//   • turnTimer.Stop() повертає false, якщо таймер уже вистрілив АБО
	//     уже був зупинений — у такому випадку у каналі може лежати тік.
	//   • Перед Reset канал ОБОВ'ЯЗКОВО має бути порожнім, інакше наступне
	//     `case <-turnTimer.C` миттєво спрацює (false-positive AFK).
	//   • select-default прибирає тік без блокування, якщо він є.
	//
	// Викликати безпечно тільки з цього goroutine — таймер не shared.
	resetTurnTimer := func() {
		r.mu.Lock()
		hasStarted := len(r.state.TurnOrder) > 0 && r.state.Phase != engine.PhaseFinished
		if hasStarted {
			r.turnStartedAt = time.Now()
		}
		r.mu.Unlock()

		if !hasStarted {
			return
		}
		if !turnTimer.Stop() {
			select {
			case <-turnTimer.C:
			default:
			}
		}
		turnTimer.Reset(engine.TurnDuration)
	}

	for {
		select {
		case <-ctx.Done():
			r.log.WithField("room_id", r.id).Info("Головний цикл зупинено через скасування контексту")
			return

		case inAct := <-r.actions:
			// Спецсигнал від StartGame/NextRound: просто рестартуємо таймер.
			if inAct.RequestID == MsgStartGame {
				resetTurnTimer()
				continue
			}
			err := r.processAction(ctx, inAct)
			if inAct.errChan != nil {
				inAct.errChan <- err
			}
			if err == nil {
				resetTurnTimer()
			}

		case <-turnTimer.C:
			r.handleAFKTick()
		}
	}
}

// handleAFKTick — ізольована обробка AFK-тіка таймера.
// Виконується ТІЛЬКИ з goroutine Start(); звідси можна тимчасово брати r.mu.Lock.
func (r *Room) handleAFKTick() {
	r.mu.Lock()
	if len(r.state.TurnOrder) == 0 || r.state.Phase == engine.PhaseFinished {
		r.mu.Unlock()
		return
	}
	currIdx := r.state.CurrentTurn
	if currIdx >= len(r.state.TurnOrder) {
		r.mu.Unlock()
		return
	}
	afkPlayerID := r.state.TurnOrder[currIdx]
	r.log.WithFields(logrus.Fields{
		"room_id":   r.id,
		"player_id": afkPlayerID,
		"phase":     r.state.Phase,
	}).Warn("Гравець перевищив ліміт часу. ГЕНЕРУЄТЬСЯ АВТО-ХІД (AFK)...")

	var afkInbound inboundAction
	if r.state.Phase == engine.PhaseResolveChancellor {
		player := r.state.Players[afkPlayerID]
		var bottomCards []engine.CardType
		if len(player.Hand) > 1 {
			bottomCards = append(bottomCards, player.Hand[1:]...)
		}
		afkInbound = inboundAction{
			PlayerID:         afkPlayerID,
			RequestID:        fmt.Sprintf("afk-timeout-chancellor-%d", time.Now().UnixNano()),
			IsChancellorType: true,
			ChancellorAction: engine.ChancellorResolveAction{
				PlayerID:      afkPlayerID,
				KeepHandIndex: 0,
				BottomOrder:   bottomCards,
			},
		}
	} else {
		afkInbound = inboundAction{
			PlayerID:  afkPlayerID,
			RequestID: fmt.Sprintf("afk-timeout-%d", time.Now().UnixNano()),
			Action: engine.Action{
				PlayerID:  afkPlayerID,
				HandIndex: 0,
				TargetID:  "",
			},
		}
	}
	r.mu.Unlock()

	select {
	case r.actions <- afkInbound:
	default:
		r.log.Warn("Черга екшенів переповнена, не вдалося поставити AFK хід")
	}
}

// =============================================================================
// Action submission API
// =============================================================================

// SubmitAction — публічна точка входу для звичайних дій гравця.
func (r *Room) SubmitAction(ctx context.Context, playerID, reqID string, act engine.Action) error {
	r.mu.Lock()
	if r.isDuplicateLocked(playerID, reqID) {
		r.mu.Unlock()
		return fmt.Errorf("duplicate request ID: %s", reqID)
	}
	r.mu.Unlock()

	// Санітизація: фронтенд інколи присилає рядкові "null"/"undefined".
	if act.TargetID == "null" || act.TargetID == "undefined" {
		act.TargetID = ""
	}

	errChan := make(chan error, 1)
	inbound := inboundAction{
		PlayerID:  playerID,
		RequestID: reqID,
		Action:    act,
		errChan:   errChan,
	}

	select {
	case r.actions <- inbound:
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SubmitChancellorAction — публічна точка входу для резолву Chancellor.
func (r *Room) SubmitChancellorAction(ctx context.Context, playerID, reqID string, act engine.ChancellorResolveAction) error {
	r.mu.Lock()
	if r.isDuplicateLocked(playerID, reqID) {
		r.mu.Unlock()
		return fmt.Errorf("duplicate request ID: %s", reqID)
	}
	r.mu.Unlock()

	errChan := make(chan error, 1)
	inbound := inboundAction{
		PlayerID:         playerID,
		RequestID:        reqID,
		IsChancellorType: true,
		ChancellorAction: act,
		errChan:          errChan,
	}

	select {
	case r.actions <- inbound:
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// AddPlayer додає гравця в лобі.
//
// Передумова виклику: playerID УЖЕ провалідований як UUID на рівні
// HandleWS. Тут ми все одно робимо defensive-перевірку, щоб дублювати
// контракт у тестах та внутрішніх викликах із Hub.
func (r *Room) AddPlayer(playerID string) error {
	playerUUID, parseErr := uuid.Parse(playerID)
	if parseErr != nil {
		return fmt.Errorf("invalid player_id, expected UUID: %w", parseErr)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	log := r.log.WithFields(logrus.Fields{"room_id": r.id, "player_id": playerID})

	if len(r.state.TurnOrder) > 0 {
		return fmt.Errorf("cannot join: game has already started")
	}
	if len(r.state.Players) >= MAX_PLAYERS {
		return fmt.Errorf("room is full")
	}

	if _, exists := r.state.Players[playerID]; exists {
		log.Debug("Гравець вже в кімнаті, пропускаємо інкремент лічильника")
	} else {
		var username string
		var avatarSeed string

		if user, dbErr := r.store.GetUserByID(context.Background(), playerUUID); dbErr == nil {
			username = user.Username
			avatarSeed = user.AvatarSeed
			log.Debug("Користувача знайдено в БД, avatarSeed: ", avatarSeed, ", username: ", username)
		} else {
			log.WithField("error", dbErr.Error()).Debug("Не вдалося отримати дані гравця з БД (не критично)")
			username = "Гравець"
			avatarSeed = "default_seed"
		}

		r.state.Players[playerID] = engine.Player{
			ID:          playerID,
			Username:    username,
			AvatarSeed:  avatarSeed,
			Hand:        []engine.CardType{},
			DiscardPile: []engine.CardType{},
			Score:       0,
		}
	}

	playersCount := len(r.state.Players)

	if err := r.saveToDB(context.Background()); err != nil {
		return fmt.Errorf("failed to save state after join: %w", err)
	}

	if r.hub != nil {
		go r.hub.NotifyLobbyUpdate()
	}

	go r.BroadcastState(UpdateTypeRoomUpdated, nil)

	log.Infof("Гравець успішно приєднався. Усього гравців: %d", playersCount)
	return nil
}

// =============================================================================
// Game lifecycle: StartGame / NextRound
// =============================================================================

// StartGame формує колоду, спалює burn-карту і роздає стартові руки.
func (r *Room) StartGame() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	log := r.log.WithField("room_id", r.id)

	if len(r.state.Players) < 2 {
		return fmt.Errorf("cannot start game: minimum 2 players required")
	}

	// ПЕРЕВІРКА НА ПЕРЕЗАПУСК: якщо гра вже йшла, але завершилась
	isRestart := len(r.state.TurnOrder) > 0 && (r.state.IsGameOver || r.state.Phase == "FINISHED")

	if len(r.state.TurnOrder) > 0 && !isRestart {
		return fmt.Errorf("game has already been started")
	}

	log.Info("Ініціалізація старту гри (або перезапуску сесії), формування колоди...")
	deck, burn := engine.PrepareNewDeck(r.rng)
	r.state.Deck = deck
	r.state.BurnCard = &burn

	// Формуємо чергу ходів на основі поточних підключених гравців
	turnOrder := make([]string, 0, len(r.state.Players))

	for id, player := range r.state.Players {
		turnOrder = append(turnOrder, id)
		if len(r.state.Deck) == 0 {
			return fmt.Errorf("not enough cards in the deck to distribute to players")
		}

		// Роздаємо стартову карту
		card := r.state.Deck[0]
		r.state.Deck = r.state.Deck[1:]

		player.Hand = []engine.CardType{card}
		player.DiscardPile = []engine.CardType{} // Очищуємо старий відбій
		player.IsOut = false                     // Повертаємо вибулих у гру
		player.IsProtected = false               // Знімаємо захист
		player.SpyPointsAwarded = false          // Скидаємо прапорці шпигуна

		// ЯКЩО ЦЕ ПЕРЕЗАПУСК — ОБНУЛЯЄМО РАХУНОК ДО 0
		if isRestart {
			player.Score = 0
		}

		r.state.Players[id] = player
	}

	// Скидаємо загальні ігрові прапорці кімнати
	r.state.TurnOrder = turnOrder
	r.state.CurrentTurn = 0
	r.state.Phase = engine.PhaseMainAction
	r.state.WinnerID = ""
	r.state.IsGameOver = false

	// Нарощуємо Sequence, щоб фронтенд побачив новий пакет даних
	r.state.Sequence++

	// Видаємо першому гравцю другу карту для початку ходу
	firstPlayerID := turnOrder[0]
	firstPlayer := r.state.Players[firstPlayerID]
	if len(r.state.Deck) == 0 {
		return fmt.Errorf("not enough cards in the deck to draw for the first player")
	}
	drawCard := r.state.Deck[0]
	r.state.Deck = r.state.Deck[1:]
	firstPlayer.Hand = append(firstPlayer.Hand, drawCard)
	r.state.Players[firstPlayerID] = firstPlayer

	// Зберігаємо оновлений стан у базу даних
	if err := r.saveToDB(context.Background()); err != nil {
		return fmt.Errorf("failed to save initialized game state: %w", err)
	}

	log.WithFields(logrus.Fields{
		"players_count": len(turnOrder),
		"deck_left":     len(r.state.Deck),
		"first_player":  firstPlayerID,
		"is_restart":    isRestart,
	}).Info("Гру успішно розпочато з нуля!")

	// Перезапускаємо таймер ходу через внутрішню чергу
	select {
	case r.actions <- inboundAction{PlayerID: systemPlayerID, RequestID: MsgStartGame}:
	default:
	}

	// Розсилаємо новий стан усім клієнтам у кімнаті
	go r.BroadcastState(UpdateTypeRoomUpdated, nil)
	return nil
}

// NextRound скидає стан раунду, перегенерує колоду та запускає новий раунд.
func (r *Room) NextRound() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	log := r.log.WithField("room_id", r.id)

	if r.state.Phase == engine.PhaseFinished {
		return fmt.Errorf("cannot start next round: the game is already finished")
	}
	if len(r.conns) < 2 {
		return fmt.Errorf("cannot start next round: insufficient connected players (%d online)", len(r.conns))
	}

	log.Info("Ініціалізація наступного раунду, перегенерація колоди...")
	deck, burn := engine.PrepareNewDeck(r.rng)
	r.state.Deck = deck
	r.state.BurnCard = &burn

	for _, id := range r.state.TurnOrder {
		player, exists := r.state.Players[id]
		if !exists {
			continue
		}

		// Скидаємо стан гравців для нового раунду
		if len(r.state.Deck) == 0 {
			return fmt.Errorf("not enough cards in the deck to distribute during next round")
		}
		card := r.state.Deck[0]
		r.state.Deck = r.state.Deck[1:]
		player.Hand = []engine.CardType{card}
		player.DiscardPile = []engine.CardType{}

		// ВАЖЛИВО: Саме тут гравець знову стає активним!
		player.IsOut = false
		player.IsProtected = false
		r.state.Players[id] = player
	}

	// Хто починає? Якщо є переможець попереднього раунду — він.
	startingIdx := 0
	if r.state.WinnerID != "" {
		for idx, id := range r.state.TurnOrder {
			if id == r.state.WinnerID {
				startingIdx = idx
				break
			}
		}
	}

	r.state.CurrentTurn = startingIdx
	r.state.Phase = engine.PhaseMainAction
	r.state.WinnerID = ""
	r.state.Sequence++

	activePlayerID := r.state.TurnOrder[startingIdx]
	activePlayer := r.state.Players[activePlayerID]
	if len(r.state.Deck) > 0 {
		drawCard := r.state.Deck[0]
		r.state.Deck = r.state.Deck[1:]
		activePlayer.Hand = append(activePlayer.Hand, drawCard)
		r.state.Players[activePlayerID] = activePlayer
	}

	if err := r.saveToDB(context.Background()); err != nil {
		return fmt.Errorf("failed to save next round state: %w", err)
	}

	log.WithFields(logrus.Fields{
		"deck_left":     len(r.state.Deck),
		"active_player": activePlayerID,
		"seq":           r.state.Sequence,
	}).Info("Наступний раунд успішно запущено!")

	select {
	case r.actions <- inboundAction{PlayerID: systemPlayerID, RequestID: MsgStartGame}:
	default:
	}

	go r.BroadcastState(UpdateTypeRoomUpdated, nil)
	return nil
}

// =============================================================================
// Action processing (під r.mu, з ігрового goroutine)
// =============================================================================

// processAction застосовує одну подію до стану.
// Викликається лише з goroutine Start() — серіалізація гарантована каналом.
func (r *Room) processAction(ctx context.Context, inAct inboundAction) error {
	if inAct.RequestID == MsgStartGame {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	log := r.log.WithFields(logrus.Fields{
		"room_id":       r.id,
		"player_id":     inAct.PlayerID,
		"request_id":    inAct.RequestID,
		"is_chancellor": inAct.IsChancellorType,
		"seq_before":    r.state.Sequence,
	})
	log.Debug("Початок обробки дії з черги кімнати")

	startEventID := r.state.Sequence * 1000
	var (
		result engine.ApplyResult
		err    error
	)

	if !inAct.IsChancellorType {
		// Уся валідація делегована engine.Apply (типізовані помилки).
		result, err = engine.Apply(r.state, inAct.Action, r.rng, r.clock, startEventID)
		if err != nil {
			log.WithField("error", err.Error()).Warn("Відхилено ігрову дію")
			return err
		}
		err = r.store.LogApplyAction(ctx, r.id, r.state.Sequence, inAct.Action, result.DomainEvents)
	} else {
		result, err = engine.ResolveChancellor(r.state, inAct.ChancellorAction, r.clock, startEventID)
		if err != nil {
			log.WithField("error", err.Error()).Warn("Відхилено вибір Канцлера")
			return err
		}
		err = r.store.LogChancellorResolveAction(ctx, r.id, r.state.Sequence, inAct.ChancellorAction, result.DomainEvents)

	}

	if err != nil {
		log.WithField("error", err.Error()).Error("Критична помилка запису логів дії в БД")
		return fmt.Errorf("failed to log action to storage: %w", err)
	}

	r.state = result.NewState
	r.state.Sequence++

	if err := r.saveToDB(ctx); err != nil {
		log.WithField("error", err.Error()).Error("Помилка збереження снапшоту")
		return fmt.Errorf("failed to save room snapshot: %w", err)
	}

	// archiveFinishedGame потребує snapshot стану — робимо clone під локом.
	stateForArchive := r.state.Clone()

	log.WithFields(logrus.Fields{
		"seq_after": r.state.Sequence,
		"phase":     r.state.Phase,
	}).Info("Дію повністю зафіксовано. Запуск розсилки стану...")

	// Виходимо з критичної секції перед side-effects.
	// `defer r.mu.Unlock()` зніме лок одразу після return; оскільки наступні
	// операції — це async-горутини та read-only операції над snapshot'ом, гонок не буде.
	go r.BroadcastState(UpdateTypeRoomUpdated, result.DomainEvents)
	r.archiveFinishedGame(stateForArchive)
	return nil
}

// =============================================================================
// Persistence helpers
// =============================================================================

// saveToDB серіалізує поточний стан у БД. Викликається ТІЛЬКИ під r.mu
// (оскільки читає r.state). Внутрішній лок storage не конфліктує з r.mu.
func (r *Room) saveToDB(ctx context.Context) error {
	stateRaw, err := json.Marshal(r.state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	return r.store.UpsertRoomState(ctx, r.id, stateRaw)
}

// archiveFinishedGame — фіксує підсумки матчу в БД, ЯКЩО матч завершено.
// Приймає snapshot стану (а не читає r.state), щоб не залежати від r.mu.
func (r *Room) archiveFinishedGame(snapshot engine.GameState) {
	if snapshot.Phase != engine.PhaseFinished && snapshot.Phase != "ROUND_END" && len(snapshot.TurnOrder) != 0 {
		return
	}

	r.log.Infof("Фіксуємо архівацію (Phase: %s) для кімнати %s...", snapshot.Phase, r.id)

	finalStateBytes, err := json.Marshal(snapshot)
	if err != nil {
		r.log.Errorf("Не вдалося серіалізувати фінальний стейт:%v", err)
		return
	}

	allUsernames := make([]string, 0, len(snapshot.Players))
	var winnerParam uuid.NullUUID
	var spyWinnerParam uuid.NullUUID

	// Парсимо UUID головного переможця, якщо він є і це не SYSTEM
	if snapshot.WinnerID != "" && snapshot.WinnerID != "SYSTEM" {
		if wUUID, parseErr := uuid.Parse(snapshot.WinnerID); parseErr == nil {
			winnerParam = uuid.NullUUID{UUID: wUUID, Valid: true}
		} else {
			r.log.Errorf("Не вдалося розпарсити UUID переможця %s: %v", snapshot.WinnerID, parseErr)
		}
	}

	for _, p := range snapshot.Players {
		allUsernames = append(allUsernames, p.Username) // Для масиву оновлення все ще збираємо юзернейми

		// Шукаємо шпигуна та парсимо його UUID
		if p.SpyPointsAwarded {
			if spyUUID, parseErr := uuid.Parse(p.ID); parseErr == nil {
				spyWinnerParam = uuid.NullUUID{UUID: spyUUID, Valid: true}
			} else {
				r.log.Errorf("Не вдалося розпарсити UUID шпигуна %s: %v", p.ID, parseErr)
			}
		}
	}

	// Створення структури інпуту з чистими UUID
	input := storage.GameResultInput{
		RoomID:       r.id,
		WinnerID:     winnerParam,    // Передаємо NullUUID
		SpyWinnerID:  spyWinnerParam, // Передаємо NullUUID
		AllPlayers:   allUsernames,   // Передаємо ["user1", "user2"]
		FinalStateJS: finalStateBytes,
	}
	roomID := r.id
	storeRef := r.store
	loggerRef := r.log

	go func() {
		err := storeRef.SaveGameResult(context.Background(), input)
		if err != nil {
			loggerRef.Errorf("Помилка збереження результатів гри %s в БД:%v", roomID, err)
		} else {
			loggerRef.Infof("Результати матчу кімнати %s успішно зафіксовані в БД через UUID.", roomID)
		}
	}()
}

// =============================================================================
// Duplicate-request guard
// =============================================================================

// isDuplicateLocked повертає true, якщо ця пара (playerID, reqID) уже бачилась.
// КОНТРАКТ: викликач уже тримає r.mu.Lock() (запис у мапу).
//
// Очищення: коли гравець виходить (UnregisterClient / HandlePlayerLeave),
// його запис видаляється з r.recentRequests, тому мапа не накопичує
// "мертвих" гравців і не тече по пам'яті.
func (r *Room) isDuplicateLocked(playerID, reqID string) bool {
	if reqID == "" {
		return false
	}
	playerRequests, exists := r.recentRequests[playerID]
	if !exists {
		r.recentRequests[playerID] = map[string]time.Time{reqID: time.Now()}
		return false
	}
	if _, duplicated := playerRequests[reqID]; duplicated {
		return true
	}
	if len(playerRequests) > recentRequestsCap {
		// LRU-вибиття найстарішого запису.
		var (
			oldestID   string
			oldestTime time.Time
		)
		for id, t := range playerRequests {
			if oldestTime.IsZero() || t.Before(oldestTime) {
				oldestTime = t
				oldestID = id
			}
		}
		delete(playerRequests, oldestID)
	}
	playerRequests[reqID] = time.Now()
	return false
}

func (r *Room) eliminatePlayer(playerID string) error {
	player, exists := r.state.Players[playerID]
	if !exists || player.IsOut {
		return nil // Гравця вже немає або він вже вибув
	}

	r.log.WithField("player_id", playerID).Info("Автоматичне виключення гравця (дисконект)")

	// Очищаємо руку гравця та міняємо статус
	for _, card := range player.Hand {
		if card != engine.CardSpy {
			player.DiscardPile = append(player.DiscardPile, card)
		}
	}
	player.Hand = nil
	player.IsOut = true
	player.IsProtected = false
	r.state.Players[playerID] = player
	r.state.Sequence++

	events := []engine.DomainEvent{
		{
			EventID:   r.state.Sequence * 1000,
			Type:      engine.EventPlayerLeft,
			Payload:   engine.PlayerLeftPayload{PlayerID: playerID},
			Timestamp: r.clock.Now(),
		},
	}

	// Перевіряємо, скільки гравців залишилось
	aliveIDs := make([]string, 0)
	for _, p := range r.state.Players {
		if !p.IsOut {
			aliveIDs = append(aliveIDs, p.ID)
		}
	}

	// Якщо залишився один або нуль гравців — завершуємо раунд
	if len(aliveIDs) <= 1 {
		startEventID := r.state.Sequence * 1000
		roundResult := engine.ResolveRoundEnd(r.state, r.clock, startEventID)
		r.state = roundResult.NewState
		events = append(events, roundResult.DomainEvents...)
		r.log.Info("Раунд автоматично завершено через дисконект гравця")

		// Оскільки це викликається з UnregisterClient, нам потрібно
		// самостійно зробити Broadcast оновленого стану
		go r.BroadcastState(UpdateTypeRoomUpdated, events)
		return nil
	}

	// КРИТИЧНЕ ВИПРАВЛЕННЯ: Якщо відключився гравець, чий зараз був хід,
	// передаємо хід наступному активному гравцю.
	currIdx := r.state.CurrentTurn
	if currIdx >= 0 && currIdx < len(r.state.TurnOrder) && r.state.TurnOrder[currIdx] == playerID {
		r.switchToNextActivePlayerAndDraw(&events)
	}

	// Перевірка на випадок, якщо закінчилися карти в колоді
	if len(r.state.Deck) == 0 {
		startEventID := r.state.Sequence * 1000
		roundResult := engine.ResolveRoundEnd(r.state, r.clock, startEventID+50)
		r.state = roundResult.NewState
		events = append(events, roundResult.DomainEvents...)
	}

	// Надсилаємо оновлений стан усім гравцям кімнати
	go r.BroadcastState(UpdateTypeRoomUpdated, events)
	return nil
}

// ПРИМІТКА: Додай цей метод до структури Room у файлі network/room.go
// Він інкапсулює логіку переходу ходу, яка дублювалася.
func (r *Room) switchToNextActivePlayerAndDraw(events *[]engine.DomainEvent) {
	currIdx := r.state.CurrentTurn
	totalPlayers := len(r.state.TurnOrder)
	if totalPlayers == 0 {
		return
	}

	// Шукаємо наступного гравця, який не вибув
	next := (currIdx + 1) % totalPlayers
	startingIdx := currIdx

	for r.state.Players[r.state.TurnOrder[next]].IsOut {
		next = (next + 1) % totalPlayers
		// Якщо пройшли повне коло і не знайшли активних гравців
		if next == startingIdx {
			return
		}
	}

	r.state.CurrentTurn = next
	newActiveID := r.state.TurnOrder[next]
	newActive := r.state.Players[newActiveID]
	newActive.IsProtected = false

	// Новий гравець бере карту з колоди
	if len(r.state.Deck) > 0 && len(newActive.Hand) < 2 {
		draw := r.state.Deck[0]
		r.state.Deck = r.state.Deck[1:]
		newActive.Hand = append(newActive.Hand, draw)

		*events = append(*events, engine.DomainEvent{
			EventID:   r.state.Sequence*1000 + 5,
			Type:      engine.EventCardDrawn,
			Payload:   engine.CardDrawnPayload{PlayerID: newActiveID, Card: draw},
			Timestamp: r.clock.Now(),
		})
	}

	r.state.Players[newActiveID] = newActive
	r.state.Phase = engine.PhaseMainAction
}
