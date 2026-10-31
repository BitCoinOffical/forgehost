package handlers

import (
	"context"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/interfaces/cache"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/interfaces/repo"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/interfaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type WorldsService interface {
	Upsert(ctx context.Context, req *dto.UpsertMinecraftServerDTO, userID string) (*models.MinecraftServer, error)
	GetAll(ctx context.Context) ([]models.MinecraftServer, error)
	GetByID(ctx context.Context, id string) (*models.MinecraftServer, error)
	Delete(ctx context.Context, id, userID string) error
}
type Services struct {
	srvc WorldsService
}

func NewServices(pool *pgxpool.Pool, rdb *redis.Client) *Services {
	repo := repo.NewWorldsRepo(pool)
	cache := cache.NewWorldsCache(rdb)
	srvc := services.NewWorldsService(repo, cache)
	return &Services{srvc: srvc}
}

type Handlers struct {
	Worlds *WorldsHandler
}

func NewHandlers(srvc *Services, logger *zap.Logger) *Handlers {
	worlds := NewWorldsHandler(srvc.srvc, logger)
	return &Handlers{Worlds: worlds}
}
