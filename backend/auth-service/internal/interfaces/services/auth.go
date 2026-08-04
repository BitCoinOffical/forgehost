package services

import (
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"BitCoinOffical/forgehost/auth-service/internal/domain/models"
	rabbitqueue "BitCoinOffical/forgehost/auth-service/internal/interfaces/queue/rabbitMQ"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/repo"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/store"
	"errors"

	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTTL       = 15 * time.Minute
	RefreshTTL      = (24 * 30) * time.Hour
	VerificationTTL = 7 * time.Minute
	ResetPassTTL    = 7 * time.Minute
	codeQueue       = "code_queue"
	codeSubject     = "Verification Code"
	resetQueue      = "reset_queue"
	resetSubject    = "Password Reset Code"
)

type AuthService struct {
	logger            *zap.Logger
	tokens            *jwtpkg.ManagerToken
	repo              *repo.AuthRepo
	queue             *rabbitqueue.RabbitQueue
	codeStore         *store.CodeStore
	resendStore       *store.ResendStore
	sessionStore      *store.SessionStore
	userStore         *store.UserStore
	WebgoogleClientID string
}

func NewAuthService(
	repo *repo.AuthRepo,
	tokens *jwtpkg.ManagerToken,
	codeStore *store.CodeStore,
	resendStore *store.ResendStore,
	sessionStore *store.SessionStore,
	userStore *store.UserStore,
	WebgoogleClientID string,
	queue *rabbitqueue.RabbitQueue,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		repo:              repo,
		tokens:            tokens,
		codeStore:         codeStore,
		resendStore:       resendStore,
		sessionStore:      sessionStore,
		userStore:         userStore,
		WebgoogleClientID: WebgoogleClientID,
		queue:             queue,
		logger:            logger,
	}
}

