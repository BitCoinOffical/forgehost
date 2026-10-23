package services

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
)

type CommentsService struct {
	repo CommentsRepo
}

func NewCommentsService(repo CommentsRepo) *CommentsService {
	return &CommentsService{repo: repo}
}

func (s *CommentsService) ListComments(ctx context.Context, postID string) ([]models.FeedComments, error) {
	comments, err := s.repo.ListComments(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.ListComments: %w", err)
	}
	return comments, nil
}

func (s *CommentsService) CreateComment(ctx context.Context, postID, userID string, comment *dto.CreateCommentDTO) error {
	comm := models.Comments{
		PostID:   postID,
		UserID:   userID,
		ParentID: comment.ParentID,
		Body:     comment.Body,
	}
	if err := s.repo.CreateComment(ctx, &comm); err != nil {
		return fmt.Errorf("s.repo.CreateComment: %w", err)
	}
	return nil
}

func (s *CommentsService) UpdateComment(ctx context.Context, postID, userID, commentID string, comment *dto.UpdateCommentDTO) (*models.Comments, error) {
	comm := models.Comments{
		ID:     commentID,
		PostID: postID,
		UserID: userID,
		Body:   comment.Body,
	}
	cmt, err := s.repo.UpdateComment(ctx, &comm)
	if err != nil {
		return nil, fmt.Errorf("s.repo.UpdateComment: %w", err)
	}
	return cmt, nil
}

func (s *CommentsService) DeleteComment(ctx context.Context, postID, userID, commentID string) error {
	comm := models.Comments{
		ID:     commentID,
		PostID: postID,
		UserID: userID,
	}

	if err := s.repo.DeleteComment(ctx, &comm); err != nil {
		return fmt.Errorf("s.repo.DeleteComment: %w", err)
	}

	return nil
}

func (s *CommentsService) ReportComment(ctx context.Context, userID, commentID string, comment *dto.ReportCommentDTO) error {
	comm := models.CommentReport{
		CommentID: commentID,
		UserID:    userID,
		Cause:     comment.Cause,
	}

	if err := s.repo.ReportComment(ctx, &comm); err != nil {
		return fmt.Errorf("s.repo.ReportComment: %w", err)
	}

	return nil
}

func (s *CommentsService) LikeComment(ctx context.Context, userID, commentID string) error {
	if err := s.repo.LikeComment(ctx, userID, commentID); err != nil {
		return fmt.Errorf("s.repo.LikeComment: %w", err)
	}

	return nil
}

func (s *CommentsService) UnlikeComment(ctx context.Context, userID, commentID string) error {
	if err := s.repo.UnlikeComment(ctx, userID, commentID); err != nil {
		return fmt.Errorf("s.repo.UnlikeComment: %w", err)
	}

	return nil
}
