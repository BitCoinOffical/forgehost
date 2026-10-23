package store

import (
	"context"
	"fmt"
	"testing"

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
func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func seedMessages(ctx context.Context, store *StreamStore, videoID string, ids ...string) error {
	for _, id := range ids {
		if err := store.SaveStreamMessage(ctx, videoID, &models.Message{ChatID: id}); err != nil {
			return err
		}
	}
	return nil
}

func TestSaveStreamMessage(t *testing.T) {
	client := setupTestDB(t)
	store := NewStreamStore(client)
	testCases := []struct {
		caseName    string
		videoID     string
		count       int
		wantLen     int
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "save stream message",
			videoID:  "video_1",
			count:    1,
			wantLen:  1,
			success:  true,
		},
		{
			caseName: "save several messages keeps order",
			videoID:  "video_2",
			count:    5,
			wantLen:  5,
			success:  true,
		},
		{
			caseName: "save exactly 200 messages",
			videoID:  "video_3",
			count:    200,
			wantLen:  200,
			success:  true,
		},
		{
			caseName: "save more than 200 messages trims oldest",
			videoID:  "video_4",
			count:    205,
			wantLen:  200,
			success:  true,
		},
		{
			caseName:    "save stream message with canceled context",
			videoID:     "video_5",
			count:       1,
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

			for i := 0; i < tc.count; i++ {
				err := store.SaveStreamMessage(ctx, tc.videoID, &models.Message{ChatID: fmt.Sprintf("msg_%d", i)})
				if !tc.success {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			}

			msgs, err := store.GetStreamMessages(context.Background(), tc.videoID)
			require.NoError(t, err)
			require.Len(t, msgs, tc.wantLen)
			assert.Equal(t, fmt.Sprintf("msg_%d", tc.count-tc.wantLen), msgs[0].ChatID)
			assert.Equal(t, fmt.Sprintf("msg_%d", tc.count-1), msgs[len(msgs)-1].ChatID)

			ttl, err := client.TTL(context.Background(), chatKey+tc.videoID).Result()
			require.NoError(t, err)
			assert.Greater(t, ttl, 0*messageTTL)
			assert.LessOrEqual(t, ttl, messageTTL)
		})
	}
}

func TestGetStreamMessages(t *testing.T) {
	client := setupTestDB(t)
	store := NewStreamStore(client)
	testCases := []struct {
		caseName    string
		videoID     string
		f           func(ctx context.Context) error
		wantIDs     []string
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "get stream messages",
			videoID:  "video_1",
			f: func(ctx context.Context) error {
				if err := seedMessages(ctx, store, "video_1", "msg_0", "msg_1"); err != nil {
					return err
				}
				return seedMessages(ctx, store, "video_other", "msg_other")
			},
			wantIDs: []string{"msg_0", "msg_1"},
			success: true,
		},
		{
			caseName: "get stream messages from empty chat",
			videoID:  "video_empty",
			wantIDs:  []string{},
			success:  true,
		},
		{
			caseName: "get stream messages with invalid json",
			videoID:  "video_invalid",
			f: func(ctx context.Context) error {
				return client.RPush(ctx, chatKey+"video_invalid", "{invalid").Err()
			},
			success: false,
		},
		{
			caseName: "get stream messages with canceled context",
			videoID:  "video_canceled",
			f: func(ctx context.Context) error {
				return seedMessages(ctx, store, "video_canceled", "msg_0")
			},
			canceledCtx: true,
			success:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			if tc.f != nil {
				require.NoError(t, tc.f(ctx))
			}
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			msgs, err := store.GetStreamMessages(ctx, tc.videoID)
			if tc.success {
				require.NoError(t, err)
				ids := make([]string, 0, len(msgs))
				for _, m := range msgs {
					ids = append(ids, m.ChatID)
				}
				assert.Equal(t, tc.wantIDs, ids)
			} else {
				require.Error(t, err)
				assert.Nil(t, msgs)
			}
		})
	}
}

func TestSaveStreamUser(t *testing.T) {
	client := setupTestDB(t)
	store := NewStreamStore(client)
	testCases := []struct {
		caseName    string
		videoID     string
		userIDs     []string
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "save stream user",
			videoID:  "video_1",
			userIDs:  []string{"user_1"},
			success:  true,
		},
		{
			caseName: "save several stream users keeps order",
			videoID:  "video_2",
			userIDs:  []string{"user_1", "user_2", "user_3"},
			success:  true,
		},
		{
			caseName:    "save stream user with canceled context",
			videoID:     "video_3",
			userIDs:     []string{"user_1"},
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

			for _, userID := range tc.userIDs {
				err := store.SaveStreamUser(ctx, userID, tc.videoID)
				if !tc.success {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			}

			users, err := store.GetStreamUsers(context.Background(), tc.videoID)
			require.NoError(t, err)
			assert.Equal(t, tc.userIDs, users)
		})
	}
}

func TestGetStreamUsers(t *testing.T) {
	client := setupTestDB(t)
	store := NewStreamStore(client)
	testCases := []struct {
		caseName    string
		videoID     string
		f           func(ctx context.Context) error
		wantUsers   []string
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "get stream users",
			videoID:  "video_1",
			f: func(ctx context.Context) error {
				for _, userID := range []string{"user_1", "user_2"} {
					if err := store.SaveStreamUser(ctx, userID, "video_1"); err != nil {
						return err
					}
				}
				return store.SaveStreamUser(ctx, "user_other", "video_other")
			},
			wantUsers: []string{"user_1", "user_2"},
			success:   true,
		},
		{
			caseName:  "get stream users from empty stream",
			videoID:   "video_empty",
			wantUsers: []string{},
			success:   true,
		},
		{
			caseName: "get stream users with canceled context",
			videoID:  "video_canceled",
			f: func(ctx context.Context) error {
				return store.SaveStreamUser(ctx, "user_1", "video_canceled")
			},
			canceledCtx: true,
			success:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			if tc.f != nil {
				require.NoError(t, tc.f(ctx))
			}
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			users, err := store.GetStreamUsers(ctx, tc.videoID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.wantUsers, users)
			} else {
				require.Error(t, err)
				assert.Nil(t, users)
			}
		})
	}
}
