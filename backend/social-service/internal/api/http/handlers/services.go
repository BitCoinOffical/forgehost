package handlers

import (
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Services struct {
	profile *services.ProfileService
}

func NewServices(pool *pgxpool.Pool) *Services {
	repo := repo.NewProfileRepo(pool)
	profile := services.NewProfileService(repo)
	return &Services{profile: profile}
}

type Handlers struct {
	Profile *ProfileHandler
}

func NewHandlers(prof *services.ProfileService, logger *zap.Logger) *Handlers {
	profile := NewProfileHandler(prof, logger)
	return &Handlers{Profile: profile}
}
