package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	role = "user"
)

func (s *AuthService) LoginUser(ctx context.Context, req *dto.UsersLoginDTO) (*models.Tokens, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetUserByEmail: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("bcrypt.CompareHashAndPassword: %w", domain.ErrInvalidCredentials)
	}

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

	event := &dto.UserKafka{
		UserId:    user.ID,
		UpdatedAt: user.UpdatedAt,
		CreatedAt: user.CreatedAt,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	if err := s.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(user.ID.String()),
		Value: data,
	}); err != nil {
		return nil, fmt.Errorf("s.writer.WriteMessages: %w", err)
	}

	s.logger.Debug("successful user login", zap.Any("user_id", user.ID))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
