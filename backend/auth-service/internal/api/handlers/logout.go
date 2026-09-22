package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/gin-gonic/gin"
)

// Logout godoc
// @Summary      Log out the current user
// @Description  Invalidates the authenticated user's session/tokens.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      204  "logged out"
// @Failure      400  {object}  map[string]string  "incorrect type value"
// @Failure      401  {object}  map[string]string  "missing or invalid token"
// @Failure      500  {object}  map[string]string
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	id, err := middleware.GetUserID(c)
	if err != nil {
		if errors.Is(err, domain.ErrValueNotFound) {
			response.Unauthorized(c, err, "not found value by key", h.logger)
			return
		}
		response.BadRequest(c, err, "incorrect type value", h.logger)
		return
	}

	if err := h.authsrvc.LogoutUser(c.Request.Context(), id); err != nil {
		response.InternalServerError(c, err, "user failed to logout", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}
