package jwtman

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	SecretKey     []byte        // Секрет для jwt
	TokenDuration time.Duration // Длительность токена доступа (вдавить 10-15 минут)
}

// Содержимое токена
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Генерируем токен
func (manager *JWTManager) GenerateAccessToken(UID int) (string, error) {
	jti := uuid.New().String()
	claims := &Claims{UserID: strconv.Itoa(UID),
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.TokenDuration)),
			IssuedAt: jwt.NewNumericDate(time.Now()), ID: jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(manager.SecretKey)
}

// Это тут наверное не нужно, но это верификация токена доступа
func (manager *JWTManager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return manager.SecretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
