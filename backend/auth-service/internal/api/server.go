package api

import (
	"github.com/BitCoinOffical/forgehost/auth-service/config"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/handlers"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/middleware"
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
			Addr:        ":" + cfg.Port,
			Handler:     engine,
			ReadTimeout: timeoutSecond * time.Second,
		},
	}
}

func (s *Server) Run() error {
	api := s.engine.Group("/api/v1")
	auth := api.Group("/auth")
	auth.Use(s.m.RateLimiter())
	{
		auth.POST("/register", s.h.Auth.Register)
		auth.POST("/login", s.h.Auth.Login)
		auth.POST("/logout", s.m.AuthMiddleware(), s.h.Auth.Logout)
		auth.POST("/refresh", s.h.Auth.UpdateAccessToken)

		auth.POST("/login/google", s.h.Auth.GoogleLoginAndroid) //android

		auth.GET("/login/google", s.h.Auth.GoogleLogin)             //web
		auth.GET("/login/google/callback", s.h.Auth.GoogleCallback) //web

		auth.POST("/verify-email", s.h.Auth.VerifyEmail)
		auth.POST("/verify-email/resend", s.h.Auth.ResendVerifyEmail)

		auth.PATCH("/password/update", s.m.AuthMiddleware(), s.h.Auth.UpdatePassword)
		auth.POST("/password/reset", s.h.Auth.PasswordReset)
		auth.POST("/password/reset/confirm", s.h.Auth.ConfirmPasswordReset)
		auth.POST("/password/reset/resend", s.h.Auth.PasswordResetResend)
	}

	return s.server.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
