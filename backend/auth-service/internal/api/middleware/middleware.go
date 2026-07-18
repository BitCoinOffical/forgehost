package middleware

import (
	"BitCoinOffical/forgehost/auth-service/internal/api/response"
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"net/http"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	limitergin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
	"go.uber.org/zap"
)

const (
	headerAuthorization = "Authorization"
	bearerSchema        = "Bearer Token"
	userContextKey      = "user_claims"
	prefix              = "rate_limiter"
)

const (
	errEmptyToken   = "missing token"
	errInvalidToken = "invalid token"
)

type Middleware struct {
	rdb    *redis.Client
	tokens *jwtpkg.ManagerToken
	logger *zap.Logger
}

func NewMiddleware(rdb *redis.Client, logger *zap.Logger, tokens *jwtpkg.ManagerToken) *Middleware {
	return &Middleware{rdb: rdb, logger: logger, tokens: tokens}
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

func (m *Middleware) RateLimiter() gin.HandlerFunc {

	rate, err := limiter.NewRateFromFormatted("5-M")
	if err != nil {
		logger.Fatal("limiter.NewRateFromFormatted", zap.Error(err))
	}

	store, err := redisstore.NewStoreWithOptions(m.rdb, limiter.StoreOptions{
		Prefix: prefix,
	})
	if err != nil {
		logger.Fatal("redisstore.NewStoreWithOptions:", zap.Error(err))
	}
	instance := limiter.New(store, rate)

	return limitergin.NewMiddleware(instance, limitergin.WithLimitReachedHandler(func(c *gin.Context) {
		response.ManyRequest(c, err, "too many requests", m.logger)
		c.Abort()
	}))

}
