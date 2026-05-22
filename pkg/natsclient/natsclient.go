package natsclient

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/videoforge/backend/pkg/logger"
)

// NATSClient defines the interface for NATS message broker operations.
// This allows for easier testing through mocking.
type NATSClient interface {
	// Connect establishes a connection to the NATS server.
	Connect() error

	// Close closes the connection.
	Close() error

	// Publish publishes a message to the given subject.
	// The data is serialized as JSON before sending.
	Publish(subject string, data []byte) error

	// PublishWithContext publishes a message with a context.
	PublishWithContext(ctx context.Context, subject string, data []byte) error

	// Subscribe registers a handler for messages on the given subject.
	Subscribe(subject string, handler func(msg *nats.Msg)) error

	// SubscribeWithHandler registers a subscription with a custom handler.
	SubscribeWithHandler(subject string, handler nats.MsgHandler) (*nats.Subscription, error)

	// IsConnected returns true if the client is currently connected.
	IsConnected() bool

	// Flush sends a protocol flush to the server.
	Flush() error
}

// Config holds the configuration for the NATS client.
type Config struct {
	// URL is the NATS server URL (e.g., nats://localhost:4222).
	URL string

	// Name is the client name.
	Name string

	// MaxReconnectAttempts is the maximum number of reconnection attempts.
	// Set to -1 for unlimited.
	MaxReconnectAttempts int

	// ReconnectWait is the time to wait between reconnection attempts.
	ReconnectWait time.Duration

	// Timeout is the timeout for operations.
	Timeout time.Duration
}

// DefaultConfig returns default configuration for NATS client.
func DefaultConfig() *Config {
	return &Config{
		URL:                 "nats://localhost:4222",
		Name:                "videoforge",
		MaxReconnectAttempts: 10,
		ReconnectWait:        2 * time.Second,
		Timeout:             5 * time.Second,
	}
}

// Client is a NATS client implementation with reconnection logic.
type Client struct {
	mu           sync.RWMutex
	conn         *nats.Conn
	config       *Config
Subscriptions map[string]*nats.Subscription
	logger       *logger.Logger
	closed       bool
}

// New creates a new NATS client with the given configuration.
func New(config *Config, log *logger.Logger) *Client {
	if config == nil {
		config = DefaultConfig()
	}
	if log == nil {
		log = logger.Default("development")
	}

	return &Client{
		config: config,
		logger: log,
		Subscriptions: make(map[string]*nats.Subscription),
	}
}

// Connect establishes a connection to the NATS server.
// It includes automatic reconnection logic.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && c.conn.IsConnected() {
		return nil
	}

	// Configure NATS options
	opts := []nats.Option{
		nats.Name(c.config.Name),
		nats.MaxReconnects(c.config.MaxReconnectAttempts),
		nats.ReconnectWait(c.config.ReconnectWait),
		nats.Timeout(c.config.Timeout),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			c.logger.Error("NATS disconnected",
				slog.String("error", err.Error()),
			)
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			url := conn.ConnectedUrl()
			c.logger.Info("NATS reconnected",
				slog.String("server", url),
			)
		}),
	}

	// Connect to NATS
	nc, err := nats.Connect(c.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	c.conn = nc
	c.logger.Info("NATS connected",
		slog.String("url", c.config.URL),
	)

	return nil
}

// ConnectWithRetry attempts to connect with retry logic.
func (c *Client) ConnectWithRetry(maxAttempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		err := c.Connect()
		if err == nil {
			return nil
		}
		lastErr = err
		c.logger.Warn("NATS connection failed, retrying...",
			slog.String("error", err.Error()),
			slog.Int("attempt", i+1),
		)
		time.Sleep(delay)
	}
	return fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, lastErr)
}

// Close closes the NATS connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	c.closed = true
	err := c.conn.Drain()
	if err != nil {
		return fmt.Errorf("failed to drain connection: %w", err)
	}

	c.conn = nil
	c.logger.Info("NATS connection closed")

	return nil
}

// Publish publishes a message to the given subject.
func (c *Client) Publish(subject string, data []byte) error {
	return c.PublishWithContext(context.Background(), subject, data)
}

// PublishWithContext publishes a message with a context.
// Note: nats.go v1.34+ uses context-enabled publish via RequestWithContext or explicit context handling
func (c *Client) PublishWithContext(ctx context.Context, subject string, data []byte) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil || !conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	// Use Publish with context via the async API, or publish directly
	// For context-aware publish in v1.34+, we use the legacy approach since newer APIs changed
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := conn.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	c.logger.Debug("NATS message published",
		slog.String("subject", subject),
		slog.Int("size", len(data)),
	)

	return nil
}

// Subscribe registers a handler for messages on the given subject.
func (c *Client) Subscribe(subject string, handler func(msg *nats.Msg)) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil || !conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	sub, err := conn.Subscribe(subject, handler)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	c.mu.Lock()
	c.Subscriptions[subject] = sub
	c.mu.Unlock()

	c.logger.Info("NATS subscription created",
		slog.String("subject", subject),
	)

	return nil
}

// SubscribeWithHandler registers a subscription with a NATS handler.
func (c *Client) SubscribeWithHandler(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil || !conn.IsConnected() {
		return nil, fmt.Errorf("NATS not connected")
	}

	sub, err := conn.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	c.mu.Lock()
	c.Subscriptions[subject] = sub
	c.mu.Unlock()

	c.logger.Info("NATS subscription created",
		slog.String("subject", subject),
	)

	return sub, nil
}

// Unsubscribe removes a subscription.
func (c *Client) Unsubscribe(subject string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub, ok := c.Subscriptions[subject]
	if !ok {
		return nil
	}

	err := sub.Unsubscribe()
	if err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	delete(c.Subscriptions, subject)
	c.logger.Info("NATS subscription removed",
		slog.String("subject", subject),
	)

	return nil
}

// IsConnected returns true if the client is currently connected.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && c.conn.IsConnected()
}

// Flush sends a protocol flush to the server.
func (c *Client) Flush() error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil || !conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	return conn.Flush()
}

// FlushTimeout sends a protocol flush with timeout.
func (c *Client) FlushTimeout(timeout time.Duration) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil || !conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	return conn.FlushTimeout(timeout)
}

// ConnectedUrl returns the URL of the connected server.
func (c *Client) ConnectedUrl() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn == nil {
		return ""
	}
	return c.conn.ConnectedUrl()
}

// QueueSubscribe creates a queue subscription.
func (c *Client) QueueSubscribe(subject, queue string, handler func(msg *nats.Msg)) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil || !conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}

	_, err := conn.QueueSubscribe(subject, queue, handler)
	if err != nil {
		return fmt.Errorf("failed to queue subscribe: %w", err)
	}

	c.logger.Info("NATS queue subscription created",
		slog.String("subject", subject),
		slog.String("queue", queue),
	)

	return nil
}