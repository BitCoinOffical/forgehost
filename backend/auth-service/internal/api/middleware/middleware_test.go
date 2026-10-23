package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	midd_mocks "github.com/BitCoinOffical/forgehost/auth-service/internal/api/middleware/mock"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	tokens := midd_mocks.NewMockManagerToken(ctrl)

	m := NewMiddleware(nil, zaptest.NewLogger(t), tokens)

	r := gin.New()
	r.GET("/protected", m.AuthMiddleware(), func(c *gin.Context) {
		_, ok := c.Get(userContextKey)
		c.JSON(http.StatusOK, gin.H{"authenticated": ok})
	})

	testCases := []struct {
		caseName string
		data     struct {
			authorization string
		}
		prepare  func()
		status   int
		wantBody string
	}{
		{
			caseName: "authenticated",
			data: struct {
				authorization string
			}{authorization: "Bearer access_token"},
			prepare: func() {
				tokens.EXPECT().ValidateToken("access_token").
					Return(&jwtpkg.Claims{UserRole: "user"}, nil).Times(1)
			},
			status:   http.StatusOK,
			wantBody: `"authenticated":true`,
		},
		{
			caseName: "missing token",
			data: struct {
				authorization string
			}{authorization: ""},
			prepare:  func() {},
			status:   http.StatusUnauthorized,
			wantBody: errEmptyToken,
		},
		{
			caseName: "invalid authorization scheme",
			data: struct {
				authorization string
			}{authorization: "Basic access_token"},
			prepare:  func() {},
			status:   http.StatusUnauthorized,
			wantBody: errEmptyToken,
		},
		{
			caseName: "invalid token",
			data: struct {
				authorization string
			}{authorization: "Bearer invalid_token"},
			prepare: func() {
				tokens.EXPECT().ValidateToken("invalid_token").
					Return(nil, errors.New("invalid token")).Times(1)
			},
			status:   http.StatusUnauthorized,
			wantBody: errInvalidToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodGet,
				"/protected",
				strings.NewReader(""),
			)
			require.NoError(t, err)

			req.Header.Set(headerAuthorization, tc.data.authorization)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantBody)
		})
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	tokens := midd_mocks.NewMockManagerToken(ctrl)

	m := NewMiddleware(nil, zaptest.NewLogger(t), tokens)

	r := gin.New()
	r.GET("/admin", m.RequireRole(), func(c *gin.Context) {
		_, ok := c.Get(userContextKey)
		c.JSON(http.StatusOK, gin.H{"authorized": ok})
	})

	testCases := []struct {
		caseName string
		data     struct {
			authorization string
		}
		prepare  func()
		status   int
		wantBody string
	}{
		{
			caseName: "admin access",
			data: struct {
				authorization string
			}{authorization: "Bearer admin_token"},
			prepare: func() {
				tokens.EXPECT().ValidateToken("admin_token").
					Return(&jwtpkg.Claims{UserRole: adminRole}, nil).Times(1)
			},
			status:   http.StatusOK,
			wantBody: `"authorized":true`,
		},
		{
			caseName: "insufficient role",
			data: struct {
				authorization string
			}{authorization: "Bearer user_token"},
			prepare: func() {
				tokens.EXPECT().ValidateToken("user_token").
					Return(&jwtpkg.Claims{UserRole: "user"}, nil).Times(1)
			},
			status:   http.StatusForbidden,
			wantBody: errRole,
		},
		{
			caseName: "missing token",
			data: struct {
				authorization string
			}{authorization: ""},
			prepare:  func() {},
			status:   http.StatusUnauthorized,
			wantBody: errEmptyToken,
		},
		{
			caseName: "invalid token",
			data: struct {
				authorization string
			}{authorization: "Bearer invalid_token"},
			prepare: func() {
				tokens.EXPECT().ValidateToken("invalid_token").
					Return(nil, errors.New("invalid token")).Times(1)
			},
			status:   http.StatusUnauthorized,
			wantBody: errInvalidToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodGet,
				"/admin",
				strings.NewReader(""),
			)
			require.NoError(t, err)

			req.Header.Set(headerAuthorization, tc.data.authorization)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.status, w.Code)
			assert.Contains(t, w.Body.String(), tc.wantBody)
		})
	}
}
