package config

import (
	"fmt"

	"github.com/caarlos0/env"
)

const (
	EnvProd = "prod"
	EnvDev  = "dev"
)

type Env string

type Config struct {
	Postgres PostgresConfig
	Google   GoogleConfig
	Resend   ResendConfig
	Redis    RedisConfig
	App      AppConfig
}

type GoogleConfig struct {
	ClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	ClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	RedirectURL  string `env:"GOOGLE_REDIRECT_URL,required"`
}

type PostgresConfig struct {
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBHost     string `env:"DB_HOST,required"`
	DBPort     string `env:"DB_PORT,required"`
	DBName     string `env:"DB_NAME,required"`
}

type RedisConfig struct {
	RDBSessionAddr string `env:"RDB_ADDR,required"`
	RDBSessionPort string `env:"RDB_PORT,required"`
	RDBSessionDB   int    `env:"RDB_TOCKEN_DB,required"`

	RDBRateAddr  string `env:"RDB_ADDR,required"`
	RDBRatePort  string `env:"RDB_PORT,required"`
	RDBLimiterDB int    `env:"RDB_RATE_LIMITER_DB,required"`

	RDBPass string `env:"RDB_PASS,required"`
}

type ResendConfig struct {
	ResendApiKey string `env:"RESEND_API_KEY,required"`
}

type AppConfig struct {
	Secret     string `env:"JWT_SECRET,required"`
	DebugLevel string `env:"DEBUG_LEVEL,required"`
	Port       string `env:"PORT,required"`
}

func NewLoad() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg.App); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Google); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Postgres); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Resend); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Redis); err != nil {
		return nil, err
	}

	var env Env = Env(cfg.App.DebugLevel)
	if env != EnvProd && env != EnvDev {
		return nil, fmt.Errorf("incorrect debug level: %s", env)
	}

	return &cfg, nil
}
