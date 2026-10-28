package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/adapters/migrations"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:18",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
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

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func strPtr(s string) *string {
	return &s
}

func requireErr(t *testing.T, err, want error) {
	t.Helper()
	require.Error(t, err)
	if want != nil {
		require.ErrorIs(t, err, want)
	}
}

func newUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	id := uuid.NewString()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO profiles (user_id, username, avatar_url) VALUES ($1, $2, $3)`,
		id, "user_"+id[:8], "https://avatar.test/"+id[:8])
	require.NoError(t, err)
	return id
}

func seedChat(t *testing.T, pool *pgxpool.Pool, repo *ChatsRepo) (chatID, ownerID string) {
	t.Helper()
	ownerID = newUser(t, pool)
	chatID, err := repo.CreateGroup(context.Background(), ownerID,
		&models.Chat{Title: strPtr("title"), Description: strPtr("description")})
	require.NoError(t, err)
	return chatID, ownerID
}

func seedMember(t *testing.T, pool *pgxpool.Pool, repo *ChatsRepo, chatID, ownerID string, role models.MemberRole) string {
	t.Helper()
	ctx := context.Background()
	userID := newUser(t, pool)
	require.NoError(t, repo.AddUserInChat(ctx, chatID, userID))
	if role != "member" {
		_, err := repo.SetUserRoleChat(ctx, ownerID, &models.ChatMember{ChatID: chatID, UserID: userID, Role: role})
		require.NoError(t, err)
	}
	return userID
}

func memberIDs(t *testing.T, repo *ChatsRepo, chatID string) []string {
	t.Helper()
	members, err := repo.GetUsersFromChat(context.Background(), chatID)
	require.NoError(t, err)
	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	return ids
}

func seedBlock(t *testing.T, pool *pgxpool.Pool, userID, targetID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO blocks (user_id, target_id) VALUES ($1, $2)`, userID, targetID)
	require.NoError(t, err)
}

func directUpdatedAt(t *testing.T, pool *pgxpool.Pool, chatID string) time.Time {
	t.Helper()
	var updatedAt time.Time
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT updated_at FROM directs WHERE id = $1`, chatID).Scan(&updatedAt))
	return updatedAt
}

func lastReadAt(t *testing.T, pool *pgxpool.Pool, chatID, userID string) time.Time {
	t.Helper()
	var readAt time.Time
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT last_read_at FROM chat_reads WHERE chat_id = $1 AND user_id = $2`, chatID, userID).Scan(&readAt))
	return readAt
}

type commentStats struct {
	likes   int
	replies int
}

func newPost(t *testing.T, pool *pgxpool.Pool, userID string) string {
	t.Helper()
	var id string
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO posts (user_id, description) VALUES ($1, $2) RETURNING id`, userID, "post_description").Scan(&id))
	return id
}

func seedComment(t *testing.T, pool *pgxpool.Pool, postID, userID string, parentID *string, body string) string {
	t.Helper()
	var id string
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO comments (post_id, user_id, parent_id, body) VALUES ($1, $2, $3, $4) RETURNING id`,
		postID, userID, parentID, body).Scan(&id))
	return id
}

func seedLike(t *testing.T, pool *pgxpool.Pool, userID, commentID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO comment_likes (user_id, comment_id) VALUES ($1, $2)`, userID, commentID)
	require.NoError(t, err)
}

func countRows(t *testing.T, pool *pgxpool.Pool, query string, args ...any) int {
	t.Helper()
	var n int
	require.NoError(t, pool.QueryRow(context.Background(), query, args...).Scan(&n))
	return n
}

func commentBody(t *testing.T, pool *pgxpool.Pool, id string) string {
	t.Helper()
	var body string
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT body FROM comments WHERE id = $1`, id).Scan(&body))
	return body
}

func seedChatBan(t *testing.T, pool *pgxpool.Pool, chatID, targetID, userID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO chat_bans (chat_id, user_id, banned_by) VALUES ($1, $2, $3)`, chatID, targetID, userID)
	require.NoError(t, err)
}

const missingTopicID = 1 << 40

func seedTopic(t *testing.T, pool *pgxpool.Pool, userID string) (id int, title string) {
	t.Helper()
	title = "topic_" + uuid.NewString()[:8]
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO topics (user_id, title) VALUES ($1, $2) RETURNING id`, userID, title).Scan(&id))
	return id, title
}

