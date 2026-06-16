package network

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"secret-message/cmd/auth"
	"secret-message/cmd/storage/gen"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserRequest struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email,omitempty"`
	AvatarSeed  string `json:"avatar_seed,omitempty"`
	Password    string `json:"password,omitempty"`
	PasswordOld string `json:"password_old,omitempty"`
}
type RegisterRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	AvatarSeed string `json:"avatar_seed"`
}

type BlockRequest struct {
	UserID string `json:"user_id"`
	Ban    bool   `json:"ban"`
	Reason string `json:"reason"`
}

func (s *Server) HandleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendHTTPError(w, http.StatusMethodNotAllowed, "Only POST allowed")
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendHTTPError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	ctx := r.Context()

	// 1. Шукаємо користувача в БД за Username
	dbUser, err := s.hub.store.Queries.GetUserByUsername(ctx, req.Username)
	if err != nil {
		// Якщо користувача не знайдено — повертаємо помилку авторизації
		s.sendHTTPError(w, http.StatusNotFound, "Користувача не знайдено")
		return
	}

	// 2. Користувач існує — перевіряємо пароль
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(req.Password))
	if err != nil {
		s.sendHTTPError(w, http.StatusUnauthorized, "Неправильний логін або пароль")
		return
	}

	// 3. ПЕРЕВІРКА НА БАН
	if dbUser.IsBanned {
		reason := "Без причини"
		if dbUser.BanReason.Valid {
			reason = dbUser.BanReason.String
		}
		s.sendHTTPError(w, http.StatusForbidden, fmt.Sprintf("Ваш акаунт заблоковано. Причина: %s", reason))
		return
	}

	// 4. ГЕНЕРАЦІЯ СЕСІЇ (JWT)
	token, err := auth.GenerateToken(dbUser.ID, dbUser.Username, dbUser.UserRole)
	if err != nil {
		s.sendHTTPError(w, http.StatusInternalServerError, "Token generation failed")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]string{
		"token":       token,
		"user_id":     dbUser.ID.String(), // Додано для повної синхронізації з UI фронтенду
		"user_role":   dbUser.UserRole,
		"username":    dbUser.Username,
		"avatar_seed": dbUser.AvatarSeed,
	})
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendHTTPError(w, http.StatusMethodNotAllowed, "Only POST allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendHTTPError(w, http.StatusBadRequest, "Некоректний формат JSON")
		return
	}

	// Базова валідація обов'язкових полів
	if req.Username == "" || req.Email == "" || req.Password == "" {
		s.sendHTTPError(w, http.StatusBadRequest, "Поля username, email та password є обов'язаковими")
		return
	}

	if len(req.AvatarSeed) > 255 {
		s.sendHTTPError(w, http.StatusBadRequest, "Занадто довгий ідентифікатор аватара")
		return
	}

	ctx := r.Context()

	// Хешування пароля користувача за допомогою bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Errorf("Помилка безпеки при хешуванні пароля для %s: %v", req.Username, err)
		s.sendHTTPError(w, http.StatusInternalServerError, "Помилка безпеки при обробці пароля")
		return
	}

	// Створення структури параметрів. Перевірте точну назву типу у вашому пакеті gen
	// (Зазвичай sqlc генерує CreateUserParams або аналогічну назву)
	newUser, err := s.hub.store.Queries.CreateUser(ctx, gen.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		AvatarSeed:   req.AvatarSeed,
	})
	if err != nil {
		// Тут можна додати перевірку на унікальність нікнейму чи пошти, якщо це критично
		s.log.Errorf("Помилка створення користувача %s в БД: %v", req.Username, err)
		s.sendHTTPError(w, http.StatusConflict, "Користувач з таким нікнеймом або email вже існує")
		return
	}

	// Генеруємо токен доступу для миттєвої авторизації після успішної реєстрації
	// Використовуємо функцію GenerateToken з вашого пакету auth
	token, err := auth.GenerateToken(newUser.ID, newUser.Username, newUser.UserRole)
	if err != nil {
		s.log.Errorf("Помилка генерації JWT токена для %s: %v", newUser.Username, err)
		s.sendHTTPError(w, http.StatusInternalServerError, "Користувача створено, але не вдалося згенерувати токен авторизації")
		return
	}

	s.log.Infof("Зареєстровано нового користувача: %s (ID: %s)", newUser.Username, newUser.ID.String())

	// Повертаємо успішну відповідь, токен та дані сесії
	s.sendJSON(w, http.StatusCreated, map[string]any{
		"status":      "success",
		"token":       token,
		"user_id":     newUser.ID.String(),
		"username":    newUser.Username,
		"avatar_seed": newUser.AvatarSeed,
		"user_role":   newUser.UserRole,
	})
}

