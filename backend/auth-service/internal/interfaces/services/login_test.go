package services

import (
	"errors"
	"testing"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	repo_mocks "github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/services/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repo_mocks.NewMockAuthRepository(ctrl)
	client := setupTestKafka(t)
	manager := repo_mocks.NewMockManagerToken(ctrl)
	sessionStore := repo_mocks.NewMockSessionStore(ctrl)

	login := NewLoginService(repo, client, manager, sessionStore, zaptest.NewLogger(t))

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)

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
			caseName: "login user",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				repo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(&models.User{PasswordHash: passwordHash}, nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("token", nil).Times(2)
				sessionStore.EXPECT().SaveToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "wrong password",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "wrong_password"},
			prepare: func() {
				repo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(&models.User{PasswordHash: passwordHash}, nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "user not found",
			data: struct {
				email    string
				password string
			}{email: "unknown@gmail.com", password: "password"},
			prepare: func() {
				repo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("user not found")).Times(1)
			},
			success: false,
		},
		{
			caseName: "generate token error",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				repo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(&models.User{PasswordHash: passwordHash}, nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("generate token error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "save token error",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				repo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(&models.User{PasswordHash: passwordHash}, nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("token", nil).Times(2)
				sessionStore.EXPECT().SaveToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("save token error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "refresh token error",
			data: struct {
				email    string
				password string
			}{email: "user@gmail.com", password: "password"},
			prepare: func() {
				repo.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Return(&models.User{PasswordHash: passwordHash}, nil).Times(1)
				first := manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("token", nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("generate refresh token error")).Times(1).After(first)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.UsersLoginDTO{
				Email:    tc.data.email,
				Password: tc.data.password,
			}
			tokens, err := login.LoginUser(t.Context(), &req)
			if tc.success {
				require.NoError(t, err)
				assert.NotEmpty(t, tokens)
			} else {
				require.Error(t, err)
			}
		})
	}
}
