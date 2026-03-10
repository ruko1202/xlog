package xlog

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
)

func setupBenchTracer() {
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanRecorder),
	)
	otel.SetTracerProvider(tracerProvider)
}

func BenchmarkWithOperationSpan(b *testing.B) {
	setupBenchTracer()
	logger := zap.NewNop()
	ctx := ContextWithLogger(context.Background(), NewZapAdapter(logger))

	b.Run("without fields", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := WithOperationSpan(ctx, "test-operation")
			span.End()
		}
	})

	b.Run("with 3 fields", func(b *testing.B) {
		fields := []Field{
			String("user_id", "12345"),
			Int("count", 42),
			Bool("active", true),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := WithOperationSpan(ctx, "test-operation", fields...)
			span.End()
		}
	})

	b.Run("with 10 fields", func(b *testing.B) {
		fields := []Field{
			String("field1", "value1"),
			String("field2", "value2"),
			String("field3", "value3"),
			Int("field4", 1),
			Int("field5", 2),
			Int("field6", 3),
			Bool("field7", true),
			Bool("field8", false),
			Float64("field9", 3.14),
			Float64("field10", 2.71),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := WithOperationSpan(ctx, "test-operation", fields...)
			span.End()
		}
	})
}

func BenchmarkSetSpanAttributes(b *testing.B) {
	setupBenchTracer()
	ctx := context.Background()

	ctx, span := WithOperationSpan(ctx, "test")
	defer span.End()

	b.Run("single attribute", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			SetSpanAttributes(ctx, attribute.String("key", "value"))
		}
	})

	b.Run("3 attributes", func(b *testing.B) {
		attrs := []attribute.KeyValue{
			attribute.String("key1", "value1"),
			attribute.Int("key2", 42),
			attribute.Bool("key3", true),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			SetSpanAttributes(ctx, attrs...)
		}
	})

	b.Run("10 attributes", func(b *testing.B) {
		attrs := []attribute.KeyValue{
			attribute.String("key1", "value1"),
			attribute.String("key2", "value2"),
			attribute.String("key3", "value3"),
			attribute.Int("key4", 1),
			attribute.Int("key5", 2),
			attribute.Int("key6", 3),
			attribute.Bool("key7", true),
			attribute.Bool("key8", false),
			attribute.Float64("key9", 3.14),
			attribute.Float64("key10", 2.71),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			SetSpanAttributes(ctx, attrs...)
		}
	})

	b.Run("no span in context", func(b *testing.B) {
		emptyCtx := context.Background()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			SetSpanAttributes(emptyCtx, attribute.String("key", "value"))
		}
	})
}

func BenchmarkAddSpanEvent(b *testing.B) {
	setupBenchTracer()
	ctx := context.Background()

	ctx, span := WithOperationSpan(ctx, "test")
	defer span.End()

	b.Run("simple event", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			AddSpanEvent(ctx, "test event")
		}
	})

	b.Run("no span in context", func(b *testing.B) {
		emptyCtx := context.Background()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			AddSpanEvent(emptyCtx, "test event")
		}
	})
}

func BenchmarkConvertFieldsToAttributes(b *testing.B) {
	b.Run("empty fields", func(b *testing.B) {
		fields := []Field{}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = fieldsToOtelAttributes(fields)
		}
	})

	b.Run("3 string fields", func(b *testing.B) {
		fields := []Field{
			String("key1", "value1"),
			String("key2", "value2"),
			String("key3", "value3"),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = fieldsToOtelAttributes(fields)
		}
	})

	b.Run("3 mixed fields", func(b *testing.B) {
		fields := []Field{
			String("string", "value"),
			Int("int", 42),
			Bool("bool", true),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = fieldsToOtelAttributes(fields)
		}
	})

	b.Run("10 mixed fields", func(b *testing.B) {
		fields := []Field{
			String("field1", "value1"),
			String("field2", "value2"),
			String("field3", "value3"),
			Int("field4", 1),
			Int("field5", 2),
			Int("field6", 3),
			Bool("field7", true),
			Bool("field8", false),
			Float64("field9", 3.14),
			Float64("field10", 2.71),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = fieldsToOtelAttributes(fields)
		}
	})

	b.Run("with unsupported types", func(b *testing.B) {
		fields := []Field{
			String("string", "value"),
			Any("any", map[string]string{"key": "value"}),
			Int("int", 42),
			Any("any2", []string{"a", "b", "c"}),
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = fieldsToOtelAttributes(fields)
		}
	})
}

func BenchmarkSpanFromContext(b *testing.B) {
	setupBenchTracer()

	b.Run("with span", func(b *testing.B) {
		ctx, span := WithOperationSpan(context.Background(), "test")
		defer span.End()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = SpanFromContext(ctx)
		}
	})

	b.Run("without span", func(b *testing.B) {
		ctx := context.Background()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = SpanFromContext(ctx)
		}
	})
}

// BenchmarkSpanCreation_Comparison: WithOperationSpan vs manual span creation.
func BenchmarkSpanCreation_Comparison(b *testing.B) {
	setupBenchTracer()
	zapLogger := zap.NewNop()
	ctx := ContextWithLogger(context.Background(), NewZapAdapter(zapLogger))
	tracer := otel.GetTracerProvider().Tracer("benchmark")

	fields := []Field{
		String("user_id", "12345"),
		Int("count", 42),
		Bool("active", true),
	}

	b.Run("xlog.WithOperationSpan", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := WithOperationSpan(ctx, "test-op", fields...)
			span.End()
		}
	})

	b.Run("manual span + logger", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			newCtx, span := tracer.Start(ctx, "test-op")
			span.SetAttributes(
				attribute.String("user_id", "12345"),
				attribute.Int("count", 42),
				attribute.Bool("active", true),
			)
			newLogger := zapLogger.Named("test-op").With(
				zap.String("user_id", "12345"),
				zap.Int("count", 42),
				zap.Bool("active", true),
			)
			_ = ContextWithLogger(newCtx, NewZapAdapter(newLogger))
			span.End()
		}
	})
}
