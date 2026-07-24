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

func (s *Session) GetToken(ctx context.Context, id string) (string, error) {
	key := fmt.Sprintf("%s%s", tokenKey, id)
	value, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("s.rdb.Get: %w", err)
	}
	return value, nil
}

func (s *Session) DeleteToken(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s%s", tokenKey, id)
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("s.rdb.Del: %w", err)
	}
	return nil
}

func (s *Session) SaveVerificationCode(ctx context.Context, value int, expiration time.Duration) error {
	key := fmt.Sprintf("%s%d", codeKey, value)
	if err := s.rdb.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("c.rdb.Set: %w", err)
	}
	return nil
}

func (s *Session) GetVerificationCode(ctx context.Context, value int) (string, error) {
	key := fmt.Sprintf("%s%d", codeKey, value)
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("s.rdb.Get: %w", err)
	}
	return val, nil
}
