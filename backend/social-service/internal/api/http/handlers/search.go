package handlers

import (
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

const (
	limit = 10
)

var searchRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "search_requests_total",
	Help: "Total number of search requests",
})

type SearchHandler struct {
	srvc   SearchService
	logger *zap.Logger
}

func NewSearchHandler(srvc SearchService, logger *zap.Logger) *SearchHandler {
	return &SearchHandler{srvc: srvc, logger: logger}
}

// Search godoc
// @Summary      Поиск
// @Description  Поиск постов, стримов или пользователей. Возвращает не более 10 результатов
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        q     query  string  false  "Поисковый запрос"
// @Param        type  query  string  true   "Тип поиска"  Enums(posts, streams, users)
// @Success      200   {object}  models.SearchResponse
// @Failure      400   {object}  response.ErrorBody
// @Failure      500   {object}  response.ErrorBody
// @Router       /social/search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	searchRequestsTotal.Inc()
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

// Suggest godoc
// @Summary      Поисковые подсказки
// @Description  Возвращает поисковые подсказки для текущего пользователя
// @Tags         search
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        q     query  string  false  "Поисковый запрос"
// @Success      200   {array}   models.Suggestion
// @Failure      401   {object}  response.ErrorBody
// @Failure      500   {object}  response.ErrorBody
// @Router       /social/search/suggest [get]
func (h *SearchHandler) Suggest(c *gin.Context) {
	searchRequestsTotal.Inc()
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
