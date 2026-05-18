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

// Config holds service configuration
type Config struct {
	Server     config.ServerConfig
	Database   DatabaseConfig
	HoldPeriod int `envconfig:"HOLD_PERIOD_DAYS" default:"14"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Pool *pgxpool.Pool
	URL  string
}

// RuulConfig holds Ruul.io API configuration
type RuulConfig struct {
	APIKey string `envconfig:"RUUL_API_KEY"`
	BaseURL string `envconfig:"RUUL_BASE_URL" default:"https://api.ruul.io/v1"`
}

// DodoConfig holds DodoPayments API configuration
type DodoConfig struct {
	APIKey string `envconfig:"DODO_API_KEY"`
	Secret string `envconfig:"DODO_SECRET"`
	BaseURL string `envconfig:"DODO_BASE_URL" default:"https://api.dodopayments.com/v1"`
}

// Load loads configuration from environment
func Load() (*Config, error) {
	serverConfig, err := config.LoadServer("")
	if err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	// Priority: NEON_DATABASE_URL_* > DATABASE_URL > local postgres
	dbURL := os.Getenv("DATABASE_URL_PAYOUT")
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
	}, nil
}

// Close closes database connections
func (c *Config) Close() error {
	if c.Database.Pool != nil {
		c.Database.Pool.Close()
	}
	return nil
}

// GetHoldPeriodDuration returns hold period as duration
func (c *Config) GetHoldPeriodDuration() time.Duration {
	return time.Duration(c.HoldPeriod) * 24 * time.Hour
}