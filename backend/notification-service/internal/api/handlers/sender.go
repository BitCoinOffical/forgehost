package handlers

import (
	"context"

	notificationv1 "github.com/BitCoinOffical/forgehost-proto/notification/v1"
	"github.com/BitCoinOffical/forgehost/notification-service/internal/adapters/email"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	notificationv1.UnimplementedNotificationServiceServer
	Rclient *email.ResendClient
	logger  *zap.Logger
}

func NewNotificationHandler(rclient *email.ResendClient, logger *zap.Logger) *Handler {
	return &Handler{
		Rclient: rclient,
		logger:  logger,
	}
}

func (h *Handler) SendEmail(ctx context.Context, req *notificationv1.SendEmailRequest) (*notificationv1.SendEmailResponse, error) {
	sendId, err := h.Rclient.SendCodeEmail([]string{req.Email}, req.Code, req.TitleSubject)
	if err != nil {
		h.logger.Error("failed to send email",
			zap.String("user_id", req.UserId),
			zap.String("error", err.Error()),
		)
		return nil, status.Errorf(codes.Internal, "failed to send email: %v", err)
	}

	h.logger.Info("email sent",
		zap.String("user_id", req.UserId),
		zap.String("email_id", sendId),
	)

	return &notificationv1.SendEmailResponse{
		EmailId: sendId,
	}, nil
}
