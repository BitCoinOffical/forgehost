package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/config"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/ws"
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
	ws     *ws.WebSockerHandlers
	h      *handlers.Handlers
	m      *middleware.Middleware
}

func NewServer(cfg *config.AppConfig, m *middleware.Middleware, h *handlers.Handlers, ws *ws.WebSockerHandlers) *Server {
	engine := gin.New()
	return &Server{
		h:      h,
		m:      m,
		ws:     ws,
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
		social.POST("/profile/:user_id/block", s.h.Profile.Block)

		social.GET("/posts", s.h.Posts.GetSubPosts)
		social.GET("/posts/global", s.h.Posts.GetGlobalPosts)
		social.GET("/posts/global/:cursor", s.h.Posts.GetGlobalPosts)
		social.GET("/posts/:post_id", s.h.Posts.GetByID)
		social.POST("/posts", s.h.Posts.CreatePost)
		social.PATCH("/posts/:post_id", s.h.Posts.Update)
		social.DELETE("/posts/:post_id", s.h.Posts.DeletePost)
		social.PATCH("/posts/:post_id/view", s.h.Posts.ViewPost)
		social.POST("/posts/:post_id/report", s.h.Posts.PostReport)
		social.POST("/posts/:post_id/like", s.h.Posts.Like)
		social.DELETE("/posts/:post_id/like", s.h.Posts.Unlike)

		social.GET("/posts/:post_id/comments", s.h.Comments.List)
		social.POST("/posts/:post_id/comments", s.h.Comments.Create)
		social.PUT("/posts/:post_id/comments/:comment_id", s.h.Comments.Update)
		social.POST("/posts/:post_id/comments/:comment_id/report", s.h.Comments.Report)
		social.DELETE("/posts/:post_id/comments/:comment_id", s.h.Comments.Delete)
		social.POST("/posts/:post_id/comments/:comment_id/like", s.h.Comments.Like)
		social.DELETE("/posts/:post_id/comments/:comment_id/like", s.h.Comments.Unlike)

		social.GET("/topics", s.h.Posts.GetTopics)
		social.GET("/topics/:topic_id", s.h.Posts.GetTopicByID)
		social.POST("/topics", s.h.Posts.CreateTopic)
		social.DELETE("/topics/:topic_id", s.h.Posts.DeleteTopic)
		social.POST("/topics/report", s.h.Posts.ReportTopic)

		social.POST("/chats", s.h.Chats.CreateChat)
		social.GET("/chats/:chat_id", s.h.Chats.GetChatByID)
		social.DELETE("/chats/:chat_id", s.h.Chats.DeleteChat)
		social.PATCH("/chats/:chat_id", s.h.Chats.UpdateChat)

		social.POST("/chats/:chat_id/leave", s.h.Chats.LeaveChat)

		social.GET("/chats/:chat_id/members", s.h.Chats.GetUsersFromChat)
		social.POST("/chats/:chat_id/members", s.h.Chats.JoinInChat)
		social.DELETE("/chats/:chat_id/members/:user_id", s.h.Chats.KickUserFromChat)
		social.PATCH("/chats/:chat_id/members/:user_id/role", s.h.Chats.SetUserRoleChat)

		social.POST("/chats/:chat_id/bans/:user_id", s.h.Chats.BanUser)
		social.DELETE("/chats/:chat_id/bans/:user_id", s.h.Chats.UnbanUser)
		social.GET("/chats/:chat_id/bans/:user_id", s.h.Chats.CheckBanUser)

		social.GET("/message/chat/:chat_id", s.ws.Msgs.GetChatHistory)
		social.GET("/message/stream/:stream_id", s.ws.Msgs.GetStreamHistory)

		social.GET("/ws/message/chat/:chat_id", s.ws.Msgs.SendMessage)
		social.GET("/ws/message/stream/:stream_id", s.ws.Msgs.SaveStreamMessage)

		social.GET("/search")
	}

	return s.server.ListenAndServe()
}

func (s *Server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("s.server.Shutdown: %w", err)
	}
	return nil
}
