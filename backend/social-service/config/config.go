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
	WebGoogle WebGoogleConfig
	Postgres  PostgresConfig
	Kafka     KafkaConfig
	Redis     RedisConfig
	Mongo     MongoConfig
	App       AppConfig
	S3        S3Config
}

type KafkaConfig struct {
	Addr string `env:"KAFKA_ADDR,required"`
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
	RDBCacheAddr string `env:"CACHE_RDB_ADDR,required"`
	RDBCachePort string `env:"CACHE_RDB_PORT,required"`
	RDBCacheDB   int    `env:"CACHE_RDB_DB,required"`

	RDBRateAddr  string `env:"RATE_RDB_ADDR,required"`
	RDBRatePort  string `env:"RATE_RDB_PORT,required"`
	RDBLimiterDB int    `env:"RATE_RDB_DB,required"`

	RDBPass string `env:"RDB_PASS,required"`
}

type S3Config struct {
	Endpoint       string `env:"S3_ENDPOINT,required"`
	PublicEndpoint string `env:"S3_PUBLIC_URL,required"`
	AccessKey      string `env:"S3_ACCESS_KEY,required"`
	SecretKey      string `env:"S3_SECRET_KEY,required"`
	Bucket         string `env:"S3_BUCKET,required"`
	Region         string `env:"S3_REGION,required"`
}

type MongoConfig struct {
	MongoHost string `env:"MONGO_HOST,required"`
	MongoPort string `env:"MONGO_PORT,required"`
}

type AppConfig struct {
	DebugLevel string `env:"DEBUG_LEVEL,required"`
	Secret     string `env:"JWT_SECRET,required"`
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

	if err := env.Parse(&cfg.WebGoogle); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Postgres); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Redis); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.S3); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Mongo); err != nil {
		return nil, err
	}

	var env Env = Env(cfg.App.DebugLevel)
	if env != EnvProd && env != EnvDev {
		return nil, fmt.Errorf("incorrect debug level: %s", env)
	}

	return &cfg, nil
}
