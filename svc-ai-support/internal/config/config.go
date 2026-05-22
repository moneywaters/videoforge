package config

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/videoforge/backend/pkg/config"
)

const (
	// Default role for new users
	DefaultRole = "client"
	DefaultStatus = "active"
)

// Config holds all configuration for the service
type Config struct {
	Server   config.ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	NATS     NATSConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Pool *pgxpool.Pool
	URL  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	PublicKey *rsa.PublicKey
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL string
}

// Load loads configuration from environment
func Load() (*Config, error) {
	serverCfg, err := config.Load("")
	if err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	serverConfig := config.ServerConfig{
		Port:        serverCfg.Port,
		Environment: serverCfg.Environment,
	}

	// Priority: NEON_DATABASE_URL_* > DATABASE_URL > local postgres
	dbURL := os.Getenv("DATABASE_URL_AI_SUPPORT")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable"
	}

	// Database pool configuration
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

	// Load JWT public key
	jwtConfig, err := loadJWTKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to load JWT keys: %w", err)
	}

	// Get NATS configuration
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	return &Config{
		Server: serverConfig,
		Database: DatabaseConfig{
			Pool: pool,
			URL:  dbURL,
		},
		JWT: jwtConfig,
		NATS: NATSConfig{
			URL: natsURL,
		},
	}, nil
}

// loadJWTKeys loads or generates RSA keys for JWT signing
func loadJWTKeys() (JWTConfig, error) {
	publicKeyPEM := os.Getenv("JWT_PUBLIC_KEY")

	var publicKey *rsa.PublicKey

	if publicKeyPEM != "" {
		// Parse provided public key
		block, _ := pem.Decode([]byte(publicKeyPEM))
		if block == nil || block.Type != "RSA PUBLIC KEY" {
			return JWTConfig{}, fmt.Errorf("invalid public key PEM")
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return JWTConfig{}, fmt.Errorf("failed to parse public key: %w", err)
		}
		publicKeyRSA, ok := pub.(*rsa.PublicKey)
		if !ok {
			return JWTConfig{}, fmt.Errorf("public key is not RSA")
		}
		publicKey = publicKeyRSA
	} else {
		// Generate new keys for development
		privateKey, err := generateRSAKey(2048)
		if err != nil {
			return JWTConfig{}, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		publicKey = &privateKey.PublicKey
	}

	return JWTConfig{
		PublicKey: publicKey,
	}, nil
}

// generateRSAKey generates an RSA key pair
func generateRSAKey(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	if err := privateKey.Validate(); err != nil {
		return nil, err
	}
	return privateKey, nil
}

// Close closes all resources
func (c *Config) Close() error {
	if c.Database.Pool != nil {
		c.Database.Pool.Close()
	}
	return nil
}