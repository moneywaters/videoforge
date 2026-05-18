package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

// Level represents the log level.
type Level int

const (
	LevelDebug Level = Level(slog.LevelDebug)
	LevelInfo  Level = Level(slog.LevelInfo)
	LevelWarn  Level = Level(slog.LevelWarn)
	LevelError Level = Level(slog.LevelError)
)

// Logger is a structured logger wrapper around slog.
// It provides convenience methods for logging at different levels
// with context and key-value pairs.
type Logger struct {
	logger *slog.Logger
	mu     sync.Mutex
}

// New creates a new Logger with the specified output and options.
// If output is nil, it defaults to os.Stderr.
// For production, use JSON handler; for development, use text handler.
func New(output io.Writer, isJSON bool, level Level) *Logger {
	if output == nil {
		output = os.Stderr
	}

	var handler slog.Handler
	if isJSON {
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: slog.Level(level),
		})
	} else {
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: slog.Level(level),
		})
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// Default creates a logger suitable for the given environment.
// It uses JSON format in production and text format in development.
func Default(environment string) *Logger {
	isJSON := environment == "production"
	return New(nil, isJSON, LevelInfo)
}

// With returns a new Logger with the given context merged into the context.
func (l *Logger) With(ctx ...any) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	newLogger := l.logger.With(ctx...)
	return &Logger{logger: newLogger}
}

// Info logs at INFO level with the given context and key-value pairs.
func (l *Logger) Info(msg string, ctx ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Info(msg, ctx...)
}

// Error logs at ERROR level with the given context and key-value pairs.
func (l *Logger) Error(msg string, ctx ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Error(msg, ctx...)
}

// Warn logs at WARN level with the given context and key-value pairs.
func (l *Logger) Warn(msg string, ctx ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Warn(msg, ctx...)
}

// Debug logs at DEBUG level with the given context and key-value pairs.
func (l *Logger) Debug(msg string, ctx ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Debug(msg, ctx...)
}

// Log logs at the specified level with the given context and key-value pairs.
func (l *Logger) Log(level Level, msg string, ctx ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Log(context.Background(), slog.Level(level), msg, ctx...)
}

// Handler returns the underlying slog handler.
func (l *Logger) Handler() slog.Handler {
	return l.logger.Handler()
}

// Context returns a context with the logger attached.
func (l *Logger) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// FromContext retrieves the logger from the context.
// Returns nil if no logger is found.
func FromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(loggerKey{}).(*Logger)
	if !ok {
		return nil
	}
	return logger
}

// loggerKey is used to store the logger in the context.
type loggerKey struct{}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if h, ok := l.logger.Handler().(slog.Handler); ok {
		return h
	}
	return nil
}