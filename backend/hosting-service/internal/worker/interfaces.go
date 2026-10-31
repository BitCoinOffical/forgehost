package worker

import (
	"context"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	redis "github.com/redis/go-redis/v9"
)

type WorldsCache interface {
	SetOnline(ctx context.Context, world *models.MinecraftServer) error
	GetAll(ctx context.Context) ([]models.MinecraftServer, error)
	GetByID(ctx context.Context, id string) (*models.MinecraftServer, error)
	Remove(ctx context.Context, id string) error
}
type WorldsQueue interface {
	Ack(ctx context.Context, messageID string) error
	CreateGroup(ctx context.Context) error
	Enqueue(ctx context.Context, world *models.MinecraftServerCheckPayload) (string, error)
	Read(ctx context.Context, consumer string) ([]redis.XMessage, error)
}

type MinecraftChecker interface {
	Check(ctx context.Context, world *models.MinecraftServer) (*models.MinecraftStatus, error)
}