func (s *Server) HandleBlockUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendHTTPError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Перевірка ролі з контексту
	claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
	if !ok || claims.UserRole != "admin" {
		s.sendHTTPError(w, http.StatusForbidden, "Доступ заборонено: потрібні права Admin")
		return
	}

	var req BlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendHTTPError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Парсимо UUID з рядка JSON запиту
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		s.sendHTTPError(w, http.StatusBadRequest, "Некоректний формат UUID користувача")
		return
	}

	// Вичитуємо користувача, щоб дізнатися його ім'я для глобального дисконекту
	targetUser, err := s.hub.store.Queries.GetUserByID(r.Context(), userUUID)
	if err != nil {
		s.sendHTTPError(w, http.StatusNotFound, "Користувача не знайдено")
		return
	}

	// Викликаємо відповідний метод sqlc залежно від прапорця
	if req.Ban {
		banReason := sql.NullString{
			String: req.Reason,
			Valid:  req.Reason != "",
		}

		err = s.hub.store.Queries.BanUser(r.Context(),
			gen.BanUserParams{
				ID:        userUUID,
				BanReason: banReason,
			})

		// Викидаємо забаненого користувача з усіх ігрових сесій у реальному часі
		s.hub.DisconnectUserGlobally(targetUser.Username)
		s.log.Infof("Адмін %s забанив користувача %s", claims.Username, targetUser.Username)
	} else {
		err = s.hub.store.Queries.UnbanUser(r.Context(), userUUID)
		s.log.Infof("Адмін %s розбанив користувача %s", claims.Username, targetUser.Username)
	}

	if err != nil {
		s.sendHTTPError(w, http.StatusInternalServerError, "Database error during status update")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// HandleHealth перевіряє статус сервера та підключення до сховища
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendHTTPError(w, http.StatusMethodNotAllowed, "Метод не підтримується")
		return
	}

	// Спробуємо зробити легкий запит до БД для перевірки лівенесу
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "OK"
	if _, err := s.hub.store.Queries.CheckHealth(ctx); err != nil {
		dbStatus = "ERROR: " + err.Error()
	}

	response := map[string]any{
		"status":    "healthy",
		"database":  dbStatus,
		"timestamp": time.Now().Unix(),
	}

	s.sendJSON(w, http.StatusOK, response)
}

