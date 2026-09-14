// network/gateway.go
package network

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Таймінги WebSocket keepalive (ping/pong).
//
// ВАЖЛИВО: http.Server.ReadTimeout/WriteTimeout встановлюють deadline
// на сирий net.Conn ЩЕ ДО виклику хендлера (main.go: startServers).
// Коли gorilla/websocket робить Upgrade → Hijack, цей conn виривається
// з-під контролю http.Server, але вже виставлений deadline (now+15s)
// нікуди не зникає — він продовжує діяти на голому net.Conn, доки його
// не перезапишуть новим SetReadDeadline/SetWriteDeadline. Без явного
// керування тут WS-з'єднання обривалося б приблизно через WriteTimeout
// незалежно від активності клієнта. Тому одразу після AddClient і на
// кожен Pong/цикл запису ми самі виставляємо власні deadlines,
// перекриваючи ті, що встановив http.Server.
const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
)

// GlobalClient представляє підключеного до сайту користувача
type GlobalClient struct {
	UserID   string
	Conn     *websocket.Conn
	SendChan chan []byte
}

type Gateway struct {
	hub     *Hub
	log     *logrus.Logger
	clients map[string]*GlobalClient
	mu      sync.RWMutex
}

func NewGateway(hub *Hub, logger *logrus.Logger) *Gateway {
	return &Gateway{
		hub:     hub,
		log:     logger,
		clients: make(map[string]*GlobalClient),
	}
}

func (g *Gateway) RemoveClient(userID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if client, exists := g.clients[userID]; exists {
		close(client.SendChan)
		client.Conn.Close()
		delete(g.clients, userID)
		g.log.Infof("[Gateway] Користувача %s видалено з глобального шлюзу", userID)
	}
}

// StartWriter запускає горутину-письменника, яка водночас відповідає за
// keepalive: раз на wsPingPeriod шле контрольний PingMessage, а перед
// кожним записом (даними чи ping) оновлює SetWriteDeadline. Це перекриває
// deadline, який http.Server виставив на сирий conn до Hijack (див.
// коментар над константами вище) — без цього WS "вмирав" би за WriteTimeout
// незалежно від активності з'єднання.
func (c *GlobalClient) StartWriter() {
	go func() {
		ticker := time.NewTicker(wsPingPeriod)
		defer ticker.Stop()

		for {
			select {
			case msg, ok := <-c.SendChan:
				if !ok {
					return
				}
				_ = c.Conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			case <-ticker.C:
				_ = c.Conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()
}

func (g *Gateway) AddClient(userID string, conn *websocket.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Перекриваємо deadline, успадкований від http.Server (ReadTimeout,
	// виставлений на conn ще до Hijack), і встановлюємо власний
	// pong-based read deadline: кожен Pong від клієнта (у відповідь на наш
	// Ping зі StartWriter) відсуває read deadline ще на wsPongWait.
	_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	client := &GlobalClient{
		UserID:   userID,
		Conn:     conn,
		SendChan: make(chan []byte, 256),
	}
	client.StartWriter()
	g.clients[userID] = client
	g.log.Infof("[Gateway] Користувача %s додано до глобального шлюзу", userID)
}

func (g *Gateway) BroadcastToDesktop(lobbyInfo any) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	packet, _ := json.Marshal(map[string]any{
		"type": UpdateTypeLobbyList,
		"data": lobbyInfo,
	})

	for _, client := range g.clients {
		select {
		case client.SendChan <- packet:
		default:
		}
	}
}
