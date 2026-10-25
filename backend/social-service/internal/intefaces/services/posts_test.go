package services

import (
	"encoding/base64"
	v2 "encoding/json/v2"
	"errors"
	"testing"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	store_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func makeCursor(t *testing.T, postID *string, createdAt *time.Time) string {
	t.Helper()
	raw, err := v2.Marshal(dto.CursorDTO{PostID: postID, CreatedAt: createdAt})
	require.NoError(t, err)
	return base64.URLEncoding.EncodeToString(raw)
}

func TestGetSubPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func()
		want    []models.FeedPost
		success bool
	}{
		{
			caseName: "get sub posts",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubPosts(gomock.Any(), "user_id").
					Return([]models.FeedPost{{PostID: "1"}, {PostID: "2"}}, []models.FeedPost{{PostID: "3"}}, nil).Times(1)
			},
			want:    []models.FeedPost{{PostID: "1"}, {PostID: "2"}, {PostID: "3"}},
			success: true,
		},
		{
			caseName: "get sub posts removes duplicates",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubPosts(gomock.Any(), "user_id").
					Return([]models.FeedPost{{PostID: "1"}, {PostID: "2"}}, []models.FeedPost{{PostID: "2"}, {PostID: "3"}}, nil).Times(1)
			},
			want:    []models.FeedPost{{PostID: "1"}, {PostID: "2"}, {PostID: "3"}},
			success: true,
		},
		{
			caseName: "get empty sub posts",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubPosts(gomock.Any(), "user_id").
					Return(nil, nil, nil).Times(1)
			},
			want:    []models.FeedPost{},
			success: true,
		},
		{
			caseName: "get sub posts error",
			data: struct {
				userID string
			}{userID: "user_id"},
			prepare: func() {
				repo.EXPECT().GetSubPosts(gomock.Any(), "user_id").
					Return(nil, nil, errors.New("get sub posts error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := post.GetSubPosts(t.Context(), tc.data.userID)
			if tc.success {
				require.NoError(t, err)
				assert.ElementsMatch(t, tc.want, got)
			} else {
				require.Error(t, err)
				assert.Nil(t, got)
			}
		})
	}
}

func TestGetGlobalPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	postIDStr := uuid.NewString()
	postUUID := uuid.MustParse(postIDStr)
	createdAt := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)

	validCursor := makeCursor(t, &postIDStr, &createdAt)
	noPostIDCursor := makeCursor(t, nil, &createdAt)
	badUUIDCursor := makeCursor(t, strPtr("not_uuid"), &createdAt)
	badJSONCursor := base64.URLEncoding.EncodeToString([]byte("{invalid"))
	badBase64Cursor := "!!!"

	postIDMatcher := gomock.Cond(func(p *uuid.UUID) bool { return p != nil && *p == postUUID })
	createdAtMatcher := gomock.Cond(func(p *time.Time) bool { return p != nil && p.Equal(createdAt) })

	feed := []models.FeedPost{{PostID: "1"}, {PostID: "2"}}

	testCases := []struct {
		caseName string
		data     struct {
			cursor string
		}
		prepare func()
		want    []models.FeedPost
		success bool
	}{
		{
			caseName: "get global posts first page",
			data: struct {
				cursor string
			}{cursor: ""},
			prepare: func() {
				cache.EXPECT().GetGlobal(gomock.Any(), "").
					Return(nil, domain.ErrNotFound).Times(1)
				repo.EXPECT().GetGlobalPosts(gomock.Any(), gomock.Nil(), gomock.Nil()).
					Return(feed, nil).Times(1)
				cache.EXPECT().SetGlobal(gomock.Any(), feed, "").
					Return(nil).Times(1)
			},
			want:    feed,
			success: true,
		},
		{
			caseName: "get global posts first page from cache",
			data: struct {
				cursor string
			}{cursor: ""},
			prepare: func() {
				cache.EXPECT().GetGlobal(gomock.Any(), "").
					Return(feed, nil).Times(1)
			},
			want:    feed,
			success: true,
		},
		{
			caseName: "get global posts with cursor",
			data: struct {
				cursor string
			}{cursor: validCursor},
			prepare: func() {
				cache.EXPECT().GetGlobal(gomock.Any(), validCursor).
					Return(nil, domain.ErrNotFound).Times(1)
				repo.EXPECT().GetGlobalPosts(gomock.Any(), postIDMatcher, createdAtMatcher).
					Return(feed, nil).Times(1)
				cache.EXPECT().SetGlobal(gomock.Any(), feed, validCursor).
					Return(nil).Times(1)
			},
			want:    feed,
			success: true,
		},
		{
			caseName: "get global posts with cursor from cache",
			data: struct {
				cursor string
			}{cursor: validCursor},
			prepare: func() {
				cache.EXPECT().GetGlobal(gomock.Any(), validCursor).
					Return(feed, nil).Times(1)
			},
			want:    feed,
			success: true,
		},
		{
			caseName: "invalid base64 cursor",
			data: struct {
				cursor string
			}{cursor: badBase64Cursor},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "invalid json in cursor",
			data: struct {
				cursor string
			}{cursor: badJSONCursor},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "cursor without post id",
			data: struct {
				cursor string
			}{cursor: noPostIDCursor},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "cursor with invalid post id",
			data: struct {
				cursor string
			}{cursor: badUUIDCursor},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "get global cache error",
			data: struct {
				cursor string
			}{cursor: ""},
			prepare: func() {
				cache.EXPECT().GetGlobal(gomock.Any(), "").
					Return(nil, errors.New("get global cache error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "get global posts error",
			data: struct {
				cursor string
			}{cursor: ""},
			prepare: func() {
				cache.EXPECT().GetGlobal(gomock.Any(), "").
					Return(nil, domain.ErrNotFound).Times(1)
				repo.EXPECT().GetGlobalPosts(gomock.Any(), gomock.Nil(), gomock.Nil()).
					Return(nil, errors.New("get global posts error")).Times(1)
			},
			success: false,
		},
		{
			caseName: "set global cache error",
			data: struct {
				cursor string
			}{cursor: ""},
			prepare: func() {
				cache.EXPECT().GetGlobal(gomock.Any(), "").
					Return(nil, domain.ErrNotFound).Times(1)
				repo.EXPECT().GetGlobalPosts(gomock.Any(), gomock.Nil(), gomock.Nil()).
					Return(feed, nil).Times(1)
				cache.EXPECT().SetGlobal(gomock.Any(), feed, "").
					Return(errors.New("set global cache error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := post.GetGlobalPosts(t.Context(), tc.data.cursor)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				assert.Nil(t, got)
			}
		})
	}
}

func TestGetTopics(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		prepare  func()
		want     []models.Topics
		success  bool
	}{
		{
			caseName: "get topics",
			prepare: func() {
				repo.EXPECT().GetTopics(gomock.Any()).
					Return([]models.Topics{{ID: 1}, {ID: 2}}, nil).Times(1)
			},
			want:    []models.Topics{{ID: 1}, {ID: 2}},
			success: true,
		},
		{
			caseName: "get empty topics",
			prepare: func() {
				repo.EXPECT().GetTopics(gomock.Any()).
					Return(nil, nil).Times(1)
			},
			want:    nil,
			success: true,
		},
		{
			caseName: "get topics error",
			prepare: func() {
				repo.EXPECT().GetTopics(gomock.Any()).
					Return(nil, errors.New("get topics error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := post.GetTopics(t.Context())
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				assert.Nil(t, got)
			}
		})
	}
}

func TestGetTopicByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			topicID string
		}
		prepare func()
		want    *models.Topics
		wantErr error
		success bool
	}{
		{
			caseName: "get topic by id",
			data: struct {
				topicID string
			}{topicID: "1"},
			prepare: func() {
				repo.EXPECT().GetTopicsByID(gomock.Any(), "1").
					Return(&models.Topics{ID: 1}, nil).Times(1)
			},
			want:    &models.Topics{ID: 1},
			success: true,
		},
		{
			caseName: "get not found topic",
			data: struct {
				topicID string
			}{topicID: "1"},
			prepare: func() {
				repo.EXPECT().GetTopicsByID(gomock.Any(), "1").
					Return(nil, domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "get topic by id error",
			data: struct {
				topicID string
			}{topicID: "1"},
			prepare: func() {
				repo.EXPECT().GetTopicsByID(gomock.Any(), "1").
					Return(nil, errors.New("get topic by id error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := post.GetTopicByID(t.Context(), tc.data.topicID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
				assert.Nil(t, got)
			}
		})
	}
}

func TestCreateTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
			title  string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "create topic",
			data: struct {
				userID string
				title  string
			}{userID: "user_id", title: "title"},
			prepare: func() {
				repo.EXPECT().CreateTopic(gomock.Any(), &models.Topics{UserID: "user_id", Title: "title"}).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "create duplicate topic",
			data: struct {
				userID string
				title  string
			}{userID: "user_id", title: "title"},
			prepare: func() {
				repo.EXPECT().CreateTopic(gomock.Any(), gomock.Any()).
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "create topic error",
			data: struct {
				userID string
				title  string
			}{userID: "user_id", title: "title"},
			prepare: func() {
				repo.EXPECT().CreateTopic(gomock.Any(), gomock.Any()).
					Return(errors.New("create topic error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.CreateTopicDTO{
				Title: tc.data.title,
			}
			err := post.CreateTopic(t.Context(), tc.data.userID, &req)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestDeleteTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			topicID string
			userID  string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "delete topic",
			data: struct {
				topicID string
				userID  string
			}{topicID: "5", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().DeleteTopic(gomock.Any(), &models.Topics{ID: 5, UserID: "user_id"}).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "delete topic with invalid id",
			data: struct {
				topicID string
				userID  string
			}{topicID: "not_a_number", userID: "user_id"},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "delete not found topic",
			data: struct {
				topicID string
				userID  string
			}{topicID: "5", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().DeleteTopic(gomock.Any(), gomock.Any()).
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "delete topic error",
			data: struct {
				topicID string
				userID  string
			}{topicID: "5", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().DeleteTopic(gomock.Any(), gomock.Any()).
					Return(errors.New("delete topic error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := post.DeleteTopic(t.Context(), tc.data.topicID, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestReportTopic(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			userID  string
			topicID string
			cause   string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "report topic",
			data: struct {
				userID  string
				topicID string
				cause   string
			}{userID: "user_id", topicID: "5", cause: "spam"},
			prepare: func() {
				repo.EXPECT().ReportTopic(gomock.Any(), &models.TopicsReport{UserID: "user_id", TopicID: 5, Cause: "spam"}).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "report topic with invalid id",
			data: struct {
				userID  string
				topicID string
				cause   string
			}{userID: "user_id", topicID: "not_a_number", cause: "spam"},
			prepare: func() {},
			success: false,
		},
		{
			caseName: "report topic twice",
			data: struct {
				userID  string
				topicID string
				cause   string
			}{userID: "user_id", topicID: "5", cause: "spam"},
			prepare: func() {
				repo.EXPECT().ReportTopic(gomock.Any(), gomock.Any()).
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "report topic error",
			data: struct {
				userID  string
				topicID string
				cause   string
			}{userID: "user_id", topicID: "5", cause: "spam"},
			prepare: func() {
				repo.EXPECT().ReportTopic(gomock.Any(), gomock.Any()).
					Return(errors.New("report topic error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.ReportTopicDTO{
				Cause: tc.data.cause,
			}
			err := post.ReportTopic(t.Context(), tc.data.userID, tc.data.topicID, &req)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestCreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			userID      string
			topicID     int
			imageURL    string
			description string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "create post",
			data: struct {
				userID      string
				topicID     int
				imageURL    string
				description string
			}{userID: "user_id", topicID: 3, imageURL: "https://image.test/1.png", description: "description"},
			prepare: func() {
				var topicID = 3
				repo.EXPECT().CreatePost(gomock.Any(), &models.Post{
					TopicID:     &topicID,
					UserID:      "user_id",
					ImageURL:    strPtr("https://image.test/1.png"),
					Description: strPtr("description"),
				}).Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "create duplicate post",
			data: struct {
				userID      string
				topicID     int
				imageURL    string
				description string
			}{userID: "user_id", topicID: 3, imageURL: "https://image.test/1.png", description: "description"},
			prepare: func() {
				repo.EXPECT().CreatePost(gomock.Any(), gomock.Any()).
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "create post error",
			data: struct {
				userID      string
				topicID     int
				imageURL    string
				description string
			}{userID: "user_id", topicID: 3, imageURL: "https://image.test/1.png", description: "description"},
			prepare: func() {
				repo.EXPECT().CreatePost(gomock.Any(), gomock.Any()).
					Return(errors.New("create post error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.CreatePostDTO{
				TopicID:     &tc.data.topicID,
				ImageUrl:    tc.data.imageURL,
				Description: tc.data.description,
			}
			err := post.CreatePost(t.Context(), &req, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestUpdatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			postID      string
			userID      string
			topicID     *int
			imageURL    *string
			description *string
		}
		prepare func()
		want    *models.Post
		wantErr error
		success bool
	}{
		{
			caseName: "update post",
			data: struct {
				postID      string
				userID      string
				topicID     *int
				imageURL    *string
				description *string
			}{postID: "post_id", userID: "user_id", topicID: intPtr(3), imageURL: strPtr("new_url"), description: strPtr("new_description")},
			prepare: func() {
				var topicID = 3
				repo.EXPECT().BuildUpdatePost(gomock.Any(), &models.Post{
					ID:          "post_id",
					TopicID:     &topicID,
					UserID:      "user_id",
					ImageURL:    strPtr("new_url"),
					Description: strPtr("new_description"),
				}).Return(&models.Post{UserID: "user_id", ImageURL: strPtr("new_url"), Description: strPtr("new_description")}, nil).Times(1)
			},
			want:    &models.Post{UserID: "user_id", ImageURL: strPtr("new_url"), Description: strPtr("new_description")},
			success: true,
		},
		{
			caseName: "update only description",
			data: struct {
				postID      string
				userID      string
				topicID     *int
				imageURL    *string
				description *string
			}{postID: "post_id", userID: "user_id", description: strPtr("new_description")},
			prepare: func() {
				repo.EXPECT().BuildUpdatePost(gomock.Any(), &models.Post{
					ID:          "post_id",
					UserID:      "user_id",
					Description: strPtr("new_description"),
				}).Return(&models.Post{UserID: "user_id", Description: strPtr("new_description")}, nil).Times(1)
			},
			want:    &models.Post{UserID: "user_id", Description: strPtr("new_description")},
			success: true,
		},
		{
			caseName: "update not found post",
			data: struct {
				postID      string
				userID      string
				topicID     *int
				imageURL    *string
				description *string
			}{postID: "post_id", userID: "user_id", description: strPtr("new_description")},
			prepare: func() {
				repo.EXPECT().BuildUpdatePost(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "update post error",
			data: struct {
				postID      string
				userID      string
				topicID     *int
				imageURL    *string
				description *string
			}{postID: "post_id", userID: "user_id", description: strPtr("new_description")},
			prepare: func() {
				repo.EXPECT().BuildUpdatePost(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("update post error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.UpdatePostDTO{
				TopicID:     tc.data.topicID,
				ImageUrl:    tc.data.imageURL,
				Description: tc.data.description,
			}
			got, err := post.UpdatePost(t.Context(), &req, tc.data.postID, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
				assert.Nil(t, got)
			}
		})
	}
}

func TestGetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			postID string
		}
		prepare func()
		want    *models.FeedPost
		wantErr error
		success bool
	}{
		{
			caseName: "get post by id",
			data: struct {
				postID string
			}{postID: "post_id"},
			prepare: func() {
				repo.EXPECT().GetPostByID(gomock.Any(), "post_id").
					Return(&models.FeedPost{PostID: "post_id"}, nil).Times(1)
			},
			want:    &models.FeedPost{PostID: "post_id"},
			success: true,
		},
		{
			caseName: "get not found post",
			data: struct {
				postID string
			}{postID: "post_id"},
			prepare: func() {
				repo.EXPECT().GetPostByID(gomock.Any(), "post_id").
					Return(nil, domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "get post by id error",
			data: struct {
				postID string
			}{postID: "post_id"},
			prepare: func() {
				repo.EXPECT().GetPostByID(gomock.Any(), "post_id").
					Return(nil, errors.New("get post by id error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := post.GetPostByID(t.Context(), tc.data.postID)
			if tc.success {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
				assert.Nil(t, got)
			}
		})
	}
}

func TestViewPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			postID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "view post",
			data: struct {
				postID string
			}{postID: "post_id"},
			prepare: func() {
				repo.EXPECT().ViewPost(gomock.Any(), "post_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "view not found post",
			data: struct {
				postID string
			}{postID: "post_id"},
			prepare: func() {
				repo.EXPECT().ViewPost(gomock.Any(), "post_id").
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "view post error",
			data: struct {
				postID string
			}{postID: "post_id"},
			prepare: func() {
				repo.EXPECT().ViewPost(gomock.Any(), "post_id").
					Return(errors.New("view post error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := post.ViewPost(t.Context(), tc.data.postID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestReportPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
			postID string
			cause  string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "report post",
			data: struct {
				userID string
				postID string
				cause  string
			}{userID: "user_id", postID: "post_id", cause: "spam"},
			prepare: func() {
				repo.EXPECT().ReportPost(gomock.Any(), &models.PostReport{UserID: "user_id", PostID: "post_id", Cause: "spam"}).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "report post twice",
			data: struct {
				userID string
				postID string
				cause  string
			}{userID: "user_id", postID: "post_id", cause: "spam"},
			prepare: func() {
				repo.EXPECT().ReportPost(gomock.Any(), gomock.Any()).
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "report post error",
			data: struct {
				userID string
				postID string
				cause  string
			}{userID: "user_id", postID: "post_id", cause: "spam"},
			prepare: func() {
				repo.EXPECT().ReportPost(gomock.Any(), gomock.Any()).
					Return(errors.New("report post error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.ReportDTO{
				Cause: tc.data.cause,
			}
			err := post.ReportPost(t.Context(), &req, tc.data.userID, tc.data.postID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			postID string
			userID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "delete post",
			data: struct {
				postID string
				userID string
			}{postID: "post_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().DeletePost(gomock.Any(), "post_id", "user_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "delete not found post",
			data: struct {
				postID string
				userID string
			}{postID: "post_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().DeletePost(gomock.Any(), "post_id", "user_id").
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "delete post error",
			data: struct {
				postID string
				userID string
			}{postID: "post_id", userID: "user_id"},
			prepare: func() {
				repo.EXPECT().DeletePost(gomock.Any(), "post_id", "user_id").
					Return(errors.New("delete post error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := post.DeletePost(t.Context(), tc.data.postID, tc.data.userID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestLikePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
			postID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "like post",
			data: struct {
				userID string
				postID string
			}{userID: "user_id", postID: "post_id"},
			prepare: func() {
				repo.EXPECT().LikePost(gomock.Any(), "user_id", "post_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "like post twice",
			data: struct {
				userID string
				postID string
			}{userID: "user_id", postID: "post_id"},
			prepare: func() {
				repo.EXPECT().LikePost(gomock.Any(), "user_id", "post_id").
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "like post error",
			data: struct {
				userID string
				postID string
			}{userID: "user_id", postID: "post_id"},
			prepare: func() {
				repo.EXPECT().LikePost(gomock.Any(), "user_id", "post_id").
					Return(errors.New("like post error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := post.LikePost(t.Context(), tc.data.userID, tc.data.postID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}

func TestUnlikePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockPostsRepo(ctrl)
	cache := store_mocks.NewMockPostsCache(ctrl)

	post := NewPostsService(repo, cache)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
			postID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "unlike post",
			data: struct {
				userID string
				postID string
			}{userID: "user_id", postID: "post_id"},
			prepare: func() {
				repo.EXPECT().UnlikePost(gomock.Any(), "user_id", "post_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "unlike not liked post",
			data: struct {
				userID string
				postID string
			}{userID: "user_id", postID: "post_id"},
			prepare: func() {
				repo.EXPECT().UnlikePost(gomock.Any(), "user_id", "post_id").
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "unlike post error",
			data: struct {
				userID string
				postID string
			}{userID: "user_id", postID: "post_id"},
			prepare: func() {
				repo.EXPECT().UnlikePost(gomock.Any(), "user_id", "post_id").
					Return(errors.New("unlike post error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := post.UnlikePost(t.Context(), tc.data.userID, tc.data.postID)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
			}
		})
	}
}
