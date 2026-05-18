package storage

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// mockStorage is a mock implementation of Storage for testing.
type mockStorage struct {
	objects            map[string][]byte
	presignedURLs      map[string]string
	presignUploadErr   error
	presignDownloadErr error
	getObjectErr       error
	deleteErr          bool
	headErr            bool
}

// NewMockStorage creates a new mock storage for testing.
func NewMockStorage() *mockStorage {
	return &mockStorage{
		objects:       make(map[string][]byte),
		presignedURLs: make(map[string]string),
	}
}

// GeneratePresignedUploadURL implements the Storage interface.
func (m *mockStorage) GeneratePresignedUploadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	if m.presignUploadErr {
		return "", &types.S3Exception{
			Message: aws.String("failed to generate presigned URL"),
		}
	}
	url := "https://storage.example.com/bucket/" + key + "?expires=" + expires.String()
	m.presignedURLs[key] = url
	return url, nil
}

// GeneratePresignedDownloadURL implements the Storage interface.
func (m *mockStorage) GeneratePresignedDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	if m.presignDownloadErr {
		return "", &types.S3Exception{
			Message: aws.String("failed to generate presigned URL"),
		}
	}
	url := "https://storage.example.com/bucket/" + key + "?expires=" + expires.String()
	m.presignedURLs[key] = url
	return url, nil
}

// DeleteObject implements the Storage interface.
func (m *mockStorage) DeleteObject(ctx context.Context, key string) error {
	if m.deleteErr {
		return &types.NoSuchKey{
			Bucket: aws.String("bucket"),
			Key:    aws.String(key),
		}
	}
	delete(m.objects, key)
	return nil
}

// ObjectExists implements the Storage interface.
func (m *mockStorage) ObjectExists(ctx context.Context, key string) (bool, error) {
	if m.headErr {
		return false, nil
	}
	_, exists := m.objects[key]
	return exists, nil
}

// GetObject implements the Storage interface.
func (m *mockStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.getObjectErr {
		return nil, &types.NoSuchKey{
			Bucket: aws.String("bucket"),
			Key:    aws.String(key),
		}
	}
	data, exists := m.objects[key]
	if !exists {
		return nil, &types.NoSuchKey{
			Bucket: aws.String("bucket"),
			Key:    aws.String(key),
		}
	}
	return &mockReader{data: data}, nil
}

// mockReader implements io.ReadCloser for testing.
type mockReader struct {
	data     []byte
	pos      int
	closed   bool
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	if r.closed {
		return 0, io.EOF
	}
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *mockReader) Close() error {
	r.closed = true
	return nil
}

