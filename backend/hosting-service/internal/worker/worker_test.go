package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	work_mocks "github.com/BitCoinOffical/forgehost/hosting-service/internal/worker/mock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestWorkerPool(t *testing.T) {
	status := &models.MinecraftStatus{
		Online:        true,
		Latency:       150,
		PlayersOnline: 4,
		PlayersMax:    20,
		Version:       "1.21.1",
		MOTD:          "ForgeHost",
	}

	validMessage := redis.XMessage{
		ID: "message-1",
		Values: map[string]interface{}{
			"payload": `{
				"id": "world-1",
				"address": "localhost",
				"port": 25565,
				"edition": "java"
			}`,
		},
	}

	testCases := []struct {
		caseName     string
		message      redis.XMessage
		readErr      error
		checkStatus  *models.MinecraftStatus
		checkErr     error
		removeErr    error
		setOnlineErr error
		ackErr       error

		wantCheck     bool
		wantRemove    bool
		wantSetOnline bool
		wantAck       bool

		cancelAt string
	}{
		{
			caseName:      "success",
			message:       validMessage,
			checkStatus:   status,
			wantCheck:     true,
			wantSetOnline: true,
			wantAck:       true,
			cancelAt:      "ack",
		},
		{
			caseName:   "checker error",
			message:    validMessage,
			checkErr:   errors.New("minecraft server unavailable"),
			wantCheck:  true,
			wantRemove: true,
			cancelAt:   "remove",
		},
		{
			caseName: "invalid payload",
			message: redis.XMessage{
				ID: "message-1",
				Values: map[string]interface{}{
					"payload": 123,
				},
			},
			cancelAt: "read",
		},
		{
			caseName: "invalid json",
			message: redis.XMessage{
				ID: "message-1",
				Values: map[string]interface{}{
					"payload": "{invalid json",
				},
			},
			cancelAt: "read",
		},
		{
			caseName:      "cache error",
			message:       validMessage,
			checkStatus:   status,
			checkErr:      nil,
			setOnlineErr:  errors.New("redis unavailable"),
			wantCheck:     true,
			wantSetOnline: true,
			cancelAt:      "set_online",
		},
		{
			caseName:      "ack error",
			message:       validMessage,
			checkStatus:   status,
			ackErr:        errors.New("redis ack failed"),
			wantCheck:     true,
			wantSetOnline: true,
			wantAck:       true,
			cancelAt:      "ack",
		},
		{
			caseName: "read error",
			readErr:  errors.New("redis unavailable"),
			cancelAt: "read",
		},
		{
			caseName:   "remove error",
			message:    validMessage,
			checkErr:   errors.New("minecraft server unavailable"),
			removeErr:  errors.New("redis remove failed"),
			wantCheck:  true,
			wantRemove: true,
			cancelAt:   "remove",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			cache := work_mocks.NewMockWorldsCache(ctrl)
			queue := work_mocks.NewMockWorldsQueue(ctrl)
			checker := work_mocks.NewMockMinecraftChecker(ctrl)

			w := NewWorker(
				zaptest.NewLogger(t),
				cache,
				queue,
				checker,
			)

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			cancelIfNeeded := func(stage string) {
				if testCase.cancelAt == stage {
					cancel()
				}
			}

			queue.EXPECT().
				Read(gomock.Any(), "worker-1").
				DoAndReturn(func(
					_ context.Context,
					consumer string,
				) ([]redis.XMessage, error) {
					cancelIfNeeded("read")

					if testCase.readErr != nil {
						return nil, testCase.readErr
					}

					return []redis.XMessage{testCase.message}, nil
				}).
				Times(1)

			if testCase.wantCheck {
				checker.EXPECT().
					Check(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						world *models.MinecraftServer,
					) (*models.MinecraftStatus, error) {
						assert.Equal(t, "world-1", world.ID)
						assert.Equal(t, "localhost", world.Address)
						assert.Equal(t, 25565, world.Port)
						assert.Equal(t, "java", world.Edition)

						return testCase.checkStatus, testCase.checkErr
					}).
					Times(1)
			}

			if testCase.wantRemove {
				cache.EXPECT().
					Remove(gomock.Any(), "world-1").
					DoAndReturn(func(
						_ context.Context,
						id string,
					) error {
						cancelIfNeeded("remove")
						return testCase.removeErr
					}).
					Times(1)
			}

			if testCase.wantSetOnline {
				cache.EXPECT().
					SetOnline(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						world *models.MinecraftServer,
					) error {
						assert.Equal(t, "world-1", world.ID)
						assert.Equal(t, true, world.Online)
						assert.Equal(t, status.Latency, world.Latency)
						assert.Equal(t, status.PlayersOnline, world.PlayersOnline)
						assert.Equal(t, status.PlayersMax, world.PlayersMax)
						assert.Equal(t, status.Version, world.Version)
						assert.Equal(t, status.MOTD, world.MOTD)

						cancelIfNeeded("set_online")
						return testCase.setOnlineErr
					}).
					Times(1)
			}

			if testCase.wantAck {
				queue.EXPECT().
					Ack(gomock.Any(), "message-1").
					DoAndReturn(func(
						_ context.Context,
						messageID string,
					) error {
						assert.Equal(t, "message-1", messageID)

						cancelIfNeeded("ack")
						return testCase.ackErr
					}).
					Times(1)
			}

			w.WorkerPool(ctx, 1)
		})
	}
}

func TestWorkerPool_CanceledContext(t *testing.T) {
	ctrl := gomock.NewController(t)

	cache := work_mocks.NewMockWorldsCache(ctrl)
	queue := work_mocks.NewMockWorldsQueue(ctrl)
	checker := work_mocks.NewMockMinecraftChecker(ctrl)

	w := NewWorker(
		zaptest.NewLogger(t),
		cache,
		queue,
		checker,
	)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	w.WorkerPool(ctx, 1)
}
