package middleware

import (
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetUserID(c *gin.Context) (string, error) {

	idStr, ok := c.Get(userContextKey)
	if !ok {
		return "", fmt.Errorf("middleware.GetUserID: %w", domain.ErrValueNotFound)
	}

	val, ok := idStr.(*jwtpkg.Claims)
	if !ok {
		return "", fmt.Errorf("middleware.GetUserID: %w type: %T", domain.ErrIncorrectType, idStr)
	}

	return val.UserID, nil
}
