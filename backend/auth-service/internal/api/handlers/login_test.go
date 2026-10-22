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

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockLoginService(ctrl)

	login := NewLoginHandler(srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/login", login.Login)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "login",
			data: struct {
				body string
			}{body: `{"email":"user@gmail.com","password":"password"}`},
			prepare: func() {
				srvc.EXPECT().LoginUser(gomock.Any(), gomock.Any()).
					Return(&models.Tokens{AccessToken: "access_token", RefreshToken: "refresh_token"}, nil).Times(1)
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
			caseName: "user not found",
			data: struct {
				body string
			}{body: `{"email":"unknown@gmail.com","password":"password"}`},
			prepare: func() {
				srvc.EXPECT().LoginUser(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("s.repo.GetUserByEmail: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusUnauthorized,
		},
		{
			caseName: "invalid credentials",
			data: struct {
				body string
			}{body: `{"email":"user@gmail.com","password":"wrong_password"}`},
			prepare: func() {
				srvc.EXPECT().LoginUser(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("bcrypt.CompareHashAndPassword: %w", domain.ErrInvalidCredentials)).Times(1)
			},
			status: http.StatusUnauthorized,
		},
		{
			caseName: "internal error",
			data: struct {
				body string
			}{body: `{"email":"user@gmail.com","password":"password"}`},
			prepare: func() {
				srvc.EXPECT().LoginUser(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("internal error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/login", strings.NewReader(tc.data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}
