package services

import (
	"errors"
	"testing"

	repo_mocks "github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/services/mock"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestUpdateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	manager := repo_mocks.NewMockManagerToken(ctrl)
	sessionStore := repo_mocks.NewMockSessionStore(ctrl)

	refresh := NewRefreshService(manager, sessionStore, zaptest.NewLogger(t))

	testCases := []struct {
		caseName string
		data     struct {
			refreshToken string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "update access token",
			data: struct {
				refreshToken string
			}{refreshToken: "refresh_token"},
			prepare: func() {
				manager.EXPECT().ValidateToken(gomock.Any()).
					Return(&jwtpkg.Claims{UserID: uuid.NewString()}, nil).Times(1)
				sessionStore.EXPECT().GetToken(gomock.Any(), gomock.Any()).
					Return("refresh_token", nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("access_token", nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "validate token error",
			data: struct {
				refreshToken string
			}{refreshToken: "invalid_token"},
			prepare: func() {
				manager.EXPECT().ValidateToken(gomock.Any()).
					Return(nil, errors.New("validate token error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "invalid user id in token",
			data: struct {
				refreshToken string
			}{refreshToken: "refresh_token"},
			prepare: func() {
				manager.EXPECT().ValidateToken(gomock.Any()).
					Return(&jwtpkg.Claims{UserID: "invalid_uuid"}, nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "get token error",
			data: struct {
				refreshToken string
			}{refreshToken: "refresh_token"},
			prepare: func() {
				manager.EXPECT().ValidateToken(gomock.Any()).
					Return(&jwtpkg.Claims{UserID: uuid.NewString()}, nil).Times(1)
				sessionStore.EXPECT().GetToken(gomock.Any(), gomock.Any()).
					Return("", errors.New("get token error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "token mismatch",
			data: struct {
				refreshToken string
			}{refreshToken: "refresh_token"},
			prepare: func() {
				manager.EXPECT().ValidateToken(gomock.Any()).
					Return(&jwtpkg.Claims{UserID: uuid.NewString()}, nil).Times(1)
				sessionStore.EXPECT().GetToken(gomock.Any(), gomock.Any()).
					Return("other_refresh_token", nil).Times(1)
			},
			success: false,
		},
		{
			caseName: "generate access token error",
			data: struct {
				refreshToken string
			}{refreshToken: "refresh_token"},
			prepare: func() {
				manager.EXPECT().ValidateToken(gomock.Any()).
					Return(&jwtpkg.Claims{UserID: uuid.NewString()}, nil).Times(1)
				sessionStore.EXPECT().GetToken(gomock.Any(), gomock.Any()).
					Return("refresh_token", nil).Times(1)
				manager.EXPECT().GenerateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("generate access token error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			tokens, err := refresh.UpdateAccessToken(t.Context(), tc.data.refreshToken)
			if tc.success {
				require.NoError(t, err)
				require.NotNil(t, tokens)
				assert.Equal(t, "access_token", tokens.AccessToken)
				assert.Equal(t, tc.data.refreshToken, tokens.RefreshToken)
			} else {
				require.Error(t, err)
				assert.Nil(t, tokens)
			}
		})
	}
}
