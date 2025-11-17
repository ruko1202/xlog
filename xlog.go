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

type ctxKey string

const (
	loggerCtxKey ctxKey = "xLoggerKey"
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
//	logger := xlog.FromContext(ctx)
//	logger.Info("direct zap logger usage")
func LoggerFromContext(ctx context.Context) *zap.Logger {
	return fromContext(ctx)
}

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
