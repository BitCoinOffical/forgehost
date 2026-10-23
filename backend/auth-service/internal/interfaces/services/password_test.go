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
	"golang.org/x/crypto/bcrypt"
)

func TestUpdatePassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repo_mocks.NewMockAuthRepository(ctrl)

	password := NewPasswordService(repo, nil, nil, nil, zaptest.NewLogger(t))

	oldPasswordHash, err := bcrypt.GenerateFromPassword([]byte("old_password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	testCases := []struct {
		caseName string
		data     struct {
			userID      string
			oldPassword string
			newPassword string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "update password",
			data: struct {
				userID      string
				oldPassword string
				newPassword string
			}{userID: uuid.NewString(), oldPassword: "old_password", newPassword: "new_password"},
			prepare: func() {
				repo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).
					Return(&models.User{PasswordHash: oldPasswordHash}, nil).Times(1)
				repo.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "invalid user id",
			data: struct {
				userID      string
				oldPassword string
				newPassword string
			}{userID: "invalid_uuid", oldPassword: "old_password", newPassword: "new_password"},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "new password too long",
			data: struct {
				userID      string
				oldPassword string
				newPassword string
			}{userID: uuid.NewString(), oldPassword: "old_password", newPassword: strings.Repeat("a", 73)},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "get user by id error",
			data: struct {
				userID      string
				oldPassword string
				newPassword string
			}{userID: uuid.NewString(), oldPassword: "old_password", newPassword: "new_password"},
			prepare: func() {
				repo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("get user by id error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "wrong old password",
			data: struct {
				userID      string
				oldPassword string
				newPassword string
			}{userID: uuid.NewString(), oldPassword: "wrong_password", newPassword: "new_password"},
			prepare: func() {
				repo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).
					Return(&models.User{PasswordHash: oldPasswordHash}, nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "update user password error",
			data: struct {
				userID      string
				oldPassword string
				newPassword string
			}{userID: uuid.NewString(), oldPassword: "old_password", newPassword: "new_password"},
			prepare: func() {
				repo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).
					Return(&models.User{PasswordHash: oldPasswordHash}, nil).Times(1)
				repo.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any()).
					Return(errors.New("update user password error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.UserPasswordDTO{
				OldPassword: tc.data.oldPassword,
				NewPassword: tc.data.newPassword,
			}
			err := password.UpdatePassword(t.Context(), &req, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPasswordReset(t *testing.T) {
	ctrl := gomock.NewController(t)
	codeStore := repo_mocks.NewMockCodeStore(ctrl)
	resendStore := repo_mocks.NewMockResendStore(ctrl)
	queue := repo_mocks.NewMockRabbitQueue(ctrl)

	password := NewPasswordService(nil, codeStore, resendStore, queue, zaptest.NewLogger(t))

	testCases := []struct {
		caseName string
		data     struct {
			email string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "password reset",
			data: struct {
				email string
			}{email: "user@gmail.com"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveResetPasswordCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				queue.EXPECT().AddEmailTaskQueue(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "too many requests",
			data: struct {
				email string
			}{email: "user@gmail.com"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "resend limit check error",
			data: struct {
				email string
			}{email: "user@gmail.com"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", errors.New("resend limit check error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "resend limit add error",
			data: struct {
				email string
			}{email: "user@gmail.com"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(errors.New("resend limit add error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "save reset password code error",
			data: struct {
				email string
			}{email: "user@gmail.com"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveResetPasswordCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("save reset password code error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "add email task queue error",
			data: struct {
				email string
			}{email: "user@gmail.com"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveResetPasswordCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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

			req := dto.PasswordResetDTO{
				Email: tc.data.email,
			}
			pendingKey, err := password.PasswordReset(t.Context(), &req)
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

func TestConfirmPasswordReset(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repo_mocks.NewMockAuthRepository(ctrl)
	codeStore := repo_mocks.NewMockCodeStore(ctrl)

	password := NewPasswordService(repo, codeStore, nil, nil, zaptest.NewLogger(t))

	testCases := []struct {
		caseName string
		data     struct {
			email       string
			newPassword string
			pendingKey  string
			code        string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "confirm password reset",
			data: struct {
				email       string
				newPassword string
				pendingKey  string
				code        string
			}{email: "user@gmail.com", newPassword: "new_password", pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetResetPasswordCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
				repo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(&models.User{ID: uuid.New()}, nil).Times(1)
				repo.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "new password too long",
			data: struct {
				email       string
				newPassword string
				pendingKey  string
				code        string
			}{email: "user@gmail.com", newPassword: strings.Repeat("a", 73), pendingKey: "pending_key", code: "123456"},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "get reset password code error",
			data: struct {
				email       string
				newPassword string
				pendingKey  string
				code        string
			}{email: "user@gmail.com", newPassword: "new_password", pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetResetPasswordCode(gomock.Any(), gomock.Any()).
					Return("", errors.New("get reset password code error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "invalid code",
			data: struct {
				email       string
				newPassword string
				pendingKey  string
				code        string
			}{email: "user@gmail.com", newPassword: "new_password", pendingKey: "pending_key", code: "000000"},
			prepare: func() {
				codeStore.EXPECT().GetResetPasswordCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "get user by email error",
			data: struct {
				email       string
				newPassword string
				pendingKey  string
				code        string
			}{email: "user@gmail.com", newPassword: "new_password", pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetResetPasswordCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
				repo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("get user by email error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "update user password error",
			data: struct {
				email       string
				newPassword string
				pendingKey  string
				code        string
			}{email: "user@gmail.com", newPassword: "new_password", pendingKey: "pending_key", code: "123456"},
			prepare: func() {
				codeStore.EXPECT().GetResetPasswordCode(gomock.Any(), gomock.Any()).
					Return("123456", nil).Times(1)
				repo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(&models.User{ID: uuid.New()}, nil).Times(1)
				repo.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any()).
					Return(errors.New("update user password error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.PasswordResetDTO{
				Email:       tc.data.email,
				NewPassword: tc.data.newPassword,
				PendingKey:  tc.data.pendingKey,
				Code:        tc.data.code,
			}
			err := password.ConfirmPasswordReset(t.Context(), &req)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPasswordResetResend(t *testing.T) {
	ctrl := gomock.NewController(t)
	codeStore := repo_mocks.NewMockCodeStore(ctrl)
	resendStore := repo_mocks.NewMockResendStore(ctrl)
	queue := repo_mocks.NewMockRabbitQueue(ctrl)

	password := NewPasswordService(nil, codeStore, resendStore, queue, zaptest.NewLogger(t))

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
			caseName: "password reset resend",
			data: struct {
				email      string
				pendingKey string
			}{email: "user@gmail.com", pendingKey: "pending_key"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveResetPasswordCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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
			caseName: "save reset password code error",
			data: struct {
				email      string
				pendingKey string
			}{email: "user@gmail.com", pendingKey: "pending_key"},
			prepare: func() {
				resendStore.EXPECT().ResendLimitCheck(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotFound).Times(1)
				resendStore.EXPECT().ResendLimitAdd(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				codeStore.EXPECT().SaveResetPasswordCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("save reset password code error")).Times(1)
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
				codeStore.EXPECT().SaveResetPasswordCode(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
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

			req := dto.PasswordResetDTO{
				Email:      tc.data.email,
				PendingKey: tc.data.pendingKey,
			}
			err := password.PasswordResetResend(t.Context(), &req)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
