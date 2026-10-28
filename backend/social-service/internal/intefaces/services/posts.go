package services

import (
	"context"
	"encoding/base64"
	v2 "encoding/json/v2"
	"errors"
	"fmt"
	"strconv"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
)

type PostsService struct {
	repo           PostsRepo
	cache          PostsCache
	PublicEndpoint string
}

func NewPostsService(repo PostsRepo, cache PostsCache, PublicEndpoint string) *PostsService {
	return &PostsService{repo: repo, cache: cache, PublicEndpoint: PublicEndpoint}
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

	posts := make([]models.FeedPost, 0, len(m))
	for _, p := range m {
		posts = append(posts, p)
	}

	res := s.withImageURLs(posts)

	return res, nil
}

func (s *PostsService) GetGlobalPosts(ctx context.Context, cursor string) ([]models.FeedPost, error) {
	var req dto.CursorDTO
	var postID *uuid.UUID
	if cursor != "" {
		decoded, err := base64.URLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, fmt.Errorf("base64.URLEncoding.DecodeString: %w", err)
		}

		if err := v2.Unmarshal(decoded, &req); err != nil {
			return nil, fmt.Errorf("v2.Unmarshal: %w", err)
		}

		if req.PostID == nil || *req.PostID == "" || req.CreatedAt == nil {
			return nil, domain.ErrInvalidCursor
		}

		id, err := uuid.Parse(*req.PostID)
		if err != nil {
			return nil, fmt.Errorf("uuid.Parse: %w", err)
		}

		postID = &id
	}

	posts, err := s.cache.GetGlobal(ctx, cursor)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("s.cache.GetCandidates: %w", err)
	}
	if posts != nil {
		res := s.withImageURLs(posts)
		return res, nil
	}

	posts, err = s.repo.GetGlobalPosts(ctx, postID, req.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetGlobalPosts: %w", err)
	}

	if err := s.cache.SetGlobal(ctx, posts, cursor); err != nil {
		return nil, fmt.Errorf("s.cache.SetGlobal: %w", err)
	}

	res := s.withImageURLs(posts)

	return res, nil
}

func (s *PostsService) imageURL(imageID *string) *string {
	if imageID == nil {
		return nil
	}
	url := s.PublicEndpoint + "/" + *imageID
	return &url
}

func (s *PostsService) withImageURLs(posts []models.FeedPost) []models.FeedPost {
	for i := range posts {
		posts[i].ImageURL = s.imageURL(posts[i].ImageID)
	}
	return posts
}

func (s *PostsService) GetTopics(ctx context.Context) ([]models.Topics, error) {
	tpcs, err := s.repo.GetTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetTopics: %w", err)
	}
	return tpcs, nil
}

func (s *PostsService) GetTopicByID(ctx context.Context, id string) (*models.Topics, error) {
	topic, err := s.repo.GetTopicsByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetTopicsByID: %w", err)
	}
	return topic, nil
}

func (s *PostsService) CreateTopic(ctx context.Context, userID string, req *dto.CreateTopicDTO) error {
	topic := models.Topics{
		UserID: userID,
		Title:  req.Title,
	}

	if err := s.repo.CreateTopic(ctx, &topic); err != nil {
		return fmt.Errorf("s.repo.CreateTopic: %w", err)
	}

	return nil
}

func (s *PostsService) DeleteTopic(ctx context.Context, strTopicID, userID string) error {
	topicID, err := strconv.Atoi(strTopicID)
	if err != nil {
		return fmt.Errorf("strconv.Atoi: %w", err)
	}
	topic := models.Topics{
		ID:     topicID,
		UserID: userID,
	}
	if err := s.repo.DeleteTopic(ctx, &topic); err != nil {
		return fmt.Errorf("s.repo.DeleteTopic: %w", err)
	}

	return nil
}

func (s *PostsService) ReportTopic(ctx context.Context, userID, strTopicID string, req *dto.ReportTopicDTO) error {
	topicID, err := strconv.Atoi(strTopicID)
	if err != nil {
		return fmt.Errorf("strconv.Atoi: %w", err)
	}
	topic := models.TopicsReport{
		UserID:  userID,
		TopicID: topicID,
		Cause:   req.Cause,
	}
	if err := s.repo.ReportTopic(ctx, &topic); err != nil {
		return fmt.Errorf("s.repo.DeleteTopic: %w", err)
	}

	return nil
}

func (s *PostsService) CreatePost(ctx context.Context, req *dto.CreatePostDTO, userID string) error {
	post := models.Post{
		TopicID:     req.TopicID,
		UserID:      userID,
		ImageID:     &req.ImageID,
		Description: &req.Description,
	}
	if err := s.repo.CreatePost(ctx, &post); err != nil {
		return fmt.Errorf("s.repo.CreatePost: %w", err)
	}

	return nil
}

func (s *PostsService) UpdatePost(ctx context.Context, req *dto.UpdatePostDTO, postID, userID string) (*models.Post, error) {
	post := models.Post{
		ID:          postID,
		TopicID:     req.TopicID,
		ImageID:     req.ImageID,
		Description: req.Description,
		UserID:      userID,
	}
	res, err := s.repo.BuildUpdatePost(ctx, &post)
	if err != nil {
		return nil, fmt.Errorf("s.repo.BuildUpdatePost: %w", err)
	}
	return res, nil
}

func (s *PostsService) GetPostByID(ctx context.Context, postID string) (*models.FeedPost, error) {
	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetPostById: %w", err)
	}
	post.ImageURL = s.imageURL(post.ImageID)
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