func (s *AuthService) LoginUser(ctx context.Context, req *dto.UsersLoginDTO) (*models.Tokens, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetUserByEmail: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("bcrypt.CompareHashAndPassword: %w", domain.ErrInvalidCredentials)
	}

	accessToken, err := s.tokens.GenerateToken(user.ID, user.EmailVerified, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}
	refreshToken, err := s.tokens.GenerateToken(user.ID, user.EmailVerified, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessionStore.SaveToken(ctx, user.ID, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.sessionStore.SaveToken: %w", err)
	}

	s.logger.Debug("successful user login", zap.Any("user_id", user.ID))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RegisterUser(ctx context.Context, req *dto.UsersRegisterDTO) (*dto.PendingKeyDTO, error) {
	_, err := s.resendStore.ResendLimitCheck(ctx, req.Email)
	if err == nil {
		return nil, fmt.Errorf("s.resendStore.ResendLimitCheck: %w", domain.ErrToManyRequest)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("s.resendStore.ResendLimitCheck: %w", err)
	}

	if err := s.resendStore.ResendLimitAdd(ctx, req.Email); err != nil {
		return nil, fmt.Errorf("s.resendStore.SaveVerificationCode: %w", err)
	}

	_, err = s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil || !errors.Is(err, domain.ErrNotFound) {
		if err == nil {
			return nil, fmt.Errorf("s.repo.GetUserByEmail: %w", domain.ErrEmailAlreadyExists)
		}
		return nil, fmt.Errorf("s.repo.GetUserByEmail: %w", err)
	}

	hashpass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	user := models.UserStored{
		Email:        req.Email,
		PasswordHash: hashpass,
	}

	randStr, err := jwtpkg.GenerateRandomString()
	if err != nil {
		return nil, fmt.Errorf("jwtpkg.GenerateRandomString: %w", err)
	}

	if err := s.userStore.SaveUser(ctx, randStr, &user); err != nil {
		return nil, fmt.Errorf("s.userStore.SaveUser: %w", err)
	}

	verify := models.VerifyEmail{
		Email: user.Email,
	}

	code := rand.Intn(900000) + 100000
	if err := s.codeStore.SaveVerificationCode(ctx, randStr, code, VerificationTTL); err != nil {
		return nil, fmt.Errorf("s.codeStore.SaveVerificationCode: %w", err)
	}

	codeStr := strconv.Itoa(code)
	body := models.RabbitQueue{
		Code:  codeStr,
		Email: verify.Email,

		TitleSubject: codeSubject,
		TaskType:     codeQueue,
		DispatchDate: time.Now().Add(VerificationTTL),
	}

	b, err := json.Marshal(&body)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	qctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.queue.AddVerificationCodeQueue(qctx, b); err != nil {
		return nil, fmt.Errorf("s.queue.AddVerificationCodeQueue: %w", err)
	}

	logger.Info("verification code send")
	return &dto.PendingKeyDTO{
		PendingKey: randStr,
	}, nil
}

func (s *AuthService) UpdateAccessToken(ctx context.Context, tokens *dto.TokensDTO) (*models.Tokens, error) {
	refreshToken := tokens.RefreshToken

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

	accessToken, err := s.tokens.GenerateToken(id, user.IsVerified, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}

	s.logger.Debug("successful token access", zap.Any("user_id", id))
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
	accessToken, err := s.tokens.GenerateToken(id, user.EmailVerified, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}
	refreshToken, err := s.tokens.GenerateToken(id, user.EmailVerified, RefreshTTL)
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

	accessToken, err := s.tokens.GenerateToken(id, user.EmailVerified, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}
	refreshToken, err := s.tokens.GenerateToken(id, user.EmailVerified, RefreshTTL)
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

func (s *AuthService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailDTO) (*models.Tokens, error) {
	if req.PendingKey == "" {
		return nil, domain.ErrEmptyValue
	}
	verify := models.VerifyEmail{
		PendingKey: req.PendingKey,
		Code:       req.Code,
	}

	rcode, err := s.codeStore.GetVerificationCode(ctx, verify.PendingKey)
	if err != nil {
		return nil, fmt.Errorf("s.codeStore.GetVerificationCode: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(verify.Code), []byte(rcode)) != 1 {
		return nil, fmt.Errorf("subtle.ConstantTimeCompare: %w", domain.ErrInvalidCode)
	}

	userStore, err := s.userStore.GetUser(ctx, verify.PendingKey)
	if err != nil {
		return nil, fmt.Errorf("s.userStore.GetUser: %w", err)
	}

	user := models.User{
		Email:         userStore.Email,
		PasswordHash:  userStore.PasswordHash,
		EmailVerified: true,
	}

	id, err := s.repo.SaveUser(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("s.repo.SaveUser: %w", err)
	}

	accessToken, err := s.tokens.GenerateToken(id, user.EmailVerified, AccessTTL)
	if err != nil {
		return nil, fmt.Errorf("accessToken s.tokens.GenerateToken: %w", err)
	}
	refreshToken, err := s.tokens.GenerateToken(id, user.EmailVerified, RefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("refreshToken s.tokens.GenerateToken: %w", err)
	}

	if err := s.sessionStore.SaveToken(ctx, id, refreshToken, RefreshTTL); err != nil {
		return nil, fmt.Errorf("s.sessionStore.SaveToken: %w", err)
	}

	s.logger.Debug("successful user verify email", zap.Any("user_id", id))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) ResendVerifyEmail(ctx context.Context, req *dto.VerifyEmailDTO) error {
	verify := models.VerifyEmail{
		PendingKey: req.PendingKey,
		Email:      req.Email,
	}

	_, err := s.resendStore.ResendLimitCheck(ctx, verify.Email)
	if err == nil {
		return fmt.Errorf("s.resendStore.ResendLimitCheck: %w", domain.ErrToManyRequest)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("s.resendStore.ResendLimitCheck: %w", err)
	}

	if err := s.resendStore.ResendLimitAdd(ctx, verify.Email); err != nil {
		return fmt.Errorf("s.resendStore.SaveVerificationCode: %w", err)
	}

	code := rand.Intn(900000) + 100000
	if err := s.codeStore.SaveVerificationCode(ctx, verify.PendingKey, code, VerificationTTL); err != nil {
		return fmt.Errorf("s.codeStore.SaveVerificationCode: %w", err)
	}

	codeStr := strconv.Itoa(code)
	body := models.RabbitQueue{
		Code:  codeStr,
		Email: verify.Email,

		TitleSubject: codeSubject,
		TaskType:     codeQueue,
		DispatchDate: time.Now().Add(VerificationTTL),
	}

	b, err := json.Marshal(&body)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	qctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.queue.AddVerificationCodeQueue(qctx, b); err != nil {
		return fmt.Errorf("s.queue.AddVerificationCodeQueue: %w", err)
	}

	s.logger.Debug("verification code resend")
	return nil
}

func (s *AuthService) UpdatePassword(ctx context.Context, req *dto.UserPasswordDTO, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("uuid.Parse: %w", err)
	}

	newPass, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("s.repo.GetUserByID: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("bcrypt.CompareHashAndPassword: %w", domain.ErrInvalidCredentials)
	}

	userPass := models.User{
		ID:           id,
		Email:        user.Email,
		PasswordHash: newPass,
	}

	if err := s.repo.UpdateUserPassword(ctx, &userPass); err != nil {
		return fmt.Errorf("s.repo.UpdateUserPassword: %w", err)
	}

	s.logger.Debug("update password", zap.Any("user_id", id))
	return nil
}

func (s *AuthService) PasswordReset(ctx context.Context, req *dto.PasswordResetDTO) (*dto.PendingKeyDTO, error) {
	_, err := s.resendStore.ResendLimitCheck(ctx, req.Email)
	if err == nil {
		return nil, fmt.Errorf("s.resendStore.ResendLimitCheck: %w", domain.ErrToManyRequest)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("s.resendStore.ResendLimitCheck: %w", err)
	}

	if err := s.resendStore.ResendLimitAdd(ctx, req.Email); err != nil {
		return nil, fmt.Errorf("s.resendStore.SaveVerificationCode: %w", err)
	}

	pendingKey, err := jwtpkg.GenerateRandomString()
	if err != nil {
		return nil, fmt.Errorf("jwtpkg.GenerateRandomString: %w", err)
	}

	code := rand.Intn(900000) + 100000
	if err := s.codeStore.SaveResetPasswordCode(ctx, pendingKey, code, ResetPassTTL); err != nil {
		return nil, fmt.Errorf("s.codeStore.SaveVerificationCode: %w", err)
	}

	codeStr := strconv.Itoa(code)
	queue := models.RabbitQueue{
		Email: req.Email,
		Code:  codeStr,

		TitleSubject: resetSubject,
		TaskType:     resetQueue,
		DispatchDate: time.Now().Add(ResetPassTTL),
	}

	body, err := json.Marshal(&queue)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	if err := s.queue.AddResetCodeQueue(ctx, body); err != nil {
		return nil, fmt.Errorf("s.queue.AddResetCodeQueue: %w", err)
	}

	return &dto.PendingKeyDTO{
		PendingKey: pendingKey,
	}, nil
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, req *dto.PasswordResetDTO) error {
	pass, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	rcode, err := s.codeStore.GetResetPasswordCode(ctx, req.PendingKey)
	if err != nil {
		return fmt.Errorf("s.codeStore.GetResetPasswordCode: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(req.Code), []byte(rcode)) != 1 {
		return fmt.Errorf("subtle.ConstantTimeCompare: %w", domain.ErrInvalidCode)
	}

	usr, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("s.repo.GetUserByEmail: %w", err)
	}

	user := models.User{
		ID:           usr.ID,
		Email:        req.Email,
		PasswordHash: pass,
	}

	if err := s.repo.UpdateUserPassword(ctx, &user); err != nil {
		return fmt.Errorf("s.repo.UpdateUserPassword: %w", err)
	}

	s.logger.Debug("reset password", zap.Any("user_id", user.ID))
	return nil
}

func (s *AuthService) PasswordResetResend(ctx context.Context, req *dto.PasswordResetDTO) error {
	_, err := s.resendStore.ResendLimitCheck(ctx, req.Email)
	if err == nil {
		return fmt.Errorf("s.resendStore.ResendLimitCheck: %w", domain.ErrToManyRequest)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("s.resendStore.ResendLimitCheck: %w", err)
	}

	if err := s.resendStore.ResendLimitAdd(ctx, req.Email); err != nil {
		return fmt.Errorf("s.resendStore.SaveVerificationCode: %w", err)
	}

	code := rand.Intn(900000) + 100000
	if err := s.codeStore.SaveResetPasswordCode(ctx, req.PendingKey, code, ResetPassTTL); err != nil {
		return fmt.Errorf("s.codeStore.SaveVerificationCode: %w", err)
	}

	codeStr := strconv.Itoa(code)
	queue := models.RabbitQueue{
		Email: req.Email,
		Code:  codeStr,

		TitleSubject: resetSubject,
		TaskType:     resetQueue,
		DispatchDate: time.Now().Add(ResetPassTTL),
	}

	body, err := json.Marshal(&queue)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	if err := s.queue.AddResetCodeQueue(ctx, body); err != nil {
		return fmt.Errorf("s.queue.AddResetCodeQueue: %w", err)
	}

	s.logger.Debug("reset password code resend")
	return nil
}
