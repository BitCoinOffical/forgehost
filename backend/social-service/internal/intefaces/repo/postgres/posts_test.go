package postgres

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func postIDs(posts []models.FeedPost) []string {
	ids := make([]string, 0, len(posts))
	for _, p := range posts {
		ids = append(ids, p.PostID)
	}
	return ids
}

func TestGetPostByID(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName  string
		f         func(t *testing.T) (postID, userID, topicTitle string)
		wantLikes int
		wantErr   error
		success   bool
	}{
		{
			caseName: "get post with topic and likes",
			f: func(t *testing.T) (string, string, string) {
				author, liker := newUser(t, pool), newUser(t, pool)
				topicID, title := seedTopic(t, pool, author)
				postID := seedPost(t, pool, author, &topicID)
				seedPostLike(t, pool, liker, postID)
				return postID, author, title
			},
			wantLikes: 1,
			success:   true,
		},
		{
			caseName: "get post without topic",
			f: func(t *testing.T) (string, string, string) {
				author := newUser(t, pool)
				return newPost(t, pool, author), author, ""
			},
			success: true,
		},
		{
			caseName: "get deleted post",
			f: func(t *testing.T) (string, string, string) {
				author := newUser(t, pool)
				postID := newPost(t, pool, author)
				softDeletePost(t, pool, postID)
				return postID, author, ""
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "get not found post",
			f: func(t *testing.T) (string, string, string) {
				return uuid.NewString(), "", ""
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "get post with invalid id",
			f: func(t *testing.T) (string, string, string) {
				return "invalid_uuid", "", ""
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			postID, userID, topicTitle := tc.f(t)

			post, err := repo.GetPostByID(context.Background(), postID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				assert.Nil(t, post)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, postID, post.PostID)
			assert.Equal(t, userID, post.UserID.String())
			require.NotNil(t, post.Username)
			assert.Equal(t, "user_"+userID[:8], *post.Username)
			assert.Equal(t, tc.wantLikes, post.LikeCount)
			if topicTitle != "" {
				require.NotNil(t, post.TopicID)
				require.NotNil(t, post.TopicName)
				assert.Equal(t, topicTitle, *post.TopicName)
			} else {
				assert.Nil(t, post.TopicID)
				assert.Nil(t, post.TopicName)
			}
		})
	}
}

func TestViewPost(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) string
		wantErr  error
		success  bool
	}{
		{
			caseName: "view post",
			f:        func(t *testing.T) string { return newPost(t, pool, newUser(t, pool)) },
			success:  true,
		},
		{
			caseName: "view deleted post",
			f: func(t *testing.T) string {
				postID := newPost(t, pool, newUser(t, pool))
				softDeletePost(t, pool, postID)
				return postID
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "view not found post",
			f:        func(t *testing.T) string { return uuid.NewString() },
			wantErr:  domain.ErrNotFound,
		},
		{
			caseName: "view post with invalid id",
			f:        func(t *testing.T) string { return "invalid_uuid" },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			postID := tc.f(t)

			err := repo.ViewPost(context.Background(), postID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, postViews(t, pool, postID))
		})
	}
}

func TestGetSubPosts(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (userID string, wantByUser, wantByTopic []string)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "subscribed to user and topic",
			f: func(t *testing.T) (string, []string, []string) {
				reader, author, topicAuthor := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				authorTopicID, _ := seedTopic(t, pool, author)
				userPost := seedPost(t, pool, author, &authorTopicID)
				seedSubscription(t, pool, reader, &author, nil)

				topicID, _ := seedTopic(t, pool, topicAuthor)
				topicPost := seedPost(t, pool, topicAuthor, &topicID)
				seedSubscription(t, pool, reader, nil, &topicID)
				return reader, []string{userPost}, []string{topicPost}
			},
			success: true,
		},
		{
			caseName: "ignores unsubscribed users and topics",
			f: func(t *testing.T) (string, []string, []string) {
				reader, author, stranger := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				topicID, _ := seedTopic(t, pool, author)
				mine := seedPost(t, pool, author, &topicID)
				seedPost(t, pool, stranger, &topicID)
				seedSubscription(t, pool, reader, &author, nil)
				return reader, []string{mine}, nil
			},
			success: true,
		},
		{
			caseName: "excludes deleted and old posts",
			f: func(t *testing.T) (string, []string, []string) {
				reader, author := newUser(t, pool), newUser(t, pool)
				topicID, _ := seedTopic(t, pool, author)
				fresh := seedPost(t, pool, author, &topicID)
				deleted := seedPost(t, pool, author, &topicID)
				old := seedPost(t, pool, author, &topicID)
				softDeletePost(t, pool, deleted)
				agePost(t, pool, old, 8)
				seedSubscription(t, pool, reader, &author, nil)
				return reader, []string{fresh}, nil
			},
			success: true,
		},
		{
			caseName: "no subscriptions",
			f: func(t *testing.T) (string, []string, []string) {
				return newUser(t, pool), nil, nil
			},
			success: true,
		},
		{
			caseName: "get sub posts with canceled context",
			f: func(t *testing.T) (string, []string, []string) {
				return newUser(t, pool), nil, nil
			},
			canceledCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, wantByUser, wantByTopic := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			byUser, byTopic, err := repo.GetSubPosts(ctx, userID)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, byUser)
				assert.Nil(t, byTopic)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, wantByUser, postIDs(byUser))
			assert.ElementsMatch(t, wantByTopic, postIDs(byTopic))
		})
	}
}

