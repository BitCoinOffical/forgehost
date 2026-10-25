package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers/mock"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

const (
	testPostID    = "post-id"
	testCommentID = "comment-id"

	commentsRoute    = "/posts/:post_id/comments"
	commentsURL      = "/posts/" + testPostID + "/comments"
	commentRoute     = "/posts/:post_id/comments/:comment_id"
	commentURL       = "/posts/" + testPostID + "/comments/" + testCommentID
	commentReportURL = commentURL + "/report"
	commentLikeURL   = commentURL + "/like"
)

func TestListCommentsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockCommentsService(ctrl)

	comments := NewCommentHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "list comments",
			prepare: func() {
				srvc.EXPECT().ListComments(gomock.Any(), testPostID).
					Return([]models.FeedComments{}, nil).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "list comments error",
			prepare: func() {
				srvc.EXPECT().ListComments(gomock.Any(), testPostID).
					Return(nil, errors.New("list comments error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodGet, commentsRoute, commentsURL, tc.data, comments.List)

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestCreateCommentHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockCommentsService(ctrl)

	comments := NewCommentHandler(srvc, zaptest.NewLogger(t))

	parentID := uuid.NewString()
	body := `{"body":"hello"}`

	testCases := []chatCase{
		{
			caseName: "create comment",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().CreateComment(gomock.Any(), testPostID, testUserID, gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, req *dto.CreateCommentDTO) error {
						assert.Equal(t, "hello", req.Body)
						assert.Nil(t, req.ParentID)
						return nil
					}).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "create reply",
			data:     chatInput{body: `{"body":"reply","parent_id":"` + parentID + `"}`, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().CreateComment(gomock.Any(), testPostID, testUserID, gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, req *dto.CreateCommentDTO) error {
						require.NotNil(t, req.ParentID)
						assert.Equal(t, parentID, *req.ParentID)
						return nil
					}).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty json fails validation",
			data:     chatInput{body: `{}`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty body text",
			data:     chatInput{body: `{"body":""}`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "body too long",
			data:     chatInput{body: `{"body":"` + strings.Repeat("a", 2001) + `"}`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "post or parent not found",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().CreateComment(gomock.Any(), testPostID, testUserID, gomock.Any()).
					Return(fmt.Errorf("s.repo.CreateComment: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "create comment error",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().CreateComment(gomock.Any(), testPostID, testUserID, gomock.Any()).
					Return(errors.New("create comment error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, commentsRoute, commentsURL, tc.data, comments.Create)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUpdateCommentHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockCommentsService(ctrl)

	comments := NewCommentHandler(srvc, zaptest.NewLogger(t))
	body := `{"body":"edited"}`

	testCases := []chatCase{
		{
			caseName: "update comment",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UpdateComment(gomock.Any(), testPostID, testUserID, testCommentID, gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _, _ string, req *dto.UpdateCommentDTO) (*models.Comments, error) {
						assert.Equal(t, "edited", req.Body)
						return &models.Comments{}, nil
					}).Times(1)
			},
			status: http.StatusOK,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty body text",
			data:     chatInput{body: `{"body":""}`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "null body",
			data:     chatInput{body: `null`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "comment not found or not yours",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UpdateComment(gomock.Any(), testPostID, testUserID, testCommentID, gomock.Any()).
					Return(nil, fmt.Errorf("s.repo.UpdateComment: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "update comment error",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UpdateComment(gomock.Any(), testPostID, testUserID, testCommentID, gomock.Any()).
					Return(nil, errors.New("update comment error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			var w *httptest.ResponseRecorder
			require.NotPanics(t, func() {
				w = serveChat(t, http.MethodPut, commentRoute, commentURL, tc.data, comments.Update)
			})

			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestDeleteCommentHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockCommentsService(ctrl)

	comments := NewCommentHandler(srvc, zaptest.NewLogger(t))

	testCases := []chatCase{
		{
			caseName: "delete comment",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().DeleteComment(gomock.Any(), testPostID, testUserID, testCommentID).
					Return(nil).Times(1)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "delete comment error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().DeleteComment(gomock.Any(), testPostID, testUserID, testCommentID).
					Return(errors.New("delete comment error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, commentRoute, commentURL, tc.data, comments.Delete)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestReportCommentHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockCommentsService(ctrl)

	comments := NewCommentHandler(srvc, zaptest.NewLogger(t))
	route := commentRoute + "/report"
	body := `{"cause":"spam"}`

	testCases := []chatCase{
		{
			caseName: "report comment",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().ReportComment(gomock.Any(), testUserID, testCommentID, gomock.Any()).
					DoAndReturn(func(_ context.Context, _, _ string, req *dto.ReportCommentDTO) error {
						assert.Equal(t, "spam", req.Cause)
						return nil
					}).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "invalid request body",
			data:     chatInput{body: `{invalid`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "empty json fails validation",
			data:     chatInput{body: `{}`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "cause too long",
			data:     chatInput{body: `{"cause":"` + strings.Repeat("a", 256) + `"}`, userID: userClaims(testUserID)},
			prepare:  func() {},
			status:   http.StatusBadRequest,
		},
		{
			caseName: "comment not found",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().ReportComment(gomock.Any(), testUserID, testCommentID, gomock.Any()).
					Return(fmt.Errorf("s.repo.ReportComment: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "already reported",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().ReportComment(gomock.Any(), testUserID, testCommentID, gomock.Any()).
					Return(fmt.Errorf("s.repo.ReportComment: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "report comment error",
			data:     chatInput{body: body, userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().ReportComment(gomock.Any(), testUserID, testCommentID, gomock.Any()).
					Return(errors.New("report comment error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases(body)...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, route, commentReportURL, tc.data, comments.Report)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestLikeCommentHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockCommentsService(ctrl)

	comments := NewCommentHandler(srvc, zaptest.NewLogger(t))
	route := commentRoute + "/like"

	testCases := []chatCase{
		{
			caseName: "like comment",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().LikeComment(gomock.Any(), testUserID, testCommentID).Return(nil).Times(1)
			},
			status: http.StatusCreated,
		},
		{
			caseName: "comment not found",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().LikeComment(gomock.Any(), testUserID, testCommentID).
					Return(fmt.Errorf("s.repo.LikeComment: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "already liked",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().LikeComment(gomock.Any(), testUserID, testCommentID).
					Return(fmt.Errorf("s.repo.LikeComment: %w", domain.ErrAlreadyExists)).Times(1)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "like comment error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().LikeComment(gomock.Any(), testUserID, testCommentID).
					Return(errors.New("like comment error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodPost, route, commentLikeURL, tc.data, comments.Like)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUnlikeCommentHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	srvc := srvc_mocks.NewMockCommentsService(ctrl)

	comments := NewCommentHandler(srvc, zaptest.NewLogger(t))
	route := commentRoute + "/like"

	testCases := []chatCase{
		{
			caseName: "unlike comment",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UnlikeComment(gomock.Any(), testUserID, testCommentID).Return(nil).Times(1)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "like not found",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UnlikeComment(gomock.Any(), testUserID, testCommentID).
					Return(fmt.Errorf("s.repo.UnlikeComment: %w", domain.ErrNotFound)).Times(1)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "unlike comment error",
			data:     chatInput{userID: userClaims(testUserID)},
			prepare: func() {
				srvc.EXPECT().UnlikeComment(gomock.Any(), testUserID, testCommentID).
					Return(errors.New("unlike comment error")).Times(1)
			},
			status: http.StatusInternalServerError,
		},
	}
	testCases = append(testCases, authCases("")...)

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			w := serveChat(t, http.MethodDelete, route, commentLikeURL, tc.data, comments.Unlike)

			assert.Equal(t, tc.status, w.Code)
		})
	}
}
