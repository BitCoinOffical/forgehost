package notification

import (
	"context"
	"fmt"

	notificationv1 "github.com/BitCoinOffical/forgehost-proto/notification/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationConfig struct {
	Addr string
}

type NotificationClient struct {
	client notificationv1.NotificationServiceClient
	conn   *grpc.ClientConn
}

func NewNotificationClient(config *NotificationConfig) (*NotificationClient, error) {
	conn, err := grpc.NewClient(
		config.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	client := notificationv1.NewNotificationServiceClient(conn)

	return &NotificationClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *NotificationClient) Close() error {
	return c.conn.Close()
}
func (c *NotificationClient) SendEmail(ctx context.Context, req *notificationv1.SendEmailRequest) (*notificationv1.SendEmailResponse, error) {
	return c.client.SendEmail(ctx, req)
}
