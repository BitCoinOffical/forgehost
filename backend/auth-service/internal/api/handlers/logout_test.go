package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/auth-service/internal/api/handlers/mock"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

const (
	userContextKey = "user_claims"
)

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockLogoutService(ctrl)

	logout := NewLogoutHandler(srvc, zaptest.NewLogger(t))

	testCases := []struct {
		caseName string
		data     struct {
			userID any
		}
		prepare func()
		status  int
	}{
		{
			caseName: "logout",
			data: struct {
				userID any
			}{userID: &jwtpkg.Claims{UserID: uuid.NewString()}},
			prepare: func() {
				srvc.EXPECT().LogoutUser(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "user id not found in context",
			data: struct {
				userID any
			}{userID: nil},
			prepare: func() {},
			status:  http.StatusUnauthorized,
		},
		{
			caseName: "incorrect user id type",
			data: struct {
				userID any
			}{userID: "not_claims"},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "logout error",
			data: struct {
				userID any
			}{userID: &jwtpkg.Claims{UserID: uuid.NewString()}},
			prepare: func() {
				srvc.EXPECT().LogoutUser(gomock.Any(), gomock.Any()).
					Return(errors.New("logout error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			r := gin.New()
			r.POST("/auth/logout", func(c *gin.Context) {
				if tc.data.userID != nil {
					c.Set(userContextKey, tc.data.userID)
				}
			}, logout.Logout)

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/logout", nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}
