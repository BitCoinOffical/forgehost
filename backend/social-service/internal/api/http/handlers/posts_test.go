package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers/mock"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestGetSubPostsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "get sub posts",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetSubPosts(gomock.Any(), "user-id").
					Return([]models.FeedPost{}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "get sub posts error",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().GetSubPosts(gomock.Any(), "user-id").
					Return(nil, errors.New("get sub posts error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/posts", "/posts", tc.data, posts.GetSubPosts)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestGetGlobalPostsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []struct {
		caseName string
		route    string
		url      string
		prepare  func()
		status   int
	}{
		{
			caseName: "first page without cursor",
			route:    "/posts/global",
			url:      "/posts/global",
			prepare: func() {
				srvc.EXPECT().GetGlobalPosts(gomock.Any(), "").
					Return([]models.FeedPost{}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "page by cursor",
			route:    "/posts/global/:cursor",
			url:      "/posts/global/abc123",
			prepare: func() {
				srvc.EXPECT().GetGlobalPosts(gomock.Any(), "abc123").
					Return([]models.FeedPost{}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid cursor",
			route:    "/posts/global/:cursor",
			url:      "/posts/global/broken",
			prepare: func() {
				srvc.EXPECT().GetGlobalPosts(gomock.Any(), "broken").
					Return(nil, fmt.Errorf("decode cursor: %w", domain.ErrInvalidCursor)).Times(1)
			},
			status: http.StatusBadRequest,
		},
		{
			caseName: "get global posts error",
			route:    "/posts/global",
			url:      "/posts/global",
			prepare: func() {
				srvc.EXPECT().GetGlobalPosts(gomock.Any(), "").
					Return(nil, errors.New("get global posts error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, tc.route, tc.url, chatInput{}, posts.GetGlobalPosts)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestGetPostByIDHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "get post by id",
			prepare: func() {
				srvc.EXPECT().GetPostByID(gomock.Any(), "post-id").
					Return(&models.FeedPost{PostID: "post-id"}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "post not found",
			prepare: func() {
				srvc.EXPECT().GetPostByID(gomock.Any(), "post-id").
					Return(nil, fmt.Errorf("r.repo.GetPostById: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "get post by id error",
			prepare: func() {
				srvc.EXPECT().GetPostByID(gomock.Any(), "post-id").
					Return(nil, errors.New("get post error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/posts/:post_id", "/posts/post-id", tc.data, posts.GetByID)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, w.Body.String(), "post-id")
			}
		})
	}
}

func TestCreatePostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))
	body := `{"description":"hello"}`

	testCases := []chatCase{
		{
			caseName: "create post with description only",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().CreatePost(gomock.Any(), gomock.Any(), "user-id").
					DoAndReturn(func(_ context.Context, req *dto.CreatePostDTO, _ string) error {
						assert.Equal(t, "hello", req.Description)
						assert.Nil(t, req.TopicID)
						return nil
					}).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "create post with image and topic",
			data: chatInput{
				body:   `{"image_url":"https://img.test/1.png","description":"d","topic_id":5}`,
				userID: userClaims("user-id"),
			},
			prepare: func() {
				srvc.EXPECT().CreatePost(gomock.Any(), gomock.Any(), "user-id").
					DoAndReturn(func(_ context.Context, req *dto.CreatePostDTO, _ string) error {
						assert.Equal(t, "https://img.test/1.png", req.ImageUrl)
						require.NotNil(t, req.TopicID)
						assert.Equal(t, 5, *req.TopicID)
						return nil
					}).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "neither image nor description",
			data:     chatInput{body: `{}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "image url over http",
			data:     chatInput{body: `{"image_url":"http://img.test/1.png"}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "image url javascript scheme",
			data:     chatInput{body: `{"image_url":"javascript:alert(1)"}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "topic id is zero",
			data:     chatInput{body: `{"description":"hello","topic_id":0}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "description too long",
			data:     chatInput{body: `{"description":"` + strings.Repeat("a", 4097) + `"}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "post already exists",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().CreatePost(gomock.Any(), gomock.Any(), "user-id").
					Return(fmt.Errorf("r.repo.CreatePost: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "create post error",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().CreatePost(gomock.Any(), gomock.Any(), "user-id").
					Return(errors.New("create post error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/posts", "/posts", tc.data, posts.CreatePost)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestGetTopicsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "get topics",
			prepare: func() {
				srvc.EXPECT().GetTopics(gomock.Any()).
					Return([]models.Topics{{ID: 1, Title: "golang"}}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "get topics error",
			prepare: func() {
				srvc.EXPECT().GetTopics(gomock.Any()).
					Return(nil, errors.New("get topics error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/topics", "/topics", tc.data, posts.GetTopics)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, w.Body.String(), "golang")
			}
		})
	}
}

func TestGetTopicByIDHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "get topic by id",
			prepare: func() {
				srvc.EXPECT().GetTopicByID(gomock.Any(), "7").
					Return(&models.Topics{ID: 7, Title: "golang"}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "topic not found",
			prepare: func() {
				srvc.EXPECT().GetTopicByID(gomock.Any(), "7").
					Return(nil, fmt.Errorf("r.repo.GetTopicsByID: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "get topic by id error",
			prepare: func() {
				srvc.EXPECT().GetTopicByID(gomock.Any(), "7").
					Return(nil, errors.New("get topic error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, "/topics/:topic_id", "/topics/7", tc.data, posts.GetTopicByID)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, w.Body.String(), "golang")
			}
		})
	}
}

func TestCreateTopicHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))
	body := `{"title":"golang"}`

	testCases := []chatCase{
		{
			caseName: "create topic",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().CreateTopic(gomock.Any(), "user-id", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, req *dto.CreateTopicDTO) error {
						assert.Equal(t, "golang", req.Title)
						return nil
					}).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty json fails validation",
			data:     chatInput{body: `{}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty title",
			data:     chatInput{body: `{"title":""}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "title too long",
			data:     chatInput{body: `{"title":"` + strings.Repeat("a", 101) + `"}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "topic already exists",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().CreateTopic(gomock.Any(), "user-id", gomock.Any()).
					Return(fmt.Errorf("r.repo.CreateTopic: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "create topic error",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().CreateTopic(gomock.Any(), "user-id", gomock.Any()).
					Return(errors.New("create topic error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/topics", "/topics", tc.data, posts.CreateTopic)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestDeleteTopicHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "delete topic",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().DeleteTopic(gomock.Any(), "7", "user-id").Return(nil).Times(1)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "topic not found",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().DeleteTopic(gomock.Any(), "7", "user-id").
					Return(fmt.Errorf("r.repo.DeleteTopic: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "delete topic error",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().DeleteTopic(gomock.Any(), "7", "user-id").
					Return(errors.New("delete topic error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, "/topics/:topic_id", "/topics/7", tc.data, posts.DeleteTopic)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestReportTopicHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))
	body := `{"cause":"spam"}`

	testCases := []chatCase{
		{
			caseName: "report topic",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().ReportTopic(gomock.Any(), "user-id", "7", gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, req *dto.ReportTopicDTO) error {
						assert.Equal(t, "spam", req.Cause)
						return nil
					}).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty json fails validation",
			data:     chatInput{body: `{}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "cause too long",
			data:     chatInput{body: `{"cause":"` + strings.Repeat("a", 256) + `"}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "topic not found",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().ReportTopic(gomock.Any(), "user-id", "7", gomock.Any()).
					Return(fmt.Errorf("r.repo.DeleteTopic: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "already reported",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().ReportTopic(gomock.Any(), "user-id", "7", gomock.Any()).
					Return(fmt.Errorf("r.repo.DeleteTopic: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "report topic error",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().ReportTopic(gomock.Any(), "user-id", "7", gomock.Any()).
					Return(errors.New("report topic error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/topics/:topic_id/report", "/topics/7/report", tc.data, posts.ReportTopic)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUpdatePostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))
	body := `{"description":"new"}`

	testCases := []chatCase{
		{
			caseName: "update post",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().UpdatePost(gomock.Any(), gomock.Any(), "post-id", "user-id").
					DoAndReturn(func(_ context.Context, req *dto.UpdatePostDTO, _, _ string) (*models.Post, error) {
						require.NotNil(t, req.Description)
						assert.Equal(t, "new", *req.Description)
						assert.Nil(t, req.ImageUrl)
						assert.Nil(t, req.TopicID)
						return &models.Post{ID: "post-id"}, nil
					}).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "image url over http",
			data:     chatInput{body: `{"image_url":"http://img.test/1.png"}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "topic id is zero",
			data:     chatInput{body: `{"topic_id":0}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "post not found or not yours",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().UpdatePost(gomock.Any(), gomock.Any(), "post-id", "user-id").
					Return(nil, fmt.Errorf("r.repo.BuildUpdatePost: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "update post error",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().UpdatePost(gomock.Any(), gomock.Any(), "post-id", "user-id").
					Return(nil, errors.New("update post error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPatch, "/posts/:post_id", "/posts/post-id", tc.data, posts.Update)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, w.Body.String(), "post-id")
			}
		})
	}
}

func TestDeletePostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "delete post",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().DeletePost(gomock.Any(), "post-id", "user-id").Return(nil).Times(1)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "post not found",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().DeletePost(gomock.Any(), "post-id", "user-id").
					Return(fmt.Errorf("s.repo.DeletePost: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "delete post error",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().DeletePost(gomock.Any(), "post-id", "user-id").
					Return(errors.New("delete post error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, "/posts/:post_id", "/posts/post-id", tc.data, posts.DeletePost)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestViewPostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "view post",
			prepare: func() {
				srvc.EXPECT().ViewPost(gomock.Any(), "post-id").Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "post not found",
			prepare: func() {
				srvc.EXPECT().ViewPost(gomock.Any(), "post-id").
					Return(fmt.Errorf("s.repo.ViewPost: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "view post error",
			prepare: func() {
				srvc.EXPECT().ViewPost(gomock.Any(), "post-id").
					Return(errors.New("view post error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPatch, "/posts/:post_id/view", "/posts/post-id/view", tc.data, posts.ViewPost)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestPostReportHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))
	body := `{"cause":"spam"}`

	testCases := []chatCase{
		{
			caseName: "report post",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().ReportPost(gomock.Any(), gomock.Any(), "user-id", "post-id").
					DoAndReturn(func(_ context.Context, req *dto.ReportDTO, _, _ string) error {
						assert.Equal(t, "spam", req.Cause)
						return nil
					}).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty json fails validation",
			data:     chatInput{body: `{}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "cause too long",
			data:     chatInput{body: `{"cause":"` + strings.Repeat("a", 2000) + `"}`, userID: userClaims("user-id")},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "post not found",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().ReportPost(gomock.Any(), gomock.Any(), "user-id", "post-id").
					Return(fmt.Errorf("s.repo.ReportPost: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "already reported",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().ReportPost(gomock.Any(), gomock.Any(), "user-id", "post-id").
					Return(fmt.Errorf("s.repo.ReportPost: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "report post error",
			data:     chatInput{body: body, userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().ReportPost(gomock.Any(), gomock.Any(), "user-id", "post-id").
					Return(errors.New("report post error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/posts/:post_id/report", "/posts/post-id/report", tc.data, posts.PostReport)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestLikePostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "like post",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().LikePost(gomock.Any(), "user-id", "post-id").Return(nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "post not found",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().LikePost(gomock.Any(), "user-id", "post-id").
					Return(fmt.Errorf("s.repo.LikePost: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "already liked",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().LikePost(gomock.Any(), "user-id", "post-id").
					Return(fmt.Errorf("s.repo.LikePost: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "like post error",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().LikePost(gomock.Any(), "user-id", "post-id").
					Return(errors.New("like post error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, "/posts/:post_id/like", "/posts/post-id/like", tc.data, posts.Like)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUnlikePostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockPostsService(ctrl)

	posts := NewPostHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "unlike post",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().UnlikePost(gomock.Any(), "user-id", "post-id").Return(nil).Times(1)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "like not found",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().UnlikePost(gomock.Any(), "user-id", "post-id").
					Return(fmt.Errorf("s.repo.UnlikePost: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "unlike post error",
			data:     chatInput{userID: userClaims("user-id")},
			prepare: func() {
				srvc.EXPECT().UnlikePost(gomock.Any(), "user-id", "post-id").
					Return(errors.New("unlike post error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, "/posts/:post_id/like", "/posts/post-id/like", tc.data, posts.Unlike)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}
