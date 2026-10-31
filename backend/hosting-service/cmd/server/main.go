package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BitCoinOffical/forgehost/hosting-service/config"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/adapters/migrations"
	postgresdb "github.com/BitCoinOffical/forgehost/hosting-service/internal/adapters/postgres"
	redisdb "github.com/BitCoinOffical/forgehost/hosting-service/internal/adapters/redis"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/api"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/api/http/handlers"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/api/middleware"
	jwtpkg "github.com/BitCoinOffical/forgehost/hosting-service/pkg/jwt"
	loggerpkg "github.com/BitCoinOffical/forgehost/hosting-service/pkg/logger"
	"go.uber.org/zap"
)

const (
	logPath = "logs/hosting.log"
)

// @title ForgeHost Hosting Service API
// @version 1.0
// @description API for managing Minecraft worlds, server hosting and availability monitoring.
// @termsOfService http://swagger.io/terms/

// @contact.name ForgeHost Support
// @contact.url https://forgehost.ru

// @host localhost:8085
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewLoad()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := loggerpkg.NewLogger(cfg.App.DebugLevel, logPath)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("config load")
	logger.Info("logger load")

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

	manager := jwtpkg.NewManagerToken(cfg.App.Secret)
	mdlwr := middleware.NewMiddleware(rate, logger, manager)
	srvcs := handlers.NewServices(pool, cache)
	handlrs := handlers.NewHandlers(srvcs, logger)

	serv := api.NewServer(&config.AppConfig{
		DebugLevel: cfg.App.DebugLevel,
		Port:       cfg.App.Port,
		Secret:     cfg.App.Secret,
	}, mdlwr, handlrs)

	go func() {
		if err := serv.Run(); err != nil {
			logger.Fatal("failed start server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	if err := serv.ShutDown(ctx); err != nil {
		logger.Error("failed shut down server", zap.Error(err))
	}
	logger.Info("kafka closed")

	postgresdb.ClosePool(pool)
	logger.Info("pool closed")

	if err := cache.Close(); err != nil {
		logger.Error("failed close cache rdb", zap.Error(err))
	}
	if err := rate.Close(); err != nil {
		logger.Error("failed close rate rdb", zap.Error(err))
	}
}
