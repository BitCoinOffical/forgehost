package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BitCoinOffical/forgehost/hosting-service/config"
	postgresdb "github.com/BitCoinOffical/forgehost/hosting-service/internal/adapters/postgres"
	redisdb "github.com/BitCoinOffical/forgehost/hosting-service/internal/adapters/redis"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/interfaces/queue"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/interfaces/repo"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/monitoring"
	loggerpkg "github.com/BitCoinOffical/forgehost/hosting-service/pkg/logger"
	"go.uber.org/zap"
)

const (
	logPath = "logs/hosting_monitoring.log"
)

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

	queuerdb, err := redisdb.NewRedis(&redisdb.RedisConfig{
		RDBAddr: cfg.Redis.RDBQueueAddr,
		RDBPort: cfg.Redis.RDBQueuePort,
		RDBDB:   cfg.Redis.RDBQueueDB,
		RDBPass: cfg.Redis.RDBPass,
	})
	if err != nil {
		logger.Fatal("redis for queue failed", zap.Error(err))
	}
	logger.Info("redis for queue applied successfully")

	q := queue.NewQueue(queuerdb)
	repo := repo.NewWorldsRepo(pool)
	monitor := monitoring.NewMonitoring(repo, q, logger)
	if err := monitor.Monitor(ctx); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()

	if err := queuerdb.Close(); err != nil {
		logger.Error("failed close queue rdb", zap.Error(err))
	}
}
