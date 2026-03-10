// Package xlog provides a wrapper around uber-go/zap for context-aware logging.
//
// The package allows storing a logger in context and automatically extracting it
// when logging functions are called. If logger is not found in context, the global logger is used.
package xlog

import (
	"context"
	"sync"
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
func ContextWithLogger(ctx context.Context, logger Logger) context.Context {
	if logger == nil {
		logger = GlobalLogger()
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
func LoggerFromContext(ctx context.Context) Logger {
	return loggerFromContext(ctx)
}

func loggerFromContext(ctx context.Context) Logger {
	logger, ok := ctx.Value(loggerCtxKey).(Logger)
	if !ok {
		return GlobalLogger()
	}

	return logger
}

var (
	_loggerMu     sync.RWMutex
	_globalLogger = NewNoopLogger() // default fallback
)

// ReplaceGlobalLogger replaces the global logger and returns a function to restore the previous logger.
// This function is thread-safe and can be called concurrently.
//
// This is similar to zap.ReplaceGlobals() but works with any Logger implementation.
//
// Example:
//
//	logger := xlog.NewZapAdapter(zapLogger)
//	restore := xlog.ReplaceGlobalLogger(logger)
//	defer restore() // Restore previous logger when done
func ReplaceGlobalLogger(logger Logger) func() {
	if logger == nil {
		logger = NewNoopLogger()
	}

	_loggerMu.Lock()
	prev := _globalLogger
	_globalLogger = logger
	_loggerMu.Unlock()

	return func() { ReplaceGlobalLogger(prev) }
}

// GlobalLogger returns the global logger.
// If no logger has been set via ReplaceGlobalLogger, returns NoopLogger.
// This function is thread-safe.
func GlobalLogger() Logger {
	_loggerMu.RLock()
	defer _loggerMu.RUnlock()
	return _globalLogger
}

// L is a shorthand for GlobalLogger().
// This is similar to zap.L() and provides quick access to the global logger.
//
// Example:
//
//	xlog.L().Info("message", xlog.String("key", "value"))
func L() Logger {
	return GlobalLogger()
}
