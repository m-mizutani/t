package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/m-mizutani/clog"
)

// LogLevel represents log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Config represents logger configuration
type Config struct {
	Level   LogLevel
	Verbose bool
}

// contextKey is the key for storing logger in context
type contextKey struct{}

var loggerKey = contextKey{}

// New creates a new logger with the given configuration
func New(cfg Config) *slog.Logger {
	// Convert LogLevel to slog.Level
	var level slog.Level
	switch cfg.Level {
	case LogLevelDebug:
		level = slog.LevelDebug
	case LogLevelInfo:
		level = slog.LevelInfo
	case LogLevelWarn:
		level = slog.LevelWarn
	case LogLevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Add source info if verbose
	if cfg.Verbose {
		opts.AddSource = true
	}

	// Create clog handler for enhanced console output
	handler := clog.New(
		clog.WithWriter(os.Stderr),
		clog.WithLevel(level),
		clog.WithSource(cfg.Verbose),
		clog.WithColor(true),
		clog.WithPrinter(clog.LinearPrinter),
	)

	return slog.New(handler)
}

// ParseLevel parses log level from string
func ParseLevel(s string) LogLevel {
	switch strings.ToLower(s) {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// ParseVerbose parses verbose flag from string
func ParseVerbose(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "1" || s == "yes" || s == "on"
}

// NewFromEnv creates a new logger from environment variables
func NewFromEnv() *slog.Logger {
	cfg := Config{
		Level:   ParseLevel(os.Getenv("T_LOG_LEVEL")),
		Verbose: ParseVerbose(os.Getenv("T_VERBOSE")),
	}
	return New(cfg)
}

// WithLogger embeds a logger into the context
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext extracts a logger from the context
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	// Return default logger if not found in context
	return slog.Default()
}
