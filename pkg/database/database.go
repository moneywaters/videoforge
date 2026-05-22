package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/videoforge/backend/pkg/logger"
)

// Pool wraps the pgxpool.Pool and provides additional functionality.
type Pool struct {
	pool  *pgxpool.Pool
	mutex sync.RWMutex
	log   *logger.Logger
}

// PoolConfig holds the configuration for the database pool.
type PoolConfig struct {
	// ConnString is the PostgreSQL connection string.
	// Format: postgres://user:password@host:port/database?sslmode=disable
	ConnString string

	// MaxConns is the maximum number of connections.
	MaxConns int32

	// MinConns is the minimum number of connections.
	MinConns int32

	// MaxConnLifetime is the maximum connection lifetime.
	MaxConnLifetime time.Duration

	// MaxConnIdleTime is the maximum connection idle time.
	MaxConnIdleTime time.Duration

	// HealthCheckPeriod is the period for health checks.
	HealthCheckPeriod time.Duration
}

// DefaultPoolConfig returns default configuration for the database pool.
func DefaultPoolConfig(connString string) *PoolConfig {
	return &PoolConfig{
		ConnString:        connString,
		MaxConns:          20,
		MinConns:          5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 5 * time.Minute,
	}
}

// NewPool creates a new database pool with the given configuration.
func NewPool(connString string, log *logger.Logger) (*Pool, error) {
	return NewPoolWithConfig(DefaultPoolConfig(connString), log)
}

// NewPoolWithConfig creates a new database pool with the given configuration.
func NewPoolWithConfig(config *PoolConfig, log *logger.Logger) (*Pool, error) {
	if log == nil {
		log = logger.Default("development")
	}

	// Parse configuration
	poolConfig, err := pgxpool.ParseConfig(config.ConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Set pool configuration
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

	// Create pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Database pool created")

	return &Pool{
		pool: pool,
		log:  log,
	}, nil
}

// Close closes the database pool.
func (p *Pool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
		p.log.Info("Database pool closed")
	}

	return nil
}

// Exec executes a query without returning rows.
func (p *Pool) Exec(ctx context.Context, query string, args ...interface{}) error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return fmt.Errorf("pool is closed")
	}

	_, err := p.pool.Exec(ctx, query, args...)
	return err
}

// Query executes a query that returns Rows
func (p *Pool) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return nil, fmt.Errorf("pool is closed")
	}

	return p.pool.Query(ctx, query, args...)
}

// QueryRow executes a query that returns a single Row
func (p *Pool) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return nil
	}

	return p.pool.QueryRow(ctx, query, args...)
}

// Acquire acquires a connection from the pool.
func (p *Pool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return nil, fmt.Errorf("pool is closed")
	}

	return p.pool.Acquire(ctx)
}

// Begin starts a transaction.
func (p *Pool) Begin(ctx context.Context) (pgx.Tx, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return nil, fmt.Errorf("pool is closed")
	}

	return p.pool.Begin(ctx)
}

// BeginFunc starts a transaction and executes the given function.
func (p *Pool) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return fmt.Errorf("pool is closed")
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	if err := f(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// Ping checks the database connection.
func (p *Pool) Ping(ctx context.Context) error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return fmt.Errorf("pool is closed")
	}

	return p.pool.Ping(ctx)
}

// Stats returns pool statistics.
func (p *Pool) Stats() *pgxpool.Stat {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return nil
	}

	return p.pool.Stat()
}

// Config returns the pool configuration.
func (p *Pool) Config() *pgxpool.Config {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return nil
	}

	return p.pool.Config()
}

// IsConnected returns true if the pool is connected.
func (p *Pool) IsConnected() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if p.pool == nil {
		return false
	}

	// Try to ping to check connection
	err := p.pool.Ping(context.Background())
	return err == nil
}

// Logger returns the logger.
func (p *Pool) Logger() *logger.Logger {
	return p.log
}