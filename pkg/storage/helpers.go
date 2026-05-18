package storage

import (
	"fmt"

	"github.com/google/uuid"
)

// GenerateKey generates a unique S3 key for storing objects.
// It creates keys in the format: "prefix/UUID/filename"
//
// Parameters:
//   - prefix: The directory prefix (e.g., "videos", "thumbnails")
//   - id: A unique identifier (e.g., video ID, user ID)
//   - filename: The filename to use (e.g., "raw.mp4", "thumbnail.jpg")
//
// Returns:
//   A formatted S3 key string like "videos/550e8400-e29b-41d4-a716-446655440000/raw.mp4"
func GenerateKey(prefix, id, filename string) string {
	// If ID is empty, generate a new UUID
	if id == "" {
		id = uuid.New().String()
	}

	// Format: prefix/uuid/filename
	return fmt.Sprintf("%s/%s/%s", prefix, id, filename)
}

// GetPublicURL returns the public URL for an object.
// If PublicURL is configured in StorjConfig, it returns a URL using the CDN.
// Otherwise, it returns a URL using the gateway endpoint.
//
// Parameters:
//   - config: The StorjConfig containing endpoint and PublicURL settings
//   - key: The S3 key of the object
//
// Returns:
//   A public URL string pointing to the object
func GetPublicURL(config StorjConfig, key string) string {
	// If PublicURL is set, use it as the base for the public URL
	if config.PublicURL != "" {
		return fmt.Sprintf("%s/%s", config.PublicURL, key)
	}

	// Otherwise, construct URL from endpoint and bucket
	return fmt.Sprintf("%s/%s/%s", config.Endpoint, config.Bucket, key)
}

// GetRawKey extracts just the filename from a full S3 key.
// For example, "videos/abc123/raw.mp4" returns "raw.mp4".
func GetRawKey(key string) string {
	// Find the last slash and return everything after it
	lastSlash := -1
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 {
		return key
	}

	return key[lastSlash+1:]
}

// GetVideoKey generates a standard video storage key.
// This is a convenience function for common video storage patterns.
func GetVideoKey(videoID, format string) string {
	// Common video formats
	validFormats := map[string]string{
		"raw":       "raw.mp4",
		"processed": "processed.mp4",
		"thumbnail": "thumbnail.jpg",
		"preview":  "preview.mp4",
		"hls":       "playlist.m3u8",
	}

	filename, ok := validFormats[format]
	if !ok {
		filename = format
	}

	return GenerateKey("videos", videoID, filename)
}

// GetThumbnailKey generates a thumbnail storage key.
func GetThumbnailKey(videoID string) string {
	return GenerateKey("thumbnails", videoID, "thumbnail.jpg")
}

// GetTempKey generates a temporary upload key.
func GetTempKey(id, filename string) string {
	return GenerateKey("temp", id, filename)
}