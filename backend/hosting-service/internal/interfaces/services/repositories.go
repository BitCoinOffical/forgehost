package services

import (
	"context"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
)

type WorldsRepo interface {
	Upsert(ctx context.Context, world *models.MinecraftServer) (*models.MinecraftServer, error)
	GetAll(ctx context.Context) ([]models.MinecraftServer, error)
	GetByID(ctx context.Context, id string) (*models.MinecraftServer, error)
	Delete(ctx context.Context, id, userID string) error
}

type WorldsCache interface {
	SetOnline(ctx context.Context, world *models.MinecraftServer) error
	GetAll(ctx context.Context) ([]models.MinecraftServer, error)
	GetByID(ctx context.Context, id string) (*models.MinecraftServer, error)
	Remove(ctx context.Context, id string) error
}
