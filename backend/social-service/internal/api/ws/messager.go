package ws

import (
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MessagerHandler struct {
	hub    *Hub
	logger *zap.Logger
	srvc   *services.MessageService
}

func NewMessageHandler(hub *Hub, srvc *services.MessageService, logger *zap.Logger) *MessagerHandler {
	return &MessagerHandler{hub: hub, logger: logger, srvc: srvc}
}

func (h *MessagerHandler) SendMessage(c *gin.Context) {
	chatID := c.Param("chat_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, nil)
	if err != nil {
		response.InternalServerError(c, err, "failed accept websocket", h.logger)
		return
	}

	defer func() {
		if err := conn.CloseNow(); err != nil {
			h.logger.Error("conn.CloseNow error", zap.Error(err))
		}
	}()

	h.hub.Register(userID, conn)
	defer h.hub.Unregister(userID)

	for {
		var msg dto.MessageDTO
		if err := wsjson.Read(c.Request.Context(), conn, &msg); err != nil {
			h.logger.Error("user disconnected", zap.Error(err))
			return
		}

		if err := h.hub.SendTo(c.Request.Context(), chatID, map[string]any{
			"from": userID,
			"text": msg.Text,
		}); err != nil {
			h.logger.Error("send to", zap.String("chat_id", chatID), zap.Error(err))
			continue
		}

		objID, err := h.srvc.SaveMessage(c.Request.Context(), userID, chatID, &msg)
		if err != nil {
			response.InternalServerError(c, err, "failed save message", h.logger)
			continue
		}

		c.JSON(http.StatusOK, objID)
	}
}

func (h *MessagerHandler) GetChatHistory(c *gin.Context) {
	chatID := c.Param("chat_id")

	msgs, err := h.srvc.GetMessages(c.Request.Context(), chatID)
	if err != nil {
		response.InternalServerError(c, err, "failed get history chat", h.logger)
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func (h *MessagerHandler) SaveStreamMessage(c *gin.Context) {
	streamId := c.Param("stream_id")

	id, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	var req dto.MessageDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if err := h.srvc.SaveStreamMessage(c.Request.Context(), id, streamId, &req); err != nil {
		response.InternalServerError(c, err, "failed save stream message", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *MessagerHandler) GetStreamHistory(c *gin.Context) {
	streamId := c.Param("stream_id")

	msgs, err := h.srvc.GetStreamHistory(c.Request.Context(), streamId)
	if err != nil {
		response.InternalServerError(c, err, "invalid get strwam history", h.logger)
		return
	}

	c.JSON(http.StatusOK, msgs)
}
