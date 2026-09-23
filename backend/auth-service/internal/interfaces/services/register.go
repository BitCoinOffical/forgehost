package services

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	v2 "encoding/json/v2"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/BitCoinOffical/forgehost/auth-service/pkg/mask"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type RegisterService struct {
	repo         AuthRepository
	tokens       ManagerToken
	userStore    UserStore
	codeStore    CodeStore
	resendStore  ResendStore
	sessionStore SessionStore
	queue        RabbitQueue
	logger       *zap.Logger
	client       *kgo.Client
}

func NewRegisterService(
	repo AuthRepository,
	tokens ManagerToken,
	userStore UserStore,
	codeStore CodeStore,
	resendStore ResendStore,
	sessionStore SessionStore,
	queue RabbitQueue,
	logger *zap.Logger,
	client *kgo.Client,
) *RegisterService {
	return &RegisterService{
		repo:         repo,
		tokens:       tokens,
		userStore:    userStore,
		codeStore:    codeStore,
		resendStore:  resendStore,
		sessionStore: sessionStore,
		queue:        queue,
		logger:       logger,
		client:       client,
	}
}

func (s *RegisterService) RegisterUser(ctx context.Context, req *dto.UsersRegisterDTO) (*dto.PendingKeyDTO, error) {
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

	code, err := jwtpkg.GenerateSecureCode()
	if err != nil {
		return nil, fmt.Errorf("jwtpkg.GenerateSecureCode: %w", err)
	}

	if err := s.codeStore.SaveVerificationCode(ctx, randStr, code, VerificationTTL); err != nil {
		return nil, fmt.Errorf("s.codeStore.SaveVerificationCode: %w", err)
	}

	codeStr := strconv.Itoa(code)
	body := models.RabbitQueue{
		Code:  codeStr,
		Email: verify.Email,

		TitleSubject: codeSubject,
		DispatchDate: time.Now().Add(VerificationTTL),
	}

	b, err := json.Marshal(&body)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	qctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.queue.AddEmailTaskQueue(qctx, b); err != nil {
		return nil, fmt.Errorf("s.queue.AddEmailTaskQueue: %w", err)
	}

	s.logger.Debug("verification code sent", zap.String("email", mask.Email(req.Email)))
	return &dto.PendingKeyDTO{
		PendingKey: randStr,
	}, nil
}

func (s *RegisterService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailDTO) (*models.Tokens, error) {
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

	event := dto.UserRegisteredEvent{
		UserID: id.String(),
	}

	data, err := v2.Marshal(&event)
	if err != nil {
		return nil, fmt.Errorf("v2.Marshal")
	}

	record := &kgo.Record{
		Topic: topic,
		Value: data,
	}
	result := s.client.ProduceSync(ctx, record)
	if err := result.FirstErr(); err != nil {
		return nil, fmt.Errorf("s.client.ProduceSync: %w", err)
	}

	s.logger.Debug("write data in kafka", zap.String("topic", topic), zap.String("user_id", id.String()))

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

	s.logger.Debug("successful user verify email", zap.String("user_id", id.String()))
	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *RegisterService) ResendVerifyEmail(ctx context.Context, req *dto.VerifyEmailDTO) error {
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

	code, err := jwtpkg.GenerateSecureCode()
	if err != nil {
		return fmt.Errorf("jwtpkg.GenerateSecureCode: %w", err)
	}

	if err := s.codeStore.SaveVerificationCode(ctx, verify.PendingKey, code, VerificationTTL); err != nil {
		return fmt.Errorf("s.codeStore.SaveVerificationCode: %w", err)
	}

	codeStr := strconv.Itoa(code)
	body := models.RabbitQueue{
		Code:  codeStr,
		Email: verify.Email,

		TitleSubject: codeSubject,
		DispatchDate: time.Now().Add(VerificationTTL),
	}

	b, err := json.Marshal(&body)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	qctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.queue.AddEmailTaskQueue(qctx, b); err != nil {
		return fmt.Errorf("s.queue.AddEmailTaskQueue: %w", err)
	}

	s.logger.Debug("verification code resend", zap.String("email", mask.Email(verify.Email)))
	return nil
}
