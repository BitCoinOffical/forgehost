package handlers

import (
	"BitCoinOffical/forgehost/auth-service/internal/api/response"
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/services"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	logger  *zap.Logger
	service *services.AuthService
}

func NewAuthHandler(logger *zap.Logger, service *services.AuthService) *AuthHandler {
	return &AuthHandler{logger: logger, service: service}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.UsersLoginDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "failed to login", h.logger)
		return
	}

	tokens, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidCredentials) {
			response.Unauthorized(c, err, "invalid credentials", h.logger)
			return
		}
		response.InternalServerError(c, "failed to login", err, h.logger)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {

}

func (h *AuthHandler) Register(c *gin.Context) {

}
