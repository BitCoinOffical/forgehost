package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var profRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "prof_requests_total",
	Help: "Total number of auth requests",
})

type ProfileHandler struct {
	srvc   ProfileService
	logger *zap.Logger
}

func NewProfileHandler(srvc ProfileService, logger *zap.Logger) *ProfileHandler {
	return &ProfileHandler{srvc: srvc, logger: logger}
}

// Me godoc
// @Summary      Получить свой профиль
// @Description  Возвращает профиль текущего авторизованного пользователя (id берётся из токена)
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.ProfileResponse
// @Failure      401  {object}  response.ErrorBody
// @Failure      404  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/profile/me [get]
func (h *ProfileHandler) Me(c *gin.Context) {
	profRequestsTotal.Inc()
	id, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	resp, err := h.srvc.GetProfileByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "profile not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get profile", h.logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProfileByID godoc
// @Summary      Получить профиль пользователя по ID
// @Description  Возвращает публичный профиль пользователя по его идентификатору
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        user_id  path      string  true  "ID пользователя"
// @Success      200      {object}  models.ProfileResponse
// @Failure      404      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/profile/{user_id} [get]
func (h *ProfileHandler) GetProfileByID(c *gin.Context) {
	profRequestsTotal.Inc()
	userId := c.Param("user_id")

	profile, err := h.srvc.GetProfileByID(c.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "profile not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get profile", h.logger)
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary      Обновить профиль
// @Description  Обновляет данные профиля текущего пользователя
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.UpdateProfileDTO  true  "Данные для обновления"
// @Success      200      {object}  models.Profile
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      404      {object}  response.ErrorBody
// @Failure      409      {object}  response.ErrorBody  "имя пользователя уже занято"
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/profile [patch]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	profRequestsTotal.Inc()
	id, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	var req dto.UpdateProfileDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	prof, err := h.srvc.UpdateProfile(c.Request.Context(), &req, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "profile not found", h.logger)
			return
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Conflict(c, err, "username already taken", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed update profile", h.logger)
		return
	}

	c.JSON(http.StatusOK, prof)
}

// GetSubscribers godoc
// @Summary      Получить подписчиков пользователя
// @Description  Возвращает список подписчиков указанного пользователя
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        user_id  path      string  true  "ID пользователя"
// @Success      200      {array}   models.Subscribes
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/profile/{user_id}/subscribers [get]
func (h *ProfileHandler) GetSubscribers(c *gin.Context) {
	profRequestsTotal.Inc()
	userId := c.Param("user_id")

	subscr, err := h.srvc.GetSubscribers(c.Request.Context(), userId)
	if err != nil {
		response.InternalServerError(c, err, "failed get subscribers", h.logger)
		return
	}

	c.JSON(http.StatusOK, subscr)
}

// GetSubscriptions godoc
// @Summary      Получить подписки пользователя
// @Description  Возвращает список пользователей, на которых подписан указанный пользователь
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        user_id  path      string  true  "ID пользователя"
// @Success      200      {array}   models.Subscriptions
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/profile/{user_id}/subscriptions [get]
func (h *ProfileHandler) GetSubscriptions(c *gin.Context) {
	profRequestsTotal.Inc()
	userId := c.Param("user_id")

	subs, err := h.srvc.GetSubscriptions(c.Request.Context(), userId)
	if err != nil {
		response.InternalServerError(c, err, "failed get subscriptions", h.logger)
		return
	}

	c.JSON(http.StatusOK, subs)
}

// Subscribe godoc
// @Summary      Подписаться на пользователя
// @Description  Оформляет подписку текущего пользователя на указанного пользователя
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path  string  true  "ID пользователя, на которого подписываемся"
// @Success      201
// @Failure      400  {object}  response.ErrorBody  "нельзя подписаться на себя"
// @Failure      401  {object}  response.ErrorBody
// @Failure      409  {object}  response.ErrorBody  "подписка уже существует"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/profile/{user_id}/subscribe [post]
func (h *ProfileHandler) Subscribe(c *gin.Context) {
	profRequestsTotal.Inc()
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	targetId := c.Param("user_id")

	if err := h.srvc.Subscribe(c.Request.Context(), userId, targetId); err != nil {
		if errors.Is(err, domain.ErrSubscribeToSelf) {
			response.BadRequest(c, err, "cannot subscribe to yourself", h.logger)
			return
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Conflict(c, err, "already subscribed", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed subscribe", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

// Unsubscribe godoc
// @Summary      Отписаться от пользователя
// @Description  Отменяет подписку текущего пользователя на указанного пользователя
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path  string  true  "ID пользователя, от которого отписываемся"
// @Success      204
// @Failure      401  {object}  response.ErrorBody
// @Failure      404  {object}  response.ErrorBody  "подписки нет"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/profile/{user_id}/unsubscribe [delete]
func (h *ProfileHandler) Unsubscribe(c *gin.Context) {
	profRequestsTotal.Inc()
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	targetId := c.Param("user_id")

	if err := h.srvc.UnSubscribe(c.Request.Context(), userId, targetId); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "subscription not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed unsubscribe", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

// Report godoc
// @Summary      Пожаловаться на пользователя
// @Description  Создаёт жалобу на указанного пользователя
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path      string         true  "ID пользователя, на которого жалуемся"
// @Param        request  body      dto.ReportDTO  true  "Причина жалобы"
// @Success      201
// @Failure      400  {object}  response.ErrorBody  "неверное тело или жалоба на себя"
// @Failure      401  {object}  response.ErrorBody
// @Failure      409  {object}  response.ErrorBody  "жалоба уже существует"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/profile/{user_id}/report [post]
func (h *ProfileHandler) Report(c *gin.Context) {
	profRequestsTotal.Inc()
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	targetId := c.Param("user_id")

	var req dto.ReportDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if err := h.srvc.Report(c.Request.Context(), userId, targetId, &req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "user not found", h.logger)
			return
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Conflict(c, err, "report already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed create profile report", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

// Block godoc
// @Summary      Заблокировать пользователя
// @Description  Блокирует указанного пользователя для текущего пользователя
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path  string  true  "ID пользователя, которого блокируем"
// @Success      200
// @Failure      401  {object}  response.ErrorBody
// @Failure      409  {object}  response.ErrorBody  "пользователь уже заблокирован"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/profile/{user_id}/block [post]
func (h *ProfileHandler) Block(c *gin.Context) {
	profRequestsTotal.Inc()
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	targetID := c.Param("user_id")

	if err := h.srvc.Block(c.Request.Context(), userID, targetID); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Conflict(c, err, "user already blocked", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed block user", h.logger)
		return
	}

	c.Status(http.StatusOK)
}
