package cache

import (
	"context"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
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
func testPosts() []models.FeedPost {
	return []models.FeedPost{{}, {}}
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestSetGlobal(t *testing.T) {
	client := setupTestDB(t)
	cache := NewCache(client)
	testCases := []struct {
		caseName string
		data     struct {
			posts  []models.FeedPost
			cursor string
		}
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "set global",
			data: struct {
				posts  []models.FeedPost
				cursor string
			}{posts: testPosts(), cursor: "cursor"},
			success: true,
		},
		{
			caseName: "set global with empty cursor",
			data: struct {
				posts  []models.FeedPost
				cursor string
			}{posts: testPosts(), cursor: ""},
			success: true,
		},
		{
			caseName: "set global with canceled context",
			data: struct {
				posts  []models.FeedPost
				cursor string
			}{posts: testPosts(), cursor: "cursor"},
			canceledCtx: true,
			success:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			err := cache.SetGlobal(ctx, tc.data.posts, tc.data.cursor)
			if tc.success {
				require.NoError(t, err)

				got, err := cache.GetGlobal(context.Background(), tc.data.cursor)
				require.NoError(t, err)
				assert.Equal(t, tc.data.posts, got)

				ttl, err := client.TTL(context.Background(), postKey+tc.data.cursor).Result()
				require.NoError(t, err)
				assert.Greater(t, ttl, time.Duration(0))
				assert.LessOrEqual(t, ttl, expiration)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetGlobal(t *testing.T) {
	client := setupTestDB(t)
	cache := NewCache(client)
	testCases := []struct {
		caseName    string
		posts       []models.FeedPost
		seedCursor  string
		cursor      string
		f           func(ctx context.Context, value []models.FeedPost, cursor string) error
		canceledCtx bool
		notFound    bool
		success     bool
	}{
		{
			caseName:   "get global",
			posts:      testPosts(),
			seedCursor: "cursor",
			cursor:     "cursor",
			f:          cache.SetGlobal,
			success:    true,
		},
		{
			caseName:   "get global with empty cursor",
			posts:      testPosts(),
			seedCursor: "",
			cursor:     "",
			f:          cache.SetGlobal,
			success:    true,
		},
		{
			caseName: "get not found global",
			cursor:   "not found",
			notFound: true,
			success:  false,
		},
		{
			caseName:   "get global by other cursor",
			posts:      testPosts(),
			seedCursor: "cursor1",
			cursor:     "cursor2",
			f:          cache.SetGlobal,
			notFound:   true,
			success:    false,
		},
		{
			caseName:   "get global with invalid json",
			seedCursor: "cursor",
			cursor:     "cursor",
			f: func(ctx context.Context, _ []models.FeedPost, cursor string) error {
				return client.Set(ctx, postKey+cursor, "{invalid", 0).Err()
			},
			success: false,
		},
		{
			caseName:    "get global with canceled context",
			posts:       testPosts(),
			seedCursor:  "cursor",
			cursor:      "cursor",
			f:           cache.SetGlobal,
			canceledCtx: true,
			success:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			if tc.f != nil {
				require.NoError(t, tc.f(ctx, tc.posts, tc.seedCursor))
			}
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			got, err := cache.GetGlobal(ctx, tc.cursor)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.posts, got)
			} else {
				require.Error(t, err)
				assert.Nil(t, got)
				if tc.notFound {
					require.ErrorIs(t, err, domain.ErrNotFound)
				} else {
					assert.NotErrorIs(t, err, domain.ErrNotFound)
				}
			}
		})
	}
}

func TestDeleteGlobal(t *testing.T) {
	client := setupTestDB(t)
	cache := NewCache(client)
	testCases := []struct {
		caseName    string
		cursor      string
		otherCursor string
		f           func(ctx context.Context, value []models.FeedPost, cursor string) error
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "delete global",
			cursor:   "cursor",
			f:        cache.SetGlobal,
			success:  true,
		},
		{
			caseName: "delete not found global",
			cursor:   "not found",
			success:  true,
		},
		{
			caseName:    "delete global keeps other cursors",
			cursor:      "cursor1",
			otherCursor: "cursor2",
			f:           cache.SetGlobal,
			success:     true,
		},
		{
			caseName:    "delete global with canceled context",
			cursor:      "cursor",
			f:           cache.SetGlobal,
			canceledCtx: true,
			success:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			if tc.f != nil {
				require.NoError(t, tc.f(ctx, testPosts(), tc.cursor))
			}
			if tc.otherCursor != "" {
				require.NoError(t, cache.SetGlobal(ctx, testPosts(), tc.otherCursor))
			}

			deleteCtx := ctx
			if tc.canceledCtx {
				deleteCtx = canceledContext()
			}

			err := cache.DeleteGlobal(deleteCtx, tc.cursor)
			if tc.success {
				require.NoError(t, err)

				_, err := cache.GetGlobal(ctx, tc.cursor)
				require.ErrorIs(t, err, domain.ErrNotFound)

				if tc.otherCursor != "" {
					got, err := cache.GetGlobal(ctx, tc.otherCursor)
					require.NoError(t, err)
					assert.Equal(t, testPosts(), got)
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}
