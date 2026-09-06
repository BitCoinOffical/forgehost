package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BitCoinOffical/forgehost/auth-service/config"
	rabbitmq "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	kafkaconn "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/kafka"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/migrations"
	postgresdb "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/postgres"
	redisdb "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/redis"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/handlers"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	loggerpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/logger"

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

	logger, err := loggerpkg.NewLogger(cfg.App.DebugLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("config load")
	logger.Info("logger start")

	pool, err := postgresdb.NewPool(&postgresdb.PostgresConfig{
		DBUser:     cfg.Postgres.DBUser,
		DBPassword: cfg.Postgres.DBPassword,
		DBHost:     cfg.Postgres.DBHost,
		DBPort:     cfg.Postgres.DBPort,
		DBName:     cfg.Postgres.DBName,
	})
	if err != nil {
		logger.Fatal("postgress pool failed", zap.Error(err))
	}
	logger.Info("postgress pool applied successfully")

	rate, err := redisdb.NewRedis(&redisdb.RedisConfig{
		RDBAddr: cfg.Redis.RDBRateAddr,
		RDBPort: cfg.Redis.RDBRatePort,
		RDBDB:   cfg.Redis.RDBLimiterDB,
		RDBPass: cfg.Redis.RDBPass,
	})
	if err != nil {
		logger.Fatal("redis for rate limitter failed", zap.Error(err))
	}
	logger.Info("redis for rate limitter applied successfully")

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

	rc := rabbitmq.NewResilientConnection(&rabbitmq.RabbitURL{
		User: cfg.RabbitMQ.RabbitUser,
		Pass: cfg.RabbitMQ.RabbitPass,
		Host: cfg.RabbitMQ.RabbitHost,
		Port: cfg.RabbitMQ.RabbitPort,
	}, logger)

	if err := migrations.RunMigrations(pool); err != nil {
		logger.Fatal("migrations failed", zap.Error(err))
	}
	logger.Info("migrations applied successfully")

	writer := kafkaconn.NewKafkaConn(&kafkaconn.KafkaConn{
		Addr: cfg.Kafka.Addr,
	})

	manager := jwtpkg.NewManagerToken(cfg.App.Secret)
	m := middleware.NewMiddleware(rate, logger, manager)
	srvs := handlers.NewServices(manager, rdb, pool, cfg.WebGoogle.WebClientID, rc, logger, writer)
	handlrs := handlers.NewHandlers(logger, srvs, &oauth2.Config{
		RedirectURL:  cfg.WebGoogle.WebRedirectURL,
		ClientID:     cfg.WebGoogle.WebClientID,
		ClientSecret: cfg.WebGoogle.WebClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	})
	serv := api.NewServer(&cfg.App, m, handlrs)
	go func() {
		if err := serv.Run(); err != nil {
			logger.Fatal("failed start server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()
	logger.Info("received a signal indicating the completion of operations")

	rc.Close()
	logger.Info("rabbitMQ conn closed")

	postgresdb.ClosePool(pool)
	logger.Info("pool closesd")

	if err := kafkaconn.KafkaClose(writer); err != nil {
		logger.Fatal("failed close kafka", zap.Error(err))
	}
	logger.Info("kafka closed")

	if err := serv.ShutDown(shutdownCtx); err != nil {
		logger.Fatal("shutdown error", zap.Error(err))
	}
	logger.Info("server shut down")
}
