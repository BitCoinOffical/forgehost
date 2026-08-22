package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	var req *dto.UserPasswordDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if req.NewPassword != req.NewPasswordRetry {
		response.BadRequest(c, domain.ErrPasswordMismatch, "passwords do not match", h.logger)
		return
	}

	idStr, err := middleware.GetUserID(c)
	if err != nil {
		if errors.Is(err, domain.ErrValueNotFound) {
			response.Unauthorized(c, err, "not found value by key", h.logger)
			return
		}
		response.BadRequest(c, err, "incorrect type value", h.logger)
		return
	}

	if err := h.authsrvc.UpdatePassword(c.Request.Context(), req, idStr); err != nil {
		response.InternalServerError(c, err, "failed update password", h.logger)
		return
	}
}

func (h *AuthHandler) PasswordReset(c *gin.Context) {
	var req dto.PasswordResetDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	key, err := h.authsrvc.PasswordReset(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, err, "failed reset password", h.logger)
		return
	}

	c.JSON(http.StatusOK, key)
}

func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	var req dto.PasswordResetDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if req.NewPassword != req.NewPasswordRetry {
		response.BadRequest(c, domain.ErrPasswordMismatch, "passwords do not match", h.logger)
		return
	}

	if err := h.authsrvc.ConfirmPasswordReset(c.Request.Context(), &req); err != nil {
		response.InternalServerError(c, err, "failed reset password", h.logger)
		return
	}

	c.Status(http.StatusOK)
}

func (h *AuthHandler) PasswordResetResend(c *gin.Context) {
	var req dto.PasswordResetDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if err := h.authsrvc.PasswordResetResend(c.Request.Context(), &req); err != nil {
		if errors.Is(err, domain.ErrToManyRequest) {
			response.ManyRequest(c, err, "too many attempts", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed update verify email", h.logger)
		return
	}

	c.Status(http.StatusOK)
}
