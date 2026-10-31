package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	"go.uber.org/zap"
)

const (
	monitorInterval = 15
)

type Monitoring struct {
	repo   WorldsRepo
	queue  WorldsQueue
	logger *zap.Logger
}

func NewMonitoring(repo WorldsRepo, queue WorldsQueue, logger *zap.Logger) *Monitoring {
	return &Monitoring{repo: repo, queue: queue, logger: logger}
}

func (m *Monitoring) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(monitorInterval * time.Second)
	defer ticker.Stop()

	if err := m.queue.CreateGroup(ctx); err != nil {
		return fmt.Errorf("m.queue.CreateGroup: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("monitoring shut down")
			return nil
		case <-ticker.C:
			worlds, err := m.repo.GetAll(ctx)
			if err != nil {
				m.logger.Error("m.repo.GetAll:", zap.Error(err))
				continue
			}

			for _, world := range worlds {
				if err := ctx.Err(); err != nil {
					return fmt.Errorf("ctx.Err(): %w", err)
				}

				playload := models.MinecraftServerCheckPayload{
					ID:      world.ID,
					Address: world.Address,
					Port:    world.Port,
					Edition: world.Edition,
				}

				id, err := m.queue.Enqueue(ctx, &playload)
				if err != nil {
					m.logger.Error("failed to enqueue world", zap.String("world_id", world.ID), zap.String("queue id", id), zap.Error(err))
				}
			}
		}
	}
}
