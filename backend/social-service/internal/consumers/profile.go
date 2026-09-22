package consumers

import (
	"context"
	v2 "encoding/json/v2"
	"fmt"
	"sync"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

const (
	consumerCount = 4
)

type Consumer struct {
	client  *kgo.Client
	logger  *zap.Logger
	service *services.ProfileService
}

func NewConsumer(client *kgo.Client, logger *zap.Logger, service *services.ProfileService) *Consumer {
	return &Consumer{client: client, logger: logger, service: service}
}

func (k *Consumer) Run(ctx context.Context) <-chan error {
	wg := &sync.WaitGroup{}
	errs := make(chan error)
	for i := range consumerCount {
		wg.Go(func() {
			k.logger.Info("consumer started", zap.Int("num", i+1))
			for {
				fetches := k.client.PollFetches(ctx)
				if fetErrs := fetches.Errors(); len(errs) > 0 {
					for _, e := range fetErrs {
						errs <- fmt.Errorf("k.client.PollFetches: %w", e.Err)
						k.logger.Error("fetch error", zap.String("topic", e.Topic), zap.Int32("partition", e.Partition), zap.Error(e.Err))
					}
					continue
				}

				fetches.EachRecord(func(r *kgo.Record) {
					var event dto.UserProfileDTO
					if err := v2.Unmarshal(r.Value, &event); err != nil {
						errs <- fmt.Errorf("v2.Unmarshal: %w", err)
						return
					}
					k.logger.Debug("event data from kafka", zap.Any("user_id", event))

					if err := k.service.SaveProfile(ctx, &event); err != nil {
						errs <- fmt.Errorf("k.service.SaveUserProfile: %w", err)
						return
					}

					k.logger.Info("data successfully received.", zap.Int("num", i))
				})
			}
		})
	}

	go func() {
		wg.Wait()
		defer close(errs)
	}()

	return errs
}
