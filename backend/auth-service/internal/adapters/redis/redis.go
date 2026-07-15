package redisdb

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	RDBAddr string
	RDBPort string
	RDBDB   int
	RDBPass string
}

func NewRedis(cfg RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RDBAddr, cfg.RDBPort),
		Password: cfg.RDBPass,
		DB:       cfg.RDBDB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}
