package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/videoforge/backend/pkg/storage"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string `envconfig:"PORT" default:"8080"`
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
}

// Config holds all configuration for the Video service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
}

// StorageConfig holds S3-compatible storage configuration (Storj)
type StorageConfig struct {
	Enabled    bool
	Client    storage.Storage
	AccessKey string `envconfig:"STORJ_ACCESS_KEY"`
	SecretKey string `envconfig:"STORJ_SECRET_KEY"`
	Endpoint string `envconfig:"STORJ_ENDPOINT" default:"https://gateway.storjshare.io"`
	Region    string `envconfig:"STORJ_REGION" default:"us-east-1"`
	Bucket   string `envconfig:"STORJ_BUCKET"`
	// PublicURL is an optional public CDN URL prefix for thumbnails/public assets
	PublicURL string `envconfig:"STORJ_PUBLIC_URL"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Pool *pgxpool.Pool
	URL  string
}

// Load loads configuration from environment
func Load() (*Config, error) {
	// Try to load .env file for local development
	_ = godotenv.Load()

	serverCfg := &ServerConfig{}
	err := envconfig.Process("server", serverCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	// Priority: NEON_DATABASE_URL_* > DATABASE_URL > local postgres
	dbURL := os.Getenv("DATABASE_URL_VIDEO")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable"
	}

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Check if this is a Neon connection (serverless PostgreSQL)
	isNeon := strings.Contains(dbURL, "neon.tech") || strings.Contains(dbURL, "ep-")
	if isNeon {
		// Neon free tier: max 10 connections
		poolConfig.MaxConns = 10
		poolConfig.MinConns = 2
		poolConfig.MaxConnLifetime = time.Hour
		poolConfig.MaxConnIdleTime = 30 * time.Minute
		poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second
	} else {
		poolConfig.MaxConns = 25
		poolConfig.MinConns = 5
		poolConfig.MaxConnLifetime = time.Hour
		poolConfig.MaxConnIdleTime = 30 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	log.Println("Database pool created successfully")

	// Load storage configuration
	storageCfg := loadStorageConfig()

	config := &Config{
		Server:   *serverCfg,
		Database: DatabaseConfig{Pool: pool, URL: dbURL},
		Storage:  storageCfg,
	}

	return config, nil
}

// loadStorageConfig loads storage configuration from environment
func loadStorageConfig() StorageConfig {
	cfg := StorageConfig{}

	// Try to load storage config
	_ = envconfig.Process("storj", &cfg)

	// Check if we have required credentials
	if cfg.AccessKey != "" && cfg.SecretKey != "" && cfg.Bucket != "" {
		cfg.Enabled = true

		// Create storage client
		storjConfig := storage.StorjConfig{
			AccessKey: cfg.AccessKey,
			SecretKey: cfg.SecretKey,
			Endpoint:  cfg.Endpoint,
			Region:    cfg.Region,
			Bucket:    cfg.Bucket,
			PublicURL: cfg.PublicURL,
		}

		client, err := storage.NewStorage(storjConfig)
		if err != nil {
			log.Printf("Warning: failed to create storage client: %v", err)
			cfg.Enabled = false
		} else {
			cfg.Client = client
			log.Println("Storage client initialized successfully")
		}
	} else {
		log.Println("Storage credentials not configured, using mock storage")
		cfg.Enabled = false
	}

	return cfg
}

// Close closes all resources
func (c *Config) Close() error {
	if c.Database.Pool != nil {
		c.Database.Pool.Close()
	}
	return nil
}

// GetAddress returns the server address
func (s *ServerConfig) GetAddress() string {
	return fmt.Sprintf(":%s", s.Port)
}