func TestGetGlobalPosts(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	future := func(days, sec int) time.Time {
		return time.Now().Add(time.Duration(days)*24*time.Hour + time.Duration(sec)*time.Second)
	}

	type want struct {
		cursorID *uuid.UUID
		cursorAt *time.Time
		prefix   []string
		absent   []string
		length   int
	}

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) want
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "newest first",
			f: func(t *testing.T) want {
				author := newUser(t, pool)
				p1, p2, p3 := newPost(t, pool, author), newPost(t, pool, author), newPost(t, pool, author)
				setPostCreatedAt(t, pool, p1, future(10, 1))
				setPostCreatedAt(t, pool, p2, future(10, 2))
				setPostCreatedAt(t, pool, p3, future(10, 3))
				return want{prefix: []string{p3, p2, p1}}
			},
			success: true,
		},
		{
			caseName: "limit 10",
			f: func(t *testing.T) want {
				author := newUser(t, pool)
				for range 11 {
					newPost(t, pool, author)
				}
				return want{length: 10}
			},
			success: true,
		},
		{
			caseName: "cursor returns older posts",
			f: func(t *testing.T) want {
				author := newUser(t, pool)
				p1, p2, p3 := newPost(t, pool, author), newPost(t, pool, author), newPost(t, pool, author)
				setPostCreatedAt(t, pool, p1, future(30, 1))
				setPostCreatedAt(t, pool, p2, future(30, 2))
				setPostCreatedAt(t, pool, p3, future(30, 3))
				id := uuid.MustParse(p2)
				at := postCreatedAt(t, pool, p2)
				return want{cursorID: &id, cursorAt: &at, prefix: []string{p1}, absent: []string{p2, p3}}
			},
			success: true,
		},
		{
			caseName: "excludes deleted posts",
			f: func(t *testing.T) want {
				author := newUser(t, pool)
				p1, p2 := newPost(t, pool, author), newPost(t, pool, author)
				setPostCreatedAt(t, pool, p1, future(40, 1))
				setPostCreatedAt(t, pool, p2, future(40, 2))
				softDeletePost(t, pool, p2)
				return want{prefix: []string{p1}, absent: []string{p2}}
			},
			success: true,
		},
		{
			caseName: "get global posts with canceled context",
			f: func(t *testing.T) want {
				return want{}
			},
			canceledCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			w := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			posts, err := repo.GetGlobalPosts(ctx, w.cursorID, w.cursorAt)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, posts)
				return
			}
			require.NoError(t, err)

			ids := postIDs(posts)
			if w.length > 0 {
				assert.Len(t, ids, w.length)
			}
			if len(w.prefix) > 0 {
				require.GreaterOrEqual(t, len(ids), len(w.prefix))
				assert.Equal(t, w.prefix, ids[:len(w.prefix)])
			}
			for _, id := range w.absent {
				assert.NotContains(t, ids, id)
			}
		})
	}
}

