// Package xlog provides a wrapper around uber-go/zap for context-aware logging.
//
// The package allows storing a logger in context and automatically extracting it
// when logging functions are called. If logger is not found in context, the global logger is used.
package xlog

import (
	"context"

	"go.uber.org/zap"
)

type xlogCtxKey int

const (
	loggerCtxKey xlogCtxKey = iota
)

// ContextWithLogger adds a logger to the context and returns a new context.
//
// Example:
//
//	logger, _ := zap.NewProduction()
//	ctx := xlog.ContextWithLogger(context.Background(), logger)
func ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	if logger == nil {
		logger = zap.L()
	}
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
	return loggerFromContext(ctx)
}

func loggerFromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(loggerCtxKey).(*zap.Logger)
	if !ok {
		return zap.L()
	}

	return logger
}
