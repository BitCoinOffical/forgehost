package services

import (
	"context"
	v2 "encoding/json/v2"
	"fmt"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type GoogleAuthService struct {
	repo         AuthRepository
	codeStore    CodeStore
	client       *kgo.Client
	tokens       ManagerToken
	sessionStore SessionStore
	logger       *zap.Logger

	WebgoogleClientID string
}

func NewGoogleAuthService(
	repo AuthRepository,
	codeStore CodeStore,
	client *kgo.Client,
	tokens ManagerToken,
	sessionStore SessionStore,
	logger *zap.Logger,
	WebgoogleClientID string,
) *GoogleAuthService {
	return &GoogleAuthService{
		repo:              repo,
		codeStore:         codeStore,
		client:            client,
		tokens:            tokens,
		sessionStore:      sessionStore,
		logger:            logger,
		WebgoogleClientID: WebgoogleClientID,
	}
}

func (s *GoogleAuthService) GoogleCallback(ctx context.Context, req *dto.GoogleUserDTO) (string, error) {
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
		return "", fmt.Errorf("s.repo.SaveGoogleUser: %w", err)
	}

	event := dto.UserRegisteredEvent{
		UserID: id.String(),
	}

	data, err := v2.Marshal(&event)
	if err != nil {
		return "", fmt.Errorf("v2.Marshal")
	}

	record := &kgo.Record{
		Topic: topic,
		Value: data,
	}
	result := s.client.ProduceSync(ctx, record)
	if err := result.FirstErr(); err != nil {
		return "", fmt.Errorf("s.client.ProduceSync: %w", err)
	}
	s.logger.Debug("write data in kafka", zap.String("topic", topic), zap.String("user_id", id.String()))

	oauthCode := uuid.New()

	if err := s.codeStore.SaveOauthCode(ctx, oauthCode.String(), id.String()); err != nil {
		return "", fmt.Errorf("s.codeStore.SaveOauthCode: %w", err)
	}

	s.logger.Debug("successful google callback", zap.String("user_id", id.String()), zap.String("source", "google"), zap.String("client", "web"))

	return oauthCode.String(), nil
}

// android
func (s *GoogleAuthService) GoogleLoginAndroid(ctx context.Context, req dto.GoogleAndroidUserDTO) (*models.Tokens, error) {
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

	s.logger.Debug("successful google callback for android", zap.String("user_id", id.String()), zap.String("source", "google"), zap.String("client", "android"))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *GoogleAuthService) Exchange(ctx context.Context, req *dto.ExchangeRequestDTO) (*models.Tokens, error) {
	userId, err := s.codeStore.GetOauthCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("s.codeStore.GetOauthCode: %w", err)
	}

	if err := s.codeStore.DeleteOauthCode(ctx, req.Code); err != nil {
		return nil, fmt.Errorf("s.codeStore.DeleteOauthCode: %w", err)
	}

	parseId, err := uuid.Parse(userId)
	if err != nil {
		return nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, parseId)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetUserByID: %w", err)
	}

	accessToken, err := s.tokens.GenerateToken(parseId, role, user.EmailVerified, user.EmailBanned, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}

	refreshToken, err := s.tokens.GenerateToken(parseId, role, user.EmailVerified, user.EmailBanned, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessionStore.SaveToken(ctx, parseId, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.sessionStore.SaveToken: %w", err)
	}

	s.logger.Debug("successful exchange", zap.String("user_id", parseId.String()), zap.String("client", "web"))

	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
