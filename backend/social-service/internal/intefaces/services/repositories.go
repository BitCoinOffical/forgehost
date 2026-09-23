package services

import (
	"context"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/google/uuid"
)

type CommentsRepo interface {
	ListComments(ctx context.Context, postId string) ([]models.FeedComments, error)
	CreateComment(ctx context.Context, comment *models.Comments) error
	UpdateComment(ctx context.Context, comment *models.Comments) (*models.Comments, error)
	DeleteComment(ctx context.Context, comment *models.Comments) error
	ReportComment(ctx context.Context, report *models.CommentReport) error
	LikeComment(ctx context.Context, userId, commentId string) error
	UnlikeComment(ctx context.Context, userId, commentId string) error
}
type PostsRepo interface {
	GetPostById(ctx context.Context, id string) (*models.FeedPost, error)
	ViewPost(ctx context.Context, id string) error
	GetSubPosts(ctx context.Context, id string) ([]models.FeedPost, []models.FeedPost, error)
	GetGlobalPosts(ctx context.Context, postId *uuid.UUID, createdAt *time.Time) ([]models.FeedPost, error)
	GetTopics(ctx context.Context) ([]models.Topics, error)
	GetTopicsByID(ctx context.Context, id string) (*models.Topics, error)
	CreatePost(ctx context.Context, post *models.Post) error
	BuildUpdatePost(ctx context.Context, post *models.Post) (*models.Post, error)
	DeletePost(ctx context.Context, postId, userId string) error
	ReportPost(ctx context.Context, report *models.PostReport) error
	LikePost(ctx context.Context, userId, postId string) error
	UnlikePost(ctx context.Context, userId, postId string) error
}

type PostsCache interface {
	GetGlobal(ctx context.Context, cursor string) ([]models.FeedPost, error)
	SetGlobal(ctx context.Context, posts []models.FeedPost, cursor string) error
}

type ProfileRepo interface {
	GetProfileByID(ctx context.Context, id string) (*models.Profile, []models.FeedPost, error)
	SaveProfile(ctx context.Context, userId string) error
	BuildUpdateProfile(ctx context.Context, profile *models.Profile) (*models.Profile, error)
	GetSubscriptions(ctx context.Context, id string) ([]models.Subscriptions, error)
	GetSubscribers(ctx context.Context, id string) ([]models.Subscribes, error)
	Subscribe(ctx context.Context, userId, targetId string) error
	UnSubscribe(ctx context.Context, userId, targetId string) error
	CreateProfileReport(ctx context.Context, userId, targetId, cause string) error
}
