package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/videoforge/backend/pkg/errors"
)

// GeneratePresignedUploadURL creates a presigned URL for uploading an object to Storj.
// The URL will be valid for the specified duration.
func (s *storjStorage) GeneratePresignedUploadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	if key == "" {
		return "", errors.BadRequest("key cannot be empty")
	}
	if expires <= 0 {
		return "", errors.BadRequest("expires must be greater than zero")
	}

	presignClient := s3.NewPresignClient(s.client)

	// Presign a PUT request for uploading
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", errors.Internal(fmt.Sprintf("failed to generate presigned upload URL: %v", err))
	}

	return req.URL, nil
}

// GeneratePresignedDownloadURL creates a presigned URL for downloading an object from Storj.
// The URL will be valid for the specified duration.
func (s *storjStorage) GeneratePresignedDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	if key == "" {
		return "", errors.BadRequest("key cannot be empty")
	}
	if expires <= 0 {
		return "", errors.BadRequest("expires must be greater than zero")
	}

	presignClient := s3.NewPresignClient(s.client)

	// Presign a GET request for downloading
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", errors.Internal(fmt.Sprintf("failed to generate presigned download URL: %v", err))
	}

	return req.URL, nil
}

// DeleteObject removes an object from Storj storage.
func (s *storjStorage) DeleteObject(ctx context.Context, key string) error {
	if key == "" {
		return errors.BadRequest("key cannot be empty")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if ok := isNoSuchKeyError(err, noSuchKey); ok {
			return errors.NotFound(fmt.Sprintf("object %q not found", key))
		}
		return errors.Internal(fmt.Sprintf("failed to delete object: %v", err))
	}

	return nil
}

// ObjectExists checks if an object exists in the Storj bucket.
func (s *storjStorage) ObjectExists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, errors.BadRequest("key cannot be empty")
	}

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if ok := isNoSuchKeyError(err, noSuchKey); ok {
			return false, nil
		}
		return false, errors.Internal(fmt.Sprintf("failed to check object existence: %v", err))
	}

	return true, nil
}

// GetObject retrieves an object from Storj storage.
// Returns an io.ReadCloser that must be closed by the caller.
func (s *storjStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if key == "" {
		return nil, errors.BadRequest("key cannot be empty")
	}

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if ok := isNoSuchKeyError(err, noSuchKey); ok {
			return nil, errors.NotFound(fmt.Sprintf("object %q not found", key))
		}
		return nil, errors.Internal(fmt.Sprintf("failed to get object: %v", err))
	}

	return resp.Body, nil
}

// isNoSuchKeyError checks if the error is a NoSuchKey error.
func isNoSuchKeyError(err error, _ *types.NoSuchKey) bool {
	if err == nil {
		return false
	}

	// Try to check if it's a 404 error
	errStr := err.Error()
	return bytes.Contains([]byte(errStr), []byte("NoSuchKey")) ||
		bytes.Contains([]byte(errStr), []byte("not found")) ||
		bytes.Contains([]byte(errStr), []byte("404"))
}