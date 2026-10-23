package postgres

import (
	"context"
	"testing"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListComments(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewCommentsRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) (postID string, want map[string]commentStats)
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "list comments with likes and replies",
			f: func(t *testing.T) (string, map[string]commentStats) {
				author, reader := newUser(t, pool), newUser(t, pool)
				postID := newPost(t, pool, author)
				first := seedComment(t, pool, postID, author, nil, "first")
				seedComment(t, pool, postID, reader, nil, "second")
				seedComment(t, pool, postID, reader, &first, "reply")
				seedLike(t, pool, reader, first)
				return postID, map[string]commentStats{
					"first":  {likes: 1, replies: 1},
					"second": {},
					"reply":  {},
				}
			},
			success: true,
		},
		{
			caseName: "list comments excludes comments of other posts",
			f: func(t *testing.T) (string, map[string]commentStats) {
				author := newUser(t, pool)
				postID := newPost(t, pool, author)
				otherPostID := newPost(t, pool, author)
				seedComment(t, pool, postID, author, nil, "mine")
				seedComment(t, pool, otherPostID, author, nil, "other")
				return postID, map[string]commentStats{"mine": {}}
			},
			success: true,
		},
		{
			caseName: "list comments of post without comments",
			f: func(t *testing.T) (string, map[string]commentStats) {
				return newPost(t, pool, newUser(t, pool)), map[string]commentStats{}
			},
			success: true,
		},
		{
			caseName: "list comments with canceled context",
			f: func(t *testing.T) (string, map[string]commentStats) {
				return newPost(t, pool, newUser(t, pool)), nil
			},
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			postID, want := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			comments, err := repo.ListComments(ctx, postID)
			if !tc.success {
				require.Error(t, err)
				assert.Nil(t, comments)
				return
			}
			require.NoError(t, err)
			require.Len(t, comments, len(want))
			for _, c := range comments {
				stats, ok := want[c.Body]
				require.True(t, ok, "unexpected comment %q", c.Body)
				assert.Equal(t, postID, c.PostID)
				assert.EqualValues(t, stats.likes, c.Likes)
				assert.EqualValues(t, stats.replies, c.CountComments)
			}
		})
	}
}

