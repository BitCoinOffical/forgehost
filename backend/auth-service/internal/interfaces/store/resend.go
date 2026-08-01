package store

import (
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	resendKey = "resend_limit:"
	resendTTL = time.Minute
)

type ResendStore struct {
	rdb *redis.Client
}

func NewResendStore(rdb *redis.Client) *ResendStore {
	return &ResendStore{rdb: rdb}
}

func (s *ResendStore) ResendLimitAdd(ctx context.Context, email string) error {
	key := fmt.Sprintf("%s%s", resendKey, email)
	if err := s.rdb.Set(ctx, key, email, resendTTL).Err(); err != nil {
		return fmt.Errorf("s.rdb.Set: %w", err)
	}
	return nil
}

func (s *ResendStore) ResendLimitCheck(ctx context.Context, email string) (string, error) {
	key := fmt.Sprintf("%s%s", resendKey, email)
	res, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) || res != "" {
			return "", fmt.Errorf("s.rdb.Set: %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("s.rdb.Set: %w", err)
	}
	return res, nil
}
