package services

import (
	"context"
	v2 "encoding/json/v2"
	"fmt"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	role  = "user"
	topic = "user.social"
)

func (s *AuthService) LoginUser(ctx context.Context, req *dto.UsersLoginDTO) (*models.Tokens, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetUserByEmail: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("bcrypt.CompareHashAndPassword: %w", domain.ErrInvalidCredentials)
	}

	event := &dto.UserKafka{
		UserId: user.ID,
	}

	data, err := v2.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("v2.Marshal: %w", err)
	}

	record := &kgo.Record{
		Topic: topic,
		Value: data,
	}
	result := s.client.ProduceSync(ctx, record)
	if err := result.FirstErr(); err != nil {
		return nil, fmt.Errorf("s.client.ProduceSync: %w", err)
	}
	s.logger.Debug("write data in kafka", zap.String("topic", topic), zap.String("user_id", user.ID.String()))

	accessToken, err := s.tokens.GenerateToken(user.ID, role, user.EmailVerified, user.EmailBanned, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}

	refreshToken, err := s.tokens.GenerateToken(user.ID, role, user.EmailVerified, user.EmailBanned, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessionStore.SaveToken(ctx, user.ID, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.sessionStore.SaveToken: %w", err)
	}

	s.logger.Debug("successful user login", zap.String("user_id", user.ID.String()))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
