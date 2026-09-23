package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
)

type AuthRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	SaveUser(ctx context.Context, req *models.User) (uuid.UUID, error)
	SaveGoogleUser(ctx context.Context, req *models.User, oathReq *models.OAuthAccount) (uuid.UUID, error)
	UpdateUserPassword(ctx context.Context, user *models.User) error
}
