package handlers

import (
	"context"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/cache"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ProfileService interface {
	GetProfileByID(ctx context.Context, id string) (*models.ProfileResponse, error)
	SaveProfile(ctx context.Context, event *dto.UserProfileDTO) error
	UpdateProfile(ctx context.Context, req *dto.UpdateProfileDTO, id string) (*models.Profile, error)
	GetSubscriptions(ctx context.Context, id string) ([]models.Subscriptions, error)
	GetSubscribers(ctx context.Context, id string) ([]models.Subscribes, error)
	Subscribe(ctx context.Context, userId, targetId string) error
	UnSubscribe(ctx context.Context, userId, targetId string) error
	Report(ctx context.Context, userId, targetId string, req *dto.ReportDTO) error
}

type CommentsService interface {
	ListComments(ctx context.Context, postId string) ([]models.FeedComments, error)
	CreateComment(ctx context.Context, postId, userId string, comment *dto.CreateCommentDTO) error
	UpdateComment(ctx context.Context, postId, userId, commentId string, comment *dto.UpdateCommentDTO) (*models.Comments, error)
	DeleteComment(ctx context.Context, postId, userId, commentId string) error
	ReportComment(ctx context.Context, userId, commentId string, comment *dto.ReportCommentDTO) error
	LikeComment(ctx context.Context, userId, commentId string) error
	UnlikeComment(ctx context.Context, userId, commentId string) error
}

type PostsService interface {
	GetSubPosts(ctx context.Context, id string) ([]models.FeedPost, error)
	GetGlobalPosts(ctx context.Context, id, cursor string) ([]models.FeedPost, error)
	GetTopics(ctx context.Context) ([]models.Topics, error)
	CreatePost(ctx context.Context, req *dto.CreatePostDTO, id string) error
	UpdatePost(ctx context.Context, req *dto.UpdatePostDTO, id string) (*models.Post, error)
	GetPostById(ctx context.Context, postId string) (*models.FeedPost, error)
	ViewPost(ctx context.Context, id string) error
	ReportPost(ctx context.Context, req *dto.ReportDTO, userId, postId string) error
	DeletePost(ctx context.Context, postId string, userId string) error
	LikePost(ctx context.Context, userId, postId string) error
	UnlikePost(ctx context.Context, userId, postId string) error
}

type Services struct {
	profile ProfileService
	post    PostsService
	coms    CommentsService
}

func NewServices(pool *pgxpool.Pool, rdb *redis.Client) *Services {
	profrepo := repo.NewProfileRepo(pool)
	profile := services.NewProfileService(profrepo)

	postrdb := cache.NewCache(rdb)
	postrepo := repo.NewPostsRepo(pool)
	post := services.NewPostsService(postrepo, postrdb)

	comrepo := repo.NewCommentsRepo(pool)
	coms := services.NewCommentsService(comrepo)

	return &Services{profile: profile, post: post, coms: coms}
}

type Handlers struct {
	Profile  *ProfileHandler
	Posts    *PostHandler
	Comments *CommentHandler
}

func NewHandlers(srvc *Services, logger *zap.Logger) *Handlers {
	comments := NewCommentHandler(srvc.coms, logger)
	profile := NewProfileHandler(srvc.profile, logger)
	posts := NewPostHandler(srvc.post, logger)
	return &Handlers{Profile: profile, Posts: posts, Comments: comments}
}
