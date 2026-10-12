package ws

import (
	"context"
	"net/http"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/coder/websocket"
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

func (h *MessagerHandler) Connect(c *gin.Context) {
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

func (h *MessagerHandler) SendMessageGroup(c *gin.Context) {
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

func (h *MessagerHandler) SendMessageDirect(c *gin.Context) {
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

func (h *MessagerHandler) GetChatHistory(c *gin.Context) {
	chatID := c.Param("chat_id")

	msgs, err := h.srvc.GetMessages(c.Request.Context(), chatID)
	if err != nil {
		response.InternalServerError(c, err, "failed get history chat", h.logger)
		return
	}

	c.JSON(http.StatusOK, msgs)
}
