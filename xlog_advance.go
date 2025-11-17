package xlog

import (
	"context"

	"go.uber.org/zap"
)

// WithOperation creates a new context with a named logger for a specific operation.
// It extracts the logger from the context, adds the operation name, and includes any additional fields.
func WithOperation(ctx context.Context, operation string, fields ...zap.Field) context.Context {
	logger := fromContext(ctx).
		Named(operation).
		With(fields...)

	return ContextWithLogger(ctx, logger)
}
