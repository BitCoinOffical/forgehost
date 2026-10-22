package services

import (
	"errors"
	"testing"

	repo_mocks "github.com/BitCoinOffical/forgehost/auth-service/internal/interfaces/services/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestLogoutUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessionStore := repo_mocks.NewMockSessionStore(ctrl)

	logout := NewLogoutService(sessionStore, zaptest.NewLogger(t))

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "logout user",
			data: struct {
				userID string
			}{userID: uuid.NewString()},
			prepare: func() {
				sessionStore.EXPECT().DeleteToken(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "invalid user id",
			data: struct {
				userID string
			}{userID: "invalid_uuid"},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "empty user id",
			data: struct {
				userID string
			}{userID: ""},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "delete token error",
			data: struct {
				userID string
			}{userID: uuid.NewString()},
			prepare: func() {
				sessionStore.EXPECT().DeleteToken(gomock.Any(), gomock.Any()).
					Return(errors.New("delete token error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := logout.LogoutUser(t.Context(), tc.data.userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
