package services

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const (
	SearchTypePosts   = "posts"
	SearchTypeStreams = "streams"
	SearchTypeUsers   = "users"
)

type SearchService struct {
	repo           SearchRepo
	PublicEndpoint string
}

func NewSearchService(repo SearchRepo, publicEndpoint string) *SearchService {
	return &SearchService{repo: repo, PublicEndpoint: publicEndpoint}
}

func (s *SearchService) imageURL(imageID *string) *string {
	if imageID == nil {
		return nil
	}
	url := s.PublicEndpoint + "/" + *imageID
	return &url
}

func (s *SearchService) Search(ctx context.Context, q, typ string, limit int) (*models.SearchResponse, error) {
	var res models.SearchResponse
	g, ctx := errgroup.WithContext(ctx)

	if typ == "" || typ == SearchTypePosts {
		g.Go(func() error {
			posts, err := s.repo.SearchPosts(ctx, q, limit)
			if err != nil {
				return fmt.Errorf("s.repo.SearchPosts: %w", err)
			}
			for i := range posts {
				posts[i].ImageURL = s.imageURL(posts[i].ImageID)
			}
			res.Posts = posts
			return nil
		})
	}
	if typ == "" || typ == SearchTypeStreams {
		g.Go(func() (err error) {
			if res.Streams, err = s.repo.SearchStreams(ctx, q, limit); err != nil {
				return fmt.Errorf("s.repo.SearchStreams: %w", err)
			}
			return nil
		})
	}
	if typ == "" || typ == SearchTypeUsers {
		g.Go(func() error {
			users, err := s.repo.SearchUsers(ctx, q, limit)
			if err != nil {
				return fmt.Errorf("s.repo.SearchUsers: %w", err)
			}
			for i := range users {
				users[i].AvatarURL = s.imageURL(users[i].AvatarURL)
			}
			res.Users = users
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("g.Wait: %w", err)
	}
	return &res, nil
}

func (s *SearchService) Suggest(ctx context.Context, userIDStr string, q string, limit int) ([]models.Suggestion, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	var users, streams []models.Suggestion
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() (err error) {
		if users, err = s.repo.SuggestUsers(ctx, userID, q, limit); err != nil {
			return fmt.Errorf("s.repo.SuggestUsers: %w", err)
		}
		return nil
	})

	g.Go(func() (err error) {
		if streams, err = s.repo.SuggestStreams(ctx, q, limit); err != nil {
			return fmt.Errorf("s.repo.SuggestStreams: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("g.Wait: %w", err)
	}

	res := append(users, streams...)
	if len(res) > limit {
		res = res[:limit]
	}
	return res, nil
}