func TestGetTopics(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (wantID int, wantTitle, wantUserID string, absentID int)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "list topics",
			f: func(t *testing.T) (int, string, string, int) {
				userID := newUser(t, pool)
				id, title := seedTopic(t, pool, userID)
				return id, title, userID, 0
			},
			success: true,
		},
		{
			caseName: "list topics excludes deleted",
			f: func(t *testing.T) (int, string, string, int) {
				id, _ := seedTopic(t, pool, newUser(t, pool))
				softDeleteTopic(t, pool, id)
				return 0, "", "", id
			},
			success: true,
		},
		{
			caseName: "list topics with canceled context",
			f: func(t *testing.T) (int, string, string, int) {
				return 0, "", "", 0
			},
			canceledCtx: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			wantID, wantTitle, wantUserID, absentID := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			topics, err := repo.GetTopics(ctx)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, topics)
				return
			}
			require.NoError(t, err)

			found := make(map[int][2]string, len(topics))
			for _, tp := range topics {
				found[tp.ID] = [2]string{tp.Title, tp.UserID}
			}
			if wantID != 0 {
				require.Contains(t, found, wantID)
				assert.Equal(t, [2]string{wantTitle, wantUserID}, found[wantID])
			}
			if absentID != 0 {
				assert.NotContains(t, found, absentID)
			}
		})
	}
}

func TestGetTopicsByID(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (input string, wantID int, title, userID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "get topic",
			f: func(t *testing.T) (string, int, string, string) {
				userID := newUser(t, pool)
				id, title := seedTopic(t, pool, userID)
				return strconv.Itoa(id), id, title, userID
			},
			success: true,
		},
		{
			caseName: "get deleted topic",
			f: func(t *testing.T) (string, int, string, string) {
				id, _ := seedTopic(t, pool, newUser(t, pool))
				softDeleteTopic(t, pool, id)
				return strconv.Itoa(id), 0, "", ""
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "get not found topic",
			f: func(t *testing.T) (string, int, string, string) {
				return strconv.Itoa(missingTopicID), 0, "", ""
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "get topic with invalid id",
			f: func(t *testing.T) (string, int, string, string) {
				return "not_a_number", 0, "", ""
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			input, wantID, title, userID := tc.f(t)

			topic, err := repo.GetTopicsByID(context.Background(), input)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				assert.Nil(t, topic)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, wantID, topic.ID)
			assert.Equal(t, title, topic.Title)
			assert.Equal(t, userID, topic.UserID)
		})
	}
}

func TestCreateTopic(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) *models.Topics
		wantErr  error
		success  bool
	}{
		{
			caseName: "create topic",
			f: func(t *testing.T) *models.Topics {
				return &models.Topics{UserID: newUser(t, pool), Title: "title_" + uuid.NewString()[:8]}
			},
			success: true,
		},
		{
			caseName: "create duplicate topic",
			f: func(t *testing.T) *models.Topics {
				topic := &models.Topics{UserID: newUser(t, pool), Title: "title_" + uuid.NewString()[:8]}
				require.NoError(t, repo.CreateTopic(context.Background(), topic))
				return topic
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "create topic with invalid user id",
			f: func(t *testing.T) *models.Topics {
				return &models.Topics{UserID: "invalid_uuid", Title: "title_" + uuid.NewString()[:8]}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			topic := tc.f(t)

			err := repo.CreateTopic(context.Background(), topic)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool,
				`SELECT COUNT(*) FROM topics WHERE user_id = $1 AND title = $2`, topic.UserID, topic.Title))
		})
	}
}

func TestDeleteTopic(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) *models.Topics
		wantDeleted bool
		wantErr     error
		success     bool
	}{
		{
			caseName: "delete topic",
			f: func(t *testing.T) *models.Topics {
				userID := newUser(t, pool)
				id, _ := seedTopic(t, pool, userID)
				return &models.Topics{ID: id, UserID: userID}
			},
			wantDeleted: true,
			success:     true,
		},
		{
			caseName: "delete topic of other user",
			f: func(t *testing.T) *models.Topics {
				id, _ := seedTopic(t, pool, newUser(t, pool))
				return &models.Topics{ID: id, UserID: newUser(t, pool)}
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "delete not found topic",
			f: func(t *testing.T) *models.Topics {
				return &models.Topics{ID: missingTopicID, UserID: newUser(t, pool)}
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			topic := tc.f(t)

			err := repo.DeleteTopic(context.Background(), topic)
			if tc.success {
				require.NoError(t, err)
			} else {
				requireErr(t, err, tc.wantErr)
			}
			n := countRows(t, pool, `SELECT COUNT(*) FROM topics WHERE id = $1 AND is_delete = true`, topic.ID)
			assert.Equal(t, tc.wantDeleted, n == 1)
		})
	}
}

