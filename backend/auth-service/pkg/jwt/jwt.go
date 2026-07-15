package jwtpkg

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	userID string
	jwt.RegisteredClaims
}

type ManagerToken struct {
	SecretKey []byte
}

func NewManagerToken(SecretKey string) *ManagerToken {
	return &ManagerToken{SecretKey: []byte(SecretKey)}
}

func (m *ManagerToken) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", t.Header["alg"])
		}
		return m.SecretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token is expired: %w", err)
		}
		return nil, fmt.Errorf("failed parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", jwt.ErrTokenInvalidClaims)
	}

	return claims, nil
}

func (m *ManagerToken) GenerateToken(userID uuid.UUID, ttl time.Duration) (string, error) {
	claims := Claims{
		userID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.SecretKey)
}
