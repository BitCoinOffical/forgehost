package cache

import (
	"context"
	v2 "encoding/json/v2"

	"errors"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

const (
	onlineWorldsKey = "hosting:worlds:online"
	worldKeyPrefix  = "hosting:worlds:"
	worldsCacheTTL  = 5 * time.Minute
)

type WorldsCache struct {
	rdb *redis.Client
}

func NewWorldsCache(rdb *redis.Client) *WorldsCache {
	return &WorldsCache{rdb: rdb}
}

func (r *WorldsCache) SetOnline(ctx context.Context, world *models.MinecraftServer) error {
	worldCopy := *world
	worldCopy.Online = true
	worldCopy.CheckedAt = time.Now()

	data, err := v2.Marshal(worldCopy)
	if err != nil {
		return fmt.Errorf("v2.Marshal: %w", err)
	}

	_, err = r.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, worldKeyPrefix+world.ID, data, worldsCacheTTL)

		pipe.ZAdd(ctx, onlineWorldsKey, redis.Z{
			Score:  float64(world.PlayersOnline),
			Member: world.ID,
		})

		pipe.Expire(ctx, onlineWorldsKey, worldsCacheTTL)

		return nil
	})
	if err != nil {
		return fmt.Errorf("r.rdb.TxPipelined: %w", err)
	}

	return nil
}

func (r *WorldsCache) GetAll(ctx context.Context) ([]models.MinecraftServer, error) {
	ids, err := r.rdb.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   onlineWorldsKey,
		Start: 0,
		Stop:  -1,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("r.rdb.ZRangeArgs: %w", err)
	}

	worlds := make([]models.MinecraftServer, 0, len(ids))
	if len(ids) == 0 {
		return worlds, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = worldKeyPrefix + id
	}

	values, err := r.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("r.rdb.MGet: %w", err)
	}

	var expiredIDs []interface{}

	for i, value := range values {
		if value == nil {
			expiredIDs = append(expiredIDs, ids[i])
			continue
		}

		raw, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected redis value type: %T", value)
		}

		var world models.MinecraftServer
		if err := v2.Unmarshal([]byte(raw), &world); err != nil {
			return nil, fmt.Errorf("v2.Unmarshal: %w", err)
		}

		worlds = append(worlds, world)
	}

	if len(expiredIDs) > 0 {
		if err := r.rdb.ZRem(ctx, onlineWorldsKey, expiredIDs...).Err(); err != nil {
			return nil, fmt.Errorf("r.rdb.ZRem: %w", err)
		}
	}

	return worlds, nil
}

func (r *WorldsCache) GetByID(ctx context.Context, id string) (*models.MinecraftServer, error) {
	data, err := r.rdb.Get(ctx, worldKeyPrefix+id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.rdb.Get: %w", err)
	}

	var world models.MinecraftServer
	if err := v2.Unmarshal(data, &world); err != nil {
		return nil, fmt.Errorf("v2.Unmarshal: %w", err)
	}

	return &world, nil
}

func (r *WorldsCache) Remove(ctx context.Context, id string) error {
	_, err := r.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, worldKeyPrefix+id)
		pipe.ZRem(ctx, onlineWorldsKey, id)

		return nil
	})
	if err != nil {
		return fmt.Errorf("r.rdb.TxPipelined: %w", err)
	}

	return nil
}
