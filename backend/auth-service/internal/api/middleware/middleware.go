package middleware

import (
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	headerAuthorization = "Authorization"
	bearerSchema        = "Bearer Token"
	userContextKey      = "user_claims"
)

const (
	errEmptyToken   = "missing token"
	errInvalidToken = "invalid token"
)

type Middleware struct {
	tokens *jwtpkg.ManagerToken
	logger *zap.Logger
}

func NewAuthMiddleware(logger *zap.Logger, tokens *jwtpkg.ManagerToken) *Middleware {
	return &Middleware{logger: logger, tokens: tokens}
}

func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(bearerSchema)
		if token == "" {
			m.logger.Warn(errEmptyToken, zap.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": errEmptyToken,
			})
			return
		}

		claims, err := m.tokens.ValidateToken(token)
		if err != nil {
			m.logger.Warn(errInvalidToken, zap.String("path", c.Request.URL.Path), zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": errInvalidToken,
			})
			return
		}

		c.Set(userContextKey, claims)
		c.Next()
	}
}
