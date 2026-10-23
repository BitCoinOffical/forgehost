package mongorepo

import (
	"context"
	"testing"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func setupTestDB(t *testing.T) *mongo.Collection {
	t.Helper()
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:7")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := mongo.Connect(options.Client().ApplyURI(connStr))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Disconnect(ctx))
	})

	return client.Database("testdb").Collection("messages")
}
func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestSaveMessage(t *testing.T) {
	coll := setupTestDB(t)
	repo := NewMessagerRepo(coll)

	testCases := []struct {
		caseName string
		data     struct {
			msg models.Message
		}
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "save message",
			data: struct {
				msg models.Message
			}{msg: models.Message{ChatID: "chat_save"}},
			success: true,
		},
		{
			caseName: "save message with canceled context",
			data: struct {
				msg models.Message
			}{msg: models.Message{ChatID: "chat_canceled"}},
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

			id, err := repo.SaveMessage(ctx, &tc.data.msg)

			saved, getErr := repo.GetMessages(context.Background(), tc.data.msg.ChatID)
			require.NoError(t, getErr)

			if tc.success {
				require.NoError(t, err)
				require.NotNil(t, id)
				assert.False(t, id.IsZero())
				assert.Len(t, saved, 1)
				return
			}
			require.Error(t, err)
			assert.Nil(t, id)
			assert.Empty(t, saved)
		})
	}
}

func TestGetMessages(t *testing.T) {
	coll := setupTestDB(t)
	repo := NewMessagerRepo(coll)

	testCases := []struct {
		caseName    string
		seed        []models.Message
		f           func(ctx context.Context, msg *models.Message) (*bson.ObjectID, error)
		chatID      string
		wantLen     int
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "get messages",
			seed: []models.Message{
				{ChatID: "chat_1"},
				{ChatID: "chat_1"},
				{ChatID: "chat_other"},
			},
			f:       repo.SaveMessage,
			chatID:  "chat_1",
			wantLen: 2,
			success: true,
		},
		{
			caseName: "get messages from chat without messages",
			chatID:   "chat_empty",
			wantLen:  0,
			success:  true,
		},
		{
			caseName: "get messages with canceled context",
			seed: []models.Message{
				{ChatID: "chat_canceled"},
			},
			f:           repo.SaveMessage,
			chatID:      "chat_canceled",
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctx := context.Background()
			if tc.f != nil {
				for i := range tc.seed {
					_, err := tc.f(ctx, &tc.seed[i])
					require.NoError(t, err)
				}
			}
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			msgs, err := repo.GetMessages(ctx, tc.chatID)
			if tc.success {
				require.NoError(t, err)
				assert.Len(t, msgs, tc.wantLen)
				for _, m := range msgs {
					assert.Equal(t, tc.chatID, m.ChatID)
				}
				return
			}
			require.Error(t, err)
			assert.Nil(t, msgs)
		})
	}
}
