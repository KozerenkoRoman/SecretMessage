package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtKey = []byte("super_secret_key_change_me_in_production")

type Claims struct {
	UserID     string `json:"user_id"` // Змінено на string для передачі UUID
	Username   string `json:"username"`
	UserRole   string `json:"user_role"`   // Змінено з Role на UserRole
	AvatarSeed string `json:"avatar_seed"` // Синхронізовано з фронтом (auth.js initUserFromToken)
	jwt.RegisteredClaims
}

// GenerateToken тепер приймає uuid.UUID
func GenerateToken(userID uuid.UUID, username, userRole, avatarSeed string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:     userID.String(), // Конвертуємо UUID у string
		Username:   username,
		UserRole:   userRole,
		AvatarSeed: avatarSeed,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
