package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BitCoinOffical/forgehost/social-service/config"
	_ "github.com/BitCoinOffical/forgehost/social-service/docs"
	kafkaread "github.com/BitCoinOffical/forgehost/social-service/internal/adapters/kafka"
	"github.com/BitCoinOffical/forgehost/social-service/internal/adapters/migrations"
	mongodb "github.com/BitCoinOffical/forgehost/social-service/internal/adapters/mongo"
	postgresdb "github.com/BitCoinOffical/forgehost/social-service/internal/adapters/postgres"
	redisdb "github.com/BitCoinOffical/forgehost/social-service/internal/adapters/redis"
	"github.com/BitCoinOffical/forgehost/social-service/internal/adapters/seaweedfs"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/http/handlers"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/middleware"
	"github.com/BitCoinOffical/forgehost/social-service/internal/api/ws"
	"github.com/BitCoinOffical/forgehost/social-service/internal/consumers"
	jwtpkg "github.com/BitCoinOffical/forgehost/social-service/pkg/jwt"
	loggerpkg "github.com/BitCoinOffical/forgehost/social-service/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	yt "google.golang.org/api/youtube/v3"
)

const (
	logPath                 = "logs/social.log"
	mongoMessagerDataBase   = "chat"
	mongoMessagerCollection = "messages"
)

// @title Social Service API
// @version 1.0
// @description API service for a social networking service
// @termsOfService http://swagger.io/terms/

// @host localhost:8084
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewLoad()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := loggerpkg.NewLogger(cfg.App.DebugLevel, logPath)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("config load")
	logger.Info("logger load")

	client, err := kafkaread.NewKafkaClient(&kafkaread.KafkaConfig{
		Addr: cfg.Kafka.Addr,
	})
	if err != nil {
		logger.Fatal("kafka client failed", zap.Error(err))
	}
	logger.Info("kafka reader applied successfully")

	pool, err := postgresdb.NewPool(&postgresdb.PostgresConfig{
		DBUser:     cfg.Postgres.DBUser,
		DBPassword: cfg.Postgres.DBPassword,
		DBHost:     cfg.Postgres.DBHost,
		DBPort:     cfg.Postgres.DBPort,
		DBName:     cfg.Postgres.DBName,
	})
	if err != nil {
		logger.Fatal("postgress pool failed", zap.Error(err))
	}
	logger.Info("postgress pool applied successfully")

	storefs, err := seaweedfs.NewSeaweedfs(ctx, seaweedfs.S3Config{
		Endpoint:       cfg.S3.Endpoint,
		PublicEndpoint: cfg.S3.PublicEndpoint,
		AccessKey:      cfg.S3.AccessKey,
		SecretKey:      cfg.S3.SecretKey,
		Bucket:         cfg.S3.Bucket,
		Region:         cfg.S3.Region,
	})
	if err != nil {
		logger.Fatal("seaweedfs failed", zap.Error(err))
	}
	logger.Info("seaweedfs applied successfully")

	if err := migrations.RunMigrations(pool); err != nil {
		logger.Fatal("failed run migrated", zap.Error(err))
	}
	logger.Info("migrated applied successfully")

	cache, err := redisdb.NewRedis(&redisdb.RedisConfig{
		RDBAddr: cfg.Redis.RDBCacheAddr,
		RDBPort: cfg.Redis.RDBCachePort,
		RDBDB:   cfg.Redis.RDBCacheDB,
		RDBPass: cfg.Redis.RDBPass,
	})
	if err != nil {
		logger.Fatal("redis for cache failed", zap.Error(err))
	}
	logger.Info("redis for cache applied successfully")

	rate, err := redisdb.NewRedis(&redisdb.RedisConfig{
		RDBAddr: cfg.Redis.RDBRateAddr,
		RDBPort: cfg.Redis.RDBRatePort,
		RDBDB:   cfg.Redis.RDBLimiterDB,
		RDBPass: cfg.Redis.RDBPass,
	})
	if err != nil {
		logger.Fatal("redis for rate limitter failed", zap.Error(err))
	}
	logger.Info("redis for rate limitter applied successfully")

	mongo, err := mongodb.NewMongoConnect(&mongodb.MongoConfig{
		MongoHost: cfg.Mongo.MongoHost,
		MongoPort: cfg.Mongo.MongoPort,
	}, mongoMessagerDataBase, mongoMessagerCollection)
	if err != nil {
		logger.Fatal("mongo failed", zap.Error(err))
	}
	logger.Info("mongo applied successfully")

	srv := consumers.NewServices(pool, storefs)
	cons := consumers.NewConsumers(srv, client, logger)

	go func() {
		errs := cons.Profile.Run(ctx)
		for err := range errs {
			logger.Error("cons.Profile.Run:", zap.Error(err))
		}
	}()

	manager := jwtpkg.NewManagerToken(cfg.App.Secret)
	mdlwr := middleware.NewMiddleware(rate, logger, manager)
	srvcs := handlers.NewServices(pool, cache, storefs)
	handlrs := handlers.NewHandlers(srvcs, logger, &oauth2.Config{
		RedirectURL:  cfg.WebGoogle.WebRedirectURL,
		ClientID:     cfg.WebGoogle.WebClientID,
		ClientSecret: cfg.WebGoogle.WebClientSecret,
		Scopes: []string{
			yt.YoutubeReadonlyScope,
		},
		Endpoint: google.Endpoint,
	})

	wsSrvc := ws.NewWebSockerServices(mongo, cache, pool, logger)
	ws := ws.NewWebSockerHandlers(wsSrvc, logger)

	serv := api.NewServer(&config.AppConfig{
		DebugLevel: cfg.App.DebugLevel,
		Port:       cfg.App.Port,
		Secret:     cfg.App.Secret,
	}, mdlwr, handlrs, ws)

	go func() {
		if err := serv.Run(); err != nil {
			logger.Fatal("failed start server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	kafkaread.KafkaClose(client)
	logger.Info("kafka closed")

	if err := serv.ShutDown(ctx); err != nil {
		logger.Error("failed shut down server", zap.Error(err))
	}
	logger.Info("kafka closed")

	postgresdb.ClosePool(pool)
	logger.Info("pool closed")

	cache.Close()
	rate.Close()
}
