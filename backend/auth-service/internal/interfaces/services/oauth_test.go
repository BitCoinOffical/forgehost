package services

import (
	"errors"
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

func TestGoogleCallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repo_mocks.NewMockAuthRepository(ctrl)
	codeStore := repo_mocks.NewMockCodeStore(ctrl)
	client := setupTestKafka(t)

	google := NewGoogleAuthService(repo, codeStore, client, nil, nil, zaptest.NewLogger(t), "client_id")

	testCases := []struct {
		caseName string
		data     struct {
			name  string
			email string
			sub   string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "google callback",
			data: struct {
				name  string
				email string
				sub   string
			}{name: "user", email: "user@gmail.com", sub: "google_sub"},
			prepare: func() {
				repo.EXPECT().SaveGoogleUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uuid.New(), nil).Times(1)
				codeStore.EXPECT().SaveOauthCode(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "save google user error",
			data: struct {
				name  string
				email string
				sub   string
			}{name: "user", email: "user@gmail.com", sub: "google_sub"},
			prepare: func() {
				repo.EXPECT().SaveGoogleUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uuid.Nil, errors.New("save google user error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "save oauth code error",
			data: struct {
				name  string
				email string
				sub   string
			}{name: "user", email: "user@gmail.com", sub: "google_sub"},
			prepare: func() {
				repo.EXPECT().SaveGoogleUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uuid.New(), nil).Times(1)
				codeStore.EXPECT().SaveOauthCode(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("save oauth code error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.GoogleUserDTO{
				Name:  tc.data.name,
				Email: tc.data.email,
				Sub:   tc.data.sub,
			}
			code, err := google.GoogleCallback(t.Context(), &req)
			if tc.success {
				require.NoError(t, err)
				assert.NotEmpty(t, code)
			} else {
				require.Error(t, err)
				assert.Empty(t, code)
			}
		})
	}
}

func TestGoogleLoginAndroid(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repo_mocks.NewMockAuthRepository(ctrl)
	manager := repo_mocks.NewMockManagerToken(ctrl)
	sessionStore := repo_mocks.NewMockSessionStore(ctrl)

	google := NewGoogleAuthService(repo, nil, nil, manager, sessionStore, zaptest.NewLogger(t), "client_id")

	testCases := []struct {
		caseName string
		data     struct {
			idToken string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "invalid id token",
			data: struct {
				idToken string
			}{idToken: "invalid_token"},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "empty id token",
			data: struct {
				idToken string
			}{idToken: ""},
			prepare: func() {},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.GoogleAndroidUserDTO{
				IDToken: tc.data.idToken,
			}
			tokens, err := google.GoogleLoginAndroid(t.Context(), req)
			if tc.success {
				require.NoError(t, err)
				assert.NotEmpty(t, tokens)
			} else {
				require.Error(t, err)
				require.ErrorIs(t, err, domain.ErrInvalidGoogleToken)
			}
		})
	}
}

func TestExchange(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repo_mocks.NewMockAuthRepository(ctrl)
	codeStore := repo_mocks.NewMockCodeStore(ctrl)
	manager := repo_mocks.NewMockManagerToken(ctrl)
	sessionStore := repo_mocks.NewMockSessionStore(ctrl)

	google := NewGoogleAuthService(repo, codeStore, nil, manager, sessionStore, zaptest.NewLogger(t), "client_id")

	testCases := []struct {
		caseName string
		data     struct {
			code string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "exchange",
			data: struct {
				code string
			}{code: "oauth_code"},
			prepare: func() {
				codeStore.EXPECT().GetOauthCode(gomock.Any(), gomock.Any()).
					Return(uuid.NewString(), nil).Times(1)
				codeStore.EXPECT().DeleteOauthCode(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).
					Return(&models.User{}, nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("token", nil).Times(2)
				sessionStore.EXPECT().SaveToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "get oauth code error",
			data: struct {
				code string
			}{code: "oauth_code"},
			prepare: func() {
				codeStore.EXPECT().GetOauthCode(gomock.Any(), gomock.Any()).
					Return("", errors.New("get oauth code error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "delete oauth code error",
			data: struct {
				code string
			}{code: "oauth_code"},
			prepare: func() {
				codeStore.EXPECT().GetOauthCode(gomock.Any(), gomock.Any()).
					Return(uuid.NewString(), nil).Times(1)
				codeStore.EXPECT().DeleteOauthCode(gomock.Any(), gomock.Any()).
					Return(errors.New("delete oauth code error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "invalid user id in store",
			data: struct {
				code string
			}{code: "oauth_code"},
			prepare: func() {
				codeStore.EXPECT().GetOauthCode(gomock.Any(), gomock.Any()).
					Return("invalid_uuid", nil).Times(1)
				codeStore.EXPECT().DeleteOauthCode(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "get user by id error",
			data: struct {
				code string
			}{code: "oauth_code"},
			prepare: func() {
				codeStore.EXPECT().GetOauthCode(gomock.Any(), gomock.Any()).
					Return(uuid.NewString(), nil).Times(1)
				codeStore.EXPECT().DeleteOauthCode(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("get user by id error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "generate access token error",
			data: struct {
				code string
			}{code: "oauth_code"},
			prepare: func() {
				codeStore.EXPECT().GetOauthCode(gomock.Any(), gomock.Any()).
					Return(uuid.NewString(), nil).Times(1)
				codeStore.EXPECT().DeleteOauthCode(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).
					Return(&models.User{}, nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("generate access token error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "generate refresh token error",
			data: struct {
				code string
			}{code: "oauth_code"},
			prepare: func() {
				codeStore.EXPECT().GetOauthCode(gomock.Any(), gomock.Any()).
					Return(uuid.NewString(), nil).Times(1)
				codeStore.EXPECT().DeleteOauthCode(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).
					Return(&models.User{}, nil).Times(1)
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
				code string
			}{code: "oauth_code"},
			prepare: func() {
				codeStore.EXPECT().GetOauthCode(gomock.Any(), gomock.Any()).
					Return(uuid.NewString(), nil).Times(1)
				codeStore.EXPECT().DeleteOauthCode(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
				repo.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).
					Return(&models.User{}, nil).Times(1)
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

			req := dto.ExchangeRequestDTO{
				Code: tc.data.code,
			}
			tokens, err := google.Exchange(t.Context(), &req)
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
