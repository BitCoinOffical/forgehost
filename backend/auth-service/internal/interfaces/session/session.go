package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	tokenKey = "refresh:"
	codeKey  = "code:"
)

type Session struct {
	rdb *redis.Client
}

func NewSession(rdb *redis.Client) *Session {
	return &Session{rdb: rdb}
}

func (s *Session) SaveToken(ctx context.Context, id uuid.UUID, value string, RefreshTTL time.Duration) error {
	key := fmt.Sprintf("%s%s", tokenKey, id.String())
	if err := s.rdb.Set(ctx, key, value, RefreshTTL).Err(); err != nil {
		return fmt.Errorf("c.rdb.Set: %w", err)
	}
	return nil
}

func (s *Session) GetToken(ctx context.Context, id uuid.UUID) (string, error) {
	key := fmt.Sprintf("%s%s", tokenKey, id.String())
	value, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("s.rdb.Get: %w", err)
	}
	return value, nil
}

func (s *Session) DeleteToken(ctx context.Context, id uuid.UUID) error {
	key := fmt.Sprintf("%s%s", tokenKey, id.String())
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("s.rdb.Del: %w", err)
	}
	return nil
}

func (s *Session) SaveVerificationCode(ctx context.Context, id uuid.UUID, value int, expiration time.Duration) error {
	key := fmt.Sprintf("%s%s", codeKey, id.String())
	if err := s.rdb.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("c.rdb.Set: %w", err)
	}
	return nil
}

func (s *Session) GetVerificationCode(ctx context.Context, id uuid.UUID) (string, error) {
	key := fmt.Sprintf("%s%s", codeKey, id.String())
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("s.rdb.Get: %w", err)
	}
	return val, nil
}
