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
		tracerName := getTracerName()
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
	originalName := getTracerName()
	t.Cleanup(func() {
		// Restore original name after test
		ReplaceTracerName(originalName)
	})

	t.Run("replaces tracer name", func(t *testing.T) {
		newName := "my-custom-service"
		restore := ReplaceTracerName(newName)
		defer restore()

		currentName := getTracerName()
		assert.Equal(t, newName, currentName)
	})

	t.Run("new tracer name is used by TracerFromContext", func(t *testing.T) {
		newName := "test-service-name"
		restore := ReplaceTracerName(newName)
		defer restore()

		ctx := context.Background()
		// Don't add custom tracer, so it uses global with custom name
		tracer := TracerFromContext(ctx)

		require.NotNil(t, tracer)

		// Verify the tracer name is used (we can't directly verify the name,
		// but we can verify the tracer is created with the global provider)
		globalTracer := otel.GetTracerProvider().Tracer(newName)
		assert.Equal(t, globalTracer, tracer)
	})

	t.Run("restore function works", func(t *testing.T) {
		original := getTracerName()

		// Change name
		restore := ReplaceTracerName("temporary-name")

		assert.Equal(t, "temporary-name", getTracerName())

		// Restore
		restore()

		assert.Equal(t, original, getTracerName())
	})
}

func TestReplaceTracerName_Concurrent(t *testing.T) {
	// Save original tracer name
	originalName := getTracerName()
	t.Cleanup(func() {
		ReplaceTracerName(originalName)
	})

	// This test should be run with -race flag to detect race conditions
	t.Run("concurrent access is safe", func(_ *testing.T) {
		done := make(chan bool)
		ctx := context.Background()

		// Writer goroutine - changes tracer name
		go func() {
			for i := 0; i < 100; i++ {
				restore := ReplaceTracerName("writer-tracer")
				restore()
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
		_ = ContextWithTracer(nil, tracer) //nolint:staticcheck // Intentionally testing nil context behavior
	}, "should panic when context is nil")
}

func TestTracerFromContext_NilContext(t *testing.T) {
	// This should panic according to context.Value documentation
	assert.Panics(t, func() {
		_ = TracerFromContext(nil) //nolint:staticcheck // Intentionally testing nil context behavior
	}, "should panic when context is nil")
}

func TestReplaceTracerName_Validation(t *testing.T) {
	// Save original
	originalName := getTracerName()
	t.Cleanup(func() {
		ReplaceTracerName(originalName)
	})

	t.Run("ignores empty tracer name", func(t *testing.T) {
		restore := ReplaceTracerName("test-name")
		defer restore()

		currentName := getTracerName()
		assert.Equal(t, "test-name", currentName)

		// Empty name should be ignored
		restoreEmpty := ReplaceTracerName("")
		defer restoreEmpty()

		currentName = getTracerName()
		assert.Equal(t, "test-name", currentName, "empty name should not replace existing name")
	})

	t.Run("accepts valid tracer name", func(t *testing.T) {
		assert.NotPanics(t, func() {
			restore := ReplaceTracerName("valid-tracer-name")
			defer restore()
		})
	})
}

func TestContextWithTracer_Validation(t *testing.T) {
	t.Run("creates global tracer when tracer is nil", func(t *testing.T) {
		ctx := context.Background()

		// Should not panic, but use global tracer provider
		assert.NotPanics(t, func() {
			newCtx := ContextWithTracer(ctx, nil)

			// Verify tracer was added from global provider
			tracer := newCtx.Value(tracerCtxKey).(trace.Tracer)
			require.NotNil(t, tracer)
		})
	})

	t.Run("panics when context is nil", func(t *testing.T) {
		tracer := noop.NewTracerProvider().Tracer("test")

		assert.Panics(t, func() {
			ContextWithTracer(nil, tracer) //nolint:staticcheck // Intentionally testing nil context behavior
		}, "should panic when context is nil")
	})
}
