package handlers

import (
	"context"

	rabbitmq "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	rabbitqueue "github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/queue/rabbitMQ"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/repo"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/services"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/store"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type LoginService interface {
	LoginUser(ctx context.Context, req *dto.UsersLoginDTO) (*models.Tokens, error)
}

type LogoutService interface {
	LogoutUser(ctx context.Context, userID string) error
}

type GoogleAuthService interface {
	GoogleCallback(ctx context.Context, req *dto.GoogleUserDTO) (string, error)
	GoogleLoginAndroid(ctx context.Context, req dto.GoogleAndroidUserDTO) (*models.Tokens, error)
	Exchange(ctx context.Context, req *dto.ExchangeRequestDTO) (*models.Tokens, error)
}

type PasswordService interface {
	UpdatePassword(ctx context.Context, req *dto.UserPasswordDTO, userID string) error
	PasswordReset(ctx context.Context, req *dto.PasswordResetDTO) (*dto.PendingKeyDTO, error)
	ConfirmPasswordReset(ctx context.Context, req *dto.PasswordResetDTO) error
	PasswordResetResend(ctx context.Context, req *dto.PasswordResetDTO) error
}

type RegisterService interface {
	RegisterUser(ctx context.Context, req *dto.UsersRegisterDTO) (*dto.PendingKeyDTO, error)
	VerifyEmail(ctx context.Context, req *dto.VerifyEmailDTO) (*models.Tokens, error)
	ResendVerifyEmail(ctx context.Context, req *dto.VerifyEmailDTO) error
}

type RefreshService interface {
	UpdateAccessToken(ctx context.Context, refreshToken string) (*models.Tokens, error)
}

type Services struct {
	login  LoginService
	logout LogoutService
	oauth  GoogleAuthService
	pass   PasswordService
	reg    RegisterService
	token  RefreshService
}

func NewServices(manager *jwtpkg.ManagerToken, rdb *redis.Client, pool *pgxpool.Pool, WebGoogleClientID string, rc *rabbitmq.ResilientConnection, logger *zap.Logger, client *kgo.Client) *Services {
	queue := rabbitqueue.NewQueue(rc)

	codeStore := store.NewCodeStore(rdb)
	resendStore := store.NewResendStore(rdb)
	sessionStore := store.NewSessionStore(rdb)
	userStore := store.NewUserStore(rdb)

	repo := repo.NewAuthRepo(pool)

	login := services.NewLoginService(repo, client, manager, sessionStore, logger)
	logout := services.NewLogoutService(sessionStore, logger)
	oauth := services.NewGoogleAuthService(repo, codeStore, client, manager, sessionStore, logger, WebGoogleClientID)
	pass := services.NewPasswordService(repo, codeStore, resendStore, queue, logger)
	reg := services.NewRegisterService(repo, manager, userStore, codeStore, resendStore, sessionStore, queue, logger, client)
	token := services.NewRefreshService(manager, sessionStore, logger)

	return &Services{login: login, logout: logout, oauth: oauth, pass: pass, reg: reg, token: token}
}

type Handlers struct {
	Login  *LoginHandler
	Logout *LogoutHandler
	Oauth  *GoogleAuthHandler
	Pass   *PasswordHandler
	Reg    *RegisterHandler
	Token  *RefreshHandler
}

func NewHandlers(logger *zap.Logger, srv *Services, cfg *oauth2.Config) *Handlers {
	login := NewLoginHandler(srv.login, logger)
	logout := NewLogoutHandler(srv.logout, logger)
	oauth := NewGoogleAuthHandler(srv.oauth, cfg, logger)
	pass := NewPasswordHandler(srv.pass, logger)
	reg := NewRegisterHandler(srv.reg, logger)
	token := NewRefreshHandler(srv.token, logger)
	return &Handlers{Login: login, Logout: logout, Oauth: oauth, Pass: pass, Reg: reg, Token: token}
}
