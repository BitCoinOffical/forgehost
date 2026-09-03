package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BitCoinOffical/forgehost/social-service/config"
	kafkaread "github.com/BitCoinOffical/forgehost/social-service/internal/adapters/kafka"
	"github.com/BitCoinOffical/forgehost/social-service/internal/adapters/migrations"
	postgresdb "github.com/BitCoinOffical/forgehost/social-service/internal/adapters/postgres"
	redisdb "github.com/BitCoinOffical/forgehost/social-service/internal/adapters/redis"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/consumers"
	jwtpkg "github.com/BitCoinOffical/forgehost/social-service/pkg/jwt"
	loggerpkg "github.com/BitCoinOffical/forgehost/social-service/pkg/logger"
	"go.uber.org/zap"
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
	logger.Info("logger load")

	reader := kafkaread.NewKafkaReaвer(&kafkaread.KafkaConfig{
		Addr: cfg.Kafka.Addr,
	})
	logger.Info("kafka reader applied successfully")

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

	if err := migrations.RunMigrations(pool); err != nil {
		logger.Fatal("failed run migrated", zap.Error(err))
	}
	logger.Info("migrated applied successfully")

	cache, err := redisdb.NewRedis(&redisdb.RedisConfig{
		RDBAddr: cfg.Redis.RDBCacheAddr,
		RDBPort: cfg.Redis.RDBCachePort,
		RDBDB:   cfg.Redis.RDBCacheDB,
		RDBPass: cfg.Redis.RDBPass,
	})
	if err != nil {
		logger.Fatal("redis for cache failed", zap.Error(err))
	}
	logger.Info("redis for cache applied successfully")

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

	srv := consumers.NewServices(pool)
	cons := consumers.NewConsumers(srv, reader, logger)

	go func() {
		errs := cons.Profile.Run(context.Background())
		for err := range errs {
			logger.Error("Profile run error", zap.Error(err))
		}
	}()

	manager := jwtpkg.NewManagerToken(cfg.App.Secret)
	m := middleware.NewMiddleware(rate, logger, manager)
	s := handlers.NewServices(pool, cache)
	h := handlers.NewHandlers(s, logger)

	serv := api.NewServer(&config.AppConfig{
		DebugLevel: cfg.App.DebugLevel,
		Port:       cfg.App.Port,
		Secret:     cfg.App.Secret,
	}, m, h)

	<-ctx.Done()
	if err := kafkaread.KafkaClose(reader); err != nil {
		logger.Fatal("failed close kafka", zap.Error(err))
	}
	logger.Info("kafka closed")

	if err := serv.ShutDown(ctx); err != nil {
		logger.Fatal("failed shut down server", zap.Error(err))
	}
	logger.Info("kafka closed")

	postgresdb.ClosePool(pool)
	logger.Info("pool closed")
}
