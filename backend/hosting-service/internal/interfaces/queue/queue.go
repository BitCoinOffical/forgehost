package queue

import (
	"context"
	v2 "encoding/json/v2"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

const (
	worldsChecksStream = "hosting:worlds:checks"
	worldsChecksGroup  = "world-checkers"
)

type WorldsQueue struct {
	rdb *redis.Client
}

func NewQueue(rdb *redis.Client) *WorldsQueue {
	return &WorldsQueue{rdb: rdb}
}
func (q *WorldsQueue) CreateGroup(ctx context.Context) error {
	err := q.rdb.XGroupCreateMkStream(
		ctx,
		worldsChecksStream,
		worldsChecksGroup,
		"0",
	).Err()

	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("q.rdb.XGroupCreateMkStream: %w", err)
	}

	return nil
}

func (q *WorldsQueue) Enqueue(ctx context.Context, world *models.MinecraftServerCheckPayload) (string, error) {
	payload, err := v2.Marshal(world)
	if err != nil {
		return "", fmt.Errorf("json.Marshal: %w", err)
	}

	id, err := q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: worldsChecksStream,
		ID:     "*",
		Values: map[string]interface{}{"payload": string(payload)}}).Result()
	if err != nil {
		return "", fmt.Errorf("q.rdb.XAdd: %w", err)
	}

	return id, nil
}

func (q *WorldsQueue) Read(ctx context.Context, consumer string) ([]redis.XMessage, error) {
	streams, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    worldsChecksGroup,
		Consumer: consumer,
		Streams:  []string{worldsChecksStream, ">"},
		Count:    1,
		Block:    5 * time.Second,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, fmt.Errorf("q.rdb.XReadGroup: %w", err)
	}

	if len(streams) == 0 {
		return nil, nil
	}

	return streams[0].Messages, nil
}

func (q *WorldsQueue) Ack(ctx context.Context, messageID string) error {
	err := q.rdb.XAck(ctx, worldsChecksStream, worldsChecksGroup, messageID).Err()
	if err != nil {
		return fmt.Errorf("q.rdb.XAck: %w", err)
	}

	return nil
}