func TestCreateComment(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewCommentsRepo(pool)

	testCases := []struct {
		caseName    string
		f           func(t *testing.T) *models.Comments
		canceledCtx bool
		success     bool
	}{
		{
			caseName: "create comment",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				return &models.Comments{PostID: newPost(t, pool, userID), UserID: userID, Body: "body"}
			},
			success: true,
		},
		{
			caseName: "create reply",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				postID := newPost(t, pool, userID)
				parent := seedComment(t, pool, postID, userID, nil, "parent")
				return &models.Comments{PostID: postID, UserID: userID, ParentID: &parent, Body: "reply body"}
			},
			success: true,
		},
		{
			caseName: "create comment for not found post",
			f: func(t *testing.T) *models.Comments {
				return &models.Comments{PostID: uuid.NewString(), UserID: newUser(t, pool), Body: "body"}
			},
			success: false,
		},
		{
			caseName: "create comment with invalid post id",
			f: func(t *testing.T) *models.Comments {
				return &models.Comments{PostID: "invalid_uuid", UserID: newUser(t, pool), Body: "body"}
			},
			success: false,
		},
		{
			caseName: "create comment with canceled context",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				return &models.Comments{PostID: newPost(t, pool, userID), UserID: userID, Body: "body"}
			},
			canceledCtx: true,
			success:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			comment := tc.f(t)
			ctx := context.Background()
			if tc.canceledCtx {
				ctx = canceledContext()
			}

			err := repo.CreateComment(ctx, comment)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var parentID *string
			require.NoError(t, pool.QueryRow(context.Background(),
				`SELECT parent_id FROM comments WHERE post_id = $1 AND user_id = $2 AND body = $3`,
				comment.PostID, comment.UserID, comment.Body).Scan(&parentID))
			if comment.ParentID == nil {
				assert.Nil(t, parentID)
			} else {
				require.NotNil(t, parentID)
				assert.Equal(t, *comment.ParentID, *parentID)
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewCommentsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) *models.Comments
		wantBody string
		success  bool
	}{
		{
			caseName: "update comment",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				postID := newPost(t, pool, userID)
				id := seedComment(t, pool, postID, userID, nil, "old")
				return &models.Comments{ID: id, PostID: postID, UserID: userID, Body: "new"}
			},
			wantBody: "new",
			success:  true,
		},
		{
			caseName: "update comment of other user",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				postID := newPost(t, pool, userID)
				id := seedComment(t, pool, postID, userID, nil, "old")
				return &models.Comments{ID: id, PostID: postID, UserID: newUser(t, pool), Body: "new"}
			},
			wantBody: "old",
			success:  false,
		},
		{
			caseName: "update comment with other post id",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				postID := newPost(t, pool, userID)
				id := seedComment(t, pool, postID, userID, nil, "old")
				return &models.Comments{ID: id, PostID: newPost(t, pool, userID), UserID: userID, Body: "new"}
			},
			wantBody: "old",
			success:  false,
		},
		{
			caseName: "update not found comment",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				return &models.Comments{ID: uuid.NewString(), PostID: newPost(t, pool, userID), UserID: userID, Body: "new"}
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			comment := tc.f(t)

			updated, err := repo.UpdateComment(context.Background(), comment)
			if tc.success {
				require.NoError(t, err)
				require.NotNil(t, updated)
				assert.Equal(t, comment.ID, updated.ID)
				assert.Equal(t, comment.PostID, updated.PostID)
				assert.Equal(t, comment.UserID, updated.UserID)
				assert.Equal(t, comment.Body, updated.Body)
				assert.Nil(t, updated.ParentID)
			} else {
				require.Error(t, err)
				assert.Nil(t, updated)
			}
			if tc.wantBody != "" {
				assert.Equal(t, tc.wantBody, commentBody(t, pool, comment.ID))
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewCommentsRepo(pool)

	testCases := []struct {
		caseName   string
		f          func(t *testing.T) *models.Comments
		wantExists bool
		success    bool
	}{
		{
			caseName: "delete comment",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				postID := newPost(t, pool, userID)
				id := seedComment(t, pool, postID, userID, nil, "body")
				return &models.Comments{ID: id, PostID: postID, UserID: userID}
			},
			wantExists: false,
			success:    true,
		},
		{
			caseName: "delete comment of other user",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				postID := newPost(t, pool, userID)
				id := seedComment(t, pool, postID, userID, nil, "body")
				return &models.Comments{ID: id, PostID: postID, UserID: newUser(t, pool)}
			},
			wantExists: true,
			success:    true,
		},
		{
			caseName: "delete comment with other post id",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				postID := newPost(t, pool, userID)
				id := seedComment(t, pool, postID, userID, nil, "body")
				return &models.Comments{ID: id, PostID: newPost(t, pool, userID), UserID: userID}
			},
			wantExists: true,
			success:    true,
		},
		{
			caseName: "delete not found comment",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				return &models.Comments{ID: uuid.NewString(), PostID: newPost(t, pool, userID), UserID: userID}
			},
			wantExists: false,
			success:    true,
		},
		{
			caseName: "delete comment with invalid id",
			f: func(t *testing.T) *models.Comments {
				userID := newUser(t, pool)
				return &models.Comments{ID: "invalid_uuid", PostID: newPost(t, pool, userID), UserID: userID}
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			comment := tc.f(t)

			err := repo.DeleteComment(context.Background(), comment)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			n := countRows(t, pool, `SELECT COUNT(*) FROM comments WHERE id = $1`, comment.ID)
			assert.Equal(t, tc.wantExists, n == 1)
		})
	}
}

