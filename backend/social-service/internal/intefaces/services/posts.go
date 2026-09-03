package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/cache"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo"
)

type PostsService struct {
	repo  *repo.PostsRepo
	cache *cache.Cache
}

func NewPostsService(repo *repo.PostsRepo, cache *cache.Cache) *PostsService {
	return &PostsService{repo: repo, cache: cache}
}

func (s *PostsService) GetSubPosts(ctx context.Context, id string) ([]models.FeedPost, error) {
	sl, st, err := s.repo.GetSubPosts(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetPostsList: %w", err)
	}

	m := make(map[int]models.FeedPost, len(sl)+len(st))
	for _, data := range sl {
		m[data.PostID] = data
	}
	for _, data := range st {
		m[data.PostID] = data
	}

	result := make([]models.FeedPost, 0, len(m))
	for _, p := range m {
		result = append(result, p)
	}

	return result, nil
}

func (r *PostsService) GetGlobalPosts(ctx context.Context, id, cursor string) ([]models.FeedPost, error) {
	var req dto.CursorDTO
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("base64.URLEncoding.DecodeString: %w", err)
	}

	if err := json.Unmarshal(decoded, &req); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	posts, err := r.cache.GetGlobal(ctx, cursor)
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("r.cache.GetCandidates: %w", err)
	}
	if posts != nil {
		return posts, nil
	}

	posts, err = r.repo.GetGlobalPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.repo.GetGlobalPosts: %w", err)
	}

	if err := r.cache.SetGlobal(ctx, posts, cursor); err != nil {
		return nil, fmt.Errorf("r.cache.SetGlobal: %w", err)
	}

	return posts, nil
}

func (r *PostsService) GetPostById(ctx context.Context, id string) (*models.Post, error) {
	post, err := r.repo.GetPostById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("r.repo.GetPostById: %w", err)
	}
	return post, nil
}

func (s *PostsService) ViewPost(ctx context.Context, id string) error {
	if err := s.repo.ViewPost(ctx, id); err != nil {
		return fmt.Errorf("s.repo.ViewPost: %w", err)
	}
	return nil
}
