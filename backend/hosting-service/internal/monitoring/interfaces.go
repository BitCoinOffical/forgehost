package monitoring

import (
	"context"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	redis "github.com/redis/go-redis/v9"
)

type WorldsRepo interface {
	Upsert(ctx context.Context, world *models.MinecraftServer) (*models.MinecraftServer, error)
	GetAll(ctx context.Context) ([]models.MinecraftServer, error)
	GetByID(ctx context.Context, id string) (*models.MinecraftServer, error)
	Delete(ctx context.Context, id, userID string) error
}
type WorldsQueue interface {
	CreateGroup(ctx context.Context) error
	Enqueue(ctx context.Context, world *models.MinecraftServerCheckPayload) (string, error)
	Read(ctx context.Context, consumer string) ([]redis.XMessage, error)
}
