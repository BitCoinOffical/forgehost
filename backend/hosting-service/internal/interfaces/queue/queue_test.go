package queue

import (
	"context"
	v2 "encoding/json/v2"
	"net"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	redistc "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	container, err := redistc.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(context.Background()))
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{
		Addr: net.JoinHostPort(host, port.Port()),
	})

	require.NoError(t, rdb.Ping(ctx).Err())

	t.Cleanup(func() {
		require.NoError(t, rdb.Close())
	})

	return rdb
}

func TestCreateGroup(t *testing.T) {
	rdb := setupTestRedis(t)
	queue := NewQueue(rdb)

	testCases := []struct {
		caseName string
		prepare  func()
		wantErr  bool
	}{
		{
			caseName: "create consumer group",
			prepare:  func() {},
			wantErr:  false,
		},
		{
			caseName: "consumer group already exists",
			prepare: func() {
				require.NoError(t, queue.CreateGroup(t.Context()))
			},
			wantErr: false,
		},
		{
			caseName: "canceled context",
			prepare:  func() {},
			wantErr:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			testCase.prepare()

			ctx := t.Context()

			if testCase.caseName == "canceled context" {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()
				ctx = canceledCtx
			}

			err := queue.CreateGroup(ctx)

			if testCase.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			groups, err := rdb.XInfoGroups(t.Context(), worldsChecksStream).Result()
			require.NoError(t, err)
			require.Len(t, groups, 1)

			assert.Equal(t, worldsChecksGroup, groups[0].Name)
		})
	}
}

func assertWorldPayload(t *testing.T, value interface{}, want *models.MinecraftServerCheckPayload) {
	t.Helper()

	payload, ok := value.(string)
	require.True(t, ok, "payload must be a string")

	var got models.MinecraftServerCheckPayload

	require.NoError(t, v2.Unmarshal([]byte(payload), &got))
	assert.Equal(t, *want, got)
}

func TestEnqueue(t *testing.T) {
	rdb := setupTestRedis(t)
	queue := NewQueue(rdb)

	world := &models.MinecraftServerCheckPayload{
		Address: "localhost",
		Port:    25565,
		Edition: "java",
	}

	testCases := []struct {
		caseName string
		prepare  func() context.Context
		wantErr  bool
	}{
		{
			caseName: "enqueue world",
			prepare: func() context.Context {
				return t.Context()
			},
			wantErr: false,
		},
		{
			caseName: "canceled context",
			prepare: func() context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			id, err := queue.Enqueue(testCase.prepare(), world)

			if testCase.wantErr {
				require.Error(t, err)
				assert.Empty(t, id)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, id)

			messages, err := rdb.XRangeN(
				t.Context(),
				worldsChecksStream,
				"-",
				"+",
				1,
			).Result()

			require.NoError(t, err)
			require.Len(t, messages, 1)

			assert.Equal(t, id, messages[0].ID)
			assert.Len(t, messages[0].Values, 1)

			assertWorldPayload(t, messages[0].Values["payload"], world)
		})
	}
}

func TestRead(t *testing.T) {
	rdb := setupTestRedis(t)
	queue := NewQueue(rdb)

	require.NoError(t, queue.CreateGroup(t.Context()))

	world := &models.MinecraftServerCheckPayload{
		Address: "localhost",
		Port:    25565,
		Edition: "java",
	}

	testCases := []struct {
		caseName string
		prepare  func() context.Context
		wantErr  bool
	}{
		{
			caseName: "read world",
			prepare: func() context.Context {
				_, err := queue.Enqueue(t.Context(), world)
				require.NoError(t, err)

				return t.Context()
			},
			wantErr: false,
		},
		{
			caseName: "canceled context",
			prepare: func() context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			messages, err := queue.Read(testCase.prepare(), "worker-1")

			if testCase.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, messages, 1)

			assert.NotEmpty(t, messages[0].ID)
			assertWorldPayload(t, messages[0].Values["payload"], world)
		})
	}
}

func TestAck(t *testing.T) {
	rdb := setupTestRedis(t)
	queue := NewQueue(rdb)
	ctx := t.Context()

	require.NoError(t, queue.CreateGroup(ctx))

	world := &models.MinecraftServerCheckPayload{
		Address: "localhost",
		Port:    25565,
		Edition: "java",
	}

	messageID, err := queue.Enqueue(ctx, world)
	require.NoError(t, err)

	messages, err := queue.Read(ctx, "worker-1")
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, messageID, messages[0].ID)

	pending, err := rdb.XPending(ctx, worldsChecksStream, worldsChecksGroup).Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), pending.Count)

	t.Run("canceled context", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		err := queue.Ack(canceledCtx, messageID)

		require.ErrorIs(t, err, context.Canceled)

		pending, err := rdb.XPending(ctx, worldsChecksStream, worldsChecksGroup).Result()
		require.NoError(t, err)
		require.Equal(t, int64(1), pending.Count)
	})

	t.Run("ack message", func(t *testing.T) {
		err := queue.Ack(ctx, messageID)
		require.NoError(t, err)

		pending, err := rdb.XPending(ctx, worldsChecksStream, worldsChecksGroup).Result()
		require.NoError(t, err)
		require.Equal(t, int64(0), pending.Count)

		streamMessages, err := rdb.XRangeN(
			ctx,
			worldsChecksStream,
			"-",
			"+",
			1,
		).Result()

		require.NoError(t, err)
		require.Len(t, streamMessages, 1)
		require.Equal(t, messageID, streamMessages[0].ID)
	})

	t.Run("ack already acknowledged message", func(t *testing.T) {
		err := queue.Ack(ctx, messageID)
		require.NoError(t, err)
	})
}