func (s *Server) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendHTTPError(w, http.StatusMethodNotAllowed, "Only POST allowed")
		return
	}

	var claims *auth.Claims
	if c, ok := r.Context().Value(UserContextKey).(*auth.Claims); ok {
		claims = c
	} else if c, ok := r.Context().Value("user_claims").(*auth.Claims); ok {
		claims = c
	}

	if claims == nil {
		s.sendHTTPError(w, http.StatusUnauthorized, "Неавторизований доступ")
		return
	}

	currentUserUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		s.sendHTTPError(w, http.StatusBadRequest, "Некоректний ID користувача в токені")
		return
	}

	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendHTTPError(w, http.StatusBadRequest, "Некоректний формат JSON")
		return
	}

	ctx := r.Context()

	dbUser, err := s.hub.store.Queries.GetUserByID(ctx, currentUserUUID)
	if err != nil {
		s.sendHTTPError(w, http.StatusNotFound, "Користувача не знайдено в базі даних")
		return
	}

	hasChanges := false
	if req.Password != "" {
		if req.PasswordOld == "" {
			s.sendHTTPError(w, http.StatusBadRequest, "Для зміни пароля необхідно вказати старий пароль")
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(req.PasswordOld))
		if err != nil {
			s.sendHTTPError(w, http.StatusUnauthorized, "Поточний пароль вказано невірно")
			return
		}

		newHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			s.sendHTTPError(w, http.StatusInternalServerError, "Помилка безпеки при обробці пароля")
			return
		}

		err = s.hub.store.Queries.UpdateUserPassword(ctx, gen.UpdateUserPasswordParams{
			ID:           dbUser.ID,
			PasswordHash: string(newHash),
		})
		if err != nil {
			s.log.Errorf("Помилка оновлення пароля для %s: %v", dbUser.Username, err)
			s.sendHTTPError(w, http.StatusInternalServerError, "Помилка бази даних при зміні пароля")
			return
		}

		hasChanges = true
		s.log.Infof("Користувач %s успішно змінив свій пароль", dbUser.Username)
	}

	usernameChanged := req.Username != "" && req.Username != dbUser.Username
	avatarChanged := req.AvatarSeed != "" && req.AvatarSeed != dbUser.AvatarSeed

	if usernameChanged || avatarChanged {
		if avatarChanged && len(req.AvatarSeed) > 255 {
			s.sendHTTPError(w, http.StatusBadRequest, "Занадто довгий ідентифікатор аватара")
			return
		}

		targetUsername := dbUser.Username
		if req.Username != "" {
			targetUsername = req.Username
		}

		targetAvatarSeed := dbUser.AvatarSeed
		if req.AvatarSeed != "" {
			targetAvatarSeed = req.AvatarSeed
		}

		err = s.hub.store.Queries.UpdateUser(ctx, gen.UpdateUserParams{
			ID:         dbUser.ID,
			Username:   targetUsername,
			AvatarSeed: targetAvatarSeed,
			Email:      dbUser.Email,
		})
		if err != nil {
			s.log.Errorf("Помилка оновлення профілю для %s: %v", dbUser.Username, err)
			s.sendHTTPError(w, http.StatusInternalServerError, "Помилка бази даних при збереженні профілю")
			return
		}

		if usernameChanged {
			s.hub.DisconnectUserGlobally(dbUser.Username)
			s.log.Infof("Користувач %s змінив нікнейм на %s (виконано реконнект)", dbUser.Username, targetUsername)
		}

		dbUser.Username = targetUsername
		dbUser.AvatarSeed = targetAvatarSeed
		hasChanges = true
	}

	if !hasChanges {
		s.sendJSON(w, http.StatusOK, map[string]string{
			"status":  "no_changes",
			"message": "Дані збігаються з тими, що вже збережені",
		})
		return
	}

	// Повертаємо успішну відповідь разом з актуальним профілем
	s.sendJSON(w, http.StatusOK, map[string]any{
		"status":      "success",
		"username":    dbUser.Username,
		"avatar_seed": dbUser.AvatarSeed,
		"user_role":   dbUser.UserRole,
	})
}

// HandleGetRooms повертає список ідентифікаторів активних кімнат у Хабі
func (s *Server) HandleGetRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendHTTPError(w, http.StatusMethodNotAllowed, "Метод не підтримується")
		return
	}

	activeLobbyRooms := s.hub.GetLobbyRooms()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"rooms":  activeLobbyRooms,
	})
}

