package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/videoforge/backend/pkg/config"
)

type Config struct {
	Server    config.ServerConfig
	Database  DatabaseConfig
	Shopify   ShopifyConfig
}

type DatabaseConfig struct {
	Pool *pgxpool.Pool
	URL  string
}

type ShopifyConfig struct {
	APIKey       string
	APISecret    string
	AccessToken  string
	StoreDomain  string
	APIVersion  string
}

func Load() (*Config, error) {
	serverConfig, err := config.Load()
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

	shopifyCfg := ShopifyConfig{
		APIKey:      os.Getenv("SHOPIFY_API_KEY"),
		APISecret:   os.Getenv("SHOPIFY_API_SECRET"),
		AccessToken: os.Getenv("SHOPIFY_ACCESS_TOKEN"),
		StoreDomain: os.Getenv("SHOPIFY_STORE_DOMAIN"),
		APIVersion:  "2024-01",
	}

	return &Config{
		Server:    serverConfig,
		Database: DatabaseConfig{Pool: pool, URL: dbURL},
		Shopify:  shopifyCfg,
	}, nil
}

func (c *Config) Close() error {
	if c.Database.Pool != nil {
		c.Database.Pool.Close()
	}
	return nil
}