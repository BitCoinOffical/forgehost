package api

import (
	"BitCoinOffical/forgehost/auth-service/config"
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
}

func NewServer(cfg *config.AppConfig, m *middleware.Middleware) *Server {
	engine := gin.New()
	return &Server{
		engine: engine,
		server: &http.Server{
			Addr:        cfg.Port,
			Handler:     engine,
			ReadTimeout: timeoutSecond * time.Second,
		},
	}
}

func (s *Server) Run() {
	auth := s.engine.Group("/auth")
	{
		auth.POST("/register")
		auth.POST("/login")
		auth.POST("/logout")
		auth.POST("/refresh")

		auth.GET("/login/google")
		auth.GET("/login/google/callback")

		auth.POST("/verify-email")
		auth.POST("/verify-email/resend")

		auth.PATCH("/password/update")
		auth.POST("/password/reset")
		auth.POST("/password/reset/confirm")
	}
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
