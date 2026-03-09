package xlog

import (
	"context"
	"math"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WithOperationSpan creates a new span for the given operation and attaches it to the context.
// It also creates a named logger with the operation name and the provided fields.
// The fields are added both to the logger and as span attributes.
// Returns the updated context with both the logger and span attached, along with the span itself.
func WithOperationSpan(ctx context.Context, operation string, fields ...zap.Field) (context.Context, trace.Span) {
	logger := loggerFromContext(ctx).
		Named(operation).
		With(fields...)

	// Trace/span metadata will be added directly in Debug/Info/etc
	tracer := tracerFromContext(ctx)
	ctx, span := tracer.Start(ctx, operation)
	if span.IsRecording() {
		span.SetAttributes(convertFieldsToAttributes(fields)...)
	}

	return ContextWithLogger(ctx, logger), span
}

// SpanFromContext extracts span from context using OpenTelemetry's standard API.
// If span is not found, returns a NoopSpan.
//
// Example:
//
//	span := xlog.SpanFromContext(ctx)
//	span.SetAttributes(attribute.String("key", "value"))
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// SetSpanAttributes adds attributes to the span extracted from the context.
// If no span is found or if the span is not recording, this is a no-op.
func SetSpanAttributes(ctx context.Context, kv ...attribute.KeyValue) {
	span := SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(kv...)
	}
}

// AddSpanEvent adds an event to the span extracted from the context.
// Events are timestamped occurrences that can include additional attributes.
// If no span is found or if the span is not recording, this is a no-op.
func AddSpanEvent(ctx context.Context, name string, options ...trace.EventOption) {
	span := SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, options...)
	}
}

// RecordSpanError records an error on the span extracted from the context.
// It also sets the span status to Error with the error message.
// If no span is found or if the span is not recording, this is a no-op.
func RecordSpanError(ctx context.Context, err error, options ...trace.EventOption) {
	span := SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err, options...)
		span.SetStatus(codes.Error, err.Error())
	}
}

// convertFieldsToAttributes converts zap fields to OpenTelemetry attributes.
//
// Supported field types:
//   - String (zapcore.StringType)
//   - Integer types (Int64, Int32, Int16, Int8, Uint64, Uint32, Uint16, Uint8)
//   - Bool (zapcore.BoolType)
//   - Float types (Float64, Float32)
//
// Unsupported types (Any, Array, Object, Binary, etc.) are silently ignored.
// This is intentional to avoid exceeding OpenTelemetry attribute size limits
// and to prevent performance issues with complex data structures.
//
// If you need to include complex types as span attributes, consider:
//   - Converting them to strings manually before passing to WithOperationSpan
//   - Using SetSpanAttributes with pre-converted attribute.KeyValue
//   - Logging the full data separately and only adding a reference ID to the span
func convertFieldsToAttributes(fields []zap.Field) []attribute.KeyValue {
	if len(fields) == 0 {
		return nil
	}

	attrs := make([]attribute.KeyValue, 0, len(fields))
	for _, f := range fields {
		switch f.Type {
		case zapcore.StringType:
			attrs = append(attrs, attribute.String(f.Key, f.String))

		case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type,
			zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
			// In zap, all integer values are stored in the Integer field
			attrs = append(attrs, attribute.Int64(f.Key, f.Integer))

		case zapcore.BoolType:
			// Zap stores boolean values as 1 (true) or 0 (false) in the Integer field
			attrs = append(attrs, attribute.Bool(f.Key, f.Integer == 1))

		case zapcore.Float64Type, zapcore.Float32Type:
			// Zap cleverly converts floats to int64 under the hood, converting back
			// #nosec G115 -- intentional bit pattern conversion from int64 to uint64 for float reconstruction
			attrs = append(attrs, attribute.Float64(f.Key, math.Float64frombits(uint64(f.Integer))))

		default:
			// Unsupported types are silently ignored.
			// See function documentation for details.
		}
	}

	return attrs
}
