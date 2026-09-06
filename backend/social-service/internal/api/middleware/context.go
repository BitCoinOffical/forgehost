package middleware

import (
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	jwtpkg "github.com/BitCoinOffical/forgehost/social-service/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func GetUserID(c *gin.Context) (string, error) {
	val, exists := c.Get(userContextKey)
	if !exists {
		return "", fmt.Errorf("middleware.GetUserID: %w", domain.ErrValueNotFound)
	}

	v, ok := val.(*jwtpkg.Claims)
	if !ok {
		return "", fmt.Errorf("middleware.GetUserID: %w type: %T", domain.ErrIncorrectType, v)
	}

	return v.UserID, nil
}