// HandleGetRoomByID повертає поточний стан конкретної кімнати
// Примітка: у стандартному net/http до версії Go 1.22 немає вбудованого парсингу {id},
// тому ми дістанемо його через просту фільтрацію префіксу або query-параметр.
// Для простоти зробимо через Query параметр: /api/room?id=game_room_777
func (s *Server) HandleGetRoomByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendHTTPError(w, http.StatusMethodNotAllowed, "Метод не підтримується")
		return
	}

	roomID := r.URL.Query().Get("id")
	if roomID == "" {
		s.sendHTTPError(w, http.StatusBadRequest, "Параметр 'id' є обов'язковим у query string (?id=...)")
		return
	}

	room, err := s.hub.GetOrCreateRoom(roomID)
	if err != nil {
		s.sendHTTPError(w, http.StatusNotFound, "Кімнату не знайдено: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"room_id": roomID,
		"state":   room.GetState(), // Тут повертається повний сирий стейт кімнати
	}

	s.sendJSON(w, http.StatusOK, response)
}

// Допоміжний метод для відправки JSON відповідей
func (s *Server) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// Допоміжний метод для відправки HTTP помилок у форматі JSON
func (s *Server) sendHTTPError(w http.ResponseWriter, statusCode int, message string) {
	s.sendJSON(w, statusCode, map[string]string{
		"error": message,
		"code":  http.StatusText(statusCode),
	})
}

func (s *Server) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendHTTPError(w, http.StatusMethodNotAllowed, "Only GET allowed")
		return
	}

	limit := 10

	leaderboard, err := s.hub.store.Queries.GetLeaderboard(r.Context(), limit)
	if err != nil {
		s.log.Errorf("Помилка отримання таблиці лідерів: %v", err)
		s.sendHTTPError(w, http.StatusInternalServerError, "Database error")
		return
	}

	s.sendJSON(w, http.StatusOK, leaderboard)
}

// HandleGetUserStats повертає статистику поточного залогіненого користувача
func (s *Server) HandleGetUserStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendHTTPError(w, http.StatusMethodNotAllowed, "Only GET allowed")
		return
	}

	// 1. Отримуємо Claims користувача, які наш Middleware (Auth) мав покласти в контекст запиту
	ctxClaims, ok := r.Context().Value("user_claims").(*auth.Claims)
	if !ok || ctxClaims == nil {
		s.sendHTTPError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Парсимо UserID з токена в UUID для запиту в БД
	userUUID, err := uuid.Parse(ctxClaims.UserID)
	if err != nil {
		s.sendHTTPError(w, http.StatusBadRequest, "Invalid user ID inside token")
		return
	}

	// 3. Зчитуємо статистику
	stats, err := s.hub.store.Queries.GetUserStats(r.Context(), userUUID)
	if err != nil {
		// Якщо користувач ще не зіграв жодної гри, запису в user_stats може не бути.
		// Замість помилки 500 повернемо "порожню" статистику користувача із нулями
		s.sendJSON(w, http.StatusOK, map[string]any{
			"username":     ctxClaims.Username,
			"games_played": 0,
			"games_won":    0,
			"total_score":  0,
		})
		return
	}

	s.sendJSON(w, http.StatusOK, stats)
}

type AddBotRequest struct {
	BotName string `json:"bot_name"`
}

func (s *Server) HandleAddBotToRoom(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(UserContextKey).(*auth.Claims)

	roomID := r.PathValue("id")
	if roomID == "" {
		s.sendHTTPError(w, http.StatusBadRequest, "room id is required")
		return
	}

	var req AddBotRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.sendHTTPError(w, http.StatusBadRequest, "invalid json body")
			return
		}
	}

	room, err := s.hub.GetOrCreateRoom(roomID)
	if err != nil {
		s.sendHTTPError(w, http.StatusNotFound, "Кімнату не знайдено: "+err.Error())
		return
	}

	if err := room.AddBotPlayer(req.BotName); err != nil {
		s.log.Errorf("Користувач %s не зміг додати бота до кімнати %s: %v", claims.Username, roomID, err)
		s.sendHTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "bot successfully added to room",
	})
}
