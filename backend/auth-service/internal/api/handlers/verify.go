package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	tokens, err := h.authsrvc.VerifyEmail(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			response.Conflict(c, err, "email already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed verificate email", h.logger)
		return
	}

	c.JSON(http.StatusOK, tokens)
}
func (h *AuthHandler) ResendVerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if err := h.authsrvc.ResendVerifyEmail(c.Request.Context(), &req); err != nil {
		if errors.Is(err, domain.ErrToManyRequest) {
			response.ManyRequest(c, err, "too many attempts", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed update verify email", h.logger)
		return
	}

	c.Status(http.StatusOK)
}