func TestReportTopic(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) *models.TopicsReport
		wantErr  error
		success  bool
	}{
		{
			caseName: "report topic",
			f: func(t *testing.T) *models.TopicsReport {
				topicID, _ := seedTopic(t, pool, newUser(t, pool))
				return &models.TopicsReport{UserID: newUser(t, pool), TopicID: topicID, Cause: "spam"}
			},
			success: true,
		},
		{
			caseName: "report topic twice",
			f: func(t *testing.T) *models.TopicsReport {
				topicID, _ := seedTopic(t, pool, newUser(t, pool))
				report := &models.TopicsReport{UserID: newUser(t, pool), TopicID: topicID, Cause: "spam"}
				require.NoError(t, repo.ReportTopic(context.Background(), report))
				return report
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "report not found topic",
			f: func(t *testing.T) *models.TopicsReport {
				return &models.TopicsReport{UserID: newUser(t, pool), TopicID: missingTopicID, Cause: "spam"}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			report := tc.f(t)

			err := repo.ReportTopic(context.Background(), report)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool,
				`SELECT COUNT(*) FROM topic_reports WHERE user_id = $1 AND topic_id = $2`, report.UserID, report.TopicID))
		})
	}
}

func TestCreatePost(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) *models.Post
		success  bool
	}{
		{
			caseName: "create post with topic",
			f: func(t *testing.T) *models.Post {
				userID := newUser(t, pool)
				topicID, _ := seedTopic(t, pool, userID)
				return &models.Post{TopicID: &topicID, UserID: userID,
					ImageURL: strPtr("https://img.test/1"), Description: strPtr("desc_" + uuid.NewString())}
			},
			success: true,
		},
		{
			caseName: "create post without topic",
			f: func(t *testing.T) *models.Post {
				return &models.Post{UserID: newUser(t, pool), Description: strPtr("desc_" + uuid.NewString())}
			},
			success: true,
		},
		{
			caseName: "create post with not found topic",
			f: func(t *testing.T) *models.Post {
				topicID := missingTopicID
				return &models.Post{TopicID: &topicID, UserID: newUser(t, pool), Description: strPtr("desc")}
			},
		},
		{
			caseName: "create post with invalid user id",
			f: func(t *testing.T) *models.Post {
				return &models.Post{UserID: "invalid_uuid", Description: strPtr("desc")}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			post := tc.f(t)

			err := repo.CreatePost(context.Background(), post)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool,
				`SELECT COUNT(*) FROM posts WHERE user_id = $1 AND description = $2`, post.UserID, *post.Description))
		})
	}
}

func TestBuildUpdatePost(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName        string
		f               func(t *testing.T) *models.Post
		wantDescription string
		success         bool
	}{
		{
			caseName: "update description and image",
			f: func(t *testing.T) *models.Post {
				userID := newUser(t, pool)
				return &models.Post{ID: newPost(t, pool, userID), UserID: userID,
					ImageURL: strPtr("https://img.test/new"), Description: strPtr("new")}
			},
			wantDescription: "new",
			success:         true,
		},
		{
			caseName: "update only description",
			f: func(t *testing.T) *models.Post {
				userID := newUser(t, pool)
				return &models.Post{ID: newPost(t, pool, userID), UserID: userID, Description: strPtr("only_desc")}
			},
			wantDescription: "only_desc",
			success:         true,
		},
		{
			caseName: "update post of other user",
			f: func(t *testing.T) *models.Post {
				id := newPost(t, pool, newUser(t, pool))
				return &models.Post{ID: id, UserID: newUser(t, pool), Description: strPtr("hacked")}
			},
			wantDescription: "post_description",
		},
		{
			caseName: "update not found post",
			f: func(t *testing.T) *models.Post {
				return &models.Post{ID: uuid.NewString(), UserID: newUser(t, pool), Description: strPtr("new")}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			post := tc.f(t)

			updated, err := repo.BuildUpdatePost(context.Background(), post)
			if tc.success {
				require.NoError(t, err)
				require.NotNil(t, updated)
				assert.Equal(t, post.UserID, updated.UserID)
				require.NotNil(t, updated.Description)
				assert.Equal(t, tc.wantDescription, *updated.Description)
			} else {
				require.Error(t, err)
				assert.Nil(t, updated)
			}
			if tc.wantDescription != "" {
				assert.Equal(t, tc.wantDescription, postDescription(t, pool, post.ID))
			}
		})
	}
}

