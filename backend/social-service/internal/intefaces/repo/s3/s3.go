package s3store

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/BitCoinOffical/forgehost/social-service/internal/adapters/seaweedfs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	storage *seaweedfs.Storage
}

func NewS3Loader(storage *seaweedfs.Storage) *S3 {
	return &S3{storage: storage}
}

func (s *S3) Upload(ctx context.Context, key, contentType string, body io.ReadSeeker) error {
	_, err := s.storage.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.storage.Bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("s.storage.Client.PutObject: %w", err)
	}
	return nil
}

func (s *S3) Delete(ctx context.Context, imageID string) error {
	_, err := s.storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.storage.Bucket),
		Key:    aws.String(imageID),
	})
	if err != nil {
		return fmt.Errorf("s.storage.Client.DeleteObject: %w", err)
	}
	return nil
}

func (s *S3) GetURL(ctx context.Context, imageID string, ttl time.Duration) (string, error) {
	req, err := s3.NewPresignClient(s.storage.Client).PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.storage.Bucket),
		Key:    aws.String(imageID),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("s3.NewPresignClient: %w", err)
	}
	return req.URL, nil
}
