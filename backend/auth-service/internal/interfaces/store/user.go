package store

import (
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/models"
	"context"
	"errors"
	"fmt"
	"time"

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
	key := fmt.Sprintf("%s%s", pendingKey, randStr)
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
	key := fmt.Sprintf("%s%s", pendingKey, randStr)
	var user models.UserStored
	err := s.rdb.HGetAll(ctx, key).Scan(&user)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("s.rdb.HGetAll: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("s.rdb.HGetAll: %w", err)
	}
	return &user, nil
}
