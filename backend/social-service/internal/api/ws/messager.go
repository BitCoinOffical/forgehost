package ws

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var msgRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "msg_requests_total",
	Help: "Total number of msg requests",
})

type MessagerHandler struct {
	hub    *Hub
	logger *zap.Logger
	srvc   MessageService
}

func NewMessageHandler(hub *Hub, srvc MessageService, logger *zap.Logger) *MessagerHandler {
	return &MessagerHandler{hub: hub, logger: logger, srvc: srvc}
}

// Connect godoc
// @Summary      Подключиться к WebSocket
// @Description  Апгрейд соединения до WebSocket. Сервер каждые 30 секунд отправляет ping. По сокету приходят JSON-события: для чатов {"id","chat_id","from","text"}, для стримов {"video_id","from","text"}. Сообщения отправляются отдельными HTTP-запросами, по сокету клиент ничего не шлёт. Браузерный WebSocket не умеет передавать заголовок Authorization
// @Tags         websocket
// @Security     BearerAuth
// @Success      101  "Switching Protocols"
// @Failure      401  {object}  response.ErrorBody
// @Failure      426  "Не WebSocket-запрос"
// @Router       /social/ws/connect [get]
func (h *MessagerHandler) Connect(c *gin.Context) {
	msgRequestsTotal.Inc()
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("failed accept websocket", zap.Error(err))
		return
	}
	ctx := conn.CloseRead(c.Request.Context())
	defer func() {
		if err := conn.CloseNow(); err != nil {
			h.logger.Debug("conn.CloseNow error", zap.Error(err))
		}
	}()

	h.hub.Register(userID, conn)
	defer h.hub.Unregister(userID, conn)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("close with context")
			return
		case <-ticker.C:
			pctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := conn.Ping(pctx)
			cancel()
			if err != nil {
				h.logger.Error("ping error", zap.Error(err))
				return
			}
		}
	}
}

// SendMessageGroup godoc
// @Summary      Отправить сообщение в групповой чат
// @Description  Сохраняет сообщение и рассылает его участникам, подключённым по WebSocket. Отправитель должен быть участником чата и не быть забаненным. Возвращает ID сообщения
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path      string         true  "ID чата"
// @Param        request  body      dto.MessageDTO  true  "Текст сообщения"
// @Success      200      {string}  string          "ID сообщения"
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      403      {object}  response.ErrorBody  "Не участник чата или забанен"
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/ws/message/group/{chat_id} [get]
func (h *MessagerHandler) SendMessageGroup(c *gin.Context) {
	msgRequestsTotal.Inc()
	chatID := c.Param("chat_id")
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

	members, err := h.srvc.GetUsersFromChat(c.Request.Context(), chatID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get members", h.logger)
		return
	}

	recipients := make([]string, 0, len(members))
	isMember := false
	for _, m := range members {
		recipients = append(recipients, m.UserID)
		if m.UserID == userID {
			isMember = true
		}
	}
	if !isMember {
		response.Forbidden(c, domain.ErrForbidden, "not a chat member", h.logger)
		return
	}

	objID, err := h.srvc.SaveGroupMessage(c.Request.Context(), userID, chatID, &msg)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed save message", h.logger)
		return
	}

	payload := map[string]any{
		"id":      objID,
		"chat_id": chatID,
		"from":    userID,
		"text":    msg.Text,
	}

	h.hub.Broadcast(c.Request.Context(), recipients, payload)

	c.JSON(http.StatusOK, objID)
}

// SendMessageDirect godoc
// @Summary      Отправить сообщение в личный чат
// @Description  Сохраняет сообщение и рассылает обоим участникам, подключённым по WebSocket. Отправитель должен быть участником чата, а собеседник не должен быть в блокировке. Возвращает ID сообщения
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path      string         true  "ID личного чата"
// @Param        request  body      dto.MessageDTO  true  "Текст сообщения"
// @Success      200      {string}  string          "ID сообщения"
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      403      {object}  response.ErrorBody  "Не участник чата или блокировка"
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/ws/message/direct/{chat_id} [get]
func (h *MessagerHandler) SendMessageDirect(c *gin.Context) {
	msgRequestsTotal.Inc()
	chatID := c.Param("chat_id")
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

	members, err := h.srvc.GetUsersFromChat(c.Request.Context(), chatID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get members", h.logger)
		return
	}

	recipients := make([]string, 0, len(members))
	isMember := false
	for _, m := range members {
		recipients = append(recipients, m.UserID)
		if m.UserID == userID {
			isMember = true
		}
	}
	if !isMember {
		response.Forbidden(c, domain.ErrForbidden, "not a chat member", h.logger)
		return
	}

	objID, err := h.srvc.SaveDirectMessage(c.Request.Context(), userID, chatID, &msg)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed save message", h.logger)
		return
	}

	payload := map[string]any{
		"id":      objID,
		"chat_id": chatID,
		"from":    userID,
		"text":    msg.Text,
	}

	h.hub.Broadcast(c.Request.Context(), recipients, payload)

	c.JSON(http.StatusOK, objID)
}

// GetChatHistory godoc
// @Summary      История сообщений чата
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path      string  true  "ID чата"
// @Success      200      {array}   models.Message
// @Failure      401      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/message/chat/{chat_id} [get]
func (h *MessagerHandler) GetChatHistory(c *gin.Context) {
	msgRequestsTotal.Inc()
	chatID := c.Param("chat_id")

	msgs, err := h.srvc.GetMessages(c.Request.Context(), chatID)
	if err != nil {
		response.InternalServerError(c, err, "failed get history chat", h.logger)
		return
	}

	c.JSON(http.StatusOK, msgs)
}
