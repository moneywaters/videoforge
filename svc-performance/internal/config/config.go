package config

import (
	"os"

	"github.com/videoforge/backend/pkg/config"
)

type Config struct {
	Server   config.ServerConfig   `mapstructure:"server"`
	Database config.DatabaseConfig `mapstructure:"database"`
	NATS     config.NATSConfig     `mapstructure:"nats"`
	Logger   config.LoggerConfig    `mapstructure:"logger"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := config.LoadFromEnv("PERFORMANCE_", &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Database.Host == "" {
		cfg.Database.Host = os.Getenv("DATABASE_HOST")
	}
	if cfg.Database.Port == "" {
		cfg.Database.Port = "5432"
	}
	if cfg.NATS.URL == "" {
		cfg.NATS.URL = os.Getenv("NATS_URL")
	}

	return &cfg, nil
}