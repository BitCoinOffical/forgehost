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
	var req dto.PostDTO

	res, err := h.srvc.GetPostById(c.Request.Context(), req.PostId)
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

func (h *PostHandler) Create(c *gin.Context) {

}

func (h *PostHandler) Update(c *gin.Context) {

}

func (h *PostHandler) Delete(c *gin.Context) {

}

func (h *PostHandler) Report(c *gin.Context) {

}

func (h *PostHandler) Like(c *gin.Context) {

}

func (h *PostHandler) Unlike(c *gin.Context) {

}
