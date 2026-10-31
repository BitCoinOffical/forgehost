package monitoring

import (
	"context"
	"errors"
	"testing"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	monitor_mocks "github.com/BitCoinOffical/forgehost/hosting-service/internal/monitoring/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestMonitor_CreateGroupError(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := monitor_mocks.NewMockWorldsRepo(ctrl)
	queue := monitor_mocks.NewMockWorldsQueue(ctrl)

	monitor := NewMonitoring(repo, queue, zaptest.NewLogger(t))

	wantErr := errors.New("redis unavailable")

	queue.EXPECT().
		CreateGroup(gomock.Any()).
		Return(wantErr).
		Times(1)

	err := monitor.Monitor(t.Context())

	require.ErrorIs(t, err, wantErr)
}

func TestMonitor_ContextCanceled(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := monitor_mocks.NewMockWorldsRepo(ctrl)
	queue := monitor_mocks.NewMockWorldsQueue(ctrl)

	monitor := NewMonitoring(repo, queue, zaptest.NewLogger(t))

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	queue.EXPECT().
		CreateGroup(gomock.Any()).
		Return(nil).
		Times(1)

	err := monitor.Monitor(ctx)

	require.NoError(t, err)
}

func TestMonitor_GetAllError(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := monitor_mocks.NewMockWorldsRepo(ctrl)
	queue := monitor_mocks.NewMockWorldsQueue(ctrl)

	monitor := NewMonitoring(repo, queue, zaptest.NewLogger(t))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	wantErr := errors.New("postgres unavailable")

	queue.EXPECT().
		CreateGroup(gomock.Any()).
		Return(nil).
		Times(1)

	repo.EXPECT().
		GetAll(gomock.Any()).
		DoAndReturn(func(context.Context) ([]models.MinecraftServer, error) {
			cancel()
			return nil, wantErr
		}).
		Times(1)

	err := monitor.Monitor(ctx)

	require.NoError(t, err)
}

func TestMonitor_EnqueueWorlds(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := monitor_mocks.NewMockWorldsRepo(ctrl)
	queue := monitor_mocks.NewMockWorldsQueue(ctrl)

	monitor := NewMonitoring(repo, queue, zaptest.NewLogger(t))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	worlds := []models.MinecraftServer{
		{
			ID:      "world-1",
			Address: "localhost",
			Port:    25565,
			Edition: "java",
		},
		{
			ID:      "world-2",
			Address: "192.168.1.100",
			Port:    19132,
			Edition: "bedrock",
		},
	}

	queue.EXPECT().
		CreateGroup(gomock.Any()).
		Return(nil).
		Times(1)

	repo.EXPECT().
		GetAll(gomock.Any()).
		Return(worlds, nil).
		Times(1)

	calls := 0

	queue.EXPECT().
		Enqueue(gomock.Any(), gomock.Any()).
		Times(len(worlds)).
		DoAndReturn(func(
			_ context.Context,
			payload *models.MinecraftServerCheckPayload,
		) (string, error) {
			expected := models.MinecraftServerCheckPayload{
				ID:      worlds[calls].ID,
				Address: worlds[calls].Address,
				Port:    worlds[calls].Port,
				Edition: worlds[calls].Edition,
			}

			assert.Equal(t, expected, *payload)

			calls++

			if calls == len(worlds) {
				cancel()
			}

			return "message-id", nil
		})

	err := monitor.Monitor(ctx)

	require.NoError(t, err)
	assert.Equal(t, len(worlds), calls)
}

func TestMonitor_EnqueueError(t *testing.T) {
	ctrl := gomock.NewController(t)

	repo := monitor_mocks.NewMockWorldsRepo(ctrl)
	queue := monitor_mocks.NewMockWorldsQueue(ctrl)

	monitor := NewMonitoring(repo, queue, zaptest.NewLogger(t))

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	worlds := []models.MinecraftServer{
		{
			ID:      "world-1",
			Address: "localhost",
			Port:    25565,
			Edition: "java",
		},
		{
			ID:      "world-2",
			Address: "localhost",
			Port:    19132,
			Edition: "bedrock",
		},
	}

	queue.EXPECT().
		CreateGroup(gomock.Any()).
		Return(nil).
		Times(1)

	repo.EXPECT().
		GetAll(gomock.Any()).
		Return(worlds, nil).
		Times(1)

	enqueueErr := errors.New("redis unavailable")
	calls := 0

	queue.EXPECT().
		Enqueue(gomock.Any(), gomock.Any()).
		Times(2).
		DoAndReturn(func(
			_ context.Context,
			_ *models.MinecraftServerCheckPayload,
		) (string, error) {
			calls++

			if calls == 1 {
				return "", enqueueErr
			}

			cancel()
			return "message-id", nil
		})

	err := monitor.Monitor(ctx)

	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}
