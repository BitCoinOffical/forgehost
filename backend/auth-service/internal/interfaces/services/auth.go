package services

import (
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/repo"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/session"
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTTL  = 15 * time.Minute
	RefreshTTL = (24 * 30) * time.Hour
)

type AuthService struct {
	tokens   *jwtpkg.ManagerToken
	repo     *repo.AuthRepo
	sessions *session.Session
}

func NewAuthService(repo *repo.AuthRepo, tokens *jwtpkg.ManagerToken, sessions *session.Session) *AuthService {
	return &AuthService{repo: repo, tokens: tokens, sessions: sessions}
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

	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) UpdateAccessToken(ctx context.Context, tokens dto.TokensDTO) (*models.Tokens, error) {
	refreshToken := tokens.RefreshToken

	user, err := s.tokens.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("s.tokens.ValidateToken: %w", err)
	}

	savedToken, err := s.sessions.GetToken(ctx, user.ID)
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

	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) GoogleCallback(ctx context.Context, req *dto.GoogleUserDTO) (*models.Tokens, error) {
	user := &models.User{
		Name:    req.Name,
		Email:   req.Email,
		Picture: req.Picture,
	}
	oauth := &models.OAuthAccount{
		Provider:       "google",
		ProviderUserID: req.Sub,
		GivenName:      req.GivenName,
		FamilyName:     req.FamilyName,
		EmailVerified:  req.EmailVerified,
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

	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) LogoutUser(ctx context.Context, id string) error {
	if err := s.sessions.DeleteToken(ctx, id); err != nil {
		return fmt.Errorf("s.sessions.DeleteToken: %w", err)
	}
	return nil
}
