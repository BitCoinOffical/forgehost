package handlers

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

var chatsRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "chats_requests_total",
	Help: "Total number of chats requests",
})

type ChatHandler struct {
	srvc   ChatService
	logger *zap.Logger
}

func NewChatHandler(srvc ChatService, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{srvc: srvc, logger: logger}
}

// CreateGroup godoc
// @Summary      Создать групповой чат
// @Description  Создаёт групповой чат, создатель становится его владельцем. Возвращает ID созданного чата
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.CreateChatDTO  true  "Данные чата"
// @Success      201      {string}  string             "ID созданного чата"
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/chats [post]
func (h *ChatHandler) CreateGroup(c *gin.Context) {
	chatsRequestsTotal.Inc()
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

	chatID, err := h.srvc.CreateGroup(c.Request.Context(), ownerID, &req)
	if err != nil {
		response.InternalServerError(c, err, "failed create group chat", h.logger)
		return
	}

	c.JSON(http.StatusCreated, chatID)
}

// CreateDirect godoc
// @Summary      Создать личный чат
// @Description  Создаёт личный чат с указанным пользователем. Если один из пользователей заблокировал другого, чат создать нельзя. Возвращает ID чата
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user_id  path      string  true  "ID собеседника"
// @Success      201      {string}  string  "ID личного чата"
// @Failure      401      {object}  response.ErrorBody
// @Failure      403      {object}  response.ErrorBody  "Один из пользователей заблокирован"
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/chats/directs/{user_id} [post]
func (h *ChatHandler) CreateDirect(c *gin.Context) {
	chatsRequestsTotal.Inc()
	targetID := c.Param("user_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	chatID, err := h.srvc.CreateDirect(c.Request.Context(), userID, targetID)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed create direct chat", h.logger)
		return
	}

	c.JSON(http.StatusCreated, chatID)
}

