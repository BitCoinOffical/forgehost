package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
)

// Register godoc
// @Summary      Register a new user
// @Description  Creates a pending registration and sends a verification code to the provided email. Returns a pending key used to confirm the email in /verify-email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UsersRegisterDTO  true  "Registration payload"
// @Success      201      {object}  dto.PendingKeyDTO
// @Failure      400      {object}  map[string]string  "invalid body or passwords do not match"
// @Failure      409      {object}  map[string]string  "email already exists"
// @Failure      500      {object}  map[string]string
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.UsersRegisterDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if req.Password != req.PasswordRetry {
		response.BadRequest(c, domain.ErrPasswordMismatch, "passwords do not match", h.logger)
		return
	}

	key, err := h.authsrvc.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			response.Conflict(c, err, "email already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed to register", h.logger)
		return
	}

	c.JSON(http.StatusCreated, key)
}
