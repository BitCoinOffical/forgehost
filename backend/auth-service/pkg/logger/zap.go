package loggerpkg

import (
	"fmt"
	"os"

	"github.com/DeRuina/timberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	maxSize    = 100
	maxBackups = 5
	maxAge     = 1
)

func NewLogger(level string, logPath string) (*zap.Logger, error) {
	var encoderCfg zapcore.EncoderConfig
	var lvl zapcore.Level

	switch level {
	case "prod":
		encoderCfg = zap.NewProductionEncoderConfig()
		lvl = zap.InfoLevel
	case "dev":
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		lvl = zap.DebugLevel
	default:
		return nil, fmt.Errorf("incorrect debug level: %s", level)
	}

	fileWriter := &timberjack.Logger{
		Filename:    logPath,
		MaxSize:     maxSize,
		MaxBackups:  maxBackups,
		MaxAge:      maxAge,
		Compression: "gzip",
	}

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), fileWriter, lvl),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), zapcore.AddSync(os.Stdout), lvl),
	)

	return zap.New(core), nil
}
