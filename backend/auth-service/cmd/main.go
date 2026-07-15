package main

import (
	"BitCoinOffical/forgehost/auth-service/config"
	"BitCoinOffical/forgehost/auth-service/internal/api"
	"BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"
	zaplogger "BitCoinOffical/forgehost/auth-service/pkg/logger"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	logger, err := zaplogger.NewLogger(cfg.App.DebugLevel)
	if err != nil {
		log.Fatal(err)
	}

	manager := jwtpkg.NewManagerToken(cfg.App.Secret)
	m := middleware.NewAuthMiddleware(logger, manager)
	serv := api.NewServer(&cfg.App, m)

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	if err := serv.ShutDown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %s", err)
	}
}
