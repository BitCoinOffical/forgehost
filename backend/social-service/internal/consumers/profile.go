package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

const (
	consumerCount = 4
)

type Consumer struct {
	reader  *kafka.Reader
	logger  *zap.Logger
	service *services.ProfileService
}

func NewConsumer(reader *kafka.Reader, logger *zap.Logger, service *services.ProfileService) *Consumer {
	return &Consumer{reader: reader, logger: logger, service: service}
}

func (k *Consumer) Run(ctx context.Context) <-chan error {
	wg := &sync.WaitGroup{}
	errs := make(chan error)
	for i := range consumerCount {
		wg.Go(func() {
			k.logger.Info("consumer started", zap.Int("num", i))
			for {
				msg, err := k.reader.ReadMessage(ctx)
				if err != nil {
					errs <- fmt.Errorf("k.reader.ReadMessage: %w", err)
					break
				}

				var event dto.UserProfileDTO
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					errs <- fmt.Errorf("json.Unmarshal: %w", err)
					continue
				}

				if err := k.service.SaveProfile(ctx, &event); err != nil {
					errs <- fmt.Errorf("k.service.SaveUserProfile: %w", err)
					continue
				}

				k.logger.Info("data successfully received.", zap.Int("num", i))
			}
		})
	}

	go func() {
		wg.Wait()
		defer close(errs)
	}()

	return errs
}
