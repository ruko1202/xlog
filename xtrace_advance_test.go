package xlog

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func setupTestTracer(t *testing.T) (*tracetest.SpanRecorder, trace.Tracer) {
	t.Helper()

	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanRecorder),
	)

	// Set as global tracer provider
	otel.SetTracerProvider(tracerProvider)

	tracer := tracerProvider.Tracer("test-tracer")
	return spanRecorder, tracer
}

func TestWithOperationSpan(t *testing.T) {
	t.Run("creates span and enriched logger", func(t *testing.T) {
		spanRecorder, _ := setupTestTracer(t)
		logger, logs := initTestLogger(t)

		ctx := ContextWithLogger(context.Background(), logger)

		// Create span with operation
		ctx, span := WithOperationSpan(ctx, "test-operation")
		defer span.End()

		// Log something
		Info(ctx, "operation started")

		// Check logger
		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		assert.Equal(t, "test-operation", entry.LoggerName)
		assert.Equal(t, "operation started", entry.Message)

		// End span to flush it
		span.End()

		// Check span
		spans := spanRecorder.Ended()
		require.Equal(t, 1, len(spans))
		assert.Equal(t, "test-operation", spans[0].Name())
	})

	t.Run("adds fields to both logger and span", func(t *testing.T) {
		spanRecorder, _ := setupTestTracer(t)
		logger, logs := initTestLogger(t)

		ctx := ContextWithLogger(context.Background(), logger)

		// Create span with fields
		ctx, span := WithOperationSpan(ctx, "payment-process",
			zap.String("user_id", "12345"),
			zap.Int("amount", 100),
		)
		defer span.End()

		// Log
		Info(ctx, "processing payment")

		// Check logger has fields
		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		// Should have user_id, amount, trace_id, span_id
		require.GreaterOrEqual(t, len(entry.Context), 2)
		assert.Equal(t, "user_id", entry.Context[0].Key)
		assert.Equal(t, "amount", entry.Context[1].Key)

		// End span to flush it
		span.End()

		// Check span has attributes
		spans := spanRecorder.Ended()
		require.Equal(t, 1, len(spans))
		attrs := spans[0].Attributes()
		require.Equal(t, 2, len(attrs))
		assert.Equal(t, attribute.String("user_id", "12345"), attrs[0])
		assert.Equal(t, attribute.Int64("amount", 100), attrs[1])
	})

	t.Run("span is stored in context", func(t *testing.T) {
		setupTestTracer(t)
		logger, _ := initTestLogger(t)

		ctx := ContextWithLogger(context.Background(), logger)

		ctx, span := WithOperationSpan(ctx, "test-op")
		defer span.End()

		// Extract span from context
		extractedSpan := trace.SpanFromContext(ctx)
		require.NotNil(t, extractedSpan)
		assert.True(t, extractedSpan.IsRecording())
	})

	t.Run("nested spans create parent-child relationship", func(t *testing.T) {
		spanRecorder, _ := setupTestTracer(t)
		logger, _ := initTestLogger(t)

		ctx := ContextWithLogger(context.Background(), logger)

		// Parent span
		ctx, parentSpan := WithOperationSpan(ctx, "parent")
		defer parentSpan.End()

		// Child span
		ctx, childSpan := WithOperationSpan(ctx, "child")
		defer childSpan.End()

		childSpan.End()
		parentSpan.End()

		spans := spanRecorder.Ended()
		require.Equal(t, 2, len(spans))

		// Child should have parent's span ID as parent
		childSpanContext := spans[0].SpanContext()
		parentSpanContext := spans[1].SpanContext()

		assert.Equal(t, "child", spans[0].Name())
		assert.Equal(t, "parent", spans[1].Name())
		assert.Equal(t, parentSpanContext.SpanID(), spans[0].Parent().SpanID())
		assert.Equal(t, parentSpanContext.TraceID(), childSpanContext.TraceID())
	})
}

func TestSpanFromContext(t *testing.T) {
	t.Run("returns span from context", func(t *testing.T) {
		setupTestTracer(t)
		ctx := context.Background()

		ctx, span := WithOperationSpan(ctx, "test")
		defer span.End()

		extractedSpan := SpanFromContext(ctx)
		require.NotNil(t, extractedSpan)
		assert.True(t, extractedSpan.IsRecording())
	})

	t.Run("returns noop span when not in context", func(t *testing.T) {
		ctx := context.Background()

		extractedSpan := SpanFromContext(ctx)
		require.NotNil(t, extractedSpan)
		// Noop span is not recording
		assert.False(t, extractedSpan.IsRecording())
	})
}

