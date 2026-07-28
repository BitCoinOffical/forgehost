package handlers

import (
	rabbitqueue "BitCoinOffical/forgehost/auth-service/internal/interfaces/queue/rabbitMQ"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/repo"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/services"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/session"
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type Services struct {
	Service *services.AuthService
}

func NewServices(manager *jwtpkg.ManagerToken, rdb *redis.Client, pool *pgxpool.Pool, WebGoogleClientID string, conn *amqp091.Connection, logger *zap.Logger) *Services {
	queue := rabbitqueue.NewQueue(conn)
	sessions := session.NewSession(rdb)
	repo := repo.NewAuthRepo(pool)
	service := services.NewAuthService(repo, manager, sessions, WebGoogleClientID, queue, logger)
	return &Services{Service: service}
}

type Handlers struct {
	Auth *AuthHandler
}

func NewHandlers(logger *zap.Logger, srv *Services, cfg *oauth2.Config) *Handlers {
	auth := NewAuthHandler(logger, srv.Service, cfg)
	return &Handlers{Auth: auth}
}
