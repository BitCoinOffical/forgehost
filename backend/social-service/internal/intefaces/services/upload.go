package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/google/uuid"
)

type UploaderService struct {
	repo           UploadRepo
	store          S3
	PublicEndpoint string
}

func NewUploaderService(repo UploadRepo, store S3, PublicEndpoint string) *UploaderService {
	return &UploaderService{repo: repo, store: store, PublicEndpoint: PublicEndpoint}
}

func (s *UploaderService) UploadImage(ctx context.Context, prefix, userID, ext, contentType string, file multipart.File) (string, error) {
	imageID := prefix + uuid.NewString() + ext

	if err := s.store.Upload(ctx, imageID, contentType, file); err != nil {
		return "", fmt.Errorf("s.store.Upload: %w", err)
	}

	if err := s.repo.CreateUpload(ctx, imageID, userID); err != nil {
		createErr := fmt.Errorf("s.repo.CreateUpload: %w", err)
		if delErr := s.store.Delete(ctx, imageID); delErr != nil {
			return "", errors.Join(createErr, fmt.Errorf("s.store.Delete: %w", delErr))
		}
		return "", createErr
	}

	return imageID, nil
}
