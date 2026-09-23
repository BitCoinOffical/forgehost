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

type RegisterHandler struct {
	srvc   RegisterService
	logger *zap.Logger
}

func NewRegisterHandler(srvc RegisterService, logger *zap.Logger) *RegisterHandler {
	return &RegisterHandler{
		srvc:   srvc,
		logger: logger,
	}
}

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
func (h *RegisterHandler) Register(c *gin.Context) {
	authRequestsTotal.Inc()
	var req dto.UsersRegisterDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if req.Password != req.PasswordRetry {
		response.BadRequest(c, domain.ErrPasswordMismatch, "passwords do not match", h.logger)
		return
	}

	key, err := h.srvc.RegisterUser(c.Request.Context(), &req)
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

// VerifyEmail godoc
// @Summary      Confirm email verification code
// @Description  Confirms the code sent to the user's email during registration, activating the account and returning an access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.VerifyEmailDTO  true  "Email, pending key and verification code"
// @Success      200      {object}  dto.TokensDTO
// @Failure      400      {object}  map[string]string  "invalid body"
// @Failure      409      {object}  map[string]string  "email already exists"
// @Failure      500      {object}  map[string]string
// @Router       /auth/verify-email [post]
func (h *RegisterHandler) VerifyEmail(c *gin.Context) {
	authRequestsTotal.Inc()
	var req dto.VerifyEmailDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	tokens, err := h.srvc.VerifyEmail(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			response.Conflict(c, err, "email already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed verificate email", h.logger)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// ResendVerifyEmail godoc
// @Summary      Resend email verification code
// @Description  Resends a fresh verification code to the user's email. Rate-limited.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  dto.VerifyEmailDTO  true  "Email and pending key"
// @Success      200  "code resent"
// @Failure      400  {object}  map[string]string  "invalid body"
// @Failure      429  {object}  map[string]string  "too many attempts"
// @Failure      500  {object}  map[string]string
// @Router       /auth/verify-email/resend [post]
func (h *RegisterHandler) ResendVerifyEmail(c *gin.Context) {
	authRequestsTotal.Inc()
	var req dto.VerifyEmailDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	if err := h.srvc.ResendVerifyEmail(c.Request.Context(), &req); err != nil {
		if errors.Is(err, domain.ErrToManyRequest) {
			response.ManyRequest(c, err, "too many attempts", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed update verify email", h.logger)
		return
	}

	c.Status(http.StatusOK)
}
