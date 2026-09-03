package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

const (
	postKey    = "global:"
	expiration = 15 * time.Minute
)

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) SetGlobal(ctx context.Context, value []models.FeedPost, cursor string) error {
	key := postKey + cursor
	if err := c.rdb.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("c.rdb.Set: %w", err)
	}
	return nil
}

func (c *Cache) GetGlobal(ctx context.Context, cursor string) ([]models.FeedPost, error) {
	key := postKey + cursor
	raw, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("c.redis.Get: %w", err)
	}

	var posts []models.FeedPost
	if err := json.Unmarshal([]byte(raw), &posts); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return posts, nil
}

func (c *Cache) DeleteGlobal(ctx context.Context, cursor string) error {
	key := postKey + cursor
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("c.rdb.Del: %w", err)
	}
	return nil
}
