package jwtpkg

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID     string `json:"user_id"`
	IsVerified bool   `json:"is_verified"`
	IsBanned   bool   `json:"is_banned"`
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

	if !claims.IsVerified {
		return nil, fmt.Errorf("email is not verify: %w", jwt.ErrTokenInvalidClaims)
	}

	if claims.IsBanned {
		return nil, fmt.Errorf("user is banned: %w", jwt.ErrTokenInvalidClaims)
	}

	return claims, nil
}

func (m *ManagerToken) GenerateToken(userID uuid.UUID, IsVerified bool, IsBanned bool, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID:     userID.String(),
		IsVerified: IsVerified,
		IsBanned:   IsBanned,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.SecretKey)
}

func GenerateRandomString() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
