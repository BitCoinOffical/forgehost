package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	repo_mocks "github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/services/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestRegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repo_mocks.NewMockAuthRepository(ctrl)
	userStore := repo_mocks.NewMockUserStore(ctrl)
	codeStore := repo_mocks.NewMockCodeStore(ctrl)
	resendStore := repo_mocks.NewMockResendStore(ctrl)
	queue := repo_mocks.NewMockRabbitQueue(ctrl)

	register := NewRegisterService(repo, nil, userStore, codeStore, resendStore, nil, queue, zaptest.NewLogger(t), nil)

	testCases := []struct {
		caseName string
		data     struct {
			email    string
			password string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "register user",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().CheckExistsUserByEmail(gomock.Any(), gomock.Any()).
					Return(false, nil).Times(1)
				userStore.EXPECT().SaveUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveVerificationCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				queue.EXPECT().AddEmailTaskQueue(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "too many requests",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "resend limit check error",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", errors.New("resend limit check error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "resend limit add error",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(errors.New("resend limit add error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "check exists user error",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().CheckExistsUserByEmail(gomock.Any(), gomock.Any()).
					Return(false, errors.New("check exists user error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "user already exists",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().CheckExistsUserByEmail(gomock.Any(), gomock.Any()).
					Return(true, nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "password too long",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: strings.Repeat("a", 73)},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().CheckExistsUserByEmail(gomock.Any(), gomock.Any()).
					Return(false, nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "save user error",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().CheckExistsUserByEmail(gomock.Any(), gomock.Any()).
					Return(false, nil).Times(1)
				userStore.EXPECT().SaveUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("save user error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "save verification code error",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().CheckExistsUserByEmail(gomock.Any(), gomock.Any()).
					Return(false, nil).Times(1)
				userStore.EXPECT().SaveUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveVerificationCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("save verification code error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "add email task queue error",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().CheckExistsUserByEmail(gomock.Any(), gomock.Any()).
					Return(false, nil).Times(1)
				userStore.EXPECT().SaveUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveVerificationCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				queue.EXPECT().AddEmailTaskQueue(gomock.Any(), gomock.Any()).
					Return(errors.New("add email task queue error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.UsersRegisterDTO{
				Email:    tc.data.email,
				Password: tc.data.password,
			}
			pendingKey, err := register.RegisterUser(t.Context(), &req)
			if tc.success {
				require.NoError(t, err)
				require.NotNil(t, pendingKey)
				assert.NotEmpty(t, pendingKey.PendingKey)
			} else {
				require.Error(t, err)
				assert.Nil(t, pendingKey)
			}
		})
	}
}

func TestVerifyEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repo_mocks.NewMockAuthRepository(ctrl)
	manager := repo_mocks.NewMockManagerToken(ctrl)
	userStore := repo_mocks.NewMockUserStore(ctrl)
	codeStore := repo_mocks.NewMockCodeStore(ctrl)
	sessionStore := repo_mocks.NewMockSessionStore(ctrl)
	client := setupTestKafka(t)

	register := NewRegisterService(repo, manager, userStore, codeStore, nil, sessionStore, nil, zaptest.NewLogger(t), client)

	testCases := []struct {
		caseName string
		data     struct {
			pendingKey string
			code       string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "verify email",
			data: struct {
				pendingKey string
				code       string
			}{pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetVerificationCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
				userStore.EXPECT().GetUser(gomock.Any(), gomock.Any()).
					Return(&models.UserStored{Email: "user@gmail.com"}, nil).Times(1)
				repo.EXPECT().SaveUser(gomock.Any(), gomock.Any()).
					Return(uuid.New(), nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("token", nil).Times(2)
				sessionStore.EXPECT().SaveToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "empty pending key",
			data: struct {
				pendingKey string
				code       string
			}{pendingKey: "", code: "123456"},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "get verification code error",
			data: struct {
				pendingKey string
				code       string
			}{pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetVerificationCode(gomock.Any(), gomock.Any()).
					Return("", errors.New("get verification code error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "invalid code",
			data: struct {
				pendingKey string
				code       string
			}{pendingKey: "pending_key", code: "000000"},
			prepare: func() {
				codeStore.EXPECT().GetVerificationCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "get user error",
			data: struct {
				pendingKey string
				code       string
			}{pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetVerificationCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
				userStore.EXPECT().GetUser(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("get user error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "save user error",
			data: struct {
				pendingKey string
				code       string
			}{pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetVerificationCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
				userStore.EXPECT().GetUser(gomock.Any(), gomock.Any()).
					Return(&models.UserStored{Email: "user@gmail.com"}, nil).Times(1)
				repo.EXPECT().SaveUser(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, errors.New("save user error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "generate access token error",
			data: struct {
				pendingKey string
				code       string
			}{pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetVerificationCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
				userStore.EXPECT().GetUser(gomock.Any(), gomock.Any()).
					Return(&models.UserStored{Email: "user@gmail.com"}, nil).Times(1)
				repo.EXPECT().SaveUser(gomock.Any(), gomock.Any()).
					Return(uuid.New(), nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("generate access token error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "generate refresh token error",
			data: struct {
				pendingKey string
				code       string
			}{pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetVerificationCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
				userStore.EXPECT().GetUser(gomock.Any(), gomock.Any()).
					Return(&models.UserStored{Email: "user@gmail.com"}, nil).Times(1)
				repo.EXPECT().SaveUser(gomock.Any(), gomock.Any()).
					Return(uuid.New(), nil).Times(1)
				first := manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("token", nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("generate refresh token error")).Times(1).After(first)
			},
			success: false,
		},
		{
			caseName: "save token error",
			data: struct {
				pendingKey string
				code       string
			}{pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetVerificationCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
				userStore.EXPECT().GetUser(gomock.Any(), gomock.Any()).
					Return(&models.UserStored{Email: "user@gmail.com"}, nil).Times(1)
				repo.EXPECT().SaveUser(gomock.Any(), gomock.Any()).
					Return(uuid.New(), nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("token", nil).Times(2)
				sessionStore.EXPECT().SaveToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("save token error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.VerifyEmailDTO{
				PendingKey: tc.data.pendingKey,
				Code:       tc.data.code,
			}
			tokens, err := register.VerifyEmail(t.Context(), &req)
			if tc.success {
				require.NoError(t, err)
				assert.NotEmpty(t, tokens)
			} else {
				require.Error(t, err)
				assert.Nil(t, tokens)
			}
		})
	}
}

func TestResendVerifyEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	codeStore := repo_mocks.NewMockCodeStore(ctrl)
	resendStore := repo_mocks.NewMockResendStore(ctrl)
	queue := repo_mocks.NewMockRabbitQueue(ctrl)

	register := NewRegisterService(nil, nil, nil, codeStore, resendStore, nil, queue, zaptest.NewLogger(t), nil)

	testCases := []struct {
		caseName string
		data     struct {
			email      string
			pendingKey string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "resend verify email",
			data: struct {
				email      string
				pendingKey string
			}{email: "user@gmail.com", pendingKey: "pending_key"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveVerificationCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				queue.EXPECT().AddEmailTaskQueue(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "too many requests",
			data: struct {
				email      string
				pendingKey string
			}{email: "user@gmail.com", pendingKey: "pending_key"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "resend limit check error",
			data: struct {
				email      string
				pendingKey string
			}{email: "user@gmail.com", pendingKey: "pending_key"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", errors.New("resend limit check error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "resend limit add error",
			data: struct {
				email      string
				pendingKey string
			}{email: "user@gmail.com", pendingKey: "pending_key"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(errors.New("resend limit add error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "save verification code error",
			data: struct {
				email      string
				pendingKey string
			}{email: "user@gmail.com", pendingKey: "pending_key"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveVerificationCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("save verification code error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "add email task queue error",
			data: struct {
				email      string
				pendingKey string
			}{email: "user@gmail.com", pendingKey: "pending_key"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveVerificationCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				queue.EXPECT().AddEmailTaskQueue(gomock.Any(), gomock.Any()).
					Return(errors.New("add email task queue error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.VerifyEmailDTO{
				Email:      tc.data.email,
				PendingKey: tc.data.pendingKey,
			}
			err := register.ResendVerifyEmail(t.Context(), &req)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
