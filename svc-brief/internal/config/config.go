package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string `envconfig:"PORT" default:"8080"`
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
}

// Config holds all configuration for the Brief service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
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

	cfg := &ServerConfig{}
	err := envconfig.Process("server", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable"
	}

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	log.Println("Database pool created successfully")

	return &Config{
		Server:   *cfg,
		Database: DatabaseConfig{Pool: pool, URL: dbURL},
	}, nil
}

// Close closes all resources
func (c *Config) Close() error {
	if c.Database.Pool != nil {
		c.Database.Pool.Close()
	}
	return nil
}