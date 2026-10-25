package ws

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var strmRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "strm_requests_total",
	Help: "Total number of strm requests",
})

type StreamHandler struct {
	hub    *Hub
	logger *zap.Logger
	srvc   StreamService
}

func NewStreamHandler(hub *Hub, srvc StreamService, logger *zap.Logger) *StreamHandler {
	return &StreamHandler{hub: hub, logger: logger, srvc: srvc}
}

// SendStreamMessage godoc
// @Summary      Отправить сообщение в чат стрима
// @Description  Сохраняет сообщение и рассылает зрителям стрима, подключённым по WebSocket. Отправитель должен быть в списке зрителей
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        video_id  path  string          true  "ID видео на YouTube"
// @Param        request   body  dto.MessageDTO  true  "Текст сообщения"
// @Success      200
// @Failure      400  {object}  response.ErrorBody
// @Failure      401  {object}  response.ErrorBody
// @Failure      403  {object}  response.ErrorBody  "Пользователь не в списке зрителей стрима"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/ws/message/stream/{video_id} [get]
func (h *StreamHandler) SendStreamMessage(c *gin.Context) {
	strmRequestsTotal.Inc()
	videoID := c.Param("video_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	var msg dto.MessageDTO
	if err := c.ShouldBindBodyWithJSON(&msg); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	users, err := h.srvc.GetStreamUsers(c.Request.Context(), videoID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get users", h.logger)
		return
	}

	recipients := make([]string, 0, len(users))
	isMember := false
	for _, user := range users {
		recipients = append(recipients, user)
		if user == userID {
			isMember = true
		}
	}
	if !isMember {
		response.Forbidden(c, domain.ErrForbidden, "not a stream member", h.logger)
		return
	}

	if err := h.srvc.SaveStreamMessage(c.Request.Context(), userID, videoID, &msg); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed save stream message", h.logger)
		return
	}

	payload := map[string]any{
		"video_id": videoID,
		"from":     userID,
		"text":     msg.Text,
	}

	h.hub.Broadcast(c.Request.Context(), recipients, payload)

	c.Status(http.StatusOK)

}

// GetStreamHistory godoc
// @Summary      История чата стрима
// @Description  Возвращает последние сообщения чата стрима (хранятся 24 часа, не более 200)
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        video_id  path      string  true  "ID видео на YouTube"
// @Success      200       {array}   models.Message
// @Failure      401       {object}  response.ErrorBody
// @Failure      500       {object}  response.ErrorBody
// @Router       /social/message/stream/{video_id} [get]
func (h *StreamHandler) GetStreamHistory(c *gin.Context) {
	strmRequestsTotal.Inc()
	videoID := c.Param("video_id")

	msgs, err := h.srvc.GetStreamHistory(c.Request.Context(), videoID)
	if err != nil {
		response.InternalServerError(c, err, "invalid get strwam history", h.logger)
		return
	}

	c.JSON(http.StatusOK, msgs)
}

// SaveStreamUser godoc
// @Summary      Присоединиться к чату стрима
// @Description  Добавляет текущего пользователя в список зрителей стрима. Только зрители из этого списка могут писать в чат стрима и получать сообщения по WebSocket
// @Tags         streams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        video_id  path  string  true  "ID видео на YouTube"
// @Success      200
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/streams/{video_id}/join [post]
func (h *StreamHandler) SaveStreamUser(c *gin.Context) {
	strmRequestsTotal.Inc()
	videoID := c.Param("video_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.SaveStreamUser(c.Request.Context(), userID, videoID); err != nil {
		response.InternalServerError(c, err, "failed save user", h.logger)
		return
	}

	c.Status(http.StatusOK)
}
