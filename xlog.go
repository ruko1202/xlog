// Package xlog provides a wrapper around uber-go/zap for context-aware logging.
//
// The package allows storing a logger in context and automatically extracting it
// when logging functions are called. If logger is not found in context, the global logger is used.
package xlog

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

var (
	globalMu     sync.Mutex
	globalLogger = zap.NewNop()
)

type xlogCtxKey string

const (
	loggerCtxKey xlogCtxKey = "xLoggerKey"
)

// ContextWithLogger adds a logger to the context and returns a new context.
//
// Example:
//
//	logger, _ := zap.NewProduction()
//	ctx := xlog.ContextWithLogger(context.Background(), logger)
func ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

// LoggerFromContext extracts logger from context.
// If logger is not found, returns the global logger.
//
// Example:
//
//	logger := xlog.LoggerFromContext(ctx)
//	logger.Info("direct zap logger usage")
func LoggerFromContext(ctx context.Context) *zap.Logger {
	return fromContext(ctx)
}

// ReplaceGlobal replaces the global logger and returns a function to restore the previous logger.
// This function is thread-safe and can be used for testing or temporarily changing the logger.
//
// Example:
//
//	logger, _ := zap.NewProduction()
//	restore := xlog.ReplaceGlobal(logger)
//	defer restore() // Restore previous logger when done
//
//	// Or for testing:
//	func TestSomething(t *testing.T) {
//	    testLogger := zaptest.NewLogger(t)
//	    restore := xlog.ReplaceGlobal(testLogger)
//	    defer restore()
//	    // test code
//	}
func ReplaceGlobal(logger *zap.Logger) func() {
	globalMu.Lock()
	defer globalMu.Unlock()

	prev := globalLogger
	globalLogger = logger

	return func() { ReplaceGlobal(prev) }
}

func fromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(loggerCtxKey).(*zap.Logger)
	if !ok {
		return globalLogger
	}

	return logger
}
