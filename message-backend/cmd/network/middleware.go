package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"secret-message/cmd/auth"
)

type contextKey string

const UserContextKey contextKey = "user_info"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Дістаємо токен із заголовка або Query-параметра (для WebSocket)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			authHeader = r.URL.Query().Get("token")
		} else {
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				authHeader = authHeader[7:]
			}
		}

		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(authHeader)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Записуємо claims в контекст запиту
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminOnlyMiddleware перевіряє, чи має авторизований користувач роль "admin"
func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Дістаємо повні Claims із контексту за правильним ключем UserContextKey
		// Оскільки ValidateToken зазвичай повертає покажчик, кастимо до *auth.Claims
		claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)

		// Якщо не вдалося дістати як покажчик, спробуємо як звичайну структуру (про всяк випадок)
		if !ok {
			if directClaims, valid := r.Context().Value(UserContextKey).(auth.Claims); valid {
				claims = &directClaims
				ok = true
			}
		}

		// 2. Якщо claims немає в контексті або роль не "admin" — повертаємо 403 Forbidden
		if !ok || claims == nil || claims.UserRole != "admin" {
			// Лог для налагодження на сервері (можна прибрати в продакшені)
			if !ok {
				fmt.Println("[ADMIN ERROR] Claims не знайдено в контексті запиту")
			} else if claims != nil {
				fmt.Println("[ADMIN ERROR] Спроба доступу користувача з роллю:", claims.UserRole)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "forbidden: administrative privileges required",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
