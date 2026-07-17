package handlers

import (
	"BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	"BitCoinOffical/forgehost/auth-service/internal/api/response"
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/services"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	userContextKey = "user_claims"
)

type AuthHandler struct {
	logger   *zap.Logger
	authsrvc *services.AuthService
}

func NewAuthHandler(logger *zap.Logger, authsrvc *services.AuthService) *AuthHandler {
	return &AuthHandler{logger: logger, authsrvc: authsrvc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.UsersLoginDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "failed to login", h.logger)
		return
	}

	tokens, err := h.authsrvc.LoginUser(c.Request.Context(), &req)
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.UsersRegisterDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "failed to register", h.logger)
		return
	}

	if req.Password != req.PasswordRetry {
		response.BadRequest(c, domain.ErrPasswordMismatch, "passwords do not match", h.logger)
		return
	}

	tokens, err := h.authsrvc.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			response.Conflict(c, err, "email already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed to register", h.logger)
		return
	}

	c.JSON(http.StatusCreated, tokens)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {

}

func (h *AuthHandler) Logout(c *gin.Context) {
	id, err := middleware.GetUserID(c)
	if err != nil {
		if errors.Is(err, domain.ErrValueNotFound) {
			response.Unauthorized(c, err, "not found value by key", h.logger)
			return
		}
		response.BadRequest(c, err, "incorrect type value", h.logger)
		return
	}

	if err := h.authsrvc.LogoutUser(c.Request.Context(), id); err != nil {
		response.InternalServerError(c, err, "user failed to logout", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) UpdateAccessToken(c *gin.Context) {
	var req dto.TokensDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "failed to login", h.logger)
		return
	}

}
