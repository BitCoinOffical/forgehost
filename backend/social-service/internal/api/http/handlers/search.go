package handlers

import (
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	limit = 10
)

type SearchHandler struct {
	srvc   SearchService
	logger *zap.Logger
}

func NewSearchHandler(srvc SearchService, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{srvc: srvc, logger: logger}
}

func (h *SearchHandler) Search(c *gin.Context) {
	searchText := c.Query("q")
	typ := c.Query("type")
	switch typ {
	case "posts", "streams", "users":
	default:
		response.BadRequest(c, domain.ErrInvalidRequest, "invalid type", h.logger)
		return
	}

	res, err := h.srvc.Search(c.Request.Context(), searchText, typ, limit)
	if err != nil {
		response.InternalServerError(c, err, "failed search", h.logger)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SearchHandler) Suggest(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	searchText := c.Query("q")

	res, err := h.srvc.Suggest(c.Request.Context(), userID, searchText, limit)
	if err != nil {
		response.InternalServerError(c, err, "failed suggest", h.logger)
		return
	}
	c.JSON(http.StatusOK, res)
}
