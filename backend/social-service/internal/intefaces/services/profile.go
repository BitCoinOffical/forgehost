package services

import (
	"context"
	"fmt"

	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/dto"
	"github.com/BitCoinOffical/forgehost/social-service/internal/domain/models"
	"github.com/BitCoinOffical/forgehost/social-service/internal/intefaces/repo"
)

type ProfileService struct {
	repo *repo.ProfileRepo
}

func NewProfileService(repo *repo.ProfileRepo) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetProfileByID(ctx context.Context, id string) (*models.Profile, error) {
	resp, err := s.repo.GetProfileByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetProfileByID: %w", err)
	}

	return resp, nil
}

func (s *ProfileService) SaveProfile(ctx context.Context, event *dto.UserProfileDTO) error {
	profile := models.Profile{
		UserID: event.UserID,
	}

	if err := s.repo.SaveProfile(ctx, &profile); err != nil {
		return fmt.Errorf("s.repo.SaveUserProfile: %w", err)
	}

	return nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, req *dto.UpdateProfileDTO) (*models.Profile, error) {
	profile := models.Profile{
		UserName:  &req.UserName,
		Bio:       &req.Bio,
		AvatarUrl: &req.AvatarUrl,
	}
	prof, err := s.repo.BuildUpdateProfile(ctx, &profile)
	if err != nil {
		return nil, fmt.Errorf("s.repo.BuildUpdateProfile: %w", err)
	}

	return prof, nil
}

func (s *ProfileService) GetSubscriptions(ctx context.Context, id string) ([]models.Subscriptions, error) {
	subs, err := s.repo.GetSubscriptions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetSubscriptions: %w", err)
	}
	return subs, nil
}

func (s *ProfileService) GetSubscribers(ctx context.Context, id string) ([]models.Subscribes, error) {
	subscr, err := s.repo.GetSubscribers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetSubscribers: %w", err)
	}
	return subscr, nil
}
