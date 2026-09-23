package services

import (
	"time"

	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/google/uuid"
)

type ManagerToken interface {
	ValidateToken(tokenString string) (*jwtpkg.Claims, error)
	GenerateToken(userID uuid.UUID, UserRole string, IsVerified bool, IsBanned bool, ttl time.Duration) (string, error)
}
