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
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockRegisterService(ctrl)

	register := NewRegisterHandler(srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/register", register.Register)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "register",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","password":"password123","password_retry":"password123"}`},
			prepare: func() {
				srvc.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).
					Return(&dto.PendingKeyDTO{PendingKey: "pending_key"}, nil).Times(1)
			},
			status: http.StatusCreated,
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
			caseName: "invalid email",
			data: struct {
				body string
			}{body: `{"email":"not_email","password":"password123","password_retry":"password123"}`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "password too short",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","password":"short","password_retry":"short"}`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "passwords do not match",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","password":"password123","password_retry":"other_password"}`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "email already exists",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","password":"password123","password_retry":"password123"}`},
			prepare: func() {
				srvc.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("s.store.Create: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "register error",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","password":"password123","password_retry":"password123"}`},
			prepare: func() {
				srvc.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("register error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/register", strings.NewReader(tc.data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusCreated {
				assert.Contains(t, w.Body.String(), "pending_key")
			}
		})
	}
}

func TestVerifyEmailHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockRegisterService(ctrl)

	register := NewRegisterHandler(srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/verify-email", register.VerifyEmail)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "verify email",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","pending_key":"pending_key","code":"123456"}`},
			prepare: func() {
				srvc.EXPECT().VerifyEmail(gomock.Any(), gomock.Any()).
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
			caseName: "code too short",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","pending_key":"pending_key","code":"123"}`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "code not numeric",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","pending_key":"pending_key","code":"12345a"}`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "email already exists",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","pending_key":"pending_key","code":"123456"}`},
			prepare: func() {
				srvc.EXPECT().VerifyEmail(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("s.store.Create: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "verify email error",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","pending_key":"pending_key","code":"123456"}`},
			prepare: func() {
				srvc.EXPECT().VerifyEmail(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("verify email error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/verify-email", strings.NewReader(tc.data.body))
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

func TestResendVerifyEmailHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockRegisterService(ctrl)

	register := NewRegisterHandler(srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/verify-email/resend", register.ResendVerifyEmail)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "resend verify email",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","pending_key":"pending_key","code":"123456"}`},
			prepare: func() {
				srvc.EXPECT().ResendVerifyEmail(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
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
			caseName: "too many requests",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","pending_key":"pending_key","code":"123456"}`},
			prepare: func() {
				srvc.EXPECT().ResendVerifyEmail(gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("s.resendStore.ResendLimitCheck: %w", domain.ErrToManyRequest)).Times(1)
			},
			status: http.StatusTooManyRequests,
		},
		{
			caseName: "resend verify email error",
			data: struct {
				body string
			}{body: `{"email":"user@example.com","pending_key":"pending_key","code":"123456"}`},
			prepare: func() {
				srvc.EXPECT().ResendVerifyEmail(gomock.Any(), gomock.Any()).
					Return(errors.New("resend verify email error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/verify-email/resend", strings.NewReader(tc.data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}
