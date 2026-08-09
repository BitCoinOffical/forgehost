package store

import (
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	tokenKey = "refresh:"
)

type SessionStore struct {
	rdb *redis.Client
}

func NewSessionStore(rdb *redis.Client) *SessionStore {
	return &SessionStore{rdb: rdb}
}

func (s *SessionStore) SaveToken(ctx context.Context, id uuid.UUID, value string, RefreshTTL time.Duration) error {
	key := fmt.Sprintf("%s%s", tokenKey, id.String())
	if err := s.rdb.Set(ctx, key, value, RefreshTTL).Err(); err != nil {
		return fmt.Errorf("s.rdb.Set: %w", err)
	}
	return nil
}

func (s *SessionStore) GetToken(ctx context.Context, id uuid.UUID) (string, error) {
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

func (s *SessionStore) DeleteToken(ctx context.Context, id uuid.UUID) error {
	key := fmt.Sprintf("%s%s", tokenKey, id.String())
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("s.rdb.Del: %w", err)
	}
	return nil
}
