package xlog

import (
	"context"
)

// WithOperation creates a new context with a named logger for a specific operation.
// It extracts the logger from the context, adds the operation name, and includes any additional fields.
func WithOperation(ctx context.Context, operation string, fields ...Field) context.Context {
	logger := fromContext(ctx).
		Named(operation).
		With(fields...)

	return ContextWithLogger(ctx, logger)
}

// WithFields creates a new context with additional fields added to the logger.
// It extracts the logger from the context and adds the specified fields.
func WithFields(ctx context.Context, fields ...Field) context.Context {
	logger := fromContext(ctx).
		With(fields...)

	return ContextWithLogger(ctx, logger)
}
