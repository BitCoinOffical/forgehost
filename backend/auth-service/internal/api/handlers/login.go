package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoginHandler struct {
	srvc   LoginService
	logger *zap.Logger
}

func NewLoginHandler(srvc LoginService, logger *zap.Logger) *LoginHandler {
	return &LoginHandler{
		srvc:   srvc,
		logger: logger,
	}
}

// Login godoc
// @Summary      Log in with email and password
// @Description  Authenticates a user and returns a new access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UsersLoginDTO  true  "Login credentials"
// @Success      200      {object}  dto.TokensDTO
// @Failure      400      {object}  map[string]string  "invalid body"
// @Failure      401      {object}  map[string]string  "invalid credentials"
// @Failure      500      {object}  map[string]string
// @Router       /auth/login [post]
func (h *LoginHandler) Login(c *gin.Context) {
	authRequestsTotal.Inc()
	var req dto.UsersLoginDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	tokens, err := h.srvc.LoginUser(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidCredentials) {
			response.Unauthorized(c, err, "invalid credentials", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed to login", h.logger)
		return
	}

	c.JSON(http.StatusOK, tokens)
}
