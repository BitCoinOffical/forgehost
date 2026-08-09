package loggerpkg

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger(level string) (*zap.Logger, error) {
	switch level {
	case "prod":
		prod, err := zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("zap.NewProduction: %w", err)
		}
		return prod, nil
	case "dev":
		dev, err := zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("zap.NewProduction: %w", err)
		}
		return dev, nil
	default:
		return nil, fmt.Errorf("incorrect debug level: %s", level)
	}
}
