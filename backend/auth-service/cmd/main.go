package main

import (
	"BitCoinOffical/forgehost/auth-service/config"
	postgresdb "BitCoinOffical/forgehost/auth-service/internal/adapters/postgres"
	redisdb "BitCoinOffical/forgehost/auth-service/internal/adapters/redis"
	"BitCoinOffical/forgehost/auth-service/internal/api"
	"BitCoinOffical/forgehost/auth-service/internal/api/handlers"
	"BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"
	zaplogger "BitCoinOffical/forgehost/auth-service/pkg/logger"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	timeoutSecond = 5
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewLoad()
	if err != nil {
		log.Fatal(err)
	}
	logger, err := zaplogger.NewLogger(cfg.App.DebugLevel)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := postgresdb.NewPool(&postgresdb.PostgresConfig{
		DBUser:     cfg.Postgres.DBUser,
		DBPassword: cfg.Postgres.DBPassword,
		DBHost:     cfg.Postgres.DBHost,
		DBPort:     cfg.Postgres.DBHost,
		DBName:     cfg.Postgres.DBName,
	})

	rdb, err := redisdb.NewRedis(&redisdb.RedisConfig{
		RDBAddr: cfg.Redis.RDBSessionAddr,
		RDBPort: cfg.Redis.RDBSessionPort,
		RDBDB:   cfg.Redis.RDBSessionDB,
		RDBPass: cfg.Redis.RDBPass,
	})
	if err != nil {
		logger.Fatal("redis failed", zap.Error(err))
	}
	logger.Info("redis applied successfully")

	manager := jwtpkg.NewManagerToken(cfg.App.Secret)
	m := middleware.NewAuthMiddleware(logger, manager)
	srvs := handlers.NewServices(manager, rdb, pool)
	handlrs := handlers.NewHandlers(logger, srvs, &oauth2.Config{
		RedirectURL:  cfg.Google.RedirectURL,
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	})
	serv := api.NewServer(&cfg.App, m, handlrs)

	go func() {
		if err := serv.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	postgresdb.ClosePool(pool)
	logger.Info("pool closesd")

	if err := serv.ShutDown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
