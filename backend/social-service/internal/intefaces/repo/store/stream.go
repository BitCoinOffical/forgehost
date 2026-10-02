package store

import (
	"context"
	v2 "encoding/json/v2"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

const (
	streamKey  = "stream:"
	messageTTL = 24 * time.Hour
)

type StreamStore struct {
	rdb *redis.Client
}

func NewStreamStore(rdb *redis.Client) *StreamStore {
	return &StreamStore{rdb: rdb}
}

func (s *StreamStore) SaveStreamMessage(ctx context.Context, streamID string, msg *models.Message) error {
	key := streamKey + streamID

	data, err := v2.Marshal(msg)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	pipe := s.rdb.TxPipeline()
	pipe.RPush(ctx, key, data)
	pipe.LTrim(ctx, key, -200, -1)
	pipe.Expire(ctx, key, messageTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipe.Exec: %w", err)
	}
	return nil
}

func (s *StreamStore) GetStreamMessages(ctx context.Context, streamID string) ([]models.Message, error) {
	key := streamKey + streamID

	items, err := s.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("s.rdb.LRange: %w", err)
	}

	msgs := make([]models.Message, 0, len(items))

	for _, item := range items {
		var m models.Message
		if err := v2.Unmarshal([]byte(item), &m); err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %w", err)
		}
		msgs = append(msgs, m)
	}

	return msgs, nil
}
