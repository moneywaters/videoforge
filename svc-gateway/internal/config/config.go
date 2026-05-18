package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	pkgconfig "github.com/videoforge/backend/pkg/config"
)

// Config holds all configuration for the Gateway service
type Config struct {
	Environment  string
	Port        string
	Address     string
	JWTPublicKey string
	JWTKeyID    string
	Services    ServicesConfig
	RateLimit   RateLimitConfig
	NATS        NATSConfig
}

// ServicesConfig holds URLs for internal services
type ServicesConfig struct {
	AuthURL         string
	UserURL         string
	BriefURL        string
	VideoURL        string
	CampaignURL    string
	ShopifyURL     string
	PerformanceURL string
	PayoutURL     string
	NotificationURL string
	AdminURL       string
	SupportURL     string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int
	Burst         int
	Enabled       bool
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL string
}

// GetAddress returns the server address
func (c *Config) GetAddress() string {
	if c.Address == "" {
		c.Address = "0.0.0.0"
	}
	if c.Port == "" {
		c.Port = "8080"
	}
	return fmt.Sprintf("%s:%s", c.Address, c.Port)
}

// Load loads configuration from environment
func Load() (*Config, error) {
	// Try to load base config from pkg/config (optional)
	_, _ = pkgconfig.Load("") // loads .env if present

	// Environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	// Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Address
	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "0.0.0.0"
	}

	// JWT configuration
	jwtPublicKey := os.Getenv("JWT_PUBLIC_KEY")
	jwtKeyID := os.Getenv("JWT_KEY_ID")
	if jwtKeyID == "" {
		jwtKeyID = "videoforge-key-1"
	}

	// Rate limiting
	rateLimit := os.Getenv("RATE_LIMIT")
	requestsPerSecond := 100
	burst := 10
	if rateLimit != "" {
		fmt.Sscanf(rateLimit, "%d", &requestsPerSecond)
		burst = requestsPerSecond / 10
	}
	rateLimitEnabled := os.Getenv("RATE_LIMIT_ENABLED") != "false"

	// NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	// Service URLs - defaults to localhost for local development
	// In production, these would be service names in the Docker/K8s network
	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:8081"
	}

	userURL := os.Getenv("USER_SERVICE_URL")
	if userURL == "" {
		userURL = "http://localhost:8082"
	}

	briefURL := os.Getenv("BRIEF_SERVICE_URL")
	if briefURL == "" {
		briefURL = "http://localhost:8083"
	}

	videoURL := os.Getenv("VIDEO_SERVICE_URL")
	if videoURL == "" {
		videoURL = "http://localhost:8084"
	}

	campaignURL := os.Getenv("CAMPAIGN_SERVICE_URL")
	if campaignURL == "" {
		campaignURL = "http://localhost:8085"
	}

	shopifyURL := os.Getenv("SHOPIFY_SERVICE_URL")
	if shopifyURL == "" {
		shopifyURL = "http://localhost:8086"
	}

	performanceURL := os.Getenv("PERFORMANCE_SERVICE_URL")
	if performanceURL == "" {
		performanceURL = "http://localhost:8087"
	}

	payoutURL := os.Getenv("PAYOUT_SERVICE_URL")
	if payoutURL == "" {
		payoutURL = "http://localhost:8088"
	}

	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationURL == "" {
		notificationURL = "http://localhost:8089"
	}

	adminURL := os.Getenv("ADMIN_SERVICE_URL")
	if adminURL == "" {
		adminURL = "http://localhost:8090"
	}

	supportURL := os.Getenv("SUPPORT_SERVICE_URL")
	if supportURL == "" {
		supportURL = "http://localhost:8091"
	}

	return &Config{
		Environment:  env,
		Port:        port,
		Address:     address,
		JWTPublicKey: jwtPublicKey,
		JWTKeyID:    jwtKeyID,
		Services: ServicesConfig{
			AuthURL:         authURL,
			UserURL:         userURL,
			BriefURL:       briefURL,
			VideoURL:       videoURL,
			CampaignURL:     campaignURL,
			ShopifyURL:      shopifyURL,
			PerformanceURL:  performanceURL,
			PayoutURL:      payoutURL,
			NotificationURL: notificationURL,
			AdminURL:       adminURL,
			SupportURL:      supportURL,
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: requestsPerSecond,
			Burst:           burst,
			Enabled:         rateLimitEnabled,
		},
		NATS: NATSConfig{
			URL: natsURL,
		},
	}, nil
}

// ParseJWTPublicKey parses the JWT public key from PEM encoded string
func (c *Config) ParseJWTPublicKey() (*rsa.PublicKey, error) {
	if c.JWTPublicKey == "" {
		return nil, nil
	}

	block, _ := pem.Decode([]byte(c.JWTPublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}