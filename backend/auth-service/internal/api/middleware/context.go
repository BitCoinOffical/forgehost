package middleware

import (
	"BitCoinOffical/forgehost/auth-service/internal/domain"

	"github.com/gin-gonic/gin"
)

func GetUserID(c *gin.Context) (string, error) {

	idStr, ok := c.Get(userContextKey)
	if !ok {
		return "", domain.ErrValueNotFound
	}

	val, ok := idStr.(string)
	if !ok {
		return "", domain.ErrIncorrectType
	}

	return val, nil
}
