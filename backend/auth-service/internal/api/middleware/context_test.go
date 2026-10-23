package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.NewString()

	tests := []struct {
		name    string
		set     any
		want    string
		wantErr error
	}{
		{"ok", &jwtpkg.Claims{UserID: id}, id, nil},
		{"not found", nil, "", domain.ErrValueNotFound},
		{"string instead of claims", "abc", "", domain.ErrIncorrectType},
		{"claims by value", jwtpkg.Claims{UserID: id}, "", domain.ErrIncorrectType},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tc.set != nil {
				c.Set(userContextKey, tc.set)
			}

			got, err := GetUserID(c)

			require.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}
