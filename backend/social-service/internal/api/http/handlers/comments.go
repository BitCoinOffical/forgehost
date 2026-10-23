package handlers

import (
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var comRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "com_requests_total",
	Help: "Total number of auth requests",
})

type CommentHandler struct {
	srvc   CommentsService
	logger *zap.Logger
}

func NewCommentHandler(srvc CommentsService, logger *zap.Logger) *CommentHandler {
	return &CommentHandler{srvc: srvc, logger: logger}
}

// List godoc
// @Summary      Получить комментарии к посту
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        post_id  path  string  true  "ID поста"
// @Success      200      {array}   object
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/comments [get]
func (h *CommentHandler) List(c *gin.Context) {
	comRequestsTotal.Inc()
	postId := c.Param("post_id")

	res, err := h.srvc.ListComments(c.Request.Context(), postId)
	if err != nil {
		response.InternalServerError(c, err, "failed get lists comments", h.logger)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Create godoc
// @Summary      Создать комментарий
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id  path  string                 true  "ID поста"
// @Param        request  body  dto.CreateCommentDTO   true  "Данные комментария"
// @Success      201
// @Failure      400  {object}  response.ErrorBody
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	comRequestsTotal.Inc()
	var req dto.CreateCommentDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	postId := c.Param("post_id")
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

// Update godoc
// @Summary      Обновить комментарий
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id     path  string                 true  "ID поста"
// @Param        comment_id  path  string                 true  "ID комментария"
// @Param        request     body  dto.UpdateCommentDTO   true  "Новый текст комментария"
// @Success      200      {object}  object
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/comments/{comment_id} [put]
func (h *CommentHandler) Update(c *gin.Context) {
	comRequestsTotal.Inc()
	var req *dto.UpdateCommentDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	postId := c.Param("post_id")
	commentId := c.Param("comment_id")
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

// Delete godoc
// @Summary      Удалить комментарий
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id     path  string  true  "ID поста"
// @Param        comment_id  path  string  true  "ID комментария"
// @Success      204
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/comments/{comment_id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	comRequestsTotal.Inc()
	postId := c.Param("post_id")
	commentId := c.Param("comment_id")
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

// Report godoc
// @Summary      Пожаловаться на комментарий
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id     path  string                 true  "ID поста"
// @Param        comment_id  path  string                 true  "ID комментария"
// @Param        request     body  dto.ReportCommentDTO   true  "Причина жалобы"
// @Success      201
// @Failure      400  {object}  response.ErrorBody
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/comments/{comment_id}/report [post]
func (h *CommentHandler) Report(c *gin.Context) {
	comRequestsTotal.Inc()
	var req dto.ReportCommentDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}
	commentId := c.Param("comment_id")
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

// Like godoc
// @Summary      Лайкнуть комментарий
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id     path  string  true  "ID поста"
// @Param        comment_id  path  string  true  "ID комментария"
// @Success      201
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/comments/{comment_id}/like [post]
func (h *CommentHandler) Like(c *gin.Context) {
	comRequestsTotal.Inc()
	commentId := c.Param("comment_id")
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

// Unlike godoc
// @Summary      Убрать лайк с комментария
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id     path  string  true  "ID поста"
// @Param        comment_id  path  string  true  "ID комментария"
// @Success      204
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/comments/{comment_id}/like [delete]
func (h *CommentHandler) Unlike(c *gin.Context) {
	comRequestsTotal.Inc()
	commentId := c.Param("comment_id")
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
