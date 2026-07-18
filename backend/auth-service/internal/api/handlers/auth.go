package handlers

import (
	"BitCoinOffical/forgehost/auth-service/internal/api/middleware"
	"BitCoinOffical/forgehost/auth-service/internal/api/response"
	"BitCoinOffical/forgehost/auth-service/internal/domain"
	"BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	"BitCoinOffical/forgehost/auth-service/internal/interfaces/services"
	jwtpkg "BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	userContextKey = "user_claims"
)

type AuthHandler struct {
	logger   *zap.Logger
	oauthCfg *oauth2.Config
	authsrvc *services.AuthService
}

func NewAuthHandler(logger *zap.Logger, authsrvc *services.AuthService, oauthCfg *oauth2.Config) *AuthHandler {
	return &AuthHandler{logger: logger, authsrvc: authsrvc, oauthCfg: oauthCfg}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.UsersLoginDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
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
		response.BadRequest(c, err, "invalid request body", h.logger)
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
	oauthState, err := jwtpkg.GenerateSessionID()
	if err != nil {
		response.InternalServerError(c, err, "failed generate session id", h.logger)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "oauth_state",
		Value:    oauthState,
		Path:     "/",
		MaxAge:   300,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	u := h.oauthCfg.AuthCodeURL(oauthState)
	c.Redirect(http.StatusTemporaryRedirect, u)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	storedId, err := c.Cookie("oauth_state")
	if err != nil || c.Query("state") != storedId {
		response.BadRequest(c, err, "invalid state", h.logger)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	code := c.Query("code")
	if code == "" {
		response.BadRequest(c, err, "empty query", h.logger)
	}

	token, err := h.oauthCfg.Exchange(c.Request.Context(), code)
	if err != nil {
		response.Unauthorized(c, err, "failed to exchange code for token", h.logger)
		return
	}

	client := h.oauthCfg.Client(c.Request.Context(), token)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		response.BadGateway(c, err, "userinfo request failed", h.logger)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.InternalServerError(c, err, "read body error", h.logger)
	}

	var req dto.GoogleUserDTO
	if err := json.Unmarshal(body, &req); err != nil {
		response.InternalServerError(c, err, "failed json unmarshal", h.logger)
		return
	}

	if !req.EmailVerified {
		response.Forbidden(c, err, "email not verified", h.logger)
		return
	}

	tokens, err := h.authsrvc.GoogleCallback(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, err, "failed to register", h.logger)
		return
	}

	c.JSON(http.StatusOK, tokens)
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
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	tokens, err := h.authsrvc.UpdateAccessToken(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, err, "failed update token", h.logger)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) GoogleLoginAndroid(c *gin.Context) {
	var req dto.GoogleAndroidUserDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	tokens, err := h.authsrvc.GoogleLoginAndroid(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidGoogleToken) {
			response.Unauthorized(c, err, "invalid google token", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed google android login", h.logger)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {

}
func (h *AuthHandler) ResendVerifyEmail(c *gin.Context) {

}
func (h *AuthHandler) UpdatePassword(c *gin.Context) {

}
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {

}
func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {

}
