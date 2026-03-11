package xlog

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/ruko1202/xlog/xfield"
)

// Debug logs a Debug level message with structured fields.
// Logger is extracted from context. If logger is not found, the global logger is used.
//
// Example:
//
//	xlog.Debug(ctx, "debug message", xlog.String("key", "value"))
func Debug(ctx context.Context, msg string, fields ...xfield.Field) {
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
//	xlog.Info(ctx, "request processed", xlog.Duration("took", time.Second))
func Info(ctx context.Context, msg string, fields ...xfield.Field) {
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
//	xlog.Warn(ctx, "slow query", xlog.Duration("took", time.Second*5))
func Warn(ctx context.Context, msg string, fields ...xfield.Field) {
	markSpanError(ctx, msg, fields)

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
//	xlog.Error(ctx, "database query error", xlog.Error(err))
func Error(ctx context.Context, msg string, fields ...xfield.Field) {
	markSpanError(ctx, msg, fields)

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
//	xlog.Fatal(ctx, "critical error", xlog.Error(err))
func Fatal(ctx context.Context, msg string, fields ...xfield.Field) {
	markSpanError(ctx, msg, fields)

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
//	xlog.Panic(ctx, "unexpected state", xlog.String("state", state))
func Panic(ctx context.Context, msg string, fields ...xfield.Field) {
	markSpanError(ctx, msg, fields)

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

func withMetadataFields(ctx context.Context, fields []xfield.Field) []xfield.Field {
	traceFields := traceMetadataFields(ctx)
	if len(traceFields) == 0 {
		return fields
	}

	totalLen := len(fields) + len(traceFields)

	result := make([]xfield.Field, 0, totalLen)
	result = append(result, fields...)
	result = append(result, traceFields...)

	return result
}

func traceMetadataFields(ctx context.Context) []xfield.Field {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return []xfield.Field{
			xfield.String("trace_id", spanCtx.TraceID().String()),
			xfield.String("span_id", spanCtx.SpanID().String()),
		}
	}

	return nil
}

func markSpanError(ctx context.Context, msg string, fields []xfield.Field) {
	span := SpanFromContext(ctx)
	if span.IsRecording() {
		for _, f := range fields {
			if f.Type == xfield.ErrorType {
				if err, ok := f.Interface.(error); ok && err != nil {
					span.SetStatus(codes.Error, msg)
					span.RecordError(err,
						trace.WithStackTrace(true),
					)
					break
				}
			}
		}
	}
}
