package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type WorldsHandler struct {
	srvc   WorldsService
	logger *zap.Logger
}

func NewWorldsHandler(srvc WorldsService, logger *zap.Logger) *WorldsHandler {
	return &WorldsHandler{
		srvc:   srvc,
		logger: logger,
	}
}

var worldsRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "worlds_requests_total",
	Help: "Total number of minecraft world requests",
})

// Create godoc
// @Summary      Создать или обновить Minecraft-сервер
// @Tags         minecraft
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.UpsertMinecraftServerDTO  true  "Данные Minecraft-сервера"
// @Success      200      {object}  models.MinecraftServer
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      409      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /minecraft [post]
func (h *WorldsHandler) Create(c *gin.Context) {
	worldsRequestsTotal.Inc()

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	var req dto.UpsertMinecraftServerDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	world, err := h.srvc.Upsert(c.Request.Context(), &req, userID)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Conflict(c, err, "minecraft server already exists", h.logger)
			return
		}

		response.InternalServerError(c, err, "failed upsert minecraft server", h.logger)
		return
	}

	c.JSON(http.StatusOK, world)
}

// GetAll godoc
// @Summary      Получить список доступных Minecraft-серверов
// @Tags         minecraft
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.MinecraftServer
// @Failure      500  {object}  response.ErrorBody
// @Router       /minecraft [get]
func (h *WorldsHandler) GetAll(c *gin.Context) {
	worldsRequestsTotal.Inc()

	worlds, err := h.srvc.GetAll(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err, "failed get minecraft servers", h.logger)
		return
	}

	c.JSON(http.StatusOK, worlds)
}

// GetByID godoc
// @Summary      Получить Minecraft-сервер по ID
// @Tags         minecraft
// @Accept       json
// @Produce      json
// @Param        world_id  path      string  true  "ID Minecraft-сервера"
// @Success      200       {object}  models.MinecraftServer
// @Failure      404       {object}  response.ErrorBody
// @Failure      500       {object}  response.ErrorBody
// @Router       /minecraft/{world_id} [get]
func (h *WorldsHandler) GetByID(c *gin.Context) {
	worldsRequestsTotal.Inc()

	worldID := c.Param("world_id")

	world, err := h.srvc.GetByID(c.Request.Context(), worldID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "minecraft server not found", h.logger)
			return
		}

		response.InternalServerError(c, err, "failed get minecraft server", h.logger)
		return
	}

	c.JSON(http.StatusOK, world)
}

// Delete godoc
// @Summary      Удалить Minecraft-сервер
// @Tags         minecraft
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        world_id  path  string  true  "ID Minecraft-сервера"
// @Success      204
// @Failure      401  {object}  response.ErrorBody
// @Failure      404  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /minecraft/{world_id} [delete]
func (h *WorldsHandler) Delete(c *gin.Context) {
	worldsRequestsTotal.Inc()

	worldID := c.Param("world_id")

	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.Delete(c.Request.Context(), worldID, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "minecraft server not found", h.logger)
			return
		}

		response.InternalServerError(c, err, "failed delete minecraft server", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}
