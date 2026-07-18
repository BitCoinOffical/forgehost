package api

import (
	"BitCoinOffical/forgehost/auth-service/config"
	"BitCoinOffical/forgehost/auth-service/internal/api/handlers"
	"BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	timeoutSecond = 5
)

type Server struct {
	engine *gin.Engine
	m      *middleware.Middleware
	server *http.Server
	h      *handlers.Handlers
}

func NewServer(cfg *config.AppConfig, m *middleware.Middleware, h *handlers.Handlers) *Server {
	engine := gin.New()
	return &Server{
		m:      m,
		h:      h,
		engine: engine,
		server: &http.Server{
			Addr:        cfg.Port,
			Handler:     engine,
			ReadTimeout: timeoutSecond * time.Second,
		},
	}
}

func (s *Server) Run() error {
	auth := s.engine.Group("/auth")
	auth.Use(s.m.RateLimiter())
	{
		auth.POST("/register", s.h.Auth.Register)
		auth.POST("/login", s.h.Auth.Login)
		auth.POST("/logout", s.m.AuthMiddleware(), s.h.Auth.Logout)
		auth.POST("/refresh", s.h.Auth.UpdateAccessToken)

		auth.POST("/login/google", s.h.Auth.GoogleLogin) //android

		auth.GET("/login/google", s.h.Auth.GoogleLogin)             //web
		auth.GET("/login/google/callback", s.h.Auth.GoogleCallback) //web

		auth.POST("/verify-email", s.m.AuthMiddleware())
		auth.POST("/verify-email/resend", s.m.AuthMiddleware())

		auth.PATCH("/password/update", s.m.AuthMiddleware())
		auth.POST("/password/reset")
		auth.POST("/password/reset/confirm")
	}

	return s.server.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
