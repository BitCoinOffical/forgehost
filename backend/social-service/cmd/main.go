package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BitCoinOffical/forgehost/social-service/config"
	kafkaread "github.com/BitCoinOffical/forgehost/social-service/internal/adapters/kafka"
	postgresdb "github.com/BitCoinOffical/forgehost/social-service/internal/adapters/postgres"
	"github.com/BitCoinOffical/forgehost/social-service/internal/consumers"
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

	srv := consumers.NewServices(pool)
	cons := consumers.NewConsumers(srv, reader, logger)

	go func() {
		errs := cons.Profile.Run(context.Background())
		for err := range errs {
			logger.Error("Profile run error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	if err := kafkaread.KafkaClose(reader); err != nil {
		logger.Fatal("failed close kafka", zap.Error(err))
	}
	logger.Info("kafka closed")
}
