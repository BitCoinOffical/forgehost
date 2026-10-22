package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
)

type AuthRepository interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	SaveUser(ctx context.Context, req *models.User) (uuid.UUID, error)
	SaveGoogleUser(ctx context.Context, req *models.User, oathReq *models.OAuthAccount) (uuid.UUID, error)
	UpdateUserPassword(ctx context.Context, user *models.User) error
	CheckExistsUserByEmail(ctx context.Context, email string) (bool, error)
}

type CodeStore interface {
	SaveVerificationCode(ctx context.Context, randStr string, value int, expiration time.Duration) error
	SaveResetPasswordCode(ctx context.Context, randStr string, value int, expiration time.Duration) error
	SaveOauthCode(ctx context.Context, oauthCode, id string) error
	GetOauthCode(ctx context.Context, oauthCode string) (string, error)
	DeleteOauthCode(ctx context.Context, oauthCode string) error
	GetResetPasswordCode(ctx context.Context, randStr string) (string, error)
	GetVerificationCode(ctx context.Context, randStr string) (string, error)
}

type ResendStore interface {
	ResendLimitAdd(ctx context.Context, email string) error
	ResendLimitCheck(ctx context.Context, email string) (string, error)
}
type SessionStore interface {
	SaveToken(ctx context.Context, userID uuid.UUID, value string, RefreshTTL time.Duration) error
	GetToken(ctx context.Context, userID uuid.UUID) (string, error)
	DeleteToken(ctx context.Context, userID uuid.UUID) error
}

type UserStore interface {
	SaveUser(ctx context.Context, randStr string, user *models.UserStored) error
	GetUser(ctx context.Context, randStr string) (*models.UserStored, error)
}

type RabbitQueue interface {
	AddEmailTaskQueue(ctx context.Context, body []byte) error
}

type ManagerToken interface {
	ValidateToken(tokenString string) (*jwtpkg.Claims, error)
	GenerateToken(userID uuid.UUID, UserRole string, IsVerified bool, IsBanned bool, ttl time.Duration) (string, error)
}
