package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/auth-service/internal/api/handlers/mock"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestUpdateAccessTokenHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockRefreshService(ctrl)

	refresh := NewRefreshHandler(srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/refresh", refresh.UpdateAccessToken)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "update access token",
			data: struct {
				body string
			}{body: `{"refresh_token":"refresh_token"}`},
			prepare: func() {
				srvc.EXPECT().UpdateAccessToken(gomock.Any(), "refresh_token").
					Return(&models.Tokens{AccessToken: "access", RefreshToken: "refresh"}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data: struct {
				body string
			}{body: `{invalid`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "empty refresh token",
			data: struct {
				body string
			}{body: `{"refresh_token":""}`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "token not found",
			data: struct {
				body string
			}{body: `{"refresh_token":"refresh_token"}`},
			prepare: func() {
				srvc.EXPECT().UpdateAccessToken(gomock.Any(), "refresh_token").
					Return(nil, fmt.Errorf("s.store.Get: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusUnauthorized,
		},
		{
			caseName: "update access token error",
			data: struct {
				body string
			}{body: `{"refresh_token":"refresh_token"}`},
			prepare: func() {
				srvc.EXPECT().UpdateAccessToken(gomock.Any(), "refresh_token").
					Return(nil, errors.New("update access token error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/refresh", strings.NewReader(tc.data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, w.Body.String(), "access_token")
				assert.Contains(t, w.Body.String(), "refresh_token")
			}
		})
	}
}