// GetChatList godoc
// @Summary      Список личных чатов
// @Description  Возвращает личные чаты текущего пользователя
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.FeedDirects
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/chats/list [get]
func (h *ChatHandler) GetChatList(c *gin.Context) {
	chatsRequestsTotal.Inc()
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	resp, err := h.srvc.GetChatList(c.Request.Context(), userID)
	if err != nil {
		response.InternalServerError(c, err, "failed get chat list", h.logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// MarkAsRead godoc
// @Summary      Отметить чат прочитанным
// @Description  Запоминает время последнего прочтения чата текущим пользователем
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path  string  true  "ID чата"
// @Success      200
// @Failure      401  {object}  response.ErrorBody
// @Failure      404  {object}  response.ErrorBody  "Чат не найден"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/chats/{chat_id}/read [post]
func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	chatsRequestsTotal.Inc()
	chatID := c.Param("chat_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.MarkAsRead(c.Request.Context(), chatID, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "message not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed mark as read", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

// GetChatByID godoc
// @Summary      Получить чат по ID
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path      string  true  "ID чата"
// @Success      200      {object}  models.Chat
// @Failure      401      {object}  response.ErrorBody
// @Failure      404      {object}  response.ErrorBody  "Чат не найден"
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/chats/{chat_id} [get]
func (h *ChatHandler) GetChatByID(c *gin.Context) {
	chatsRequestsTotal.Inc()
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

// DeleteChat godoc
// @Summary      Удалить чат
// @Description  Удалить чат может только его владелец
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path  string  true  "ID чата"
// @Success      204
// @Failure      401  {object}  response.ErrorBody
// @Failure      403  {object}  response.ErrorBody  "Недостаточно прав"
// @Failure      404  {object}  response.ErrorBody  "Чат не найден"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/chats/{chat_id} [delete]
func (h *ChatHandler) DeleteChat(c *gin.Context) {
	chatsRequestsTotal.Inc()
	chatID := c.Param("chat_id")
	ownerID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	if err := h.srvc.DeleteChat(c.Request.Context(), chatID, ownerID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "chat not found", h.logger)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "failed delete chat", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed delete chat", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

// LeaveChat godoc
// @Summary      Выйти из чата
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path  string  true  "ID чата"
// @Success      200
// @Failure      401  {object}  response.ErrorBody
// @Failure      404  {object}  response.ErrorBody  "Чат не найден"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/chats/{chat_id}/leave [post]
func (h *ChatHandler) LeaveChat(c *gin.Context) {
	chatsRequestsTotal.Inc()
	chatID := c.Param("chat_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.LeaveChat(c.Request.Context(), chatID, userID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "chat not found", h.logger)
			return
		}

		response.InternalServerError(c, err, "failed leave chat", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

// UpdateChat godoc
// @Summary      Обновить чат
// @Description  Изменяет название и описание чата. Передаются только изменяемые поля. Доступно владельцу и модераторам
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path      string             true  "ID чата"
// @Param        request  body      dto.UpdateChatDTO  true  "Изменяемые поля"
// @Success      200      {object}  models.Chat
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      403      {object}  response.ErrorBody  "Недостаточно прав"
// @Failure      404      {object}  response.ErrorBody  "Чат не найден"
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/chats/{chat_id} [patch]
func (h *ChatHandler) UpdateChat(c *gin.Context) {
	chatsRequestsTotal.Inc()
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
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "chat not found", h.logger)
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed update chat", h.logger)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SetUserRoleChat godoc
// @Summary      Назначить роль участнику чата
// @Description  Меняет роль участника (member или moderator). Доступно только владельцу чата
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path      string                 true  "ID чата"
// @Param        user_id  path      string                 true  "ID участника"
// @Param        request  body      dto.SetMemberRoleDTO   true  "Новая роль"
// @Success      200      {object}  models.ChatMember
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      403      {object}  response.ErrorBody  "Недостаточно прав"
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/chats/{chat_id}/members/{user_id}/role [patch]
func (h *ChatHandler) SetUserRoleChat(c *gin.Context) {
	chatsRequestsTotal.Inc()
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
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed set role", h.logger)
		return
	}

	c.JSON(http.StatusOK, member)
}

// JoinInChat godoc
// @Summary      Вступить в чат
// @Description  Добавляет текущего пользователя в чат. Забаненный в чате пользователь вступить не может
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path  string  true  "ID чата"
// @Success      200
// @Failure      401  {object}  response.ErrorBody
// @Failure      403  {object}  response.ErrorBody  "Пользователь забанен в чате"
// @Failure      409  {object}  response.ErrorBody  "Пользователь уже в чате"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/chats/{chat_id}/members [post]
func (h *ChatHandler) JoinInChat(c *gin.Context) {
	chatsRequestsTotal.Inc()
	chatID := c.Param("chat_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	if err := h.srvc.AddUserInChat(c.Request.Context(), chatID, userID); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Conflict(c, err, "user already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed add user in chat", h.logger)
		return
	}
	c.Status(http.StatusOK)
}

// KickUserFromChat godoc
// @Summary      Исключить участника из чата
// @Description  Доступно владельцу и модераторам. Владельца исключить нельзя
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path  string  true  "ID чата"
// @Param        user_id  path  string  true  "ID исключаемого участника"
// @Success      200
// @Failure      401  {object}  response.ErrorBody
// @Failure      403  {object}  response.ErrorBody  "Недостаточно прав или цель — владелец"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/chats/{chat_id}/members/{user_id} [delete]
func (h *ChatHandler) KickUserFromChat(c *gin.Context) {
	chatsRequestsTotal.Inc()
	chatID := c.Param("chat_id")
	targetID := c.Param("user_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.KickUserFromChat(c.Request.Context(), chatID, targetID, userID); err != nil {
		if errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrCannotKickOwner) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed kick user from chat", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

// GetUsersFromChat godoc
// @Summary      Список участников чата
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path      string  true  "ID чата"
// @Success      200      {array}   models.FeedChatMember
// @Failure      401      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/chats/{chat_id}/members [get]
func (h *ChatHandler) GetUsersFromChat(c *gin.Context) {
	chatsRequestsTotal.Inc()
	chatID := c.Param("chat_id")
	members, err := h.srvc.GetUsersFromChat(c.Request.Context(), chatID)
	if err != nil {
		response.InternalServerError(c, err, "failed get users from chat", h.logger)
		return
	}

	c.JSON(http.StatusOK, members)
}

// BanUser godoc
// @Summary      Забанить пользователя в чате
// @Description  Исключает пользователя и запрещает ему повторное вступление. Доступно владельцу и модераторам. Владельца забанить нельзя
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path  string  true  "ID чата"
// @Param        user_id  path  string  true  "ID банимого пользователя"
// @Success      200
// @Failure      401  {object}  response.ErrorBody
// @Failure      403  {object}  response.ErrorBody  "Недостаточно прав или цель — владелец"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/chats/{chat_id}/bans/{user_id} [post]
func (h *ChatHandler) BanUser(c *gin.Context) {
	chatsRequestsTotal.Inc()
	chatID := c.Param("chat_id")
	targetID := c.Param("user_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.BanUser(c.Request.Context(), chatID, targetID, userID); err != nil {
		if errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrCannotKickOwner) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed ban user from chat", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

// UnbanUser godoc
// @Summary      Разбанить пользователя в чате
// @Description  Доступно владельцу и модераторам
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path  string  true  "ID чата"
// @Param        user_id  path  string  true  "ID разбаниваемого пользователя"
// @Success      200
// @Failure      401  {object}  response.ErrorBody
// @Failure      403  {object}  response.ErrorBody  "Недостаточно прав"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/chats/{chat_id}/bans/{user_id} [delete]
func (h *ChatHandler) UnbanUser(c *gin.Context) {
	chatsRequestsTotal.Inc()
	chatID := c.Param("chat_id")
	targetID := c.Param("user_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.UnbanUser(c.Request.Context(), chatID, targetID, userID); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.Forbidden(c, err, "not enough rights", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed unban user", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

// CheckBanUser godoc
// @Summary      Проверить, забанен ли пользователь в чате
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        chat_id  path      string  true  "ID чата"
// @Param        user_id  path      string  true  "ID проверяемого пользователя"
// @Success      200      {boolean}  boolean  "true, если пользователь забанен"
// @Failure      401      {object}  response.ErrorBody
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/chats/{chat_id}/bans/{user_id} [get]
func (h *ChatHandler) CheckBanUser(c *gin.Context) {
	chatsRequestsTotal.Inc()
	chatID := c.Param("chat_id")
	targetID := c.Param("user_id")
	ok, err := h.srvc.CheckBanUser(c.Request.Context(), chatID, targetID)
	if err != nil {
		response.InternalServerError(c, err, "failed check ban user", h.logger)
		return
	}

	c.JSON(http.StatusOK, ok)
}
