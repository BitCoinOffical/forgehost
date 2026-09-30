package services

import (
	"context"
	"encoding/base64"
	v2 "encoding/json/v2"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
)

type PostsService struct {
	repo  PostsRepo
	cache PostsCache
}

func NewPostsService(repo PostsRepo, cache PostsCache) *PostsService {
	return &PostsService{repo: repo, cache: cache}
}

func (s *PostsService) GetSubPosts(ctx context.Context, id string) ([]models.FeedPost, error) {
	sl, st, err := s.repo.GetSubPosts(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetPostsList: %w", err)
	}

	m := make(map[string]models.FeedPost, len(sl)+len(st))
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
	var postID *uuid.UUID
	if cursor != "" {
		decoded, err := base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, fmt.Errorf("base64.URLEncoding.DecodeString: %w", err)
		}

		if err := v2.Unmarshal(decoded, req); err != nil {
			return nil, fmt.Errorf("v2.Unmarshal: %w", err)
		}

		id, err := uuid.Parse(*req.PostID)
		if err != nil {
			return nil, fmt.Errorf("uuid.Parse: %w", err)
		}

		postID = &id
	}

	posts, err := r.cache.GetGlobal(ctx, cursor)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("r.cache.GetCandidates: %w", err)
	}
	if posts != nil {
		return posts, nil
	}

	posts, err = r.repo.GetGlobalPosts(ctx, postID, req.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("r.repo.GetGlobalPosts: %w", err)
	}

	if err := r.cache.SetGlobal(ctx, posts, cursor); err != nil {
		return nil, fmt.Errorf("r.cache.SetGlobal: %w", err)
	}

	return posts, nil
}

func (r *PostsService) GetTopics(ctx context.Context) ([]models.Topics, error) {
	tpcs, err := r.repo.GetTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.repo.GetTopics: %w", err)
	}
	return tpcs, nil
}

func (r *PostsService) GetTopicByID(ctx context.Context, id string) (*models.Topics, error) {
	topic, err := r.repo.GetTopicsByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("r.repo.GetTopicsByID: %w", err)
	}
	return topic, nil
}

func (r *PostsService) CreateTopic(ctx context.Context, userID string, req *dto.CreateTopicDTO) error {
	topic := models.Topics{
		UserID: userID,
		Title:  req.Title,
	}

	if err := r.repo.CreateTopic(ctx, &topic); err != nil {
		return fmt.Errorf("r.repo.CreateTopic: %w", err)
	}

	return nil
}

func (r *PostsService) DeleteTopic(ctx context.Context, topicID, userID string) error {
	topic := models.Topics{
		ID:     topicID,
		UserID: userID,
	}
	if err := r.repo.DeleteTopic(ctx, &topic); err != nil {
		return fmt.Errorf("r.repo.DeleteTopic: %w", err)
	}

	return nil
}

func (r *PostsService) ReportTopic(ctx context.Context, userID, topicID string, req *dto.ReportTopicDTO) error {
	topic := models.TopicsReport{
		UserID:  userID,
		TopicID: topicID,
		Cause:   req.Cause,
	}
	if err := r.repo.ReportTopic(ctx, &topic); err != nil {
		return fmt.Errorf("r.repo.DeleteTopic: %w", err)
	}

	return nil
}

func (r *PostsService) CreatePost(ctx context.Context, req *dto.CreatePostDTO, id string) error {
	post := models.Post{
		TopicID:     req.TopicID,
		UserID:      id,
		ImageURL:    &req.ImageUrl,
		Description: &req.Description,
	}
	if err := r.repo.CreatePost(ctx, &post); err != nil {
		return fmt.Errorf("r.repo.CreatePost: %w", err)
	}

	return nil
}

func (r *PostsService) UpdatePost(ctx context.Context, req *dto.UpdatePostDTO, id string) (*models.Post, error) {
	post := models.Post{
		ID:          req.PostID,
		TopicID:     req.TopicID,
		ImageURL:    &req.ImageUrl,
		Description: &req.Description,
		UserID:      id,
	}
	res, err := r.repo.BuildUpdatePost(ctx, &post)
	if err != nil {
		return nil, fmt.Errorf("r.repo.BuildUpdatePost: %w", err)
	}
	return res, nil
}

func (r *PostsService) GetPostByID(ctx context.Context, postID string) (*models.FeedPost, error) {
	post, err := r.repo.GetPostByID(ctx, postID)
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

func (s *PostsService) ReportPost(ctx context.Context, req *dto.ReportDTO, userID, postID string) error {
	report := models.PostReport{
		UserID: userID,
		PostID: postID,
		Cause:  req.Cause,
	}

	if err := s.repo.ReportPost(ctx, &report); err != nil {
		return fmt.Errorf("s.repo.ReportPost: %w", err)
	}

	return nil
}

func (s *PostsService) DeletePost(ctx context.Context, postID string, userID string) error {
	if err := s.repo.DeletePost(ctx, postID, userID); err != nil {
		return fmt.Errorf("s.repo.DeletePost: %w", err)
	}
	return nil
}

func (s *PostsService) LikePost(ctx context.Context, userID, postID string) error {
	if err := s.repo.LikePost(ctx, userID, postID); err != nil {
		return fmt.Errorf("s.repo.LikePost: %w", err)
	}
	return nil
}
func (s *PostsService) UnlikePost(ctx context.Context, userID, postID string) error {
	if err := s.repo.UnlikePost(ctx, userID, postID); err != nil {
		return fmt.Errorf("s.repo.UnlikePost: %w", err)
	}
	return nil
}
