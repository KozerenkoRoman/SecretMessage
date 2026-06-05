// network/gateway.go
package network

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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

func (c *GlobalClient) StartWriter() {
	go func() {
		for msg := range c.SendChan {
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()
}

func (g *Gateway) AddClient(userID string, conn *websocket.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()

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
