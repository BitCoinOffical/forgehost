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
	Postgres  PostgresConfig
	WebGoogle WebGoogleConfig
	Resend    ResendConfig
	RabbitMQ  RabbitConfig
	Redis     RedisConfig
	App       AppConfig
}

type RabbitConfig struct {
	RabbitUser string `env:"RABBITMQ_DEFAULT_USER,required"`
	RabbitPass string `env:"RABBITMQ_DEFAULT_PASS,required"`
	RabbitHost string `env:"RABBITMQ_DEFAULT_HOST,required"`
	RabbitPort string `env:"RABBITMQ_DEFAULT_PORT,required"`
}

type WebGoogleConfig struct {
	WebClientID     string `env:"WEB_GOOGLE_CLIENT_ID,required"`
	WebClientSecret string `env:"WEB_GOOGLE_CLIENT_SECRET,required"`
	WebRedirectURL  string `env:"WEB_GOOGLE_REDIRECT_URL,required"`
}

type PostgresConfig struct {
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	DBHost     string `env:"DB_HOST,required"`
	DBPort     string `env:"DB_PORT,required"`
	DBName     string `env:"DB_NAME,required"`
}

type RedisConfig struct {
	RDBSessionAddr string `env:"SESSION_RDB_ADDR,required"`
	RDBSessionPort string `env:"SESSION_RDB_PORT,required"`
	RDBSessionDB   int    `env:"SESSION_RDB_DB,required"`

	RDBRateAddr  string `env:"RATE_RDB_ADDR,required"`
	RDBRatePort  string `env:"RATE_RDB_PORT,required"`
	RDBLimiterDB int    `env:"RATE_RDB_DB,required"`

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

	if err := env.Parse(&cfg.WebGoogle); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Postgres); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Resend); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.RabbitMQ); err != nil {
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
