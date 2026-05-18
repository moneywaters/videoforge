package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/videoforge/backend/pkg/config"
)

// Config holds all application configuration
type Config struct {
	Server   config.ServerConfig
	Database DatabaseConfig
	JWTKey   string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Pool *pgxpool.Pool
	URL  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	serverConfig, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	// Load JWT key
	jwtKey := os.Getenv("JWT_PUBLIC_KEY")
	if jwtKey == "" {
		jwtKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
		if jwtKeyPath != "" {
			if data, err := os.ReadFile(jwtKeyPath); err == nil {
				jwtKey = string(data)
			}
		}
	}

	// Priority: NEON_DATABASE_URL_* > DATABASE_URL > local postgres
	dbURL := os.Getenv("DATABASE_URL_CAMPAIGN")
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

	return &Config{
		Server:   serverConfig,
		Database: DatabaseConfig{Pool: pool, URL: dbURL},
		JWTKey:   jwtKey,
	}, nil
}

// Close closes the configuration
func (c *Config) Close() error {
	if c.Database.Pool != nil {
		c.Database.Pool.Close()
	}
	return nil
}