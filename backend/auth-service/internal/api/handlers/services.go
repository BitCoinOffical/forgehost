package handlers

import (
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/repo"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/services"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/session"
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Services struct {
	service *services.AuthService
}

func NewServices(manager *jwtpkg.ManagerToken, rdb *redis.Client, pool *pgxpool.Pool) *Services {
	sessions := session.NewSession(rdb)
	repo := repo.NewAuthRepo(pool)
	service := services.NewAuthService(repo, manager, sessions)
	return &Services{service: service}
}

type Handlers struct {
	auth AuthHandler
}

func NewHandlers(logger *zap.Logger, srv Services) *Handlers {
	auth := NewAuthHandler(logger, srv.service)
	return &Handlers{auth: *auth}
}
