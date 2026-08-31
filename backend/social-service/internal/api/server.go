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

		social.GET("/posts")
		social.GET("/posts/:post_id")
		social.POST("/posts")
		social.PATCH("/posts/:post_id")
		social.DELETE("/posts/:post_id")
		social.PATCH("/posts/:post_id/view")
		social.POST("/posts/:post_id/report")
		social.POST("/posts/:post_id/like")
		social.DELETE("/posts/:post_id/like")

		social.GET("/posts/:post_id/comments")
		social.POST("/posts/:post_id/comments")
		social.PUT("/posts/:post_id/comments/:commentID")
		social.DELETE("/posts/:post_id/comments/:commentID")
	}

	return s.server.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
