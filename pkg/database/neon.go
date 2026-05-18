package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NeonConfig holds configuration for Neon serverless PostgreSQL connections.
// Neon uses serverless connection pooling with specific optimizations.
type NeonConfig struct {
	// ConnectionString is the Neon PostgreSQL connection string.
	// Format: postgres://user:password@host/database?sslmode=require
	ConnectionString string

	// MaxConns is the maximum number of connections in the pool.
	// Neon free tier has a limit of 10 connections.
	MaxConns int32

	// MinConns is the minimum number of connections to maintain.
	MinConns int32

	// MaxConnLifetime is the maximum lifetime of a connection.
	MaxConnLifetime time.Duration

	// MaxConnIdleTime is the maximum idle time of a connection.
	MaxConnIdleTime time.Duration

	// ConnectTimeout is the timeout for establishing a connection.
	ConnectTimeout time.Duration

	// QueryTimeout is the timeout for queries.
	QueryTimeout time.Duration

	// HealthCheckPeriod is the period for health checks.
	HealthCheckPeriod time.Duration
}

// DefaultNeonConfig returns default configuration optimized for Neon serverless.
func DefaultNeonConfig(connString string) *NeonConfig {
	return &NeonConfig{
		ConnectionString:   connString,
		MaxConns:          10,              // Neon free tier limit
		MinConns:          2,              // Keep some connections warm
		MaxConnLifetime:   time.Hour,      // Connections survive 1 hour
		MaxConnIdleTime:   30 * time.Minute, // Idle connections timeout
		ConnectTimeout:    10 * time.Second, // Connection establishment timeout
		QueryTimeout:      30 * time.Second, // Query timeout
		HealthCheckPeriod: 5 * time.Minute, // Health check period
	}
}

// NewNeonPool creates a connection pool optimized for Neon serverless PostgreSQL.
// This is the main entry point for creating Neon database connections.
func NewNeonPool(connString string) (*pgxpool.Pool, error) {
	return NewNeonPoolWithConfig(DefaultNeonConfig(connString))
}

// NewNeonPoolWithConfig creates a connection pool with the given Neon configuration.
func NewNeonPoolWithConfig(config *NeonConfig) (*pgxpool.Pool, error) {
	if config == nil {
		return nil, fmt.Errorf("NeonConfig is required")
	}

	// Parse the connection string
	poolConfig, err := pgxpool.ParseConfig(config.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Apply Neon-specific optimizations
	// These settings are tuned for Neon serverless PostgreSQL
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime

	// Set connection timeout
	poolConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	// Set statement timeout for serverless (prevents long-running queries)
	// Note: pgx doesn't support runtime params via config, so we set it per-session via AfterConnect
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// Set statement timeout
		_, err := conn.Exec(ctx, fmt.Sprintf("SET statement_timeout = '%s'", config.QueryTimeout))
		return err
	}

	// Create the pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Verify the connection works
	if err := pool.Ping(context.Background()); err != nil {
		pool.Dispose()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// NeonHealthCheckResult represents the result of a Neon health check.
type NeonHealthCheckResult struct {
	Healthy         bool          `json:"healthy"`
	Latency         time.Duration `json:"latency"`
	MaxConnections  int32         `json:"max_connections"`
	AvailableConns int32         `json:"available_connections"`
	ProjectID      string        `json:"project_id,omitempty"`
	DatabaseName   string        `json:"database_name,omitempty"`
	Error          string        `json:"error,omitempty"`
}

// CheckNeonHealth performs a health check on the Neon connection.
// It returns a NeonHealthCheckResult with details about the connection health.
func CheckNeonHealth(ctx context.Context, pool *pgxpool.Pool) (*NeonHealthCheckResult, error) {
	result := &NeonHealthCheckResult{
		Healthy: false,
	}

	start := time.Now()

	// Perform a simple query to check connectivity
	err := pool.Ping(ctx)
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}

	result.Latency = time.Since(start)
	result.Healthy = true

	// Get pool stats
	stats := pool.Stat()
	result.MaxConnections = stats.MaxConns()
	result.AvailableConns = stats.AvailableConns()

	// Try to extract project/database info from config
	if pool.Config() != nil {
		connStr := pool.Config().ConnString()
		parsed, err := url.Parse(connStr)
		if err == nil && parsed != nil {
			result.DatabaseName = strings.TrimPrefix(parsed.Path, "/")
		}
	}

	return result, nil
}

// ParseNeonConnectionString parses a Neon connection string and returns its components.
// This is useful for extracting metadata like project ID, database name, etc.
func ParseNeonConnectionString(connString string) (map[string]string, error) {
	result := make(map[string]string)

	// Parse as URL
	parsed, err := url.Parse(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Extract components
	result["scheme"] = parsed.Scheme
	result["host"] = parsed.Host
	result["database"] = strings.TrimPrefix(parsed.Path, "/")

	// Parse query parameters
	if parsed.RawQuery != "" {
		queryParams, err := url.ParseQuery(parsed.RawQuery)
		if err == nil {
			result["sslmode"] = queryParams.Get("sslmode")
			result["connect_timeout"] = queryParams.Get("connect_timeout")
		}
	}

	// Extract user info (contains username)
	if parsed.User != nil {
		result["username"] = parsed.User.Username()
	}

	return result, nil
}

// IsNeonConnectionString checks if a connection string appears to be a Neon connection.
func IsNeonConnectionString(connString string) bool {
	// Neon connections typically have these characteristics:
	// - Use postgres:// protocol
	// - Host ends with .neon.tech or contains ep- prefix (endpoint)
	// - Have sslmode=require
	//
	// For now, we check if it contains "neon.tech" or has the ep- host pattern
	return strings.Contains(connString, "neon.tech") ||
		strings.Contains(connString, "ep-") ||
		strings.Contains(connString, ".cloud.local") // For local development
}

// GetNeonDatabaseName extracts the database name from a Neon connection string.
func GetNeonDatabaseName(connString string) (string, error) {
	parsed, err := url.Parse(connString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	return strings.TrimPrefix(parsed.Path, "/"), nil
}

// NeonPoolWrapper wraps a pgxpool.Pool with Neon-specific optimizations.
// This provides a higher-level interface similar to the regular Pool.
type NeonPoolWrapper struct {
	pool     *pgxpool.Pool
	databse string
}

// NewNeonPoolWrapper creates a new NeonPoolWrapper.
func NewNeonPoolWrapper(connString string) (*NeonPoolWrapper, error) {
	pool, err := NewNeonPool(connString)
	if err != nil {
		return nil, err
	}

	dbName, err := GetNeonDatabaseName(connString)
	if err != nil {
		dbName = "unknown"
	}

	return &NeonPoolWrapper{
		pool:     pool,
		databse: dbName,
	}, nil
}

// Close closes the underlying pool.
func (n *NeonPoolWrapper) Close() {
	if n.pool != nil {
		n.pool.Dispose()
		n.pool = nil
	}
}

// Ping checks the connection.
func (n *NeonPoolWrapper) Ping(ctx context.Context) error {
	if n.pool == nil {
		return fmt.Errorf("pool is closed")
	}
	return n.pool.Ping(ctx)
}

// Pool returns the underlying pgxpool.Pool.
func (n *NeonPoolWrapper) Pool() *pgxpool.Pool {
	return n.pool
}

// Database returns the database name.
func (n *NeonPoolWrapper) Database() string {
	return n.databse
}