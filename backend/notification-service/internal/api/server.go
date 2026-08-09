package api

import (
	"fmt"
	"net"

	notificationv1 "github.com/BitCoinOffical/forgehost-proto/notification/v1"
	"github.com/BitCoinOffical/forgehost/notification-service/internal/adapters/email"
	"github.com/BitCoinOffical/forgehost/notification-service/internal/api/handlers"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ServerConfig struct {
	Network string
	Address string
}

func NewNotificationServer(cfg *ServerConfig, rclient *email.ResendClient, logger *zap.Logger) error {
	lis, err := net.Listen(cfg.Network, cfg.Address)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	grpcServer := grpc.NewServer()

	notificationv1.RegisterNotificationServiceServer(grpcServer, handlers.NewNotificationHandler(rclient, logger))

	logger.Info("notification server start")

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("grpcServer.Serve: %w", err)
	}

	return nil
}
