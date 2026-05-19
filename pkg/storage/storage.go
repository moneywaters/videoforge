// Package storage provides S3-compatible object storage integration for Storj.
//
// This package implements a Storage interface that can be used with any S3-compatible
// backend (e.g., Storj Gateway, AWS S3, MinIO). It uses AWS SDK for Go v2.
package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/videoforge/backend/pkg/errors"
)

// Storage defines the interface for S3-compatible object storage operations.
// This interface can be mocked for testing.
type Storage interface {
	// GeneratePresignedUploadURL creates a presigned URL for uploading an object.
	// The URL will be valid for the specified duration.
	GeneratePresignedUploadURL(ctx context.Context, key string, expires time.Duration) (string, error)

	// GeneratePresignedDownloadURL creates a presigned URL for downloading an object.
	// The URL will be valid for the specified duration.
	GeneratePresignedDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error)

	// DeleteObject removes an object from storage.
	DeleteObject(ctx context.Context, key string) error

	// ObjectExists checks if an object exists in the bucket.
	ObjectExists(ctx context.Context, key string) (bool, error)

	// GetObject retrieves an object from storage.
	// Returns an io.ReadCloser that must be closed by the caller.
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)

	// FileExists is an alias for ObjectExists for compatibility.
	FileExists(ctx context.Context, key string) (bool, error)
}

// StorjConfig holds the configuration for connecting to Storj S3-compatible storage.
type StorjConfig struct {
	AccessKey string `envconfig:"STORJ_ACCESS_KEY" required:"true"`
	SecretKey string `envconfig:"STORJ_SECRET_KEY" required:"true"`
	Endpoint  string `envconfig:"STORJ_ENDPOINT" default:"https://gateway.storjshare.io"`
	Region    string `envconfig:"STORJ_REGION" default:"us-east-1"`
	Bucket   string `envconfig:"STORJ_BUCKET" required:"true"`
	// PublicURL is an optional public CDN URL prefix.
	// If set, GetPublicURL will return URLs pointing to this CDN instead of the gateway.
	PublicURL string `envconfig:"STORJ_PUBLIC_URL"`
}

// storjStorage implements the Storage interface for Storj S3-compatible storage.
type storjStorage struct {
	client *s3.Client
	config StorjConfig
}

// NewStorage creates a new Storage instance configured with the provided StorjConfig.
// It establishes a connection to the Storj S3-compatible gateway.
func NewStorage(cfg StorjConfig) (Storage, error) {
	// Validate required configuration
	if cfg.AccessKey == "" {
		return nil, errors.NewProblem(http.StatusBadRequest, "Bad Request", "STORJ_ACCESS_KEY is required")
	}
	if cfg.SecretKey == "" {
		return nil, errors.NewProblem(http.StatusBadRequest, "Bad Request", "STORJ_SECRET_KEY is required")
	}
	if cfg.Bucket == "" {
		return nil, errors.NewProblem(http.StatusBadRequest, "Bad Request", "STORJ_BUCKET is required")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://gateway.storjshare.io"
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	// Create AWS SDK v2 configuration
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &storjStorage{
		client: client,
		config: cfg,
	}, nil
}