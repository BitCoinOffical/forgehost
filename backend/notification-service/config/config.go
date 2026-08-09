package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	Resend ResendConfig
	App    AppConfig
}

type AppConfig struct {
	Port       string `env:"PORT,required"`
	DebugLevel string `env:"DEBUG_LEVEL,required"`
	Network    string `env:"NETWORK,required"`
	Address    string `env:"ADDRESS,required"`
}

type ResendConfig struct {
	ResendApiKey string `env:"RESEND_API_KEY,required"`
}

func NewLoad() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg.App); err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	if err := env.Parse(&cfg.Resend); err != nil {
		return nil, fmt.Errorf("failed to load Resend config: %w", err)
	}

	return &cfg, nil
}