func TestDeletePost(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (postID, userID string)
		wantDeleted bool
		wantErr     error
		success     bool
	}{
		{
			caseName: "delete post",
			f: func(t *testing.T) (string, string) {
				userID := newUser(t, pool)
				return newPost(t, pool, userID), userID
			},
			wantDeleted: true,
			success:     true,
		},
		{
			caseName: "delete post of other user",
			f: func(t *testing.T) (string, string) {
				return newPost(t, pool, newUser(t, pool)), newUser(t, pool)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "delete not found post",
			f: func(t *testing.T) (string, string) {
				return uuid.NewString(), newUser(t, pool)
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			postID, userID := tc.f(t)

			err := repo.DeletePost(context.Background(), postID, userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				requireErr(t, err, tc.wantErr)
			}
			n := countRows(t, pool, `SELECT COUNT(*) FROM posts WHERE id = $1 AND is_delete = true`, postID)
			assert.Equal(t, tc.wantDeleted, n == 1)
		})
	}
}

func TestReportPost(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) *models.PostReport
		wantErr  error
		success  bool
	}{
		{
			caseName: "report post",
			f: func(t *testing.T) *models.PostReport {
				postID := newPost(t, pool, newUser(t, pool))
				return &models.PostReport{UserID: newUser(t, pool), PostID: postID, Cause: "spam"}
			},
			success: true,
		},
		{
			caseName: "report post twice",
			f: func(t *testing.T) *models.PostReport {
				postID := newPost(t, pool, newUser(t, pool))
				report := &models.PostReport{UserID: newUser(t, pool), PostID: postID, Cause: "spam"}
				require.NoError(t, repo.ReportPost(context.Background(), report))
				return report
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "report not found post",
			f: func(t *testing.T) *models.PostReport {
				return &models.PostReport{UserID: newUser(t, pool), PostID: uuid.NewString(), Cause: "spam"}
			},
		},
		{
			caseName: "report post with invalid id",
			f: func(t *testing.T) *models.PostReport {
				return &models.PostReport{UserID: newUser(t, pool), PostID: "invalid_uuid", Cause: "spam"}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			report := tc.f(t)

			err := repo.ReportPost(context.Background(), report)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool,
				`SELECT COUNT(*) FROM post_reports WHERE user_id = $1 AND post_id = $2`, report.UserID, report.PostID))
		})
	}
}

func TestLikePost(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (userID, postID string)
		wantErr  error
		success  bool
	}{
		{
			caseName: "like post",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), newPost(t, pool, newUser(t, pool))
			},
			success: true,
		},
		{
			caseName: "like post twice",
			f: func(t *testing.T) (string, string) {
				liker, postID := newUser(t, pool), newPost(t, pool, newUser(t, pool))
				require.NoError(t, repo.LikePost(context.Background(), liker, postID))
				return liker, postID
			},
			wantErr: domain.ErrAlreadyExists,
		},
		{
			caseName: "like not found post",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), uuid.NewString()
			},
		},
		{
			caseName: "like post with invalid id",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), "invalid_uuid"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, postID := tc.f(t)

			err := repo.LikePost(context.Background(), userID, postID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 1, countRows(t, pool,
				`SELECT COUNT(*) FROM post_likes WHERE user_id = $1 AND post_id = $2`, userID, postID))
		})
	}
}

func TestUnlikePost(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewPostsRepo(pool)

	testCases := []struct {
		caseName  string
		f         func(t *testing.T) (userID, postID string)
		wantLikes int
		wantErr   error
		success   bool
	}{
		{
			caseName: "unlike post",
			f: func(t *testing.T) (string, string) {
				liker, postID := newUser(t, pool), newPost(t, pool, newUser(t, pool))
				seedPostLike(t, pool, liker, postID)
				return liker, postID
			},
			wantLikes: 0,
			success:   true,
		},
		{
			caseName: "unlike keeps likes of other users",
			f: func(t *testing.T) (string, string) {
				liker, other, postID := newUser(t, pool), newUser(t, pool), newPost(t, pool, newUser(t, pool))
				seedPostLike(t, pool, liker, postID)
				seedPostLike(t, pool, other, postID)
				return liker, postID
			},
			wantLikes: 1,
			success:   true,
		},
		{
			caseName: "unlike not liked post",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), newPost(t, pool, newUser(t, pool))
			},
			wantErr: domain.ErrNotFound,
		},
		{
			caseName: "unlike post with invalid id",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), "invalid_uuid"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, postID := tc.f(t)

			err := repo.UnlikePost(context.Background(), userID, postID)
			if !tc.success {
				requireErr(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantLikes,
				countRows(t, pool, `SELECT COUNT(*) FROM post_likes WHERE post_id = $1`, postID))
			assert.Equal(t, 0,
				countRows(t, pool, `SELECT COUNT(*) FROM post_likes WHERE user_id = $1 AND post_id = $2`, userID, postID))
		})
	}
}
