package services

import (
	"context"
	"crypto/subtle"
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
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

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
		DispatchDate: time.Now().Add(ResetPassTTL),
	}

	body, err := json.Marshal(&queue)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	if err := s.queue.AddEmailTaskQueue(ctx, body); err != nil {
		return nil, fmt.Errorf("s.queue.AddEmailTaskQueue: %w", err)
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
		DispatchDate: time.Now().Add(ResetPassTTL),
	}

	body, err := json.Marshal(&queue)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	if err := s.queue.AddEmailTaskQueue(ctx, body); err != nil {
		return fmt.Errorf("s.queue.AddEmailTaskQueue: %w", err)
	}

	s.logger.Debug("reset password code resend")
	return nil
}
