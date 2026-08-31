package services

import (
	"context"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo"
)

type PostsService struct {
	repo *repo.PostsRepo
}

func NewPostsService(repo *repo.PostsRepo) *PostsService {
	return &PostsService{repo: repo}
}

func (s *PostsService) GetPostsList(ctx context.Context, id string) ([]models.FeedPost, []models.FeedPost, error) {
	sl, st, err := s.repo.GetPostsList(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return sl, st, nil
}
