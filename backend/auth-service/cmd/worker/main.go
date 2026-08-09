package main

import (
	"github.com/BitCoinOffical/forgehost/auth-service/config"
	rabbitmq "github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/adapters/notification"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/worker/sender"

	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	loggerpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/logger"

	"go.uber.org/zap"
)

const (
	maxWorkers = 4
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

	rc := rabbitmq.NewResilientConnection(&rabbitmq.RabbitURL{
		User: cfg.RabbitMQ.RabbitUser,
		Pass: cfg.RabbitMQ.RabbitPass,
		Host: cfg.RabbitMQ.RabbitHost,
		Port: cfg.RabbitMQ.RabbitPort,
	}, logger)

	client, err := notification.NewNotificationClient(&notification.NotificationConfig{
		Addr: cfg.Notification.NotificationAddr,
	})
	if err != nil {
		logger.Fatal("failed connect grpc notificatin ")
	}
	logger.Info("successful connection grpc notificatin")

	Work := sender.NewWorker(logger, client, rc)
	Errs := Work.WorkerPool(maxWorkers)
	for err := range Errs {
		logger.Error("worker error", zap.Error(err))
	}

	<-ctx.Done()
	logger.Info("received a signal indicating the completion of operations")

	client.Close()
	logger.Info("gRPC conn closed")

	rc.Close()
	logger.Info("rabbitMQ conn closed")
}
