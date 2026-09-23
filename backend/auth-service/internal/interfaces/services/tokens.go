package services

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RefreshService struct {
	tokens       ManagerToken
	sessionStore SessionStore
	logger       *zap.Logger
}

func NewRefreshService(
	tokens ManagerToken,
	sessionStore SessionStore,
	logger *zap.Logger,
) *RefreshService {
	return &RefreshService{
		tokens:       tokens,
		sessionStore: sessionStore,
		logger:       logger,
	}
}

func (s *RefreshService) UpdateAccessToken(ctx context.Context, refreshToken string) (*models.Tokens, error) {
	user, err := s.tokens.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("s.tokens.ValidateToken: %w", err)
	}

	id, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	savedToken, err := s.sessionStore.GetToken(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.sessionStore.GetToken: %w", err)
	}

	if savedToken != refreshToken {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.tokens.GenerateToken(id, role, user.IsVerified, user.IsBanned, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}

	s.logger.Debug("successful token access", zap.String("user_id", id.String()))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
