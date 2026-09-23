package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BitCoinOffical/forgehost/auth-service/config"
	_ "github.com/BitCoinOffical/forgehost/auth-service/docs"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/handlers"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

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
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s.engine.GET("/metrics", s.m.RequireRole(), gin.WrapH(promhttp.Handler()))
	api := s.engine.Group("/api/v1")
	auth := api.Group("/auth")
	auth.Use(s.m.RateLimiter())
	{

		auth.POST("/register", s.h.Reg.Register)
		auth.POST("/login", s.h.Login.Login)
		auth.POST("/logout", s.m.AuthMiddleware(), s.h.Logout.Logout)
		auth.POST("/refresh", s.h.Token.UpdateAccessToken)

		auth.POST("/login/google", s.h.Oauth.GoogleLoginAndroid) //android

		auth.GET("/login/google", s.h.Oauth.GoogleLogin)             //web
		auth.GET("/login/google/callback", s.h.Oauth.GoogleCallback) //web

		auth.POST("/verify-email", s.h.Reg.VerifyEmail)
		auth.POST("/verify-email/resend", s.h.Reg.ResendVerifyEmail)

		auth.PATCH("/password/update", s.m.AuthMiddleware(), s.h.Pass.UpdatePassword)
		auth.POST("/password/reset", s.h.Pass.PasswordReset)
		auth.POST("/password/reset/confirm", s.h.Pass.ConfirmPasswordReset)
		auth.POST("/password/reset/resend", s.h.Pass.PasswordResetResend)

		auth.POST("/exchange", s.h.Oauth.Exchange)
	}

	return s.server.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
