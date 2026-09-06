package services

import (
	rabbitqueue "github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/queue/rabbitMQ"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/repo"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/store"
	"github.com/segmentio/kafka-go"

	"time"

	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"

	"go.uber.org/zap"
)

const (
	AccessTTL       = 5 * time.Minute
	RefreshTTL      = (24 * 30 * 12) * time.Hour
	VerificationTTL = 7 * time.Minute
	ResetPassTTL    = 7 * time.Minute
	codeSubject     = "Verification Code"
	resetSubject    = "Password Reset Code"
)

type AuthService struct {
	logger            *zap.Logger
	writer            *kafka.Writer
	tokens            *jwtpkg.ManagerToken
	repo              *repo.AuthRepo
	queue             *rabbitqueue.RabbitQueue
	codeStore         *store.CodeStore
	resendStore       *store.ResendStore
	sessionStore      *store.SessionStore
	userStore         *store.UserStore
	WebgoogleClientID string
}

func NewAuthService(
	repo *repo.AuthRepo,
	tokens *jwtpkg.ManagerToken,
	codeStore *store.CodeStore,
	resendStore *store.ResendStore,
	sessionStore *store.SessionStore,
	userStore *store.UserStore,
	WebgoogleClientID string,
	queue *rabbitqueue.RabbitQueue,
	logger *zap.Logger,
	writer *kafka.Writer,
) *AuthService {
	return &AuthService{
		repo:              repo,
		tokens:            tokens,
		codeStore:         codeStore,
		resendStore:       resendStore,
		sessionStore:      sessionStore,
		userStore:         userStore,
		WebgoogleClientID: WebgoogleClientID,
		queue:             queue,
		logger:            logger,
		writer:            writer,
	}
}
