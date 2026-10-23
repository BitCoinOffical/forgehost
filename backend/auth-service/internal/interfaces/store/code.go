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
	codeKey      = "code:"
	resetKey     = "reset:"
	oauthKey     = "oauth_code:"
	oauthCodeTTL = 60 * time.Second
)

type CodeStore struct {
	rdb *redis.Client
}

func NewCodeStore(rdb *redis.Client) *CodeStore {
	return &CodeStore{rdb: rdb}
}

func (s *CodeStore) SaveVerificationCode(ctx context.Context, randStr string, code int, expiration time.Duration) error {
	key := codeKey + randStr
	if err := s.rdb.Set(ctx, key, code, expiration).Err(); err != nil {
		return fmt.Errorf("c.rdb.Set: %w", err)
	}
	return nil
}

func (s *CodeStore) SaveResetPasswordCode(ctx context.Context, randStr string, code int, expiration time.Duration) error {
	key := resetKey + randStr
	if err := s.rdb.Set(ctx, key, code, expiration).Err(); err != nil {
		return fmt.Errorf("c.rdb.Set: %w", err)
	}
	return nil
}

func (s *CodeStore) SaveOauthCode(ctx context.Context, oauthCode, userID string) error {
	key := oauthKey + oauthCode
	if err := s.rdb.Set(ctx, key, userID, oauthCodeTTL).Err(); err != nil {
		return fmt.Errorf("c.rdb.Set: %w", err)
	}
	return nil
}

func (s *CodeStore) GetOauthCode(ctx context.Context, oauthCode string) (string, error) {
	key := oauthKey + oauthCode
	userID, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("c.rdb.Get: %w", err)
	}
	return userID, nil
}

func (s *CodeStore) DeleteOauthCode(ctx context.Context, oauthCode string) error {
	key := oauthKey + oauthCode
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("c.rdb.Det: %w", err)
	}
	return nil
}

func (s *CodeStore) GetResetPasswordCode(ctx context.Context, randStr string) (string, error) {
	key := resetKey + randStr
	code, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("s.rdb.Get: %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("s.rdb.Get: %w", err)
	}
	return code, nil
}

func (s *CodeStore) GetVerificationCode(ctx context.Context, randStr string) (string, error) {
	key := codeKey + randStr
	code, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("s.rdb.Get: %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("s.rdb.Get: %w", err)
	}
	return code, nil
}
