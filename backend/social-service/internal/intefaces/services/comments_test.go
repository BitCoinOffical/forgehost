package services

import (
	"errors"
	"testing"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	store_mocks "github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestListComments(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockCommentsRepo(ctrl)

	comments := NewCommentsService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			postID string
		}
		prepare func()
		want    []models.FeedComments
		success bool
	}{
		{
			caseName: "list comments",
			data: struct {
				postID string
			}{postID: "post_id"},
			prepare: func() {
				repo.EXPECT().ListComments(gomock.Any(), "post_id").
					Return([]models.FeedComments{{Body: "first"}, {Body: "second"}}, nil).Times(1)
			},
			want:    []models.FeedComments{{Body: "first"}, {Body: "second"}},
			success: true,
		},
		{
			caseName: "list empty comments",
			data: struct {
				postID string
			}{postID: "post_id"},
			prepare: func() {
				repo.EXPECT().ListComments(gomock.Any(), "post_id").
					Return(nil, nil).Times(1)
			},
			want:    nil,
			success: true,
		},
		{
			caseName: "list comments error",
			data: struct {
				postID string
			}{postID: "post_id"},
			prepare: func() {
				repo.EXPECT().ListComments(gomock.Any(), "post_id").
					Return(nil, errors.New("list comments error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			got, err := comments.ListComments(t.Context(), tc.data.postID)
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

func TestCreateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockCommentsRepo(ctrl)

	comments := NewCommentsService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			postID   string
			userID   string
			parentID *string
			body     string
		}
		prepare func()
		success bool
	}{
		{
			caseName: "create comment",
			data: struct {
				postID   string
				userID   string
				parentID *string
				body     string
			}{postID: "post_id", userID: "user_id", body: "body"},
			prepare: func() {
				repo.EXPECT().CreateComment(gomock.Any(),
					&models.Comments{PostID: "post_id", UserID: "user_id", Body: "body"}).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "create reply",
			data: struct {
				postID   string
				userID   string
				parentID *string
				body     string
			}{postID: "post_id", userID: "user_id", parentID: strPtr("parent_id"), body: "reply"},
			prepare: func() {
				repo.EXPECT().CreateComment(gomock.Any(),
					&models.Comments{PostID: "post_id", UserID: "user_id", ParentID: strPtr("parent_id"), Body: "reply"}).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "create comment error",
			data: struct {
				postID   string
				userID   string
				parentID *string
				body     string
			}{postID: "post_id", userID: "user_id", body: "body"},
			prepare: func() {
				repo.EXPECT().CreateComment(gomock.Any(), gomock.Any()).
					Return(errors.New("create comment error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.CreateCommentDTO{
				ParentID: tc.data.parentID,
				Body:     tc.data.body,
			}
			err := comments.CreateComment(t.Context(), tc.data.postID, tc.data.userID, &req)
			if tc.success {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockCommentsRepo(ctrl)

	comments := NewCommentsService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			postID    string
			userID    string
			commentID string
			body      string
		}
		prepare func()
		want    *models.Comments
		wantErr error
		success bool
	}{
		{
			caseName: "update comment",
			data: struct {
				postID    string
				userID    string
				commentID string
				body      string
			}{postID: "post_id", userID: "user_id", commentID: "comment_id", body: "new_body"},
			prepare: func() {
				repo.EXPECT().UpdateComment(gomock.Any(),
					&models.Comments{ID: "comment_id", PostID: "post_id", UserID: "user_id", Body: "new_body"}).
					Return(&models.Comments{ID: "comment_id", PostID: "post_id", UserID: "user_id", Body: "new_body"}, nil).Times(1)
			},
			want:    &models.Comments{ID: "comment_id", PostID: "post_id", UserID: "user_id", Body: "new_body"},
			success: true,
		},
		{
			caseName: "update not found comment",
			data: struct {
				postID    string
				userID    string
				commentID string
				body      string
			}{postID: "post_id", userID: "user_id", commentID: "comment_id", body: "new_body"},
			prepare: func() {
				repo.EXPECT().UpdateComment(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "update comment error",
			data: struct {
				postID    string
				userID    string
				commentID string
				body      string
			}{postID: "post_id", userID: "user_id", commentID: "comment_id", body: "new_body"},
			prepare: func() {
				repo.EXPECT().UpdateComment(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("update comment error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.UpdateCommentDTO{
				Body: tc.data.body,
			}
			got, err := comments.UpdateComment(t.Context(), tc.data.postID, tc.data.userID, tc.data.commentID, &req)
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

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockCommentsRepo(ctrl)

	comments := NewCommentsService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			postID    string
			userID    string
			commentID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "delete comment",
			data: struct {
				postID    string
				userID    string
				commentID string
			}{postID: "post_id", userID: "user_id", commentID: "comment_id"},
			prepare: func() {
				repo.EXPECT().DeleteComment(gomock.Any(),
					&models.Comments{ID: "comment_id", PostID: "post_id", UserID: "user_id"}).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "delete not found comment",
			data: struct {
				postID    string
				userID    string
				commentID string
			}{postID: "post_id", userID: "user_id", commentID: "comment_id"},
			prepare: func() {
				repo.EXPECT().DeleteComment(gomock.Any(), gomock.Any()).
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "delete comment error",
			data: struct {
				postID    string
				userID    string
				commentID string
			}{postID: "post_id", userID: "user_id", commentID: "comment_id"},
			prepare: func() {
				repo.EXPECT().DeleteComment(gomock.Any(), gomock.Any()).
					Return(errors.New("delete comment error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := comments.DeleteComment(t.Context(), tc.data.postID, tc.data.userID, tc.data.commentID)
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

func TestReportComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockCommentsRepo(ctrl)

	comments := NewCommentsService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID    string
			commentID string
			cause     string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "report comment",
			data: struct {
				userID    string
				commentID string
				cause     string
			}{userID: "user_id", commentID: "comment_id", cause: "spam"},
			prepare: func() {
				repo.EXPECT().ReportComment(gomock.Any(),
					&models.CommentReport{UserID: "user_id", CommentID: "comment_id", Cause: "spam"}).
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "report comment twice",
			data: struct {
				userID    string
				commentID string
				cause     string
			}{userID: "user_id", commentID: "comment_id", cause: "spam"},
			prepare: func() {
				repo.EXPECT().ReportComment(gomock.Any(), gomock.Any()).
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "report comment error",
			data: struct {
				userID    string
				commentID string
				cause     string
			}{userID: "user_id", commentID: "comment_id", cause: "spam"},
			prepare: func() {
				repo.EXPECT().ReportComment(gomock.Any(), gomock.Any()).
					Return(errors.New("report comment error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			req := dto.ReportCommentDTO{
				Cause: tc.data.cause,
			}
			err := comments.ReportComment(t.Context(), tc.data.userID, tc.data.commentID, &req)
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

func TestLikeComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockCommentsRepo(ctrl)

	comments := NewCommentsService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID    string
			commentID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "like comment",
			data: struct {
				userID    string
				commentID string
			}{userID: "user_id", commentID: "comment_id"},
			prepare: func() {
				repo.EXPECT().LikeComment(gomock.Any(), "user_id", "comment_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "like comment twice",
			data: struct {
				userID    string
				commentID string
			}{userID: "user_id", commentID: "comment_id"},
			prepare: func() {
				repo.EXPECT().LikeComment(gomock.Any(), "user_id", "comment_id").
					Return(domain.ErrAlreadyExists).Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
			success: false,
		},
		{
			caseName: "like comment error",
			data: struct {
				userID    string
				commentID string
			}{userID: "user_id", commentID: "comment_id"},
			prepare: func() {
				repo.EXPECT().LikeComment(gomock.Any(), "user_id", "comment_id").
					Return(errors.New("like comment error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := comments.LikeComment(t.Context(), tc.data.userID, tc.data.commentID)
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

func TestUnlikeComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := store_mocks.NewMockCommentsRepo(ctrl)

	comments := NewCommentsService(repo)

	testCases := []struct {
		caseName string
		data     struct {
			userID    string
			commentID string
		}
		prepare func()
		wantErr error
		success bool
	}{
		{
			caseName: "unlike comment",
			data: struct {
				userID    string
				commentID string
			}{userID: "user_id", commentID: "comment_id"},
			prepare: func() {
				repo.EXPECT().UnlikeComment(gomock.Any(), "user_id", "comment_id").
					Return(nil).Times(1)
			},
			success: true,
		},
		{
			caseName: "unlike not liked comment",
			data: struct {
				userID    string
				commentID string
			}{userID: "user_id", commentID: "comment_id"},
			prepare: func() {
				repo.EXPECT().UnlikeComment(gomock.Any(), "user_id", "comment_id").
					Return(domain.ErrNotFound).Times(1)
			},
			wantErr: domain.ErrNotFound,
			success: false,
		},
		{
			caseName: "unlike comment error",
			data: struct {
				userID    string
				commentID string
			}{userID: "user_id", commentID: "comment_id"},
			prepare: func() {
				repo.EXPECT().UnlikeComment(gomock.Any(), "user_id", "comment_id").
					Return(errors.New("unlike comment error")).Times(1)
			},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			tc.prepare()

			err := comments.UnlikeComment(t.Context(), tc.data.userID, tc.data.commentID)
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
