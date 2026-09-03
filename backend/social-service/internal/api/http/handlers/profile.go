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

type ProfileHandler struct {
	srvc   *services.ProfileService
	logger *zap.Logger
}

func NewProfileHandler(srvc *services.ProfileService, logger *zap.Logger) *ProfileHandler {
	return &ProfileHandler{srvc: srvc, logger: logger}
}

func (h *ProfileHandler) Me(c *gin.Context) {
	id, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	profile, err := h.srvc.GetProfileByID(c.Request.Context(), id)
	if err != nil {
		response.InternalServerError(c, err, "failed get profile", h.logger)
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *ProfileHandler) GetProfileByID(c *gin.Context) {
	userId := c.Param("user_id")

	profile, err := h.srvc.GetProfileByID(c.Request.Context(), userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "profile not found", h.logger)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	prof, err := h.srvc.UpdateProfile(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, err, "failed update profile", h.logger)
		return
	}

	c.JSON(http.StatusOK, prof)
}

func (h *ProfileHandler) GetSubscribers(c *gin.Context) {
	userId := c.Param("user_id")

	subscr, err := h.srvc.GetSubscribers(c.Request.Context(), userId)
	if err != nil {
		response.InternalServerError(c, err, "failed get subscribers", h.logger)
		return
	}

	c.JSON(http.StatusOK, subscr)
}

func (h *ProfileHandler) GetSubscriptions(c *gin.Context) {
	userId := c.Param("user_id")

	subs, err := h.srvc.GetSubscriptions(c.Request.Context(), userId)
	if err != nil {
		response.InternalServerError(c, err, "failed get GetSubscriptions", h.logger)
		return
	}

	c.JSON(http.StatusOK, subs)
}

func (h *ProfileHandler) Subscribe(c *gin.Context) {
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	targetId := c.Param("user_id")

	if err := h.srvc.Subscribe(c.Request.Context(), userId, targetId); err != nil {
		response.InternalServerError(c, err, "failed subscribe", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *ProfileHandler) Unsubscribe(c *gin.Context) {
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	targetId := c.Param("user_id")

	if err := h.srvc.UnSubscribe(c.Request.Context(), userId, targetId); err != nil {
		response.InternalServerError(c, err, "failed unsubscribe", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProfileHandler) Report(c *gin.Context) {
	userId, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	targetId := c.Param("user_id")

	var req dto.ProfileReportDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if err := h.srvc.Report(c.Request.Context(), userId, targetId, req); err != nil {
		response.InternalServerError(c, err, "failed create profile report", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}
