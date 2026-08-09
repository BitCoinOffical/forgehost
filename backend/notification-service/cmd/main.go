package main

import (
	"log"

	"github.com/BitCoinOffical/forgehost/notification-service/config"
	"github.com/BitCoinOffical/forgehost/notification-service/internal/adapters/email"
	"github.com/BitCoinOffical/forgehost/notification-service/internal/api"

	loggerpkg "github.com/BitCoinOffical/forgehost/notification-service/pkg/logger"
)

func main() {
	cfg, err := config.NewLoad()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := loggerpkg.NewLogger(cfg.App.DebugLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	logger.Info("logger started")
	logger.Info("config loaded")

	rclient := email.NewResendSender(&email.ResendConfig{
		ResendApiKey: cfg.Resend.ResendApiKey,
	})
	logger.Info("resend started")

	if err := api.NewNotificationServer(&api.ServerConfig{
		Network: cfg.App.Network,
		Address: cfg.App.Address,
	}, rclient, logger); err != nil {
		logger.Fatal("notification server start failed")
	}
	logger.Fatal("notification server stoped")
}
