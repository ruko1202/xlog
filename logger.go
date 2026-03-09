package xlog

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Debug logs a Debug level message with structured fields.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Debug(ctx, "debug message", zap.String("key", "value"))
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	logger := loggerFromContext(ctx)
	logger.Debug(msg, withMetadataFields(ctx, fields)...)
}

// Debugf logs a formatted Debug level message.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Debugf(ctx, "value: %d, status: %s", 42, "ok")
func Debugf(ctx context.Context, template string, args ...any) {
	Debug(ctx, fmt.Sprintf(template, args...))
}

// Info logs an Info level message with structured fields.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Info(ctx, "request processed", zap.Duration("took", time.Second))
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	logger := loggerFromContext(ctx)
	logger.Info(msg, withMetadataFields(ctx, fields)...)
}

// Infof logs a formatted Info level message.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Infof(ctx, "user %s logged in", userID)
func Infof(ctx context.Context, template string, args ...any) {
	Info(ctx, fmt.Sprintf(template, args...))
}

// Warn logs a Warn level message with structured fields.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Warn(ctx, "slow query", zap.Duration("took", time.Second*5))
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	logger := loggerFromContext(ctx)
	logger.Warn(msg, withMetadataFields(ctx, fields)...)
}

// Warnf logs a formatted Warn level message.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Warnf(ctx, "retry attempts: %d", retryCount)
func Warnf(ctx context.Context, template string, args ...any) {
	Warn(ctx, fmt.Sprintf(template, args...))
}

// Error logs an Error level message with structured fields.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Error(ctx, "database query error", zap.Error(err))
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	logger := loggerFromContext(ctx)
	logger.Error(msg, withMetadataFields(ctx, fields)...)
}

// Errorf logs a formatted Error level message.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Errorf(ctx, "failed to process request: %v", err)
func Errorf(ctx context.Context, template string, args ...any) {
	Error(ctx, fmt.Sprintf(template, args...))
}

// Fatal logs a Fatal level message with structured fields and terminates the program.
// Logger is extracted from context. If logger is not found, the global logger is used.
// Calls os.Exit(1) after logging.
//
// Example:
//
//	xlog.Fatal(ctx, "critical error", zap.Error(err))
func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	logger := loggerFromContext(ctx)
	logger.Fatal(msg, withMetadataFields(ctx, fields)...)
}

// Fatalf logs a formatted Fatal level message and terminates the program.
// Logger is extracted from context. If logger is not found, the global logger is used.
// Calls os.Exit(1) after logging.
//
// Example:
//
//	xlog.Fatalf(ctx, "failed to start server: %v", err)
func Fatalf(ctx context.Context, template string, args ...any) {
	Fatal(ctx, fmt.Sprintf(template, args...))
}

// Panic logs a Panic level message with structured fields and panics.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Panic(ctx, "unexpected state", zap.String("state", state))
func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	logger := loggerFromContext(ctx)
	logger.Panic(msg, withMetadataFields(ctx, fields)...)
}

// Panicf logs a formatted Panic level message and panics.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Panicf(ctx, "invalid value: %v", value)
func Panicf(ctx context.Context, template string, args ...any) {
	Panic(ctx, fmt.Sprintf(template, args...))
}

func withMetadataFields(ctx context.Context, fields []zap.Field) []zap.Field {
	slices := [][]zap.Field{
		fields,
	}

	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]zap.Field, 0, totalLen)
	for _, s := range slices {
		result = append(result, s...)
	}

	return result
}
