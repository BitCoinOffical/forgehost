package handlers

import (
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CommentHandler struct {
	srvc   *services.CommentsService
	logger *zap.Logger
}

func NewCommentHandler(srvc *services.CommentsService, logger *zap.Logger) *CommentHandler {
	return &CommentHandler{srvc: srvc, logger: logger}
}
func (h *CommentHandler) List(c *gin.Context) {
	postId := c.Query("post_id")

	res, err := h.srvc.ListComments(c.Request.Context(), postId)
	if err != nil {
		response.InternalServerError(c, err, "failed get lists comments", h.logger)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *CommentHandler) Create(c *gin.Context) {
	var req dto.CreateCommentDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	postId := c.Query("post_id")
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.CreateComment(c.Request.Context(), postId, userId, &req); err != nil {
		response.InternalServerError(c, err, "failed create comments", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *CommentHandler) Update(c *gin.Context) {
	var req *dto.UpdateCommentDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	postId := c.Query("post_id")
	commentId := c.Query("comment_id")
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	res, err := h.srvc.UpdateComment(c.Request.Context(), postId, userId, commentId, req)
	if err != nil {
		response.InternalServerError(c, err, "failed create comments", h.logger)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	postId := c.Query("post_id")
	commentId := c.Query("comment_id")
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.DeleteComment(c.Request.Context(), postId, commentId, userId); err != nil {
		response.InternalServerError(c, err, "failed delete comments", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CommentHandler) Report(c *gin.Context) {
	var req dto.ReportCommentDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}
	commentId := c.Query("comment_id")
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.ReportComment(c.Request.Context(), userId, commentId, &req); err != nil {
		response.InternalServerError(c, err, "failed create report comments", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *CommentHandler) Like(c *gin.Context) {
	commentId := c.Query("comment_id")
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.LikeComment(c.Request.Context(), userId, commentId); err != nil {
		response.InternalServerError(c, err, "failed create report comments", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *CommentHandler) Unlike(c *gin.Context) {
	commentId := c.Query("comment_id")
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.UnlikeComment(c.Request.Context(), userId, commentId); err != nil {
		response.InternalServerError(c, err, "failed create report comments", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}
