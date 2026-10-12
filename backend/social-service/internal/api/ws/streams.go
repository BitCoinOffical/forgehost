package ws

import (
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StreamHandler struct {
	hub    *Hub
	logger *zap.Logger
	srvc   *services.StreamService
}

func NewStreamHandler(hub *Hub, srvc *services.StreamService, logger *zap.Logger) *StreamHandler {
	return &StreamHandler{hub: hub, logger: logger, srvc: srvc}
}

func (h *StreamHandler) SendStreamMessage(c *gin.Context) {
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
func (h *StreamHandler) GetStreamHistory(c *gin.Context) {
	videoID := c.Param("video_id")

	msgs, err := h.srvc.GetStreamHistory(c.Request.Context(), videoID)
	if err != nil {
		response.InternalServerError(c, err, "invalid get strwam history", h.logger)
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func (h *StreamHandler) SaveStreamUser(c *gin.Context) {
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
