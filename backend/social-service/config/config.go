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
	Kafka    KafkaConfig
	Redis    RedisConfig
	App      AppConfig
}

type KafkaConfig struct {
	Addr string `env:"KAFKA_ADDR,required"`
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

type AppConfig struct {
	DebugLevel string `env:"DEBUG_LEVEL,required"`
	Port       string `env:"PORT,required"`
}

func NewLoad() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg.App); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Kafka); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Postgres); err != nil {
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
