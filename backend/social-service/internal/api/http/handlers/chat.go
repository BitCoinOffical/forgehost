package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChatHandler struct {
	srvc   *services.ChatService
	logger *zap.Logger
}

func NewChatHandler(srvc *services.ChatService, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{srvc: srvc, logger: logger}
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	ownerID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	var req dto.CreateChatDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}
	chatID, err := h.srvc.CreateChat(c.Request.Context(), ownerID, &req)
	if err != nil {
		response.InternalServerError(c, err, "failed create chat", h.logger)
		return
	}

	c.JSON(http.StatusCreated, chatID)
}

func (h *ChatHandler) GetChatByID(c *gin.Context) {
	chatID := c.Param("chat_id")

	chat, err := h.srvc.GetChatByID(c.Request.Context(), chatID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "chat not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get chat", h.logger)
		return
	}

	c.JSON(http.StatusOK, chat)
}

func (h *ChatHandler) DeleteChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	ownerID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	if err := h.srvc.DeleteChat(c.Request.Context(), chatID, ownerID); err != nil {
		response.InternalServerError(c, err, "failed delete chat", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ChatHandler) LeaveChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.LeaveChat(c.Request.Context(), chatID, userID); err != nil {
		response.InternalServerError(c, err, "failed leave chat", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

func (h *ChatHandler) UpdateChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	var req dto.UpdateChatDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}
	resp, err := h.srvc.BuildUpdateChat(c.Request.Context(), chatID, userID, &req)
	if err != nil {
		response.InternalServerError(c, err, "failed update chat", h.logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) SetUserRoleChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	ownerID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	var req dto.SetMemberRoleDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}
	member, err := h.srvc.SetUserRoleChat(c.Request.Context(), chatID, ownerID, &req)
	if err != nil {
		response.InternalServerError(c, err, "failed set role", h.logger)
		return
	}

	c.JSON(http.StatusOK, member)
}

func (h *ChatHandler) JoinInChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	if err := h.srvc.AddUserInChat(c.Request.Context(), chatID, userID); err != nil {
		response.InternalServerError(c, err, "failed add user in chat", h.logger)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatHandler) KickUserFromChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	targetID := c.Param("user_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.KickUserFromChat(c.Request.Context(), chatID, targetID, userID); err != nil {
		response.InternalServerError(c, err, "failed kick user from chat", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

func (h *ChatHandler) GetUsersFromChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	members, err := h.srvc.GetUsersFromChat(c.Request.Context(), chatID)
	if err != nil {
		response.InternalServerError(c, err, "failed get users from chat", h.logger)
		return
	}

	c.JSON(http.StatusOK, members)
}

func (h *ChatHandler) BanUser(c *gin.Context) {
	chatID := c.Param("chat_id")
	targetID := c.Param("user_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.BanUser(c.Request.Context(), chatID, targetID, userID); err != nil {
		response.InternalServerError(c, err, "failed ban user from chat", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

func (h *ChatHandler) UnbanUser(c *gin.Context) {
	chatID := c.Param("chat_id")
	targetID := c.Param("user_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.UnbanUser(c.Request.Context(), chatID, targetID, userID); err != nil {
		response.InternalServerError(c, err, "failed unban user", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

func (h *ChatHandler) CheckBanUser(c *gin.Context) {
	chatID := c.Param("chat_id")
	targetID := c.Param("user_id")
	ok, err := h.srvc.CheckBanUser(c.Request.Context(), chatID, targetID)
	if err != nil {
		response.InternalServerError(c, err, "failed check ban user", h.logger)
		return
	}

	c.JSON(http.StatusOK, ok)
}
