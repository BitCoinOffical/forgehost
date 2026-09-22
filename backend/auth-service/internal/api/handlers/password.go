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

// UpdatePassword godoc
// @Summary      Update password (authenticated)
// @Description  Changes the password of the currently authenticated user after verifying the old password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  dto.UserPasswordDTO  true  "Old and new password"
// @Success      200  "password updated"
// @Failure      400  {object}  map[string]string  "invalid body, incorrect type value, or passwords do not match"
// @Failure      401  {object}  map[string]string  "missing token or invalid old password"
// @Failure      500  {object}  map[string]string
// @Router       /auth/password/update [patch]
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	authRequestsTotal.Inc()
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

// PasswordReset godoc
// @Summary      Request password reset
// @Description  Sends a password reset code to the given email and returns a pending key used to confirm the reset.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.PasswordResetDTO  true  "Email"
// @Success      200      {object}  dto.PendingKeyDTO
// @Failure      400      {object}  map[string]string  "invalid body"
// @Failure      500      {object}  map[string]string
// @Router       /auth/password/reset [post]
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

// ConfirmPasswordReset godoc
// @Summary      Confirm password reset
// @Description  Confirms the reset code sent to the user's email and sets the new password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  dto.PasswordResetDTO  true  "Email, pending key, code and new password"
// @Success      200  "password reset"
// @Failure      400  {object}  map[string]string  "invalid body or passwords do not match"
// @Failure      500  {object}  map[string]string
// @Router       /auth/password/reset/confirm [post]
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

// PasswordResetResend godoc
// @Summary      Resend password reset code
// @Description  Resends a fresh password reset code to the user's email. Rate-limited.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  dto.PasswordResetDTO  true  "Email and pending key"
// @Success      200  "code resent"
// @Failure      400  {object}  map[string]string  "invalid body"
// @Failure      429  {object}  map[string]string  "too many attempts"
// @Failure      500  {object}  map[string]string
// @Router       /auth/password/reset/resend [post]
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
