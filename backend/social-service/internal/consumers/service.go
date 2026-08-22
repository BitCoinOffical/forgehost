package consumers

import (
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Services struct {
	service *services.ProfileService
}

func NewServices(pool *pgxpool.Pool) *Services {
	repo := repo.NewProfileRepo(pool)
	service := services.NewProfileService(repo)

	return &Services{
		service: service,
	}
}

type Consumers struct {
	Profile *Consumer
}

func NewConsumers(srv *Services, reader *kafka.Reader, logger *zap.Logger) *Consumers {

	prof := NewConsumer(reader, logger, srv.service)
	return &Consumers{Profile: prof}
}
