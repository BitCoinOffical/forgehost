package handlers

import (
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/cache"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Services struct {
	profile  *services.ProfileService
	post     *services.PostsService
	commSrvc *services.CommentsService
}

func NewServices(pool *pgxpool.Pool, rdb *redis.Client) *Services {
	profrepo := repo.NewProfileRepo(pool)
	profile := services.NewProfileService(profrepo)

	postrdb := cache.NewCache(rdb)
	postrepo := repo.NewPostsRepo(pool)
	post := services.NewPostsService(postrepo, postrdb)

	commRepo := repo.NewCommentsRepo(pool)
	commSrvc := services.NewCommentsService(commRepo)
	return &Services{profile: profile, post: post, commSrvc: commSrvc}
}

type Handlers struct {
	Profile  *ProfileHandler
	Posts    *PostHandler
	Comments *CommentHandler
}

func NewHandlers(srvc *Services, logger *zap.Logger) *Handlers {
	profile := NewProfileHandler(srvc.profile, logger)
	posts := NewPostHandler(srvc.post, logger)
	comments := NewCommentHandler(srvc.commSrvc, logger)
	return &Handlers{Profile: profile, Posts: posts, Comments: comments}
}
