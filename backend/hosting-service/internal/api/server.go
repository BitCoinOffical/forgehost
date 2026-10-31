package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BitCoinOffical/forgehost/hosting-service/config"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/api/http/handlers"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	timeoutSecond = 5
)

type Server struct {
	engine *gin.Engine
	server *http.Server
	m      *middleware.Middleware
	h      *handlers.Handlers
}

func NewServer(cfg *config.AppConfig, m *middleware.Middleware, h *handlers.Handlers) *Server {
	engine := gin.New()
	return &Server{
		engine: engine,
		m:      m,
		h:      h,
		server: &http.Server{
			Addr:        ":" + cfg.Port,
			Handler:     engine,
			ReadTimeout: timeoutSecond * time.Second,
		},
	}
}

func (s *Server) Run() error {
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	api := s.engine.Group("/api/v1")
	worlds := api.Group("/world")
	worlds.Use(s.m.AuthMiddleware())
	worlds.Use(s.m.RateLimiter())
	{
		worlds.POST("/minecraft", s.h.Worlds.Create)
		worlds.GET("/minecraft", s.h.Worlds.GetAll)
		worlds.GET("/minecraft/:world_id", s.h.Worlds.GetByID)
		worlds.DELETE("/minecraft/:world_id", s.h.Worlds.Delete)
	}

	return s.server.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
