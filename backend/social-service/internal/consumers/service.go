package consumers

import (
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
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

func NewConsumers(srv *Services, client *kgo.Client, logger *zap.Logger) *Consumers {

	prof := NewConsumer(client, logger, srv.service)
	return &Consumers{Profile: prof}
}
