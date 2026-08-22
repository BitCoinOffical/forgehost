package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) UpdateAccessToken(c *gin.Context) {
	var req dto.TokensDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	tokens, err := h.authsrvc.UpdateAccessToken(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Unauthorized(c, err, "not found token", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed update token", h.logger)
		return
	}

	c.JSON(http.StatusOK, tokens)
}
