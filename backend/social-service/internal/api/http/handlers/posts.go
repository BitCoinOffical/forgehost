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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var postsRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "post_requests_total",
	Help: "Total number of auth requests",
})

type PostHandler struct {
	srvc   *services.PostsService
	logger *zap.Logger
}

func NewPostHandler(srvc *services.PostsService, logger *zap.Logger) *PostHandler {
	return &PostHandler{srvc: srvc, logger: logger}
}

// GetSubPosts godoc
// @Summary      Лента постов по подпискам
// @Description  Возвращает посты авторов, на которых подписан текущий пользователь
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   dto.PostDTO
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts [get]
func (h *PostHandler) GetSubPosts(c *gin.Context) {
	postsRequestsTotal.Inc()
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

// GetGlobalPosts godoc
// @Summary      Глобальная лента постов
// @Description  Возвращает глобальную ленту постов с пагинацией по курсору
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cursor  path      string  false  "Курсор пагинации"
// @Success      200     {array}   dto.PostDTO
// @Failure      401     {object}  response.ErrorBody
// @Failure      500     {object}  response.ErrorBody
// @Router       /social/posts/global/{cursor} [get]
func (h *PostHandler) GetGlobalPosts(c *gin.Context) {
	postsRequestsTotal.Inc()
	cursor := c.Param("cursor")
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

// GetByID godoc
// @Summary      Получить пост по ID
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post_id  path      string  true  "ID поста"
// @Success      200      {object}  dto.PostDTO
// @Failure      404      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/posts/{post_id} [get]
func (h *PostHandler) GetByID(c *gin.Context) {
	postsRequestsTotal.Inc()
	postId := c.Param("post_id")

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

// CreatePost godoc
// @Summary      Создать пост
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  dto.CreatePostDTO  true  "Данные поста"
// @Success      201
// @Failure      400  {object}  response.ErrorBody
// @Failure      401  {object}  response.ErrorBody
// @Failure      409  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	postsRequestsTotal.Inc()
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

// GetTopics godoc
// @Summary      Получить список топиков
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      200  {array}   string
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/topics [get]
func (h *PostHandler) GetTopics(c *gin.Context) {
	postsRequestsTotal.Inc()
	res, err := h.srvc.GetTopics(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err, "failed get topics", h.logger)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Update godoc
// @Summary      Обновить пост
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id  path      string             true  "ID поста"
// @Param        request  body      dto.UpdatePostDTO  true  "Данные для обновления"
// @Success      200      {object}  dto.PostDTO
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/posts/{post_id} [patch]
func (h *PostHandler) Update(c *gin.Context) {
	postsRequestsTotal.Inc()
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

// DeletePost godoc
// @Summary      Удалить пост
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id  path  string  true  "ID поста"
// @Success      204
// @Failure      401  {object}  response.ErrorBody
// @Failure      404  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	postsRequestsTotal.Inc()
	postId := c.Param("post_id")

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

// ViewPost godoc
// @Summary      Зафиксировать просмотр поста
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post_id  path  string  true  "ID поста"
// @Success      200
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/view [patch]
func (h *PostHandler) ViewPost(c *gin.Context) {
	postsRequestsTotal.Inc()
	postId := c.Param("post_id")
	if err := h.srvc.ViewPost(c.Request.Context(), postId); err != nil {
		response.InternalServerError(c, err, "failed view post", h.logger)
		return
	}
	c.Status(http.StatusOK)
}

// PostReport godoc
// @Summary      Пожаловаться на пост
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id  path  string         true  "ID поста"
// @Param        request  body  dto.ReportDTO  true  "Причина жалобы"
// @Success      201
// @Failure      400  {object}  response.ErrorBody
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/report [post]
func (h *PostHandler) PostReport(c *gin.Context) {
	postsRequestsTotal.Inc()
	postId := c.Param("post_id")
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

// Like godoc
// @Summary      Лайкнуть пост
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id  path  string  true  "ID поста"
// @Success      200
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/like [post]
func (h *PostHandler) Like(c *gin.Context) {
	postsRequestsTotal.Inc()
	postId := c.Param("post_id")
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

// Unlike godoc
// @Summary      Убрать лайк с поста
// @Tags         posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        post_id  path  string  true  "ID поста"
// @Success      204
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/posts/{post_id}/like [delete]
func (h *PostHandler) Unlike(c *gin.Context) {
	postsRequestsTotal.Inc()
	postId := c.Param("post_id")
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