func seedPost(t *testing.T, pool *pgxpool.Pool, userID string, topicID *int) string {
	t.Helper()
	var id string
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO posts (topic_id, user_id, description) VALUES ($1, $2, $3) RETURNING id`,
		topicID, userID, "post_description").Scan(&id))
	return id
}

func seedPostLike(t *testing.T, pool *pgxpool.Pool, userID, postID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO post_likes (user_id, post_id) VALUES ($1, $2)`, userID, postID)
	require.NoError(t, err)
}

func seedSubscription(t *testing.T, pool *pgxpool.Pool, userID string, targetUserID *string, targetTopicID *int) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO subscriptions (user_id, target_user_id, target_topic_id) VALUES ($1, $2, $3)`,
		userID, targetUserID, targetTopicID)
	require.NoError(t, err)
}

func softDeletePost(t *testing.T, pool *pgxpool.Pool, postID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `UPDATE posts SET is_delete = true WHERE id = $1`, postID)
	require.NoError(t, err)
}

func softDeleteTopic(t *testing.T, pool *pgxpool.Pool, topicID int) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `UPDATE topics SET is_delete = true WHERE id = $1`, topicID)
	require.NoError(t, err)
}

func agePost(t *testing.T, pool *pgxpool.Pool, postID string, days int) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE posts SET created_at = NOW() - make_interval(days => $2) WHERE id = $1`, postID, days)
	require.NoError(t, err)
}

func setPostCreatedAt(t *testing.T, pool *pgxpool.Pool, postID string, at time.Time) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `UPDATE posts SET created_at = $2 WHERE id = $1`, postID, at.UTC())
	require.NoError(t, err)
}

func postCreatedAt(t *testing.T, pool *pgxpool.Pool, postID string) time.Time {
	t.Helper()
	var at time.Time
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT created_at FROM posts WHERE id = $1`, postID).Scan(&at))
	return at
}

func postDescription(t *testing.T, pool *pgxpool.Pool, postID string) string {
	t.Helper()
	var d string
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT description FROM posts WHERE id = $1`, postID).Scan(&d))
	return d
}

func postViews(t *testing.T, pool *pgxpool.Pool, postID string) int {
	t.Helper()
	return countRows(t, pool, `SELECT views FROM posts WHERE id = $1`, postID)
}

func banProfile(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `UPDATE profiles SET is_banned = true WHERE user_id = $1`, userID)
	require.NoError(t, err)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
func seedStream(t *testing.T, pool *pgxpool.Pool, userID, videoID, status string) string {
	t.Helper()
	var id string
	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO streams (user_id, video_id, title, status, embed_url, viewers, likes)
		 VALUES ($1, $2, 'title', $3::stream_status, $4, 0, 0) RETURNING id`,
		userID, videoID, status, "https://embed.test/"+videoID).Scan(&id))
	return id
}
func setChannel(t *testing.T, pool *pgxpool.Pool, userID, channelID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE profiles SET channel_id = $1 WHERE user_id = $2`, channelID, userID)
	require.NoError(t, err)
}

func channelOf(t *testing.T, pool *pgxpool.Pool, userID string) *string {
	t.Helper()
	var ch *string
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT channel_id FROM profiles WHERE user_id = $1`, userID).Scan(&ch))
	return ch
}

func newChannelID() string {
	return "UC" + uuid.NewString()[:22]
}

func newVideoID() string {
	return "vid_" + uuid.NewString()[:8]
}

func seedUpload(t *testing.T, pool *pgxpool.Pool, imageID, userID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO uploads (image_id, user_id) VALUES ($1, $2)`, imageID, userID)
	require.NoError(t, err)
}
func postIDs(posts []models.FeedPost) []string {
	ids := make([]string, 0, len(posts))
	for _, p := range posts {
		ids = append(ids, p.PostID)
	}
	return ids
}
