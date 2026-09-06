package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/config"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

const (
	timeoutSecond = 5
)

type Server struct {
	engine *gin.Engine
	server *http.Server
	h      *handlers.Handlers
	m      *middleware.Middleware
}

func NewServer(cfg *config.AppConfig, m *middleware.Middleware, h *handlers.Handlers) *Server {
	engine := gin.New()
	return &Server{
		h:      h,
		m:      m,
		engine: engine,
		server: &http.Server{
			Addr:        ":" + cfg.Port,
			Handler:     engine,
			ReadTimeout: timeoutSecond * time.Second,
		},
	}
}

func (s *Server) Run() error {
	api := s.engine.Group("/api/v1")
	social := api.Group("/social")
	social.Use(s.m.AuthMiddleware())
	social.Use(s.m.RateLimiter())
	{
		social.GET("/profile/me", s.h.Profile.Me)
		social.GET("/profile/:user_id", s.h.Profile.GetProfileByID)
		social.PATCH("/profile", s.h.Profile.UpdateProfile)
		social.GET("/profile/:user_id/subscribers", s.h.Profile.GetSubscribers)
		social.GET("/profile/:user_id/subscriptions", s.h.Profile.GetSubscriptions)
		social.POST("/profile/:user_id/subscribe", s.h.Profile.Subscribe)
		social.DELETE("/profile/:user_id/unsubscribe", s.h.Profile.Unsubscribe)
		social.POST("/profile/:user_id/report", s.h.Profile.Report)

		social.GET("/posts", s.h.Posts.GetSubPosts)
		social.GET("/posts/global/:cursor", s.h.Posts.GetGlobalPosts)
		social.GET("/posts/:post_id", s.h.Posts.GetByID)
		social.POST("/posts", s.h.Posts.CreatePost)
		social.PATCH("/posts/:post_id", s.h.Posts.Update)
		social.DELETE("/posts/:post_id", s.h.Posts.DeletePost)
		social.PATCH("/posts/:post_id/view", s.h.Posts.ViewPost)
		social.POST("/posts/:post_id/report", s.h.Posts.PostReport)
		social.POST("/posts/:post_id/like", s.h.Posts.Like)
		social.DELETE("/posts/:post_id/like", s.h.Posts.Unlike)

		social.GET("/posts/:post_id/comments", s.h.Comments.Like)
		social.POST("/posts/:post_id/comments", s.h.Comments.Create)
		social.PUT("/posts/:post_id/comments/:comment_id", s.h.Comments.Update)
		social.POST("/posts/:post_id/comments/:comment_id/report", s.h.Comments.Report)
		social.DELETE("/posts/:post_id/comments/:comment_id", s.h.Comments.Delete)
		social.POST("/posts/:post_id/comments/:comment_id/like", s.h.Comments.Like)
		social.DELETE("/posts/:post_id/comments/:comment_id/like", s.h.Comments.Unlike)

		social.GET("/topics", s.h.Posts.GetTopics)
	}

	return s.server.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
