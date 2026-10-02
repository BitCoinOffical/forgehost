package handlers

import (
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MessagerHandler struct {
	srvc   *services.MessageService
	logger *zap.Logger
}

func NewMessagerHandler(srvc *services.MessageService, logger *zap.Logger) *MessagerHandler {
	return &MessagerHandler{srvc: srvc, logger: logger}
}

func (h *MessagerHandler) GetRoomHistory(c *gin.Context) {
	//roomId := c.Param("room_id")
}

func (h *MessagerHandler) GetGroupHistory(c *gin.Context) {
	//groupId := c.Param("group_id")
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

func (h *MessagerHandler) SaveRoomMessage(c *gin.Context) {
	//roomId := c.Param("room_id")

}

func (h *MessagerHandler) SaveGroupMessage(c *gin.Context) {
	//groupId := c.Param("group_id")
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
