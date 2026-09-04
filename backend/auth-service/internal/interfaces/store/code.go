package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

const (
	codeKey  = "code:"
	resetKey = "reset:"
)

type CodeStore struct {
	rdb *redis.Client
}

func NewCodeStore(rdb *redis.Client) *CodeStore {
	return &CodeStore{rdb: rdb}
}

func (s *CodeStore) SaveVerificationCode(ctx context.Context, randStr string, value int, expiration time.Duration) error {
	key := codeKey + randStr
	if err := s.rdb.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("c.rdb.Set: %w", err)
	}
	return nil
}

func (s *CodeStore) SaveResetPasswordCode(ctx context.Context, randStr string, value int, expiration time.Duration) error {
	key := resetKey + randStr
	if err := s.rdb.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("c.rdb.Set: %w", err)
	}
	return nil
}

func (s *CodeStore) GetResetPasswordCode(ctx context.Context, randStr string) (string, error) {
	key := resetKey + randStr
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("s.rdb.Get: %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("s.rdb.Get: %w", err)
	}
	return val, nil
}

func (s *CodeStore) GetVerificationCode(ctx context.Context, randStr string) (string, error) {
	key := codeKey + randStr
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("s.rdb.Get: %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("s.rdb.Get: %w", err)
	}
	return val, nil
}
