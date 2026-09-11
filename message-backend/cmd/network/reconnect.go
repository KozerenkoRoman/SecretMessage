// =============================================================================
// network/reconnect.go
//
// RECONNECTION SESSION REGISTRY
// -----------------------------------------------------------------------------
// Мобільні клієнти часто втрачають WebSocket, коли застосунок іде у фон
// (ОС призупиняє сокети). Щоб гра НЕ завершувалась передчасно, ми:
//
//   1. На старті гри видаємо гравцю унікальний reconnection_token і
//      прив'язуємо його до пари {userID, roomID}.
//   2. Клієнт зберігає токен у localStorage.
//   3. При поверненні у foreground клієнт перепідключає WS і шле RECONNECT
//      з цим токеном. Сервер валідовує токен, скасовує grace-таймер кімнати
//      й повертає повний GAME_STATE_SNAPSHOT.
//
// Реєстр живе на Hub (глобально), бо WS-з'єднання не прив'язане до однієї
// кімнати, а токен має пережити повний обрив сокета.
// =============================================================================

package network

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// reconnectTokenTTL — скільки живе виданий токен без активності.
// Достатньо довгий для типової партії; протухлі записи чистяться лениво.
const reconnectTokenTTL = 6 * time.Hour

// ReconnectSession — серверний запис про можливість перепідключення.
type ReconnectSession struct {
	Token     string
	UserID    string
	RoomID    string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// reconnectRegistry — потокобезпечний реєстр токенів.
type reconnectRegistry struct {
	mu         sync.RWMutex
	byToken    map[string]*ReconnectSession
	byUserRoom map[string]string // "userID|roomID" -> token (для реюзу/інвалідизації)
}

func newReconnectRegistry() *reconnectRegistry {
	return &reconnectRegistry{
		byToken:    make(map[string]*ReconnectSession),
		byUserRoom: make(map[string]string),
	}
}

func userRoomKey(userID, roomID string) string {
	return userID + "|" + roomID
}

// generateReconnectToken повертає криптографічно випадковий hex-токен.
func generateReconnectToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Дуже малоймовірно; fallback на timestamp, щоб не панікувати.
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

// Issue створює (або переюзає) токен для пари {userID, roomID}.
// Якщо для цієї пари вже є валідний токен — продовжуємо його TTL і повертаємо
// той самий рядок, щоб уникнути "мертвих" записів при повторних Issue.
func (rr *reconnectRegistry) Issue(userID, roomID string) string {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	now := time.Now()
	key := userRoomKey(userID, roomID)

	if existingToken, ok := rr.byUserRoom[key]; ok {
		if sess, ok2 := rr.byToken[existingToken]; ok2 && now.Before(sess.ExpiresAt) {
			sess.ExpiresAt = now.Add(reconnectTokenTTL)
			return existingToken
		}
		delete(rr.byToken, existingToken)
	}

	token := generateReconnectToken()
	rr.byToken[token] = &ReconnectSession{
		Token:     token,
		UserID:    userID,
		RoomID:    roomID,
		IssuedAt:  now,
		ExpiresAt: now.Add(reconnectTokenTTL),
	}
	rr.byUserRoom[key] = token
	return token
}

// Validate перевіряє токен і повертає прив'язану сесію.
// Токен вважається валідним лише якщо він збігається З ПОТОЧНИМ userID
// підключеного сокета (запобігаємо викраденню чужої сесії).
func (rr *reconnectRegistry) Validate(token, userID string) (*ReconnectSession, bool) {
	rr.mu.RLock()
	sess, ok := rr.byToken[token]
	rr.mu.RUnlock()

	if !ok {
		return nil, false
	}
	if time.Now().After(sess.ExpiresAt) {
		rr.Revoke(token)
		return nil, false
	}
	if sess.UserID != userID {
		return nil, false
	}
	return sess, true
}

// Revoke видаляє токен (напр., коли гравець свідомо вийшов або гра завершилась).
func (rr *reconnectRegistry) Revoke(token string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	if sess, ok := rr.byToken[token]; ok {
		delete(rr.byUserRoom, userRoomKey(sess.UserID, sess.RoomID))
		delete(rr.byToken, token)
	}
}

// RevokeForRoomUser знімає токен за парою {userID, roomID} (без знання самого токена).
func (rr *reconnectRegistry) RevokeForRoomUser(userID, roomID string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	key := userRoomKey(userID, roomID)
	if token, ok := rr.byUserRoom[key]; ok {
		delete(rr.byToken, token)
		delete(rr.byUserRoom, key)
	}
}
