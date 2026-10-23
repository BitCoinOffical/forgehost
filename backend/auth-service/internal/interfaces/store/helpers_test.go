package store

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	redistc "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()
	container, err := redistc.Run(ctx,
		"redis:7", testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections"),
		))
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
