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
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestUpdatePassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPasswordService(ctrl)

	password := NewPasswordHandler(srvc, zaptest.NewLogger(t))

	testCases := []struct {
		caseName string
		data     struct {
			body   string
			userID any
		}
		prepare func()
		status  int
	}{
		{
			caseName: "update password",
			data: struct {
				body   string
				userID any
			}{
				body:   `{"old_password":"old_password","new_password":"new_password","new_password_retry":"new_password"}`,
				userID: &jwtpkg.Claims{UserID: uuid.NewString()},
			},
			prepare: func() {
				srvc.EXPECT().UpdatePassword(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data: struct {
				body   string
				userID any
			}{body: `{invalid`, userID: &jwtpkg.Claims{UserID: uuid.NewString()}},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "passwords do not match",
			data: struct {
				body   string
				userID any
			}{
				body:   `{"old_password":"old_password","new_password":"new_password","new_password_retry":"other_password"}`,
				userID: &jwtpkg.Claims{UserID: uuid.NewString()},
			},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "user id not found in context",
			data: struct {
				body   string
				userID any
			}{
				body:   `{"old_password":"old_password","new_password":"new_password","new_password_retry":"new_password"}`,
				userID: nil,
			},
			prepare: func() {},
			status:  http.StatusUnauthorized,
		},
		{
			caseName: "incorrect user id type",
			data: struct {
				body   string
				userID any
			}{
				body:   `{"old_password":"old_password","new_password":"new_password","new_password_retry":"new_password"}`,
				userID: "not_claims",
			},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "update password error",
			data: struct {
				body   string
				userID any
			}{
				body:   `{"old_password":"old_password","new_password":"new_password","new_password_retry":"new_password"}`,
				userID: &jwtpkg.Claims{UserID: uuid.NewString()},
			},
			prepare: func() {
				srvc.EXPECT().UpdatePassword(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("update password error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			r := gin.New()
			r.PATCH("/auth/password/update", func(c *gin.Context) {
				if tc.data.userID != nil {
					c.Set(userContextKey, tc.data.userID)
				}
			}, password.UpdatePassword)

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPatch, "/auth/password/update", strings.NewReader(tc.data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestPasswordResetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPasswordService(ctrl)

	password := NewPasswordHandler(srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/password/reset", password.PasswordReset)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "password reset",
			data: struct {
				body string
			}{body: `{"email": "user@example.com","new_password": "newpassword123","new_password_retry": "newpassword123","pending_key": "any-non-empty-string","code": "123456"}`},
			prepare: func() {
				srvc.EXPECT().PasswordReset(gomock.Any(), gomock.Any()).
					Return(&dto.PendingKeyDTO{PendingKey: "pending_key"}, nil).Times(1)
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
			caseName: "password reset error",
			data: struct {
				body string
			}{body: `{"email": "user@example.com","new_password": "newpassword123","new_password_retry": "newpassword123","pending_key": "any-non-empty-string","code": "123456"}`},
			prepare: func() {
				srvc.EXPECT().PasswordReset(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("password reset error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/password/reset", strings.NewReader(tc.data.body))
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

func TestConfirmPasswordResetHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPasswordService(ctrl)

	password := NewPasswordHandler(srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/password/reset/confirm", password.ConfirmPasswordReset)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "confirm password reset",
			data: struct {
				body string
			}{body: `{"email":"user@gmail.com","pending_key":"pending_key","code":"123456","new_password":"new_password","new_password_retry":"new_password"}`},
			prepare: func() {
				srvc.EXPECT().ConfirmPasswordReset(gomock.Any(), gomock.Any()).
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
			caseName: "passwords do not match",
			data: struct {
				body string
			}{body: `{"email":"user@gmail.com","pending_key":"pending_key","code":"123456","new_password":"new_password","new_password_retry":"other_password"}`},
			prepare: func() {},
			status:  http.StatusBadRequest,
		},
		{
			caseName: "confirm password reset error",
			data: struct {
				body string
			}{body: `{"email":"user@gmail.com","pending_key":"pending_key","code":"123456","new_password":"new_password","new_password_retry":"new_password"}`},
			prepare: func() {
				srvc.EXPECT().ConfirmPasswordReset(gomock.Any(), gomock.Any()).
					Return(errors.New("confirm password reset error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/password/reset/confirm", strings.NewReader(tc.data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestPasswordResetResendHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPasswordService(ctrl)

	password := NewPasswordHandler(srvc, zaptest.NewLogger(t))

	r := gin.New()
	r.POST("/auth/password/reset/resend", password.PasswordResetResend)

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func()
		status  int
	}{
		{
			caseName: "password reset resend",
			data: struct {
				body string
			}{body: `{"email": "user@example.com","new_password": "newpassword123","new_password_retry": "newpassword123","pending_key": "key","code": "123456"}`},
			prepare: func() {
				srvc.EXPECT().PasswordResetResend(gomock.Any(), gomock.Any()).
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
			}{body: `{"email": "user@example.com","new_password": "newpassword123","new_password_retry": "newpassword123","pending_key": "key","code": "123456"}`},
			prepare: func() {
				srvc.EXPECT().PasswordResetResend(gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("s.resendStore.ResendLimitCheck: %w", domain.ErrToManyRequest)).Times(1)
			},
			status: http.StatusTooManyRequests,
		},
		{
			caseName: "password reset resend error",
			data: struct {
				body string
			}{body: `{"email": "user@example.com","new_password": "newpassword123","new_password_retry": "newpassword123","pending_key": "key","code": "123456"}`},
			prepare: func() {
				srvc.EXPECT().PasswordResetResend(gomock.Any(), gomock.Any()).
					Return(errors.New("password reset resend error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/password/reset/resend", strings.NewReader(tc.data.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}
