package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
)

type WorldsService struct {
	repo  WorldsRepo
	cache WorldsCache
}

func NewWorldsService(repo WorldsRepo, cache WorldsCache) *WorldsService {
	return &WorldsService{repo: repo, cache: cache}
}

func (s *WorldsService) Upsert(ctx context.Context, req *dto.UpsertMinecraftServerDTO, userID string) (*models.MinecraftServer, error) {
	world := models.MinecraftServer{
		UserID:  userID,
		Name:    strings.TrimSpace(req.Name),
		Address: strings.ToLower(strings.TrimSpace(req.Address)),
		Port:    req.Port,
		Edition: req.Edition,
	}

	res, err := s.repo.Upsert(ctx, &world)
	if err != nil {
		return nil, fmt.Errorf("s.repo.Upsert: %w", err)
	}

	if err := s.cache.Remove(ctx, res.ID); err != nil {
		return nil, fmt.Errorf("s.cache.Remove: %w", err)
	}

	return res, nil
}

func (s *WorldsService) GetAll(ctx context.Context) ([]models.MinecraftServer, error) {
	worlds, err := s.cache.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.cache.GetAll: %w", err)
	}

	return worlds, nil
}

func (s *WorldsService) GetByID(ctx context.Context, id string) (*models.MinecraftServer, error) {
	world, err := s.cache.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.cache.GetByID: %w", err)
	}

	return world, nil
}

func (s *WorldsService) Delete(ctx context.Context, id, userID string) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("s.repo.Delete: %w", err)
	}

	if err := s.cache.Remove(ctx, id); err != nil {
		return fmt.Errorf("s.cache.Remove: %w", err)
	}

	return nil
}
