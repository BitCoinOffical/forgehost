package services

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo"
)

type CommentsService struct {
	repo *repo.CommentsRepo
}

func NewCommentsService(repo *repo.CommentsRepo) *CommentsService {
	return &CommentsService{repo: repo}
}

func (s *CommentsService) ListComments(ctx context.Context, postId string) ([]models.FeedComments, error) {
	coments, err := s.repo.ListComments(ctx, postId)
	if err != nil {
		return nil, fmt.Errorf("s.repo.ListComments: %w", err)
	}
	return coments, nil
}

func (s *CommentsService) CreateComment(ctx context.Context, postId, userId string, comment *dto.CreateCommentDTO) error {
	comm := models.Comments{
		PostID:   postId,
		UserID:   userId,
		ParentID: comment.ParentID,
		Body:     comment.Body,
	}
	s.repo.CreateComment(ctx, &comm)
	return nil
}

func (s *CommentsService) UpdateComment(ctx context.Context, postId, userId, commentId string, comment *dto.UpdateCommentDTO) (*models.Comments, error) {
	comm := models.Comments{
		ID:     commentId,
		PostID: postId,
		UserID: userId,
		Body:   comment.Body,
	}
	cmt, err := s.repo.UpdateComment(ctx, &comm)
	if err != nil {
		return nil, fmt.Errorf("s.repo.UpdateComment: %w", err)
	}
	return cmt, nil
}

func (s *CommentsService) DeleteComment(ctx context.Context, postId, userId, commentId string) error {
	comm := models.Comments{
		ID:     commentId,
		PostID: postId,
		UserID: userId,
	}

	if err := s.repo.DeleteComment(ctx, &comm); err != nil {
		return fmt.Errorf("s.repo.DeleteComment: %w", err)
	}

	return nil
}

func (s *CommentsService) ReportComment(ctx context.Context, userId, commentId string, comment *dto.ReportCommentDTO) error {
	comm := models.CommentReport{
		CommentId: commentId,
		UserId:    userId,
		Cause:     comment.Cause,
	}

	if err := s.repo.ReportComment(ctx, &comm); err != nil {
		return fmt.Errorf("s.repo.ReportComment: %w", err)
	}

	return nil
}

func (s *CommentsService) LikeComment(ctx context.Context, userId, comentId string) error {
	if err := s.repo.LikeComment(ctx, userId, comentId); err != nil {
		return fmt.Errorf("s.repo.LikeComment: %w", err)
	}

	return nil
}
func (s *CommentsService) UnlikeComment(ctx context.Context, userId, comentId string) error {
	if err := s.repo.UnlikeComment(ctx, userId, comentId); err != nil {
		return fmt.Errorf("s.repo.UnlikeComment: %w", err)
	}

	return nil
}
