package repo

import (
	"context"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/adapters/migrations"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrestc "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupWorldsTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := postgrestc.Run(
		ctx,
		"postgres:18",
		postgrestc.WithDatabase("testdb"),
		postgrestc.WithUsername("test"),
		postgrestc.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	t.Cleanup(pool.Close)

	require.NoError(t, migrations.RunMigrations(pool))

	return pool
}

func worldsCanceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func newTestWorld(userID, address string, port int) *models.MinecraftServer {
	return &models.MinecraftServer{
		UserID:  userID,
		Name:    "test-world",
		Address: address,
		Port:    port,
		Edition: "java",
	}
}

func worldsCount(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()

	var count int
	require.NoError(t, pool.QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM servers`,
	).Scan(&count))

	return count
}

func TestUpsert(t *testing.T) {
	pool := setupWorldsTestDB(t)
	repo := NewWorldsRepo(pool)

	testCases := []struct {
		caseName string
		prepare  func(t *testing.T, world *models.MinecraftServer) *models.MinecraftServer
		wantErr  error
		success  bool
	}{
		{
			caseName: "create world",
			success:  true,
		},
		{
			caseName: "update existing world",
			prepare: func(t *testing.T, world *models.MinecraftServer) *models.MinecraftServer {
				old := *world
				old.Name = "old-world"
				old.Address = "old.test.com"

				existing, err := repo.Upsert(context.Background(), &old)
				require.NoError(t, err)

				world.Name = "updated-world"
				world.Address = "new.test.com"
				world.Port = 25566

				return existing
			},
			success: true,
		},
		{
			caseName: "create world with existing endpoint",
			prepare: func(t *testing.T, world *models.MinecraftServer) *models.MinecraftServer {
				other := *world
				other.UserID = uuid.NewString()

				existing, err := repo.Upsert(context.Background(), &other)
				require.NoError(t, err)

				return existing
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "upsert with canceled context",
			wantErr:  context.Canceled,
			success:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			_, err := pool.Exec(context.Background(), `TRUNCATE TABLE servers`)
			require.NoError(t, err)

			world := newTestWorld(uuid.NewString(), "mc.test.com", 25565)

			var existing *models.MinecraftServer
			if tc.prepare != nil {
				existing = tc.prepare(t, world)
			}

			ctx := context.Background()
			if tc.wantErr == context.Canceled {
				ctx = worldsCanceledContext()
			}

			got, err := repo.Upsert(ctx, world)

			if !tc.success {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, got)

				if tc.wantErr == domain.ErrAlreadyExists {
					assert.Equal(t, 1, worldsCount(t, pool))
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)

			assert.NotEmpty(t, got.ID)
			assert.Equal(t, world.UserID, got.UserID)
			assert.Equal(t, world.Name, got.Name)
			assert.Equal(t, world.Address, got.Address)
			assert.Equal(t, world.Port, got.Port)
			assert.Equal(t, world.Edition, got.Edition)

			require.NotEmpty(t, got.CreatedAt)
			require.NotEmpty(t, got.UpdatedAt)

			if existing != nil {
				assert.Equal(t, existing.ID, got.ID)
				assert.True(t, got.CreatedAt.Equal(existing.CreatedAt))
				assert.Equal(t, 1, worldsCount(t, pool))
			} else {
				assert.Equal(t, 1, worldsCount(t, pool))
			}

			saved, err := repo.GetByID(context.Background(), got.ID)
			require.NoError(t, err)
			assert.Equal(t, got, saved)
		})
	}
}

func TestGetAll(t *testing.T) {
	pool := setupWorldsTestDB(t)
	repo := NewWorldsRepo(pool)

	testCases := []struct {
		caseName    string
		prepare     func(t *testing.T) []string
		canceledCtx bool
		wantErr     bool
	}{
		{
			caseName: "get all worlds",
			prepare: func(t *testing.T) []string {
				world1 := newTestWorld(uuid.NewString(), "mc1.test.com", 25565)
				world2 := newTestWorld(uuid.NewString(), "mc2.test.com", 25565)

				got1, err := repo.Upsert(context.Background(), world1)
				require.NoError(t, err)

				got2, err := repo.Upsert(context.Background(), world2)
				require.NoError(t, err)

				return []string{got1.ID, got2.ID}
			},
		},
		{
			caseName: "get empty worlds",
			prepare: func(t *testing.T) []string {
				return []string{}
			},
		},
		{
			caseName:    "get all with canceled context",
			canceledCtx: true,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			_, err := pool.Exec(context.Background(), `TRUNCATE TABLE servers`)
			require.NoError(t, err)

			var wantIDs []string
			if tc.prepare != nil {
				wantIDs = tc.prepare(t)
			}

			ctx := context.Background()
			if tc.canceledCtx {
				ctx = worldsCanceledContext()
			}

			got, err := repo.GetAll(ctx)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Len(t, got, len(wantIDs))

			ids := make([]string, 0, len(got))
			for _, world := range got {
				ids = append(ids, world.ID)
				assert.NotEmpty(t, world.UserID)
				assert.NotEmpty(t, world.Name)
				assert.NotEmpty(t, world.Address)
				assert.NotEmpty(t, world.CreatedAt)
				assert.NotEmpty(t, world.UpdatedAt)
			}

			assert.ElementsMatch(t, wantIDs, ids)
		})
	}
}

func TestGetByID(t *testing.T) {
	pool := setupWorldsTestDB(t)
	repo := NewWorldsRepo(pool)

	testCases := []struct {
		caseName    string
		prepare     func(t *testing.T) string
		worldID     string
		canceledCtx bool
		wantErr     error
		success     bool
	}{
		{
			caseName: "get world by id",
			prepare: func(t *testing.T) string {
				world := newTestWorld(uuid.NewString(), "mc.test.com", 25565)

				got, err := repo.Upsert(context.Background(), world)
				require.NoError(t, err)

				return got.ID
			},
			success: true,
		},
		{
			caseName: "get not found world",
			worldID:  uuid.NewString(),
			wantErr:  domain.ErrNotFound,
			success:  false,
		},
		{
			caseName: "get world with invalid id",
			worldID:  "invalid_uuid",
			success:  false,
		},
		{
			caseName: "get world with canceled context",
			prepare: func(t *testing.T) string {
				world := newTestWorld(uuid.NewString(), "mc.test.com", 25565)

				got, err := repo.Upsert(context.Background(), world)
				require.NoError(t, err)

				return got.ID
			},
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			_, err := pool.Exec(context.Background(), `TRUNCATE TABLE servers`)
			require.NoError(t, err)

			worldID := tc.worldID
			if tc.prepare != nil {
				worldID = tc.prepare(t)
			}

			ctx := context.Background()
			if tc.canceledCtx {
				ctx = worldsCanceledContext()
			}

			got, err := repo.GetByID(ctx, worldID)

			if tc.success {
				require.NoError(t, err)
				require.NotNil(t, got)

				assert.Equal(t, worldID, got.ID)
				assert.Equal(t, "test-world", got.Name)
				assert.Equal(t, "mc.test.com", got.Address)
				assert.Equal(t, 25565, got.Port)
				assert.Equal(t, "java", got.Edition)
				assert.NotEmpty(t, got.CreatedAt)
				assert.NotEmpty(t, got.UpdatedAt)

				return
			}

			require.Error(t, err)
			assert.Nil(t, got)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NotErrorIs(t, err, domain.ErrNotFound)
			}

			if tc.canceledCtx {
				require.ErrorIs(t, err, context.Canceled)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	pool := setupWorldsTestDB(t)
	repo := NewWorldsRepo(pool)

	testCases := []struct {
		caseName    string
		prepare     func(t *testing.T) (worldID, userID string)
		wantErr     error
		stillExists bool
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "delete world",
			prepare: func(t *testing.T) (string, string) {
				world := newTestWorld(uuid.NewString(), "mc.test.com", 25565)

				got, err := repo.Upsert(context.Background(), world)
				require.NoError(t, err)

				return got.ID, world.UserID
			},
			success: true,
		},
		{
			caseName: "delete world by another user",
			prepare: func(t *testing.T) (string, string) {
				world := newTestWorld(uuid.NewString(), "mc.test.com", 25565)

				got, err := repo.Upsert(context.Background(), world)
				require.NoError(t, err)

				return got.ID, uuid.NewString()
			},
			wantErr:     domain.ErrNotFound,
			stillExists: true,
			success:     false,
		},
		{
			caseName: "delete not found world",
			prepare: func(t *testing.T) (string, string) {
				return uuid.NewString(), uuid.NewString()
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "delete world with invalid id",
			prepare: func(t *testing.T) (string, string) {
				return "invalid_uuid", uuid.NewString()
			},
			success: false,
		},
		{
			caseName: "delete world with canceled context",
			prepare: func(t *testing.T) (string, string) {
				world := newTestWorld(uuid.NewString(), "mc.test.com", 25565)

				got, err := repo.Upsert(context.Background(), world)
				require.NoError(t, err)

				return got.ID, world.UserID
			},
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			_, err := pool.Exec(context.Background(), `TRUNCATE TABLE servers`)
			require.NoError(t, err)

			worldID, userID := tc.prepare(t)

			ctx := context.Background()
			if tc.canceledCtx {
				ctx = worldsCanceledContext()
			}

			err = repo.Delete(ctx, worldID, userID)

			if !tc.success {
				require.Error(t, err)

				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}

				if tc.canceledCtx {
					require.ErrorIs(t, err, context.Canceled)
				}

				if tc.stillExists {
					got, err := repo.GetByID(context.Background(), worldID)
					require.NoError(t, err)
					assert.Equal(t, worldID, got.ID)
				}

				return
			}

			require.NoError(t, err)

			_, err = repo.GetByID(context.Background(), worldID)
			require.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}

func TestUpsertUpdatedAt(t *testing.T) {
	pool := setupWorldsTestDB(t)
	repo := NewWorldsRepo(pool)

	world := newTestWorld(uuid.NewString(), "mc.test.com", 25565)

	first, err := repo.Upsert(context.Background(), world)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	world.Name = "updated-world"

	second, err := repo.Upsert(context.Background(), world)
	require.NoError(t, err)

	assert.Equal(t, first.ID, second.ID)
	assert.True(t, second.UpdatedAt.After(first.UpdatedAt))
	assert.True(t, second.CreatedAt.Equal(first.CreatedAt))
}
