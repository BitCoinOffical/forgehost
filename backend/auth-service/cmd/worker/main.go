package main

import (
	"BitCoinOffical/forgehost/auth-service/config"
	rabbitmq "BitCoinOffical/forgehost/auth-service/internal/adapters/RabbitMQ"
	"BitCoinOffical/forgehost/auth-service/internal/adapters/email"
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
	codeQueue     = "code_queue"
	resetQueue    = "reset_queue"
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

	rc := rabbitmq.NewResilientConnection(&rabbitmq.RabbitURL{
		User: cfg.RabbitMQ.RabbitUser,
		Pass: cfg.RabbitMQ.RabbitPass,
		Host: cfg.RabbitMQ.RabbitHost,
		Port: cfg.RabbitMQ.RabbitPort,
	}, logger)

	verifyWork := sender.NewWorker(logger, rclient, rc, codeQueue)
	verifyErrs := verifyWork.WorkerPool(2)

	resetWork := sender.NewWorker(logger, rclient, rc, resetQueue)
	resetErrs := resetWork.WorkerPool(2)

	go func() {
		for err := range verifyErrs {
			logger.Error("verify worker error", zap.Error(err))
		}
	}()
	go func() {
		for err := range resetErrs {
			logger.Error("reset worker error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("received a signal indicating the completion of operations")

	rc.Close()
	logger.Info("rabbitMQ conn closed")
}
