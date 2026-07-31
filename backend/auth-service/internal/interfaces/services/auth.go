package services

import (
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"BitCoinOffical/forgehost/auth-service/internal/domain/models"
	rabbitqueue "BitCoinOffical/forgehost/auth-service/internal/interfaces/queue/rabbitMQ"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/repo"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/session"
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTTL       = 15 * time.Minute
	RefreshTTL      = (24 * 30) * time.Hour
	VerificationTTL = 5 * time.Minute
)

type AuthService struct {
	logger            *zap.Logger
	tokens            *jwtpkg.ManagerToken
	repo              *repo.AuthRepo
	queue             *rabbitqueue.RabbitQueue
	sessions          *session.Session
	WebgoogleClientID string
}

func NewAuthService(repo *repo.AuthRepo, tokens *jwtpkg.ManagerToken, sessions *session.Session, WebgoogleClientID string, queue *rabbitqueue.RabbitQueue, logger *zap.Logger) *AuthService {
	return &AuthService{repo: repo, tokens: tokens, sessions: sessions, WebgoogleClientID: WebgoogleClientID, queue: queue, logger: logger}
}

func (s *AuthService) LoginUser(ctx context.Context, req *dto.UsersLoginDTO) (*models.Tokens, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetUserByEmail: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("bcrypt.CompareHashAndPassword: %w", domain.ErrInvalidCredentials)
	}

	accessToken, err := s.tokens.GenerateToken(user.ID, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}
	refreshToken, err := s.tokens.GenerateToken(user.ID, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessions.SaveToken(ctx, user.ID, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.session.SaveToken: %w", err)
	}

	s.logger.Debug("successful user login", zap.Any("user_id", user.ID))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RegisterUser(ctx context.Context, req *dto.UsersRegisterDTO) (*models.Tokens, error) {
	newpass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: newpass,
	}

	id, err := s.repo.SaveUser(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("s.repo.SaveUser: %w", err)
	}

	accessToken, err := s.tokens.GenerateToken(id, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}

	refreshToken, err := s.tokens.GenerateToken(id, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessions.SaveToken(ctx, id, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.session.SaveToken: %w", err)
	}

	s.logger.Debug("successful user registration", zap.Any("user_id", id))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) UpdateAccessToken(ctx context.Context, tokens *dto.TokensDTO) (*models.Tokens, error) {
	refreshToken := tokens.RefreshToken

	user, err := s.tokens.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("s.tokens.ValidateToken: %w", err)
	}

	id, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	savedToken, err := s.sessions.GetToken(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.sessions.GetToken: %w", err)
	}

	if savedToken != refreshToken {
		return nil, domain.ErrInvalidCredentials
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	accessToken, err := s.tokens.GenerateToken(userID, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}

	s.logger.Debug("successful token access", zap.Any("user_id", user.ID))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

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
	accessToken, err := s.tokens.GenerateToken(id, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}
	refreshToken, err := s.tokens.GenerateToken(id, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessions.SaveToken(ctx, id, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.session.SaveToken: %w", err)
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

	accessToken, err := s.tokens.GenerateToken(id, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}
	refreshToken, err := s.tokens.GenerateToken(id, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessions.SaveToken(ctx, id, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.session.SaveToken: %w", err)
	}

	s.logger.Debug("successful google callback for android", zap.Any("user_id", user.ID), zap.String("source", "google"), zap.String("client", "android"))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) LogoutUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("uuid.Parse: %w", err)
	}

	if err := s.sessions.DeleteToken(ctx, id); err != nil {
		return fmt.Errorf("s.sessions.DeleteToken: %w", err)
	}
	s.logger.Debug("user exited", zap.Any("user_id", id))
	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailDTO, userID string) error {
	verify := models.VerifyEmail{
		Email: req.Email,
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("uuid.Parse: %w", err)
	}

	code := rand.Intn(900000) + 100000
	if err := s.sessions.SaveVerificationCode(ctx, id, code, VerificationTTL); err != nil {
		return fmt.Errorf("s.sessions.SaveVerificationCode: %w", err)
	}

	codeStr := strconv.Itoa(code)
	body := models.RabbitQueue{
		UserID: userID,
		Code:   codeStr,
		Email:  verify.Email,

		DispatchDate: time.Now().Add(VerificationTTL),
	}

	b, err := json.Marshal(&body)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	qctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.queue.AddQueue(qctx, b); err != nil {
		return fmt.Errorf("s.queue.AddQueue: %w", err)
	}

	s.logger.Debug("code send", zap.Any("user_id", userID))
	return nil
}

func (s *AuthService) ResendVerifyEmail(ctx context.Context, req *dto.VerifyEmailDTO, userID string) error {
	verify := models.VerifyEmail{
		Email: req.Email,
		Code:  req.Code,
	}

	id, err := uuid.Parse(req.Code)
	if err != nil {
		return fmt.Errorf("uuid.Parse: %w", err)
	}

	code, err := s.sessions.GetVerificationCode(ctx, id)
	if err != nil {
		return fmt.Errorf("s.sessions.GetVerificationCode: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(verify.Email), []byte(code)) != 1 {
		return fmt.Errorf("subtle.ConstantTimeCompare: %w", domain.ErrInvalidCode)
	}

	if err := s.repo.UpdateVerifyEmail(ctx, userID, verify.Email); err != nil {
		return fmt.Errorf("s.repo.UpdateVerifyEmail: %w", err)
	}

	s.logger.Debug("update email verified", zap.Any("user_id", userID))
	return nil
}
