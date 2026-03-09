package xlog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestContextWithTracer(t *testing.T) {
	ctx := context.Background()
	tracer := noop.NewTracerProvider().Tracer("test-tracer")

	// Add tracer to context
	ctx = ContextWithTracer(ctx, tracer)

	// Verify it's stored
	value := ctx.Value(tracerCtxKey)
	require.NotNil(t, value, "tracer should be stored in context")

	storedTracer, ok := value.(trace.Tracer)
	require.True(t, ok, "stored value should be a trace.Tracer")
	assert.Equal(t, tracer, storedTracer, "stored tracer should match the one we set")
}

func TestTracerFromContext(t *testing.T) {
	t.Run("returns tracer from context when present", func(t *testing.T) {
		ctx := context.Background()
		customTracer := noop.NewTracerProvider().Tracer("custom-tracer")

		// Add custom tracer to context
		ctx = ContextWithTracer(ctx, customTracer)

		// Extract it
		extractedTracer := TracerFromContext(ctx)
		require.NotNil(t, extractedTracer)

		// Verify it's the same tracer we put in
		assert.Equal(t, customTracer, extractedTracer)
	})

	t.Run("returns global tracer when not in context", func(t *testing.T) {
		ctx := context.Background()

		// Don't add any tracer to context
		extractedTracer := TracerFromContext(ctx)
		require.NotNil(t, extractedTracer, "should return global tracer, not nil")

		// Verify it uses the global tracer provider
		tracerName := _tracerName.Load().(string)
		globalTracer := otel.GetTracerProvider().Tracer(tracerName)
		assert.Equal(t, globalTracer, extractedTracer)
	})
}

func TestTracerFromContext_Internal(t *testing.T) {
	t.Run("internal function works correctly", func(t *testing.T) {
		ctx := context.Background()
		customTracer := noop.NewTracerProvider().Tracer("internal-test")

		ctx = ContextWithTracer(ctx, customTracer)

		// Test internal function
		extractedTracer := tracerFromContext(ctx)
		assert.Equal(t, customTracer, extractedTracer)
	})
}

func TestReplaceTracerName(t *testing.T) {
	// Save original tracer name
	originalName := _tracerName.Load().(string)
	t.Cleanup(func() {
		// Restore original name after test
		_tracerName.Store(originalName)
	})

	t.Run("replaces tracer name", func(t *testing.T) {
		newName := "my-custom-service"
		ReplaceTracerName(newName)

		currentName := _tracerName.Load().(string)
		assert.Equal(t, newName, currentName)
	})

	t.Run("new tracer name is used by TracerFromContext", func(t *testing.T) {
		newName := "test-service-name"
		ReplaceTracerName(newName)

		ctx := context.Background()
		// Don't add custom tracer, so it uses global with custom name
		tracer := TracerFromContext(ctx)

		require.NotNil(t, tracer)

		// Verify the tracer name is used (we can't directly verify the name,
		// but we can verify the tracer is created with the global provider)
		globalTracer := otel.GetTracerProvider().Tracer(newName)
		assert.Equal(t, globalTracer, tracer)
	})
}

func TestReplaceTracerName_Concurrent(t *testing.T) {
	// Save original tracer name
	originalName := _tracerName.Load().(string)
	t.Cleanup(func() {
		_tracerName.Store(originalName)
	})

	// This test should be run with -race flag to detect race conditions
	t.Run("concurrent access is safe", func(t *testing.T) {
		done := make(chan bool)
		ctx := context.Background()

		// Writer goroutine - changes tracer name
		go func() {
			for i := 0; i < 100; i++ {
				ReplaceTracerName("writer-tracer")
			}
			done <- true
		}()

		// Reader goroutine - reads tracer from context
		go func() {
			for i := 0; i < 100; i++ {
				_ = TracerFromContext(ctx)
			}
			done <- true
		}()

		// Wait for both goroutines
		<-done
		<-done
	})
}

func TestContextWithTracer_NilContext(t *testing.T) {
	// This should panic according to context.WithValue documentation
	tracer := noop.NewTracerProvider().Tracer("test")

	assert.Panics(t, func() {
		_ = ContextWithTracer(nil, tracer)
	}, "should panic when context is nil")
}

func TestTracerFromContext_NilContext(t *testing.T) {
	// This should panic according to context.Value documentation
	assert.Panics(t, func() {
		_ = TracerFromContext(nil)
	}, "should panic when context is nil")
}
