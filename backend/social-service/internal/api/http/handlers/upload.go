package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/response"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

const (
	postPrefix   = "posts/"
	avatarPrefix = "avatars/"
	maxImageSize = 10 << 20
	size         = 512
	offset       = 0
)

var uploaderRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "uploader_requests_total",
	Help: "Total number of uploader requests",
})

type UploaderHandler struct {
	srvc   UploaderService
	logger *zap.Logger
}

func NewUploaderHandler(srvc UploaderService, logger *zap.Logger) *UploaderHandler {
	return &UploaderHandler{srvc: srvc, logger: logger}
}

// UploadImage godoc
// @Summary      Загрузить изображение
// @Description  Загружает изображение для публикации. Поддерживаются JPEG, PNG и WebP. Максимальный размер — 10 МБ.
// @Tags         posts
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        image  formData  file  true  "Файл изображения"
// @Success      201  {string}  string  "ID загруженного изображения"
// @Failure      400  {object}  response.ErrorBody  "Некорректный файл или неподдерживаемый формат"
// @Failure      401  {object}  response.ErrorBody  "Пользователь не авторизован"
// @Failure      500  {object}  response.ErrorBody  "Ошибка загрузки изображения"
// @Router       /social/upload/post [post]
func (h *UploaderHandler) UploadPostImage(c *gin.Context) {
	uploaderRequestsTotal.Inc()
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageSize+1<<20)

	fileHeader, err := c.FormFile("image")
	if err != nil {
		response.BadRequest(c, err, "image is required", h.logger)
		return
	}
	if fileHeader.Size > maxImageSize {
		response.BadRequest(c, domain.ErrInvalidImage, "image is too large", h.logger)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InternalServerError(c, err, "failed open file", h.logger)
		return
	}
	defer file.Close()

	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		response.InternalServerError(c, err, "failed read file", h.logger)
		return
	}

	var allowedImageTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	contentType := http.DetectContentType(buf[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		response.BadRequest(c, domain.ErrInvalidImage, "unsupported image type", h.logger)
		return
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		response.InternalServerError(c, err, "failed seek file", h.logger)
		return
	}

	imageID, err := h.srvc.UploadImage(c.Request.Context(), postPrefix, userID, ext, contentType, file)
	if err != nil {
		response.InternalServerError(c, err, "failed upload image", h.logger)
		return
	}

	c.JSON(http.StatusCreated, imageID)
}

// UploadImage godoc
// @Summary      Загрузить изображение
// @Description  Загружает изображение для публикации. Поддерживаются JPEG, PNG и WebP. Максимальный размер — 10 МБ.
// @Tags         posts
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        image  formData  file  true  "Файл изображения"
// @Success      201  {string}  string  "ID загруженного изображения"
// @Failure      400  {object}  response.ErrorBody  "Некорректный файл или неподдерживаемый формат"
// @Failure      401  {object}  response.ErrorBody  "Пользователь не авторизован"
// @Failure      500  {object}  response.ErrorBody  "Ошибка загрузки изображения"
// @Router       /social/upload/avatar [post]
func (h *UploaderHandler) UploadAvatarImage(c *gin.Context) {
	uploaderRequestsTotal.Inc()
	userID, err := middleware.GetUserID(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user id", h.logger)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageSize+1<<20)

	fileHeader, err := c.FormFile("image")
	if err != nil {
		response.BadRequest(c, err, "image is required", h.logger)
		return
	}
	if fileHeader.Size > maxImageSize {
		response.BadRequest(c, domain.ErrInvalidImage, "image is too large", h.logger)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.InternalServerError(c, err, "failed open file", h.logger)
		return
	}
	defer file.Close()

	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		response.InternalServerError(c, err, "failed read file", h.logger)
		return
	}

	var allowedImageTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	contentType := http.DetectContentType(buf[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		response.BadRequest(c, domain.ErrInvalidImage, "unsupported image type", h.logger)
		return
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		response.InternalServerError(c, err, "failed seek file", h.logger)
		return
	}

	imageID, err := h.srvc.UploadImage(c.Request.Context(), avatarPrefix, userID, ext, contentType, file)
	if err != nil {
		response.InternalServerError(c, err, "failed upload image", h.logger)
		return
	}

	c.JSON(http.StatusCreated, imageID)
}
