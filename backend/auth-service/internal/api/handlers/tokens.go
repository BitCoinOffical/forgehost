package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
)

// UpdateAccessToken godoc
// @Summary      Refresh access token
// @Description  Exchanges a valid refresh token for a new access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UpdateTokensDTO  true  "Refresh token payload"
// @Success      200      {object}  dto.TokensDTO
// @Failure      400      {object}  map[string]string  "invalid body"
// @Failure      401      {object}  map[string]string  "token not found or expired"
// @Failure      500      {object}  map[string]string
// @Router       /auth/refresh [post]
func (h *AuthHandler) UpdateAccessToken(c *gin.Context) {
	authRequestsTotal.Inc()
	var req dto.UpdateTokensDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	tokens, err := h.authsrvc.UpdateAccessToken(c.Request.Context(), req.RefreshToken)
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
