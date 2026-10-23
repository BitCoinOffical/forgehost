package store

import (
	"context"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/models"

	"github.com/redis/go-redis/v9"
)

const (
	registerUserTTL = 24 * time.Hour
	pendingKey      = "pending_key:"
)

type UserStore struct {
	rdb *redis.Client
}

func NewUserStore(rdb *redis.Client) *UserStore {
	return &UserStore{rdb: rdb}
}

func (s *UserStore) SaveUser(ctx context.Context, randStr string, user *models.UserStored) error {
	key := pendingKey + randStr
	pipe := s.rdb.Pipeline()
	if err := pipe.HSet(ctx, key, user).Err(); err != nil {
		return fmt.Errorf("s.rdb.HSet: %w", err)
	}
	pipe.Expire(ctx, key, registerUserTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipe.Exec: %w", err)
	}
	return nil
}

func (s *UserStore) GetUser(ctx context.Context, randStr string) (*models.UserStored, error) {
	key := pendingKey + randStr
	cmd := s.rdb.HGetAll(ctx, key)
	res, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("s.rdb.HGetAll: %w", err)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("s.rdb.HGetAll: %w", domain.ErrNotFound)
	}

	var user models.UserStored
	if err := cmd.Scan(&user); err != nil {
		return nil, fmt.Errorf("cmd.Scan: %w", err)
	}
	return &user, nil
}
