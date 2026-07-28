package main

import (
	"BitCoinOffical/forgehost/auth-service/config"
	rabbitmq "BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	"BitCoinOffical/forgehost/auth-service/internal/adapters/email"
	"BitCoinOffical/forgehost/auth-service/internal/worker/connect"
	"BitCoinOffical/forgehost/auth-service/internal/worker/sender"

	loggerpkg "BitCoinOffical/forgehost/auth-service/pkg/logger"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

const (
	maxWorkers    = 4
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

	rclient := email.NewResendSender(&email.ResendConfig{
		ResendApiKey: cfg.Resend.ResendApiKey,
	})
	logger.Info("resend started")

	conn, err := rabbitmq.NewRabbitMQ(&rabbitmq.RabbitURL{
		User: cfg.RabbitMQ.RabbitUser,
		Pass: cfg.RabbitMQ.RabbitPass,
		Host: cfg.RabbitMQ.RabbitHost,
		Port: cfg.RabbitMQ.RabbitPort,
	})
	if err != nil {
		logger.Fatal("rabbitMQ failed", zap.Error(err))
	}
	logger.Info("rabbitMQ applied successfully")

	consumer := connect.NewRunConsumer(conn, logger)
	work := sender.NewWorker(logger, rclient, consumer)
	errs := work.WorkerPool(4)
	for err := range errs {
		logger.Error("worker error", zap.Error(err))
	}

	<-ctx.Done()
	logger.Info("received a signal indicating the completion of operations")

	if err := rabbitmq.CloseRabbitMQ(conn); err != nil {
		logger.Fatal("CloseRabbitMQ error", zap.Error(err))
	}
	logger.Info("rabbitMQ conn closed")
}
