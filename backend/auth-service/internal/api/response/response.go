package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Unauthorized(c *gin.Context, err error, msg string, logger *zap.Logger) {
	logger.Error(msg, zap.Error(err), zap.String("patch", c.FullPath()))
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": err.Error(),
	})
}

func BadRequest(c *gin.Context, err error, msg string, logger *zap.Logger) {
	logger.Error(msg, zap.Error(err), zap.String("path", c.FullPath()))
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}

func InternalServerError(c *gin.Context, err error, msg string, logger *zap.Logger) {
	logger.Error(msg, zap.Error(err), zap.String("patch", c.FullPath()))
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
	})
}

func Conflict(c *gin.Context, err error, msg string, logger *zap.Logger) {
	logger.Error(msg, zap.Error(err), zap.String("patch", c.FullPath()))
	c.JSON(http.StatusConflict, gin.H{
		"error": err.Error(),
	})
}

func BadGateway(c *gin.Context, err error, msg string, logger *zap.Logger) {
	logger.Error(msg, zap.Error(err), zap.String("patch", c.FullPath()))
	c.JSON(http.StatusBadGateway, gin.H{
		"error": err.Error(),
	})
}

func Forbidden(c *gin.Context, err error, msg string, logger *zap.Logger) {
	logger.Error(msg, zap.Error(err), zap.String("patch", c.FullPath()))
	c.JSON(http.StatusForbidden, gin.H{
		"error": err.Error(),
	})
}