func TestReportComment(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewCommentsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) *models.CommentReport
		success  bool
	}{
		{
			caseName: "report comment",
			f: func(t *testing.T) *models.CommentReport {
				author, reporter := newUser(t, pool), newUser(t, pool)
				commentID := seedComment(t, pool, newPost(t, pool, author), author, nil, "body")
				return &models.CommentReport{UserID: reporter, CommentID: commentID, Cause: "spam"}
			},
			success: true,
		},
		{
			caseName: "report not found comment",
			f: func(t *testing.T) *models.CommentReport {
				return &models.CommentReport{UserID: newUser(t, pool), CommentID: uuid.NewString(), Cause: "spam"}
			},
			success: false,
		},
		{
			caseName: "report comment with invalid id",
			f: func(t *testing.T) *models.CommentReport {
				return &models.CommentReport{UserID: newUser(t, pool), CommentID: "invalid_uuid", Cause: "spam"}
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			report := tc.f(t)

			err := repo.ReportComment(context.Background(), report)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			n := countRows(t, pool,
				`SELECT COUNT(*) FROM comment_reports WHERE user_id = $1 AND comment_id = $2`,
				report.UserID, report.CommentID)
			assert.Equal(t, 1, n)
		})
	}
}

func TestLikeComment(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewCommentsRepo(pool)

	testCases := []struct {
		caseName string
		f        func(t *testing.T) (userID, commentID string)
		success  bool
	}{
		{
			caseName: "like comment",
			f: func(t *testing.T) (string, string) {
				author := newUser(t, pool)
				return newUser(t, pool), seedComment(t, pool, newPost(t, pool, author), author, nil, "body")
			},
			success: true,
		},
		{
			caseName: "like comment twice",
			f: func(t *testing.T) (string, string) {
				author, liker := newUser(t, pool), newUser(t, pool)
				commentID := seedComment(t, pool, newPost(t, pool, author), author, nil, "body")
				require.NoError(t, repo.LikeComment(context.Background(), liker, commentID))
				return liker, commentID
			},
			success: false,
		},
		{
			caseName: "like not found comment",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), uuid.NewString()
			},
			success: false,
		},
		{
			caseName: "like comment with invalid id",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), "invalid_uuid"
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, commentID := tc.f(t)

			err := repo.LikeComment(context.Background(), userID, commentID)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			n := countRows(t, pool,
				`SELECT COUNT(*) FROM comment_likes WHERE user_id = $1 AND comment_id = $2`, userID, commentID)
			assert.Equal(t, 1, n)
		})
	}
}

func TestUnlikeComment(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewCommentsRepo(pool)

	testCases := []struct {
		caseName  string
		f         func(t *testing.T) (userID, commentID string)
		wantLikes int
		success   bool
	}{
		{
			caseName: "unlike comment",
			f: func(t *testing.T) (string, string) {
				author, liker := newUser(t, pool), newUser(t, pool)
				commentID := seedComment(t, pool, newPost(t, pool, author), author, nil, "body")
				seedLike(t, pool, liker, commentID)
				return liker, commentID
			},
			wantLikes: 0,
			success:   true,
		},
		{
			caseName: "unlike not liked comment",
			f: func(t *testing.T) (string, string) {
				author := newUser(t, pool)
				return newUser(t, pool), seedComment(t, pool, newPost(t, pool, author), author, nil, "body")
			},
			wantLikes: 0,
			success:   true,
		},
		{
			caseName: "unlike keeps likes of other users",
			f: func(t *testing.T) (string, string) {
				author, liker, other := newUser(t, pool), newUser(t, pool), newUser(t, pool)
				commentID := seedComment(t, pool, newPost(t, pool, author), author, nil, "body")
				seedLike(t, pool, liker, commentID)
				seedLike(t, pool, other, commentID)
				return liker, commentID
			},
			wantLikes: 1,
			success:   true,
		},
		{
			caseName: "unlike comment with invalid id",
			f: func(t *testing.T) (string, string) {
				return newUser(t, pool), "invalid_uuid"
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			userID, commentID := tc.f(t)

			err := repo.UnlikeComment(context.Background(), userID, commentID)
			if !tc.success {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.wantLikes,
				countRows(t, pool, `SELECT COUNT(*) FROM comment_likes WHERE comment_id = $1`, commentID))
			assert.Equal(t, 0,
				countRows(t, pool, `SELECT COUNT(*) FROM comment_likes WHERE user_id = $1 AND comment_id = $2`, userID, commentID))
		})
	}
}
