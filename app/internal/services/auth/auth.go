package auth

import (
	jwtman "auth_service/internal/JWT/access"
	"auth_service/internal/JWT/refresh"
	"auth_service/internal/models"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/redis/go-redis/v9"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	CreateNewUser(ctx context.Context, newUser models.NewUser) error
	IsAdmin(ctx context.Context, UID int) bool
}

type Auth struct {
	Logger  *slog.Logger
	Storage UserRepository
	JWT     *jwtman.JWTManager
	Redis   *redis.Client
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewAuth(logger *slog.Logger, Storage UserRepository, Redis *redis.Client, JWT *jwtman.JWTManager) *Auth {
	return &Auth{Logger: logger, Storage: Storage, JWT: JWT, Redis: Redis}
}

func HashPassword(password []byte) ([]byte, error) {
	bytes, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	return bytes, err
}

func CheckPasswordHash(password, hash []byte) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (auth *Auth) Register(ctx context.Context, user models.NewUser) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Хешируем пароль
	hashed, err := HashPassword(user.HashPass)
	if err != nil {
		return err
	}
	user.HashPass = hashed

	return auth.Storage.CreateNewUser(ctx, user)
}

func (auth *Auth) Login(ctx context.Context, user models.NewUser) (*AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	storedUser, err := auth.Storage.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	if ok := CheckPasswordHash(user.HashPass, storedUser.HashPass); !ok {
		auth.Logger.Info("Неверный пароль от пользователя", slog.String(user.Email, ""))
		return nil, errors.New("неправильный пароль")
	}

	accessToken, err := auth.JWT.GenerateAccessToken(storedUser.UID)
	if err != nil {
		return nil, err
	}

	refreshToken := refresh.GenerateRefreshToken()
	struid := strconv.Itoa(storedUser.UID)
	if err := auth.StoreRefreshToken(ctx, struid, refreshToken); err != nil {
		return nil, err
	}

	auth.Logger.Debug("Успешно создан токен", slog.String("user_id", struid))
	return &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// Сохраняем в Redis рефреш токен
func (auth *Auth) StoreRefreshToken(ctx context.Context, UID, refreshToken string) error {
	key := fmt.Sprintf("refresh:%s", refreshToken)
	return auth.Redis.Set(ctx, key, UID, auth.JWT.TokenDuration).Err()
}

// Проверка рефреш токена
func (auth *Auth) VerifyRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	key := fmt.Sprintf("refresh:%s", refreshToken)
	userID, err := auth.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("invalid or expired refresh token")
	}
	return userID, err
}

func (auth *Auth) Refresh(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	userID, err := auth.VerifyRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	uid, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid stored user id: %w", err)
	}

	// Генерация нового Access Token
	accessToken, err := auth.JWT.GenerateAccessToken(uid)
	if err != nil {
		return nil, err
	}
	auth.Logger.Debug("Создан новый токен", slog.String("user_id", userID))
	oldKey := fmt.Sprintf("refresh:%s", refreshToken)
	if err := auth.Redis.Del(ctx, oldKey).Err(); err != nil {
		auth.Logger.Warn("Не удалось удалить старый refresh токен", slog.String("refresh", refreshToken))
	}
	// Ротация Refresh Token (по желанию)
	newRefreshToken := refresh.GenerateRefreshToken()

	if err := auth.StoreRefreshToken(ctx, userID, newRefreshToken); err != nil {
		return nil, err
	}
	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (auth *Auth) Logout(ctx context.Context, refreshToken string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh:%s", refreshToken)
	err := auth.Redis.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	auth.Logger.Debug("Пользователь разлогинился", slog.String("refresh_token", refreshToken))
	return nil
}