// TestGenerateKey tests the GenerateKey helper function.
func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		id       string
		filename string
		want     string
	}{
		{
			name:     "standard video key",
			prefix:   "videos",
			id:       "550e8400-e29b-41d4-a716-446655440000",
			filename: "raw.mp4",
			want:     "videos/550e8400-e29b-41d4-a716-446655440000/raw.mp4",
		},
		{
			name:     "thumbnail key",
			prefix:   "thumbnails",
			id:       "abc123",
			filename: "thumbnail.jpg",
			want:     "thumbnails/abc123/thumbnail.jpg",
		},
		{
			name:     "auto-generate UUID when ID is empty",
			prefix:   "videos",
			id:       "",
			filename: "raw.mp4",
			want:     "", // Will be random UUID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateKey(tt.prefix, tt.id, tt.filename)
			if tt.id == "" {
				// Check that we get a valid UUID after the prefix
				if !strings.HasPrefix(got, tt.prefix+"/") {
					t.Errorf("GenerateKey() = %v, want prefix %s/", got, tt.prefix)
				}
				// Check format: prefix/uuid/filename
				parts := strings.Split(got, "/")
				if len(parts) != 3 {
					t.Errorf("GenerateKey() = %v, want format prefix/uuid/filename", got)
				}
				// Verify UUID format
				if _, err := uuid.Parse(parts[1]); err != nil {
					t.Errorf("GenerateKey() generated invalid UUID: %v", err)
				}
				if parts[2] != tt.filename {
					t.Errorf("GenerateKey() = %v, want filename %s", got, tt.filename)
				}
			} else {
				if got != tt.want {
					t.Errorf("GenerateKey() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestGetPublicURL tests the GetPublicURL helper function.
func TestGetPublicURL(t *testing.T) {
	tests := []struct {
		name    string
		config  StorjConfig
		key     string
		want    string
	}{
		{
			name: "with public URL",
			config: StorjConfig{
				PublicURL: "https://cdn.example.com",
				Bucket:    "my-bucket",
			},
			key:  "videos/abc123/raw.mp4",
			want: "https://cdn.example.com/videos/abc123/raw.mp4",
		},
		{
			name: "without public URL uses endpoint",
			config: StorjConfig{
				Endpoint: "https://gateway.storjshare.io",
				Bucket:   "my-bucket",
			},
			key:  "videos/abc123/raw.mp4",
			want: "https://gateway.storjshare.io/my-bucket/videos/abc123/raw.mp4",
		},
		{
			name: "empty config with endpoint",
			config: StorjConfig{
				Endpoint: "https://gateway.storjshare.io",
				Bucket:   "my-bucket",
			},
			key:  "videos/abc123/raw.mp4",
			want: "https://gateway.storjshare.io/my-bucket/videos/abc123/raw.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPublicURL(tt.config, tt.key)
			if got != tt.want {
				t.Errorf("GetPublicURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetVideoKey tests the GetVideoKey helper function.
func TestGetVideoKey(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "raw format",
			format: "raw",
			want:   "raw.mp4",
		},
		{
			name:   "processed format",
			format: "processed",
			want:   "processed.mp4",
		},
		{
			name:   "thumbnail format",
			format: "thumbnail",
			want:   "thumbnail.jpg",
		},
		{
			name:   "unknown format passes through",
			format: "custom.mp4",
			want:   "custom.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			videoID := "test-video-id"
			got := GetVideoKey(videoID, tt.format)
			want := "videos/" + videoID + "/" + tt.want
			if got != want {
				t.Errorf("GetVideoKey() = %v, want %v", got, want)
			}
		})
	}
}

// TestGetRawKey tests the GetRawKey helper function.
func TestGetRawKey(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		want  string
	}{
		{
			name:  "simple key",
			key:   "video.mp4",
			want:  "video.mp4",
		},
		{
			name:  "nested key",
			key:   "videos/abc123/raw.mp4",
			want:  "raw.mp4",
		},
		{
			name:  "deeply nested key",
			key:   "a/b/c/d/file.txt",
			want:  "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRawKey(tt.key)
			if got != tt.want {
				t.Errorf("GetRawKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStorjConfig tests the StorjConfig struct.
func TestStorjConfig(t *testing.T) {
	cfg := StorjConfig{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Endpoint:  "https://gateway.storjshare.io",
		Region:    "us-east-1",
		Bucket:    "test-bucket",
		PublicURL: "https://cdn.example.com",
	}

	if cfg.AccessKey != "test-access-key" {
		t.Errorf("AccessKey = %v, want test-access-key", cfg.AccessKey)
	}
	if cfg.SecretKey != "test-secret-key" {
		t.Errorf("SecretKey = %v, want test-secret-key", cfg.SecretKey)
	}
	if cfg.Endpoint != "https://gateway.storjshare.io" {
		t.Errorf("Endpoint = %v, want https://gateway.storjshare.io", cfg.Endpoint)
	}
	if cfg.Region != "us-east-1" {
		t.Errorf("Region = %v, want us-east-1", cfg.Region)
	}
	if cfg.Bucket != "test-bucket" {
		t.Errorf("Bucket = %v, want test-bucket", cfg.Bucket)
	}
	if cfg.PublicURL != "https://cdn.example.com" {
		t.Errorf("PublicURL = %v, want https://cdn.example.com", cfg.PublicURL)
	}
}

// TestMockStorage tests the mock storage implementation.
func TestMockStorage(t *testing.T) {
	ctx := context.Background()
	storage := NewMockStorage()

	// Test ObjectExists for non-existent object
	exists, err := storage.ObjectExists(ctx, "test-key")
	if err != nil {
		t.Errorf("ObjectExists() error = %v", err)
	}
	if exists {
		t.Error("ObjectExists() = true, want false")
	}

	// Add an object
	storage.objects["test-key"] = []byte("test data")

	// Test ObjectExists for existing object
	exists, err = storage.ObjectExists(ctx, "test-key")
	if err != nil {
		t.Errorf("ObjectExists() error = %v", err)
	}
	if !exists {
		t.Error("ObjectExists() = false, want true")
	}

	// Test GetObject
	reader, err := storage.GetObject(ctx, "test-key")
	if err != nil {
		t.Errorf("GetObject() error = %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("ReadAll() error = %v", err)
	}
	if string(data) != "test data" {
		t.Errorf("GetObject() = %v, want test data", string(data))
	}
	reader.Close()

	// Test GeneratePresignedUploadURL
	url, err := storage.GeneratePresignedUploadURL(ctx, "test-key", time.Hour)
	if err != nil {
		t.Errorf("GeneratePresignedUploadURL() error = %v", err)
	}
	if !strings.Contains(url, "test-key") {
		t.Errorf("GeneratePresignedUploadURL() = %v, want to contain test-key", url)
	}

	// Test GeneratePresignedDownloadURL
	url, err = storage.GeneratePresignedDownloadURL(ctx, "test-key", time.Hour)
	if err != nil {
		t.Errorf("GeneratePresignedDownloadURL() error = %v", err)
	}
	if !strings.Contains(url, "test-key") {
		t.Errorf("GeneratePresignedDownloadURL() = %v, want to contain test-key", url)
	}

	// Test DeleteObject
	err = storage.DeleteObject(ctx, "test-key")
	if err != nil {
		t.Errorf("DeleteObject() error = %v", err)
	}

	// Verify object was deleted
	exists, err = storage.ObjectExists(ctx, "test-key")
	if err != nil {
		t.Errorf("ObjectExists() error = %v", err)
	}
	if exists {
		t.Error("ObjectExists() = true, want false after deletion")
	}
}

// TestMockStorageErrors tests error handling in mock storage.
func TestMockStorageErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("GetObject non-existent", func(t *testing.T) {
		storage := NewMockStorage()
		_, err := storage.GetObject(ctx, "non-existent")
		if err == nil {
			t.Error("GetObject() error = nil, want error for non-existent object")
		}
	})

	t.Run("DeleteObject non-existent", func(t *testing.T) {
		storage := NewMockStorage()
		err := storage.DeleteObject(ctx, "non-existent")
		if err != nil {
			t.Errorf("DeleteObject() error = %v, want nil", err)
		}
	})
}

// TestNewStorage tests the NewStorage constructor with various configurations.
func TestNewStorage(t *testing.T) {
	tests := []struct {
		name      string
		config    StorjConfig
		wantErr  bool
		errType  string // "access_key", "secret_key", "bucket", or "" for success
	}{
		{
			name: "valid configuration",
			config: StorjConfig{
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
				Endpoint:  "https://gateway.storjshare.io",
				Region:    "us-east-1",
				Bucket:    "test-bucket",
			},
			wantErr: false,
		},
		{
			name: "missing access key",
			config: StorjConfig{
				SecretKey: "test-secret-key",
				Endpoint:  "https://gateway.storjshare.io",
				Region:    "us-east-1",
				Bucket:    "test-bucket",
			},
			wantErr: true,
			errType: "access_key",
		},
		{
			name: "missing secret key",
			config: StorjConfig{
				AccessKey: "test-access-key",
				Endpoint:  "https://gateway.storjshare.io",
				Region:    "us-east-1",
				Bucket:    "test-bucket",
			},
			wantErr: true,
			errType: "secret_key",
		},
		{
			name: "missing bucket",
			config: StorjConfig{
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
				Endpoint:  "https://gateway.storjshare.io",
				Region:    "us-east-1",
			},
			wantErr: true,
			errType: "bucket",
		},
		{
			name: "default endpoint and region",
			config: StorjConfig{
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
				Bucket:    "test-bucket",
			},
			wantErr: false,
		},
		{
			name: "with public URL",
			config: StorjConfig{
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
				Endpoint:  "https://gateway.storjshare.io",
				Region:    "us-east-1",
				Bucket:    "test-bucket",
				PublicURL: "https://cdn.example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStorage(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewStorage() error = nil, want error for %s", tt.errType)
				}
			} else {
				if err != nil {
					t.Errorf("NewStorage() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestGetThumbnailKey tests the GetThumbnailKey helper function.
func TestGetThumbnailKey(t *testing.T) {
	tests := []struct {
		name     string
		videoID  string
		want     string
	}{
		{
			name:    "standard thumbnail key",
			videoID: "550e8400-e29b-41d4-a716-446655440000",
			want:    "thumbnails/550e8400-e29b-41d4-a716-446655440000/thumbnail.jpg",
		},
		{
			name:    "short video ID",
			videoID: "abc123",
			want:    "thumbnails/abc123/thumbnail.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetThumbnailKey(tt.videoID)
			if got != tt.want {
				t.Errorf("GetThumbnailKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetTempKey tests the GetTempKey helper function.
func TestGetTempKey(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		filename string
		want     string
	}{
		{
			name:     "standard temp key",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			filename: "raw.mp4",
			want:     "temp/550e8400-e29b-41d4-a716-446655440000/raw.mp4",
		},
		{
			name:     "with different filename",
			id:      "abc123",
			filename: "preview.mp4",
			want:     "temp/abc123/preview.mp4",
		},
		{
			name:     "empty id generates UUID",
			id:       "",
			filename: "test.mp4",
			want:     "", // Will be random
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTempKey(tt.id, tt.filename)
			if tt.id == "" {
				// Check that format is correct: temp/uuid/filename
				if !strings.HasPrefix(got, "temp/") {
					t.Errorf("GetTempKey() = %v, want prefix temp/", got)
				}
				parts := strings.Split(got, "/")
				if len(parts) != 3 {
					t.Errorf("GetTempKey() = %v, want format temp/uuid/filename", got)
				}
				if _, err := uuid.Parse(parts[1]); err != nil {
					t.Errorf("GetTempKey() generated invalid UUID: %v", err)
				}
				if parts[2] != tt.filename {
					t.Errorf("GetTempKey() = %v, want filename %s", got, tt.filename)
				}
			} else {
				if got != tt.want {
					t.Errorf("GetTempKey() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}