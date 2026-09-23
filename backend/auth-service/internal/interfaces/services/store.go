package services

import (
	"context"
	"time"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/google/uuid"
)

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
	SaveToken(ctx context.Context, id uuid.UUID, value string, RefreshTTL time.Duration) error
	GetToken(ctx context.Context, id uuid.UUID) (string, error)
	DeleteToken(ctx context.Context, id uuid.UUID) error
}

type UserStore interface {
	SaveUser(ctx context.Context, randStr string, user *models.UserStored) error
	GetUser(ctx context.Context, randStr string) (*models.UserStored, error)
}
