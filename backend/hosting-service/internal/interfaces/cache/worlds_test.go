package cache

import (
	"context"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	redistc "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()

	container, err := redistc.Run(ctx,
		"redis:7",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections"),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	opt, err := redis.ParseURL(connStr)
	require.NoError(t, err)

	client := redis.NewClient(opt)

	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})

	return client
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func testMinecraftWorld(id string, players int64) models.MinecraftServer {
	return models.MinecraftServer{
		ID:            id,
		UserID:        uuid.NewString(),
		Name:          "test-world",
		Address:       "mc.test.com",
		Port:          25565,
		Edition:       "java",
		PlayersOnline: players,
		PlayersMax:    100,
		Version:       "1.21.10",
		MOTD:          "Test server",
	}
}

func seedWorld(t *testing.T, cache *WorldsCache, world *models.MinecraftServer) {
	t.Helper()
	require.NoError(t, cache.SetOnline(context.Background(), world))
}

func TestSetOnline(t *testing.T) {
	client := setupTestDB(t)
	cache := NewWorldsCache(client)

	testCases := []struct {
		caseName    string
		prepare     func(t *testing.T, world *models.MinecraftServer)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "set online",
			success:  true,
		},
		{
			caseName: "update existing world",
			prepare: func(t *testing.T, world *models.MinecraftServer) {
				old := testMinecraftWorld(world.ID, 5)
				seedWorld(t, cache, &old)

				world.Name = "updated-world"
				world.PlayersOnline = 25
			},
			success: true,
		},
		{
			caseName:    "set online with canceled context",
			canceledCtx: true,
			success:     false,
		},
		{
			caseName: "set online with redis wrong type",
			prepare: func(t *testing.T, _ *models.MinecraftServer) {
				require.NoError(t, client.Set(
					context.Background(),
					onlineWorldsKey,
					"wrong-type",
					0,
				).Err())
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			world := testMinecraftWorld(uuid.NewString(), 10)

			if tc.prepare != nil {
				tc.prepare(t, &world)
			}

			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			err := cache.SetOnline(ctx, &world)

			if !tc.success {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			got, err := cache.GetByID(context.Background(), world.ID)
			require.NoError(t, err)

			assert.Equal(t, world.ID, got.ID)
			assert.Equal(t, world.Name, got.Name)
			assert.Equal(t, world.Address, got.Address)
			assert.Equal(t, world.Port, got.Port)
			assert.Equal(t, world.Edition, got.Edition)
			assert.Equal(t, world.PlayersOnline, got.PlayersOnline)
			assert.True(t, got.Online)
			assert.False(t, got.CheckedAt.IsZero())
			assert.False(t, world.Online)
			assert.True(t, world.CheckedAt.IsZero())

			ttl, err := client.TTL(
				context.Background(),
				worldKeyPrefix+world.ID,
			).Result()
			require.NoError(t, err)

			assert.Greater(t, ttl, time.Duration(0))
			assert.LessOrEqual(t, ttl, worldsCacheTTL)

			score, err := client.ZScore(
				context.Background(),
				onlineWorldsKey,
				world.ID,
			).Result()
			require.NoError(t, err)

			assert.Equal(t, float64(world.PlayersOnline), score)
		})
	}
}

