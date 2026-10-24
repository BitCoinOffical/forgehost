package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	jwtpkg "github.com/BitCoinOffical/forgehost/social-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

const (
	redirect_url = "http://localhost:3000/oauth/settings?youtube=connected"
)

type StreamHandler struct {
	srvc     *services.StreamService
	oauthCfg *oauth2.Config
	logger   *zap.Logger
}

func NewStreamHandler(logger *zap.Logger, oauthCfg *oauth2.Config, srvc *services.StreamService) *StreamHandler {
	return &StreamHandler{logger: logger, oauthCfg: oauthCfg, srvc: srvc}
}

func (h *StreamHandler) GoogleLogin(c *gin.Context) {
	oauthState, err := jwtpkg.GenerateRandomString()
	if err != nil {
		response.InternalServerError(c, err, "failed gen random string", h.logger)
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

func (h *StreamHandler) GoogleCallback(c *gin.Context) {
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

	svc, err := yt.NewService(c.Request.Context(), option.WithTokenSource(h.oauthCfg.TokenSource(c.Request.Context(), token)))
	if err != nil {
		response.BadGateway(c, err, "youtube client failed", h.logger)
		return
	}

	resp, err := svc.Channels.List([]string{"id"}).Mine(true).Do()
	if err != nil || len(resp.Items) == 0 {
		response.BadRequest(c, err, "no youtube channel", h.logger)
		return
	}

	channelID := resp.Items[0].Id
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.SaveChannel(c.Request.Context(), channelID, userID); err != nil {
		response.InternalServerError(c, err, "failed save channel id", h.logger)
		return
	}

	c.Redirect(http.StatusSeeOther, redirect_url)
}

func (h *StreamHandler) GetChannel(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}
	channel, err := h.srvc.GetChannel(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "user not found", h.logger)
			return
		}
		if errors.Is(err, domain.ErrChannelNotLinked) {
			response.NotFound(c, err, "channel not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get channel", h.logger)
		return
	}

	c.JSON(http.StatusOK, channel)
}

func (h *StreamHandler) SaveStream(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	var req dto.StreamDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	streamID, err := h.srvc.SaveStream(c.Request.Context(), userID, &req)
	if err != nil {
		response.InternalServerError(c, err, "failed save stream", h.logger)
		return
	}

	c.JSON(http.StatusCreated, streamID)
}

func (h *StreamHandler) BuildUpdateStream(c *gin.Context) {
	videoID := c.Param("video_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	var req dto.UpdateStreamDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body", h.logger)
		return
	}

	stream, err := h.srvc.BuildUpdateStream(c.Request.Context(), userID, videoID, &req)
	if err != nil {
		response.InternalServerError(c, err, "failed update stream", h.logger)
		return
	}

	c.JSON(http.StatusOK, stream)
}

func (h *StreamHandler) DeleteStream(c *gin.Context) {
	videoID := c.Param("video_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.DeleteStream(c.Request.Context(), videoID, userID); err != nil {
		response.InternalServerError(c, err, "failed delete stream", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *StreamHandler) GetStream(c *gin.Context) {
	videoID := c.Param("video_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	stream, err := h.srvc.GetStream(c.Request.Context(), videoID, userID)
	if err != nil {
		response.InternalServerError(c, err, "failed get stream", h.logger)
		return
	}

	c.JSON(http.StatusOK, stream)
}

func (h *StreamHandler) GetCursorStreams(c *gin.Context) {
	cursor := c.Param("cursor")
	streams, err := h.srvc.GetCursorStreams(c.Request.Context(), cursor)
	if err != nil {
		response.InternalServerError(c, err, "failed get streams", h.logger)
		return
	}
	c.JSON(http.StatusOK, streams)
}
