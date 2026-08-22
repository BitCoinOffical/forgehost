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

	subscr, err := h.srvc.GetSubscriptions(c.Request.Context(), userId)
	if err != nil {
		response.InternalServerError(c, err, "failed get GetSubscriptions", h.logger)
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *ProfileHandler) Subscribe(c *gin.Context) {

}

func (h *ProfileHandler) Unsubscribe(c *gin.Context) {

}

func (h *ProfileHandler) Report(c *gin.Context) {

}
