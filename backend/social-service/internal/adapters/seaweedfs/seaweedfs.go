package seaweedfs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	Bucket         string
	Region         string
}

type Storage struct {
	Client         *s3.Client
	Bucket         string
	PublicEndpoint string
}

func NewSeaweedfs(ctx context.Context, cfg S3Config) (*Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))

	if err != nil {
		return nil, fmt.Errorf("config.LoadDefaultConfig: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &Storage{Client: client, Bucket: cfg.Bucket, PublicEndpoint: cfg.PublicEndpoint}, nil
}
