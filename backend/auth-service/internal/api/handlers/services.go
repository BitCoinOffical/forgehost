package handlers

import (
	rabbitmq "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	rabbitqueue "github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/queue/rabbitMQ"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/repo"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/services"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/store"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/segmentio/kafka-go"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type Services struct {
	Service *services.AuthService
}

func NewServices(manager *jwtpkg.ManagerToken, rdb *redis.Client, pool *pgxpool.Pool, WebGoogleClientID string, rc *rabbitmq.ResilientConnection, logger *zap.Logger, writer *kafka.Writer) *Services {
	queue := rabbitqueue.NewQueue(rc)

	codeStore := store.NewCodeStore(rdb)
	resendStore := store.NewResendStore(rdb)
	sessionStore := store.NewSessionStore(rdb)
	userStore := store.NewUserStore(rdb)

	repo := repo.NewAuthRepo(pool)
	service := services.NewAuthService(repo, manager, codeStore, resendStore, sessionStore, userStore, WebGoogleClientID, queue, logger, writer)
	return &Services{Service: service}
}

type Handlers struct {
	Auth *AuthHandler
}

func NewHandlers(logger *zap.Logger, srv *Services, cfg *oauth2.Config) *Handlers {
	auth := NewAuthHandler(logger, srv.Service, cfg)
	return &Handlers{Auth: auth}
}
