package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *AuthService) LogoutUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("uuid.Parse: %w", err)
	}

	if err := s.sessionStore.DeleteToken(ctx, id); err != nil {
		return fmt.Errorf("s.sessionStore.DeleteToken: %w", err)
	}
	s.logger.Debug("user exited", zap.Any("user_id", id))
	return nil
}
