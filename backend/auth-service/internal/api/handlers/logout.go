package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LogoutHandler struct {
	srvc   LogoutService
	logger *zap.Logger
}

func NewLogoutHandler(srvc LogoutService, logger *zap.Logger) *LogoutHandler {
	return &LogoutHandler{
		srvc:   srvc,
		logger: logger,
	}
}

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
func (h *LogoutHandler) Logout(c *gin.Context) {
	authRequestsTotal.Inc()
	id, err := middleware.GetUserID(c)
	if err != nil {
		if errors.Is(err, domain.ErrValueNotFound) {
			response.Unauthorized(c, err, "not found value by key", h.logger)
			return
		}
		response.BadRequest(c, err, "incorrect type value", h.logger)
		return
	}

	if err := h.srvc.LogoutUser(c.Request.Context(), id); err != nil {
		response.InternalServerError(c, err, "user failed to logout", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}
