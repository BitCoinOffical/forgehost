package consumers

import (
	"context"

	"github.com/BitCoinOffical/forgehost/social-service/internal/adapters/seaweedfs"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo/postgres"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type ProfileService interface {
	GetProfileByID(ctx context.Context, id string) (*models.ProfileResponse, error)
	SaveProfile(ctx context.Context, event *dto.UserProfileDTO) error
	UpdateProfile(ctx context.Context, req *dto.UpdateProfileDTO, id string) (*models.Profile, error)
	GetSubscriptions(ctx context.Context, id string) ([]models.Subscriptions, error)
	GetSubscribers(ctx context.Context, id string) ([]models.Subscribes, error)
	Subscribe(ctx context.Context, userId, targetId string) error
	UnSubscribe(ctx context.Context, userId, targetId string) error
	Report(ctx context.Context, userId, targetId string, req *dto.ReportDTO) error
	Block(ctx context.Context, userId, targetId string) error
}

type Services struct {
	service ProfileService
}

func NewServices(pool *pgxpool.Pool, storefs *seaweedfs.Storage) *Services {
	repo := postgres.NewProfileRepo(pool)
	service := services.NewProfileService(repo, storefs.PublicEndpoint)

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
