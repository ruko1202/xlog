package xlog

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/ruko1202/xlog/xfield"
)

// WithOperationSpan creates a new span for the given operation and attaches it to the context.
// It also creates a named logger with the operation name and the provided fields.
// The fields are added both to the logger and as span attributes.
// Returns the updated context with both the logger and span attached, along with the span itself.
//
// Example:
//
//	ctx, span := xlog.WithOperationSpan(ctx, "process-payment",
//	    xlog.String("user_id", "123"),
//	    xlog.String("payment_id", "pay_xyz"),
//	)
//	defer span.End()
func WithOperationSpan(ctx context.Context, operation string, fields ...xfield.Field) (context.Context, trace.Span) {
	logger := loggerFromContext(ctx).
		Named(operation).
		With(fields...)

	// Trace/span metadata will be added directly in Debug/Info/etc
	tracer := tracerFromContext(ctx)
	ctx, span := tracer.Start(ctx, operation)
	if span.IsRecording() {
		span.SetAttributes(fieldsToOtelAttributes(fields)...)
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
//
// Example:
//
//	xlog.SetSpanAttributes(ctx,
//	    attribute.String("user_id", "123"),
//	    attribute.Int64("count", 42),
//	)
func SetSpanAttributes(ctx context.Context, kv ...attribute.KeyValue) {
	span := SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(kv...)
	}
}

// AddSpanEvent adds an event to the span extracted from the context.
// Events are timestamped occurrences that can include additional attributes.
// If no span is found or if the span is not recording, this is a no-op.
//
// Example:
//
//	xlog.AddSpanEvent(ctx, "cache-miss")
//	xlog.AddSpanEvent(ctx, "retry", trace.WithAttributes(
//	    attribute.Int("attempt", 3),
//	))
func AddSpanEvent(ctx context.Context, name string, options ...trace.EventOption) {
	span := SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, options...)
	}
}

// RecordSpanError records an error on the span extracted from the context.
// It also sets the span status to Error with the error message.
// If no span is found or if the span is not recording, this is a no-op.
//
// Example:
//
//	if err := doSomething(); err != nil {
//	    xlog.RecordSpanError(ctx, err, trace.WithStackTrace(true))
//	    return err
//	}
func RecordSpanError(ctx context.Context, err error, options ...trace.EventOption) {
	span := SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err, options...)
		span.SetStatus(codes.Error, err.Error())
	}
}
