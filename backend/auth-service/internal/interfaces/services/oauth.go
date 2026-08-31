package services

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"go.uber.org/zap"
)

func (s *AuthService) GoogleCallback(ctx context.Context, req *dto.GoogleUserDTO) (*models.Tokens, error) {
	user := &models.User{
		Name:          &req.Name,
		Email:         req.Email,
		Picture:       &req.Picture,
		EmailVerified: req.EmailVerified,
	}
	oauth := &models.OAuthAccount{
		Provider:       "google",
		ProviderUserID: req.Sub,
		GivenName:      req.GivenName,
		FamilyName:     req.FamilyName,
	}
	id, err := s.repo.SaveGoogleUser(ctx, user, oauth)
	if err != nil {
		return nil, fmt.Errorf("s.repo.SaveGoogleUser: %w", err)
	}
	accessToken, err := s.tokens.GenerateToken(id, role, user.EmailVerified, user.EmailBanned, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}
	refreshToken, err := s.tokens.GenerateToken(id, role, user.EmailVerified, user.EmailBanned, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessionStore.SaveToken(ctx, id, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.sessionStore.SaveToken: %w", err)
	}

	s.logger.Debug("successful google callback", zap.Any("user_id", user.ID), zap.String("source", "google"), zap.String("client", "web"))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// android
func (s *AuthService) GoogleLoginAndroid(ctx context.Context, req dto.GoogleAndroidUserDTO) (*models.Tokens, error) {
	payload, err := idtoken.Validate(ctx, req.IdToken, s.WebgoogleClientID) //android
	if err != nil {
		return nil, fmt.Errorf("idtoken.Validate: %w error: %v", domain.ErrInvalidGoogleToken, err)
	}

	sub, ok := payload.Claims["sub"].(string)
	if !ok || sub == "" {
		return nil, fmt.Errorf("missing sub claim")
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("missing email claim")
	}

	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	givenName, _ := payload.Claims["given_name"].(string)
	familyName, _ := payload.Claims["family_name"].(string)
	user := &models.User{
		Name:          &name,
		Email:         email,
		Picture:       &picture,
		EmailVerified: true,
	}
	oauth := &models.OAuthAccount{
		Provider:       "google",
		ProviderUserID: sub,
		GivenName:      givenName,
		FamilyName:     familyName,
	}

	id, err := s.repo.SaveGoogleUser(ctx, user, oauth)
	if err != nil {
		return nil, fmt.Errorf("s.repo.SaveGoogleUser: %w", err)
	}

	accessToken, err := s.tokens.GenerateToken(id, role, user.EmailVerified, user.EmailBanned, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}
	refreshToken, err := s.tokens.GenerateToken(id, role, user.EmailVerified, user.EmailBanned, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessionStore.SaveToken(ctx, id, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.sessionStore.SaveToken: %w", err)
	}

	s.logger.Debug("successful google callback for android", zap.Any("user_id", user.ID), zap.String("source", "google"), zap.String("client", "android"))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
