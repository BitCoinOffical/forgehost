package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PostHandler struct {
	srvc   *services.PostsService
	logger *zap.Logger
}

func NewPostHandler(srvc *services.PostsService, logger *zap.Logger) *PostHandler {
	return &PostHandler{srvc: srvc, logger: logger}
}

func (h *PostHandler) GetSubPosts(c *gin.Context) {
	id, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	posts, err := h.srvc.GetSubPosts(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, err, "failed get subscription posts", h.logger)
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) GetGlobalPosts(c *gin.Context) {
	cursor := c.Query("cursor")
	id, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	posts, err := h.srvc.GetGlobalPosts(c.Request.Context(), id, cursor)
	if err != nil {
		response.InternalServerError(c, err, "failed get global posts", h.logger)
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (h *PostHandler) GetByID(c *gin.Context) {
	postId := c.Query("post_id")

	res, err := h.srvc.GetPostById(c.Request.Context(), postId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "post not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get global posts", h.logger)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	id, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	var req dto.CreatePostDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if err := h.srvc.CreatePost(c.Request.Context(), &req, id); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Conflict(c, err, "such a post already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed create post", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *PostHandler) GetTopics(c *gin.Context) {
	res, err := h.srvc.GetTopics(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err, "failed get topics", h.logger)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *PostHandler) Update(c *gin.Context) {
	id, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	var req dto.UpdatePostDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	res, err := h.srvc.UpdatePost(c.Request.Context(), &req, id)
	if err != nil {
		response.InternalServerError(c, err, "failed update post by id", h.logger)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	postId := c.Query("post_id")

	id, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.DeletePost(c.Request.Context(), postId, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "post not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed delete post", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PostHandler) ViewPost(c *gin.Context) {
	postId := c.Query("post_id")
	if err := h.srvc.ViewPost(c.Request.Context(), postId); err != nil {
		response.InternalServerError(c, err, "failed view post", h.logger)
		return
	}
	c.Status(http.StatusOK)
}

func (h *PostHandler) PostReport(c *gin.Context) {
	postId := c.Query("post_id")
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	var req dto.ReportDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if err := h.srvc.ReportPost(c.Request.Context(), &req, userId, postId); err != nil {
		response.InternalServerError(c, err, "failed create report post", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *PostHandler) Like(c *gin.Context) {
	postId := c.Query("post_id")
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.LikePost(c.Request.Context(), userId, postId); err != nil {
		response.InternalServerError(c, err, "failed create report post", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

func (h *PostHandler) Unlike(c *gin.Context) {
	postId := c.Query("post_id")
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.UnlikePost(c.Request.Context(), userId, postId); err != nil {
		response.InternalServerError(c, err, "failed create report post", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}
