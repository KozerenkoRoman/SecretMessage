// cmd/network/admin.go
package network

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Структура для мапінгу тіла запиту блокування
type BlockUserPayload struct {
	Reason string `json:"reason"`
}

// HandleAdminGetUsers обробляє GET /api/admin/users
// Повертає повний список зареєстрованих користувачів для адмін-панелі
func (s *Server) HandleAdminGetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := s.store.ListUsers(ctx)
	if err != nil {
		s.log.WithError(err).Error("Помилка отримання списку користувачів для адмін-панелі")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to retrieve users"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"users":  users,
	})
}

// HandleAdminGetGames обробляє GET /api/admin/games
// Повертає історію всіх зіграних матчів із бази даних
func (s *Server) HandleAdminGetGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Припускаємо, що у твоєму s.store (Storage) є метод для отримання завершених ігор.
	// Наприклад, store.Queries.GetAllGameResults(ctx) або кастомний метод історії.
	// Якщо його ще немає, для компіляції можна використати заглушку або прямий виклик sqlc:
	games, err := s.store.GetAllGameResults(ctx)
	if err != nil {
		s.log.WithError(err).Error("Помилка отримання логів ігор для адмін-панелі")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to retrieve game logs"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"games":  games,
	})
}

// HandleAdminBlockUser обробляє POST /api/admin/users/block
// Стандартний http.ServeMux із Go 1.22+ дозволяє діставати параметри з URL за допомогою r.PathValue("id")
func (s *Server) HandleAdminBlockUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Отримуємо ID користувача з URL-маршруту (наприклад, /api/admin/users/{id}/block)
	targetUserID := r.PathValue("id")
	if targetUserID == "" {
		// Якщо використовується старіша версія або query-параметр, спробуємо дістати з query:
		targetUserID = r.URL.Query().Get("user_id")
	}

	targetUserID = strings.TrimSpace(targetUserID)
	if targetUserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing target user id"})
		return
	}

	// Парсимо причину блокування з JSON
	var payload BlockUserPayload
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&payload)
	}
	if payload.Reason == "" {
		payload.Reason = "Порушення правил платформи (Адміністративне блокування)"
	}

	// 1. Оновлюємо прапорець IsBlocked у базі даних через Storage
	// Припускаємо, що sqlc згенерував метод BlockUser або UpdateUserStatus
	err := s.store.BlockUser(ctx, targetUserID, payload.Reason)
	if err != nil {
		s.log.WithError(err).Errorf("Не вдалося заблокувати користувача %s в БД", targetUserID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to block user in database"})
		return
	}

	// 2. КРИТИЧНО ДЛЯ БЕЗПЕКИ: Викидаємо користувача з усіх активних ігор та закриваємо його сокети!
	// У нас у Hub вже є чудовий готовий метод DisconnectUserGlobally
	s.hub.DisconnectUserGlobally(targetUserID)

	s.log.WithFields(map[string]any{
		"admin_id":       r.Context().Value("user_id"),
		"target_user_id": targetUserID,
		"reason":         payload.Reason,
	}).Warn("Адміністратор успішно заблокував користувача та розірвав його активні сесії")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "user has been successfully blocked and disconnected globally",
	})
}
