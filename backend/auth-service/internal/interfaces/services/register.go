package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/bytedance/gopkg/util/logger"
	"golang.org/x/crypto/bcrypt"
)

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

	logger.Info("verification code send")
	return &dto.PendingKeyDTO{
		PendingKey: randStr,
	}, nil
}
