package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	jwtpkg "github.com/BitCoinOffical/forgehost/social-service/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

var streamRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "stream_requests_total",
	Help: "Total number of stream requests",
})

const (
	redirect_url = "http://localhost:3000/oauth/settings?youtube=connected"
)

type StreamHandler struct {
	srvc     StreamService
	oauthCfg *oauth2.Config
	logger   *zap.Logger
}

func NewStreamHandler(logger *zap.Logger, oauthCfg *oauth2.Config, srvc StreamService) *StreamHandler {
	return &StreamHandler{logger: logger, oauthCfg: oauthCfg, srvc: srvc}
}

// GoogleLogin godoc
// @Summary      Начать подключение YouTube-канала
// @Description  Ставит cookie oauth_state на 5 минут и перенаправляет пользователя на страницу согласия Google (доступ youtube.readonly)
// @Tags         youtube
// @Security     BearerAuth
// @Success      307  "Перенаправление на Google OAuth"
// @Failure      401  {object}  response.ErrorBody
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/youtube/connect [get]
func (h *StreamHandler) GoogleLogin(c *gin.Context) {
	streamRequestsTotal.Inc()
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

// GoogleCallback godoc
// @Summary      Завершить подключение YouTube-канала
// @Description  Проверяет state из cookie, обменивает code на токен, получает ID канала пользователя через YouTube API и сохраняет его в профиле. В конце перенаправляет на страницу настроек фронтенда
// @Tags         youtube
// @Security     BearerAuth
// @Param        state  query  string  true  "Значение state, выданное при старте OAuth"
// @Param        code   query  string  true  "Код авторизации Google"
// @Success      303  "Перенаправление на страницу настроек"
// @Failure      400  {object}  response.ErrorBody  "Неверный state, пустой code или у аккаунта нет YouTube-канала"
// @Failure      401  {object}  response.ErrorBody  "Не удалось обменять code на токен"
// @Failure      409  {object}  response.ErrorBody  "Канал уже привязан к другому пользователю"
// @Failure      500  {object}  response.ErrorBody
// @Failure      502  {object}  response.ErrorBody  "Ошибка запроса к YouTube API"
// @Router       /social/youtube/callback [get]
func (h *StreamHandler) GoogleCallback(c *gin.Context) {
	streamRequestsTotal.Inc()
	storedID, err := c.Cookie("oauth_state")
	if err != nil {
		response.BadRequest(c, err, "invalid state", h.logger)
		return
	}
	if c.Query("state") != storedID {
		response.BadRequest(c, domain.ErrInvalidState, "invalid state", h.logger)
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
		response.BadRequest(c, domain.ErrEmptyValue, "empty query", h.logger)
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
	if err != nil {
		response.BadGateway(c, err, "youtube request failed", h.logger)
		return
	}
	if len(resp.Items) == 0 {
		response.BadRequest(c, errors.New("no youtube channel"), "no youtube channel", h.logger)
		return
	}

	channelID := resp.Items[0].Id
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.SaveChannel(c.Request.Context(), channelID, userID); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Conflict(c, err, "stream already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed save channel id", h.logger)
		return
	}

	c.Redirect(http.StatusSeeOther, redirect_url)
}

// GetChannel godoc
// @Summary      Получить привязанный YouTube-канал
// @Description  Возвращает ID YouTube-канала текущего пользователя
// @Tags         youtube
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {string}  string  "ID канала"
// @Failure      401  {object}  response.ErrorBody
// @Failure      404  {object}  response.ErrorBody  "Пользователь не найден или канал не привязан"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/youtube/channel [get]
func (h *StreamHandler) GetChannel(c *gin.Context) {
	streamRequestsTotal.Inc()
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

// SaveStream godoc
// @Summary      Сохранить стрим
// @Description  Регистрирует YouTube-трансляцию текущего пользователя. Возвращает ID созданной записи
// @Tags         streams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.StreamDTO  true  "Данные трансляции"
// @Success      201      {string}  string         "ID созданного стрима"
// @Failure      400      {object}  response.ErrorBody
// @Failure      401      {object}  response.ErrorBody
// @Failure      409      {object}  response.ErrorBody  "Такой стрим уже существует"
// @Failure      500      {object}  response.ErrorBody
// @Router       /social/streams [post]
func (h *StreamHandler) SaveStream(c *gin.Context) {
	streamRequestsTotal.Inc()
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
		if errors.Is(err, domain.ErrAlreadyExists) {
			response.Conflict(c, err, "stream already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed save stream", h.logger)
		return
	}

	c.JSON(http.StatusCreated, streamID)
}

// BuildUpdateStream godoc
// @Summary      Обновить стрим
// @Description  Изменяет данные стрима текущего пользователя. Передаются только изменяемые поля
// @Tags         streams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        video_id  path      string             true  "ID видео на YouTube"
// @Param        request   body      dto.UpdateStreamDTO  true  "Изменяемые поля"
// @Success      200       {object}  models.Stream
// @Failure      400       {object}  response.ErrorBody
// @Failure      401       {object}  response.ErrorBody
// @Failure      404       {object}  response.ErrorBody  "Стрим не найден"
// @Failure      500       {object}  response.ErrorBody
// @Router       /social/streams/{video_id} [patch]
func (h *StreamHandler) BuildUpdateStream(c *gin.Context) {
	streamRequestsTotal.Inc()
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
		if errors.Is(err, domain.ErrStreamNotFound) {
			response.NotFound(c, err, "stream not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed update stream", h.logger)
		return
	}

	c.JSON(http.StatusOK, stream)
}

// DeleteStream godoc
// @Summary      Удалить стрим
// @Tags         streams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        video_id  path  string  true  "ID видео на YouTube"
// @Success      204
// @Failure      401  {object}  response.ErrorBody
// @Failure      404  {object}  response.ErrorBody  "Стрим не найден"
// @Failure      500  {object}  response.ErrorBody
// @Router       /social/streams/{video_id} [delete]
func (h *StreamHandler) DeleteStream(c *gin.Context) {
	streamRequestsTotal.Inc()
	videoID := c.Param("video_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	if err := h.srvc.DeleteStream(c.Request.Context(), videoID, userID); err != nil {
		if errors.Is(err, domain.ErrStreamNotFound) {
			response.NotFound(c, err, "stream not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed delete stream", h.logger)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetStream godoc
// @Summary      Получить стрим
// @Description  Возвращает стрим текущего пользователя по ID видео
// @Tags         streams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        video_id  path      string  true  "ID видео на YouTube"
// @Success      200       {object}  models.Stream
// @Failure      401       {object}  response.ErrorBody
// @Failure      404       {object}  response.ErrorBody  "Стрим не найден"
// @Failure      500       {object}  response.ErrorBody
// @Router       /social/streams/{video_id} [get]
func (h *StreamHandler) GetStream(c *gin.Context) {
	streamRequestsTotal.Inc()
	videoID := c.Param("video_id")
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	stream, err := h.srvc.GetStream(c.Request.Context(), videoID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrStreamNotFound) {
			response.NotFound(c, err, "stream not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get stream", h.logger)
		return
	}

	c.JSON(http.StatusOK, stream)
}

// GetCursorStreams godoc
// @Summary      Лента стримов
// @Description  Возвращает до 10 стримов, от новых к старым. Для следующей страницы передаётся курсор из последнего элемента. Без курсора возвращается первая страница
// @Tags         streams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cursor  query     string  false  "Курсор пагинации (base64)"
// @Success      200     {array}   models.FeedStream
// @Failure      400     {object}  response.ErrorBody  "Неверный курсор"
// @Failure      401     {object}  response.ErrorBody
// @Failure      500     {object}  response.ErrorBody
// @Router       /social/streams/list [get]
func (h *StreamHandler) GetCursorStreams(c *gin.Context) {
	streamRequestsTotal.Inc()
	cursor := c.Param("cursor")
	streams, err := h.srvc.GetCursorStreams(c.Request.Context(), cursor)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCursor) {
			response.BadRequest(c, err, "invalid cursor", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get streams", h.logger)
		return
	}
	c.JSON(http.StatusOK, streams)
}