func TestGetAll(t *testing.T) {
	client := setupTestDB(t)
	cache := NewWorldsCache(client)

	testCases := []struct {
		caseName    string
		prepare     func(t *testing.T)
		canceledCtx bool
		wantIDs     []string
		wantErr     bool
	}{
		{
			caseName: "get all worlds",
			prepare: func(t *testing.T) {
				world1 := testMinecraftWorld("world-1", 10)
				world2 := testMinecraftWorld("world-2", 20)

				seedWorld(t, cache, &world1)
				seedWorld(t, cache, &world2)
			},
			wantIDs: []string{"world-1", "world-2"},
		},
		{
			caseName: "get empty worlds",
			wantIDs:  []string{},
		},
		{
			caseName: "get all removes expired ids",
			prepare: func(t *testing.T) {
				require.NoError(t, client.ZAdd(
					context.Background(),
					onlineWorldsKey,
					redis.Z{
						Score:  10,
						Member: "expired-world",
					},
				).Err())
			},
			wantIDs: []string{},
		},
		{
			caseName: "get all with invalid json",
			prepare: func(t *testing.T) {
				ctx := context.Background()

				require.NoError(t, client.ZAdd(ctx, onlineWorldsKey, redis.Z{
					Score:  10,
					Member: "invalid-world",
				}).Err())

				require.NoError(t, client.Set(
					ctx,
					worldKeyPrefix+"invalid-world",
					"{invalid",
					worldsCacheTTL,
				).Err())
			},
			wantErr: true,
		},
		{
			caseName:    "get all with canceled context",
			canceledCtx: true,
			wantErr:     true,
		},
		{
			caseName: "get all with redis wrong type",
			prepare: func(t *testing.T) {
				require.NoError(t, client.Set(
					context.Background(),
					onlineWorldsKey,
					"wrong-type",
					0,
				).Err())
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			require.NoError(t, client.FlushDB(context.Background()).Err())

			if tc.prepare != nil {
				tc.prepare(t)
			}

			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			got, err := cache.GetAll(ctx)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)

			ids := make([]string, 0, len(got))
			for _, world := range got {
				ids = append(ids, world.ID)
				assert.True(t, world.Online)
			}

			assert.ElementsMatch(t, tc.wantIDs, ids)

			if tc.caseName == "get all removes expired ids" {
				count, err := client.ZCard(context.Background(), onlineWorldsKey).Result()
				require.NoError(t, err)
				assert.EqualValues(t, 0, count)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	client := setupTestDB(t)
	cache := NewWorldsCache(client)

	testCases := []struct {
		caseName    string
		worldID     string
		prepare     func(t *testing.T)
		canceledCtx bool
		wantErr     error
		success     bool
	}{
		{
			caseName: "get world by id",
			worldID:  "world-1",
			prepare: func(t *testing.T) {
				world := testMinecraftWorld("world-1", 10)
				seedWorld(t, cache, &world)
			},
			success: true,
		},
		{
			caseName: "get not found world",
			worldID:  "not-found",
			wantErr:  domain.ErrNotFound,
			success:  false,
		},
		{
			caseName: "get world with invalid json",
			worldID:  "invalid-world",
			prepare: func(t *testing.T) {
				require.NoError(t, client.Set(
					context.Background(),
					worldKeyPrefix+"invalid-world",
					"{invalid",
					worldsCacheTTL,
				).Err())
			},
			success: false,
		},
		{
			caseName: "get world with canceled context",
			worldID:  "world-2",
			prepare: func(t *testing.T) {
				world := testMinecraftWorld("world-2", 5)
				seedWorld(t, cache, &world)
			},
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}

			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			got, err := cache.GetByID(ctx, tc.worldID)

			if tc.success {
				require.NoError(t, err)
				require.NotNil(t, got)

				assert.Equal(t, tc.worldID, got.ID)
				assert.Equal(t, "test-world", got.Name)
				assert.True(t, got.Online)
				assert.False(t, got.CheckedAt.IsZero())
				return
			}

			require.Error(t, err)
			assert.Nil(t, got)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NotErrorIs(t, err, domain.ErrNotFound)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	client := setupTestDB(t)
	cache := NewWorldsCache(client)

	testCases := []struct {
		caseName    string
		worldID     string
		prepare     func(t *testing.T)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "remove world",
			worldID:  "world-1",
			prepare: func(t *testing.T) {
				world := testMinecraftWorld("world-1", 10)
				seedWorld(t, cache, &world)
			},
			success: true,
		},
		{
			caseName: "remove not found world",
			worldID:  "not-found",
			success:  true,
		},
		{
			caseName: "remove world keeps other worlds",
			worldID:  "world-2",
			prepare: func(t *testing.T) {
				world1 := testMinecraftWorld("world-2", 10)
				world2 := testMinecraftWorld("world-3", 20)

				seedWorld(t, cache, &world1)
				seedWorld(t, cache, &world2)
			},
			success: true,
		},
		{
			caseName: "remove world with canceled context",
			worldID:  "world-4",
			prepare: func(t *testing.T) {
				world := testMinecraftWorld("world-4", 10)
				seedWorld(t, cache, &world)
			},
			canceledCtx: true,
			success:     false,
		},
		{
			caseName: "remove world with redis wrong type",
			worldID:  "world-5",
			prepare: func(t *testing.T) {
				ctx := context.Background()

				world := testMinecraftWorld("world-5", 10)
				seedWorld(t, cache, &world)

				require.NoError(t, client.Set(
					ctx,
					onlineWorldsKey,
					"wrong-type",
					0,
				).Err())
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}

			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			err := cache.Remove(ctx, tc.worldID)

			if !tc.success {
				require.Error(t, err)

				if tc.canceledCtx {
					_, getErr := cache.GetByID(context.Background(), tc.worldID)
					require.NoError(t, getErr)
				}
				return
			}

			require.NoError(t, err)

			_, err = cache.GetByID(context.Background(), tc.worldID)
			require.ErrorIs(t, err, domain.ErrNotFound)

			_, err = client.ZScore(
				context.Background(),
				onlineWorldsKey,
				tc.worldID,
			).Result()
			require.ErrorIs(t, err, redis.Nil)

			if tc.caseName == "remove world keeps other worlds" {
				other, err := cache.GetByID(context.Background(), "world-3")
				require.NoError(t, err)
				assert.Equal(t, "world-3", other.ID)

				worlds, err := cache.GetAll(context.Background())
				require.NoError(t, err)
				require.Len(t, worlds, 1)
				assert.Equal(t, "world-3", worlds[0].ID)
			}
		})
	}
}
