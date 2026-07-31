package store

import (
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	tokenKey   = "refresh:"
	codeKey    = "code:"
	pendingKey = "pending_key:"
)

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

func (s *RedisStore) SaveToken(ctx context.Context, id uuid.UUID, value string, RefreshTTL time.Duration) error {
	key := fmt.Sprintf("%s%s", tokenKey, id.String())
	if err := s.rdb.Set(ctx, key, value, RefreshTTL).Err(); err != nil {
		return fmt.Errorf("s.rdb.Set: %w", err)
	}
	return nil
}

func (s *RedisStore) GetToken(ctx context.Context, id uuid.UUID) (string, error) {
	key := fmt.Sprintf("%s%s", tokenKey, id.String())
	value, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("s.rdb.Get: %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("s.rdb.Get: %w", err)
	}
	return value, nil
}

func (s *RedisStore) DeleteToken(ctx context.Context, id uuid.UUID) error {
	key := fmt.Sprintf("%s%s", tokenKey, id.String())
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("s.rdb.Del: %w", err)
	}
	return nil
}