func TestSetSpanAttributes(t *testing.T) {
	t.Run("sets attributes on active span", func(t *testing.T) {
		spanRecorder, _ := setupTestTracer(t)
		ctx := context.Background()

		ctx, span := WithOperationSpan(ctx, "test")
		defer span.End()

		// Set attributes
		SetSpanAttributes(ctx,
			attribute.String("key1", "value1"),
			attribute.Int("key2", 42),
			attribute.Bool("key3", true),
		)

		span.End()

		spans := spanRecorder.Ended()
		require.Equal(t, 1, len(spans))
		attrs := spans[0].Attributes()
		require.Equal(t, 3, len(attrs))
		assert.Equal(t, attribute.String("key1", "value1"), attrs[0])
		assert.Equal(t, attribute.Int("key2", 42), attrs[1])
		assert.Equal(t, attribute.Bool("key3", true), attrs[2])
	})

	t.Run("no-op when no span in context", func(t *testing.T) {
		ctx := context.Background()

		// Should not panic
		assert.NotPanics(t, func() {
			SetSpanAttributes(ctx, attribute.String("key", "value"))
		})
	})
}

func TestAddSpanEvent(t *testing.T) {
	t.Run("adds event to active span", func(t *testing.T) {
		spanRecorder, _ := setupTestTracer(t)
		ctx := context.Background()

		ctx, span := WithOperationSpan(ctx, "test")
		defer span.End()

		// Add events
		AddSpanEvent(ctx, "step1 started")
		AddSpanEvent(ctx, "step2 completed")

		span.End()

		spans := spanRecorder.Ended()
		require.Equal(t, 1, len(spans))
		events := spans[0].Events()
		require.Equal(t, 2, len(events))
		assert.Equal(t, "step1 started", events[0].Name)
		assert.Equal(t, "step2 completed", events[1].Name)
	})

	t.Run("no-op when no span in context", func(t *testing.T) {
		ctx := context.Background()

		// Should not panic
		assert.NotPanics(t, func() {
			AddSpanEvent(ctx, "test event")
		})
	})
}

func TestRecordSpanError(t *testing.T) {
	t.Run("records error on active span", func(t *testing.T) {
		spanRecorder, _ := setupTestTracer(t)
		ctx := context.Background()

		ctx, span := WithOperationSpan(ctx, "test")
		defer span.End()

		// Record error
		testErr := errors.New("test error")
		RecordSpanError(ctx, testErr)

		span.End()

		spans := spanRecorder.Ended()
		require.Equal(t, 1, len(spans))

		// Check status is error
		assert.Equal(t, codes.Error, spans[0].Status().Code)
		assert.Equal(t, "test error", spans[0].Status().Description)

		// Check error event
		events := spans[0].Events()
		require.Equal(t, 1, len(events))
		assert.Equal(t, "exception", events[0].Name)
	})

	t.Run("no-op when no span in context", func(t *testing.T) {
		ctx := context.Background()

		// Should not panic
		assert.NotPanics(t, func() {
			RecordSpanError(ctx, errors.New("test"))
		})
	})
}

