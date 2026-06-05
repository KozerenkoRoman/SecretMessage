// cmd/network/handlers_lobby.go
package network

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"secret-message/cmd/auth"
	"secret-message/cmd/util"
)

// Структура для мапінгу тіла запиту створення кімнати
type CreateRoomRequest struct {
	RoomID string `json:"room_id"` // Опціонально, якщо клієнт хоче кастомне ім'я кімнати
}

// HandleCreateRoom обробляє POST /api/rooms
func (s *Server) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Отримуємо Claims через правильний ключ контексту, який використовує твій AuthMiddleware
	// Примітка: Перевірте, як саме називається ключ у вашому проекті (наприклад, UserContextKey чи "user")
	// Якщо UserContextKey лежить у пакеті network, пиши просто UserContextKey.

	var hostID string

	// Варіант А: Спроба дістати об'єкт *auth.Claims (найімовірніший варіант для твого проекту)
	if claims, ok := r.Context().Value(UserContextKey).(*auth.Claims); ok {
		hostID = claims.UserID
	}

	// Резервний Варіант Б: Якщо раптом ключ контексту — це просто рядок "user"
	if hostID == "" {
		if claims, ok := r.Context().Value("user").(*auth.Claims); ok {
			hostID = claims.UserID
		}
	}

	// Резервний Варіант В: Перевірка на випадок, якщо мідлвар пише туди суто ID користувача (рядок)
	if hostID == "" {
		if id, ok := r.Context().Value("user_id").(string); ok {
			hostID = id
		}
	}

	// Якщо жоден варіант не зміг витягнути hostID з контексту
	if hostID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized: missing user identification in context",
		})
		return
	}

	// 2. Парсимо тіло запиту
	var req CreateRoomRequest
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	roomID := strings.TrimSpace(req.RoomID)
	if roomID == "" {
		roomID = util.EncodeName(generateRandomRoomID() + "-" + hostID[:6]) // Додаємо частину hostID для більшої унікальності
	}

	// 3. Створюємо кімнату через Hub
	room, err := s.hub.CreateNewRoom(roomID, hostID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// 4. Повертаємо успішну відповідь з даними створеного лобі
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(room.GetLobbyInfo())
}

// generateRandomRoomID створює короткий унікальний ID (наприклад, "LL-A3F8B2")
func generateRandomRoomID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}
