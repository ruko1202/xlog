package xlog

import "github.com/ruko1202/xlog/xfield"

// NoopLogger is a logger that does nothing.
// This is used as a fallback when no logger is configured.
type NoopLogger struct{}

// NewNoopLogger creates a new no-op logger.
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

// Debug is a no-op implementation.
func (l *NoopLogger) Debug(_ string, _ ...xfield.Field) {}

// Info is a no-op implementation.
func (l *NoopLogger) Info(_ string, _ ...xfield.Field) {}

// Warn is a no-op implementation.
func (l *NoopLogger) Warn(_ string, _ ...xfield.Field) {}

// Error is a no-op implementation.
func (l *NoopLogger) Error(_ string, _ ...xfield.Field) {}

// Fatal is a no-op implementation.
func (l *NoopLogger) Fatal(_ string, _ ...xfield.Field) {}

// Panic panics with the given message.
func (l *NoopLogger) Panic(msg string, _ ...xfield.Field) { panic(msg) }

// With returns the same logger instance.
func (l *NoopLogger) With(_ ...xfield.Field) Logger { return l }

// Named returns the same logger instance.
func (l *NoopLogger) Named(_ string) Logger { return l }

// Sync flushes any buffered log entries.
func (l *NoopLogger) Sync() error { return nil }