func TestConvertFieldsToAttributes(t *testing.T) {
	t.Run("converts string fields", func(t *testing.T) {
		fields := []zap.Field{
			zap.String("key1", "value1"),
			zap.String("key2", "value2"),
		}

		attrs := convertFieldsToAttributes(fields)
		require.Equal(t, 2, len(attrs))
		assert.Equal(t, attribute.String("key1", "value1"), attrs[0])
		assert.Equal(t, attribute.String("key2", "value2"), attrs[1])
	})

	t.Run("converts integer fields", func(t *testing.T) {
		fields := []zap.Field{
			zap.Int("int", 42),
			zap.Int64("int64", 9999),
			zap.Int32("int32", 100),
		}

		attrs := convertFieldsToAttributes(fields)
		require.Equal(t, 3, len(attrs))
		assert.Equal(t, attribute.Int64("int", 42), attrs[0])
		assert.Equal(t, attribute.Int64("int64", 9999), attrs[1])
		assert.Equal(t, attribute.Int64("int32", 100), attrs[2])
	})

	t.Run("converts bool fields", func(t *testing.T) {
		fields := []zap.Field{
			zap.Bool("bool1", true),
			zap.Bool("bool2", false),
		}

		attrs := convertFieldsToAttributes(fields)
		require.Equal(t, 2, len(attrs))
		assert.Equal(t, attribute.Bool("bool1", true), attrs[0])
		assert.Equal(t, attribute.Bool("bool2", false), attrs[1])
	})

	t.Run("converts float fields", func(t *testing.T) {
		fields := []zap.Field{
			zap.Float64("float64", 3.14),
			zap.Float64("another_float", 2.71),
		}

		attrs := convertFieldsToAttributes(fields)
		require.Equal(t, 2, len(attrs))
		assert.Equal(t, attribute.Key("float64"), attrs[0].Key)
		assert.InDelta(t, 3.14, attrs[0].Value.AsFloat64(), 0.001)
		assert.Equal(t, attribute.Key("another_float"), attrs[1].Key)
		assert.InDelta(t, 2.71, attrs[1].Value.AsFloat64(), 0.001)
	})

	t.Run("handles empty fields", func(t *testing.T) {
		fields := []zap.Field{}

		attrs := convertFieldsToAttributes(fields)
		assert.Nil(t, attrs)
	})

	t.Run("ignores complex types", func(t *testing.T) {
		fields := []zap.Field{
			zap.String("string", "value"),
			zap.Any("any", map[string]string{"key": "value"}),
			zap.Int("int", 42),
		}

		attrs := convertFieldsToAttributes(fields)
		// Should only have string and int, Any is ignored
		require.Equal(t, 2, len(attrs))
		assert.Equal(t, attribute.String("string", "value"), attrs[0])
		assert.Equal(t, attribute.Int64("int", 42), attrs[1])
	})

	t.Run("handles mixed field types", func(t *testing.T) {
		fields := []zap.Field{
			zap.String("name", "test"),
			zap.Int("count", 10),
			zap.Bool("active", true),
			zap.Float64("rate", 0.95),
		}

		attrs := convertFieldsToAttributes(fields)
		require.Equal(t, 4, len(attrs))
		assert.Equal(t, attribute.String("name", "test"), attrs[0])
		assert.Equal(t, attribute.Int64("count", 10), attrs[1])
		assert.Equal(t, attribute.Bool("active", true), attrs[2])
		assert.InDelta(t, 0.95, attrs[3].Value.AsFloat64(), 0.001)
	})
}

func TestTraceMetadataFields(t *testing.T) {
	t.Run("adds trace_id and span_id when span is present", func(t *testing.T) {
		spanRecorder, _ := setupTestTracer(t)
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		// Create span
		ctx, span := WithOperationSpan(ctx, "test")
		defer span.End()

		// Log something
		Info(ctx, "test message")

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]

		// Should have trace_id and span_id fields
		var hasTraceID, hasSpanID bool
		for _, field := range entry.Context {
			if field.Key == "trace_id" {
				hasTraceID = true
				assert.NotEmpty(t, field.String)
			}
			if field.Key == "span_id" {
				hasSpanID = true
				assert.NotEmpty(t, field.String)
			}
		}

		assert.True(t, hasTraceID, "should have trace_id field")
		assert.True(t, hasSpanID, "should have span_id field")

		span.End()
		spans := spanRecorder.Ended()
		require.Equal(t, 1, len(spans))
	})

	t.Run("no trace fields when no span", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		// Log without span
		Info(ctx, "test message")

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]

		// Should not have trace_id or span_id
		for _, field := range entry.Context {
			assert.NotEqual(t, "trace_id", field.Key)
			assert.NotEqual(t, "span_id", field.Key)
		}
	})
}

func TestMarkSpanError(t *testing.T) {
	t.Run("marks span as error when error field is present", func(t *testing.T) {
		spanRecorder, _ := setupTestTracer(t)
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		ctx, span := WithOperationSpan(ctx, "test")
		defer span.End()

		// Log error
		testErr := errors.New("test error")
		Error(ctx, "operation failed", zap.Error(testErr))

		require.Equal(t, 1, logs.Len())

		span.End()

		spans := spanRecorder.Ended()
		require.Equal(t, 1, len(spans))

		// Check span status
		assert.Equal(t, codes.Error, spans[0].Status().Code)
		assert.Equal(t, "operation failed", spans[0].Status().Description)
	})

	t.Run("does not mark span as error without error field", func(t *testing.T) {
		spanRecorder, _ := setupTestTracer(t)
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		ctx, span := WithOperationSpan(ctx, "test")
		defer span.End()

		// Log error without error field
		Error(ctx, "operation failed", zap.String("reason", "timeout"))

		require.Equal(t, 1, logs.Len())

		span.End()

		spans := spanRecorder.Ended()
		require.Equal(t, 1, len(spans))

		// Check span status should still be OK (Unset)
		assert.Equal(t, codes.Unset, spans[0].Status().Code)
	})
}
