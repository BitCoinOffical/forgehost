package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"sync"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	"go.uber.org/zap"
)

type Worker struct {
	logger  *zap.Logger
	cache   WorldsCache
	queue   WorldsQueue
	checker MinecraftChecker
}

func NewWorker(logger *zap.Logger, cache WorldsCache, queue WorldsQueue, checker MinecraftChecker) *Worker {
	return &Worker{
		logger:  logger,
		cache:   cache,
		queue:   queue,
		checker: checker,
	}
}

func (w *Worker) WorkerPool(ctx context.Context, worker int) {
	wg := &sync.WaitGroup{}

	for num := range worker {
		consumer := fmt.Sprintf("worker-%d", num+1)
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					w.logger.Info("worker shut down", zap.Int("num", num+1))
					return
				default:
					if ctx.Err() != nil {
						return
					}
					messages, err := w.queue.Read(ctx, consumer)
					if err != nil {
						w.logger.Error("w.queue.Read", zap.Error(err))
						continue
					}

					for _, message := range messages {
						payload, ok := message.Values["payload"].(string)
						if !ok {
							w.logger.Error("invalid message payload", zap.String("message_id", message.ID))
							continue
						}

						var world models.MinecraftServer

						if err := json.Unmarshal([]byte(payload), &world); err != nil {
							w.logger.Error("failed to unmarshal world", zap.String("message_id", message.ID), zap.Error(err))
							continue
						}

						status, err := w.checker.Check(ctx, &world)
						if err != nil {
							w.logger.Error("w.checker.Check", zap.Error(err))
							if removeErr := w.cache.Remove(ctx, world.ID); removeErr != nil {
								w.logger.Error("w.cache.Remove", zap.Error(removeErr))
							}
							continue
						}

						world.Online = status.Online
						world.Latency = status.Latency
						world.PlayersOnline = status.PlayersOnline
						world.PlayersMax = status.PlayersMax
						world.Version = status.Version
						world.MOTD = status.MOTD

						if err := w.cache.SetOnline(ctx, &world); err != nil {
							w.logger.Error("w.cache.SetOnline", zap.Error(err))
							continue
						}

						if err := w.queue.Ack(ctx, message.ID); err != nil {
							w.logger.Error("w.queue.Ack", zap.Error(err))
							continue
						}
					}
				}
			}
		})
	}

	wg.Wait()

}
