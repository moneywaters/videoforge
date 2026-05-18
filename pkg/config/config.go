package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// ServerConfig holds server-specific config
type ServerConfig struct {
	Port        string `envconfig:"PORT" default:"8080"`
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
}

// DatabaseConfig holds database configuration with support for Neon
type DatabaseConfig struct {
	Host     string `envconfig:"DATABASE_HOST" default:"localhost"`
	Port     string `envconfig:"DATABASE_PORT" default:"5432"`
	Username string `envconfig:"DATABASE_USERNAME" default:"videoforge"`
	Password string `envconfig:"DATABASE_PASSWORD" default:"password"`
	Database string `envconfig:"DATABASE_NAME" default:"videoforge"`
	URL      string `envconfig:"DATABASE_URL"` // Full connection URL (takes precedence)
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL string `envconfig:"NATS_URL"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level string `envconfig:"LOG_LEVEL" default:"info"`
}

// Config holds all application configuration loaded from environment variables.
// It supports loading from .env file for local development.
type Config struct {
	// Server configuration
	Port string `envconfig:"PORT" default:"8080"`

	// Database configuration
	// Priority: NEON_DATABASE_URL_* > DATABASE_URL > individual params
	DatabaseURL string `envconfig:"DATABASE_URL"`

	// NATS configuration
	NatsURL string `envconfig:"NATS_URL" default:"nats://localhost:4222"`

	// Authentication
	JWTSecret string `envconfig:"JWT_SECRET"`

	// Environment
	Environment string `envconfig:"ENVIRONMENT" default:"development"`

	// Optional: App name for logging
	AppName string `envconfig:"APP_NAME" default:"videoforge"`
}

// LoadFromEnv loads configuration with a specific prefix
func LoadFromEnv(prefix string, cfg interface{}) error {
	_ = godotenv.Load()
	return envconfig.Process(prefix, cfg)
}

// GetDatabaseURL returns the effective database URL.
// It checks for service-specific Neon URLs first, then falls back to DATABASE_URL.
func GetDatabaseURL(serviceName string) string {
	// Check for service-specific Neon URL (e.g., DATABASE_URL_USER)
	if serviceName != "" {
		neonURL := os.Getenv("DATABASE_URL_" + strings.ToUpper(serviceName))
		if neonURL != "" {
			return neonURL
		}
	}
	// Fall back to standard DATABASE_URL
	return os.Getenv("DATABASE_URL")
}

// Load loads configuration from environment variables.
// It first loads from .env file if present (for local development),
// then overwrites with environment variables.
func Load(prefix string) (*Config, error) {
	// Try to load .env file for local development
	_ = godotenv.Load()

	cfg := &Config{}
	err := envconfig.Process(prefix, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// For development, use default JWT secret if not set
	if cfg.JWTSecret == "" {
		log.Println("WARNING: JWT_SECRET not set, using insecure default for development")
		cfg.JWTSecret = "insecure-dev-secret-change-in-production"
	}

	return cfg, nil
}

// LoadServer loads just server config
func LoadServer(prefix string) (*ServerConfig, error) {
	// Try to load .env file for local development
	_ = godotenv.Load()

	cfg := &ServerConfig{}
	err := envconfig.Process(prefix, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	return cfg, nil
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if the environment is development.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// GetDSN returns the database DSN from the connection URL.
// This is useful for tools that require a DSN format.
func (c *Config) GetDSN() string {
	return c.DatabaseURL
}

// GetAddress returns the server address in the format host:port.
func (c *Config) GetAddress() string {
	return fmt.Sprintf(":%s", c.Port)
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}