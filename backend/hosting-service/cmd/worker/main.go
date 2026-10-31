package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BitCoinOffical/forgehost/hosting-service/config"
	redisdb "github.com/BitCoinOffical/forgehost/hosting-service/internal/adapters/redis"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/interfaces/cache"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/interfaces/checker"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/interfaces/queue"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/worker"
	loggerpkg "github.com/BitCoinOffical/forgehost/hosting-service/pkg/logger"
	"go.uber.org/zap"
)

const (
	logPath = "logs/hosting_worker.log"
	workers = 4
	timeout = 30
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

	cacherdb, err := redisdb.NewRedis(&redisdb.RedisConfig{
		RDBAddr: cfg.Redis.RDBCacheAddr,
		RDBPort: cfg.Redis.RDBCachePort,
		RDBDB:   cfg.Redis.RDBCacheDB,
		RDBPass: cfg.Redis.RDBPass,
	})
	if err != nil {
		logger.Fatal("redis for cache failed", zap.Error(err))
	}
	logger.Info("redis for cache applied successfully")

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

	ch := checker.NewMinecraftChecker(timeout)
	q := queue.NewQueue(queuerdb)
	c := cache.NewWorldsCache(cacherdb)
	w := worker.NewWorker(logger, c, q, ch)
	w.WorkerPool(ctx, workers)

	<-ctx.Done()

	if err := cacherdb.Close(); err != nil {
		logger.Error("failed close cache rdb", zap.Error(err))
	}

	if err := queuerdb.Close(); err != nil {
		logger.Error("failed close queue rdb", zap.Error(err))
	}

}
