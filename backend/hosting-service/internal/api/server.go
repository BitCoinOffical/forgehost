package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
}

func NewServer() *Server {
	engine := gin.New()
	return &Server{
		engine: engine,
		server: &http.Server{
			Addr:        "",
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
	{
		worlds.GET("/minecraft")
		worlds.GET("/minecraft/:world_id")
		worlds.POST("/minecraft")
		worlds.DELETE("/minecraft/:world_id")
		worlds.PUT("/minecraft/:world_id/heartbeat")

		worlds.GET("/search")
	}

	return s.server.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
