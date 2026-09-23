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
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	redirect_url = "http://localhost:3000/oauth/callback?code="
)

type GoogleAuthHandler struct {
	srvc     GoogleAuthService
	oauthCfg *oauth2.Config
	logger   *zap.Logger
}

func NewGoogleAuthHandler(srvc GoogleAuthService, oauthCfg *oauth2.Config, logger *zap.Logger) *GoogleAuthHandler {
	return &GoogleAuthHandler{
		srvc:     srvc,
		oauthCfg: oauthCfg,
		logger:   logger,
	}
}

// GoogleLogin godoc
// @Summary      Start Google OAuth login (web)
// @Description  Redirects the browser to Google's consent screen. Sets an "oauth_state" CSRF-protection cookie.
// @Tags         auth
// @Produce      json
// @Success      307  "redirect to Google consent screen"
// @Failure      500  {object}  map[string]string
// @Router       /auth/login/google [get]
func (h *GoogleAuthHandler) GoogleLogin(c *gin.Context) {
	authRequestsTotal.Inc()
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

// GoogleCallback godoc
// @Summary      Google OAuth callback (web)
// @Description  Handles Google's redirect after consent, validates state, exchanges the code, fetches the Google profile, then redirects to the frontend with a short-lived one-time exchange code.
// @Tags         auth
// @Produce      json
// @Param        state  query  string  true  "OAuth state, must match the oauth_state cookie"
// @Param        code   query  string  true  "Authorization code issued by Google"
// @Success      307  "redirect to frontend with one-time exchange code"
// @Failure      400  {object}  map[string]string  "invalid state or empty code"
// @Failure      401  {object}  map[string]string  "failed to exchange code for token"
// @Failure      403  {object}  map[string]string  "google email not verified"
// @Failure      500  {object}  map[string]string
// @Failure      502  {object}  map[string]string  "userinfo request failed"
// @Router       /auth/login/google/callback [get]
func (h *GoogleAuthHandler) GoogleCallback(c *gin.Context) {
	authRequestsTotal.Inc()
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

	oauthCode, err := h.srvc.GoogleCallback(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, err, "failed to register", h.logger)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirect_url+oauthCode)
}

// Exchange godoc
// @Summary      Exchange one-time OAuth code for tokens
// @Description  Called by the frontend right after the Google OAuth redirect. Exchanges the short-lived one-time code (from the callback redirect) for a real access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ExchangeRequestDTO  true  "One-time exchange code"
// @Success      200      {object}  dto.TokensDTO
// @Failure      400      {object}  map[string]string  "invalid body"
// @Failure      500      {object}  map[string]string
// @Router       /auth/exchange [post]
func (h *GoogleAuthHandler) Exchange(c *gin.Context) {
	authRequestsTotal.Inc()
	var req dto.ExchangeRequestDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	tokens, err := h.srvc.Exchange(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, err, "failed get tokens", h.logger)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// GoogleLoginAndroid godoc
// @Summary      Log in with Google (Android)
// @Description  Verifies a Google ID token obtained natively on Android and returns an access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.GoogleAndroidUserDTO  true  "Google ID token"
// @Success      200      {object}  dto.TokensDTO
// @Failure      400      {object}  map[string]string  "invalid body"
// @Failure      401      {object}  map[string]string  "invalid google token"
// @Failure      500      {object}  map[string]string
// @Router       /auth/login/google [post]
func (h *GoogleAuthHandler) GoogleLoginAndroid(c *gin.Context) {
	authRequestsTotal.Inc()
	var req dto.GoogleAndroidUserDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	tokens, err := h.srvc.GoogleLoginAndroid(c.Request.Context(), req)
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
