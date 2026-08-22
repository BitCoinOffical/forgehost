package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/BitCoinOffical/forgehost/auth-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/auth-service/internal/domain/dto"
	jwtpkg "github.com/BitCoinOffical/forgehost/auth-service/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	oauthState, err := jwtpkg.GenerateRandomString()
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
		return
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
		return
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
