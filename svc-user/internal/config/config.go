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
	"github.com/videoforge/backend/pkg/auth"
	"github.com/videoforge/backend/pkg/config"
)

const (
	// Default role for new users
	DefaultRole = "client"
	DefaultStatus = "active"

	// JWT settings
	AccessTokenExpiry  = 15 * time.Minute
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

// Config holds all configuration for the service
type Config struct {
	Server   config.ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	NeonAuth *auth.NeonAuthClient
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Pool *pgxpool.Pool
	URL  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	KeyID      string
}

// Load loads configuration from environment
func Load() (*Config, error) {
	serverConfig, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	// Get database URL - prefer Neon-specific URL if set, otherwise fallback to DATABASE_URL or local
	// Priority: NEON_DATABASE_URL_* > DATABASE_URL > local postgres
	dbURL := os.Getenv("DATABASE_URL_USER")
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
	// Neon has lower connection limits and requires different pool settings
	isNeon := strings.Contains(dbURL, "neon.tech") || strings.Contains(dbURL, "ep-")
	if isNeon {
		// Neon free tier: max 10 connections
		poolConfig.MaxConns = 10
		poolConfig.MinConns = 2
		poolConfig.MaxConnLifetime = time.Hour
		poolConfig.MaxConnIdleTime = 30 * time.Minute
		poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second
	} else {
		// Local/traditional PostgreSQL
		poolConfig.MaxConns = 25
		poolConfig.MinConns = 5
		poolConfig.MaxConnLifetime = time.Hour
		poolConfig.MaxConnIdleTime = 30 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Load JWT keys
	jwtConfig, err := loadJWTKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to load JWT keys: %w", err)
	}

	// Initialize Neon Auth client (optional)
	neonAPIKey := os.Getenv("NEON_API_KEY")
	neonProjectID := os.Getenv("NEON_PROJECT_ID")
	neonBranchID := os.Getenv("NEON_BRANCH_ID")
	var neonAuthClient *auth.NeonAuthClient
	if neonAPIKey != "" && neonProjectID != "" && neonBranchID != "" {
		neonAuthConfig := auth.NeonAuthConfig{
			APIKey:       neonAPIKey,
			ProjectID:    neonProjectID,
			BranchID:     neonBranchID,
			CookieSecret: os.Getenv("NEON_AUTH_COOKIE_SECRET"),
		}
		neonAuthClient = auth.NewNeonAuthClient(neonAuthConfig)
		fmt.Println("Neon Auth configured")
	}

	return &Config{
		Server:   serverConfig,
		Database: DatabaseConfig{
			Pool: pool,
			URL:  dbURL,
		},
		JWT:      jwtConfig,
		NeonAuth: neonAuthClient,
	}, nil
}

// loadJWTKeys loads or generates RSA keys for JWT signing
func loadJWTKeys() (JWTConfig, error) {
	privateKeyPEM := os.Getenv("JWT_PRIVATE_KEY")
	publicKeyPEM := os.Getenv("JWT_PUBLIC_KEY")

	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey

	if privateKeyPEM != "" && publicKeyPEM != "" {
		// Parse provided keys
		block, _ := pem.Decode([]byte(privateKeyPEM))
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			return JWTConfig{}, fmt.Errorf("invalid private key PEM")
		}
		var err error
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return JWTConfig{}, fmt.Errorf("failed to parse private key: %w", err)
		}

		block, _ = pem.Decode([]byte(publicKeyPEM))
		if block == nil || block.Type != "RSA PUBLIC KEY" {
			return JWTConfig{}, fmt.Errorf("invalid public key PEM")
		}
		publicKey, err = x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return JWTConfig{}, fmt.Errorf("failed to parse public key: %w", err)
		}
		publicKeyRSA, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			return JWTConfig{}, fmt.Errorf("public key is not RSA")
		}
		publicKey = publicKeyRSA
	} else {
		// Generate new keys for development
		var err error
		privateKey, err = generateRSAKey(2048)
		if err != nil {
			return JWTConfig{}, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		publicKey = &privateKey.PublicKey
	}

	keyID := os.Getenv("JWT_KEY_ID")
	if keyID == "" {
		keyID = "videoforge-key-001"
	}

	return JWTConfig{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		KeyID:      keyID,
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