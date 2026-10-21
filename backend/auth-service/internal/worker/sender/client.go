package sender

import (
	"context"

	notificationv1 "github.com/BitCoinOffical/forgehost-proto/notification/v1"
)

type NotificationClient interface {
	Close() error
	SendEmail(ctx context.Context, req *notificationv1.SendEmailRequest) (*notificationv1.SendEmailResponse, error)
}
