package xlog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/ruko1202/xlog/xfield"
)

func TestWithOperation(t *testing.T) {
	t.Run("creates named logger with operation", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		// Create context with operation
		ctx = WithOperation(ctx, "payment-processing")

		// Log something
		Info(ctx, "processing payment")

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		assert.Equal(t, "payment-processing", entry.LoggerName)
		assert.Equal(t, "processing payment", entry.Message)
	})

	t.Run("creates named logger with operation and fields", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		// Create context with operation and fields
		ctx = WithOperation(ctx, "user-auth",
			xfield.String("user_id", "12345"),
			xfield.String("session_id", "sess_xyz"),
		)

		// Log something
		Info(ctx, "user authenticated")

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		assert.Equal(t, "user-auth", entry.LoggerName)
		assert.Equal(t, "user authenticated", entry.Message)

		// Check fields
		require.Equal(t, 2, len(entry.Context))
		assert.Equal(t, "user_id", entry.Context[0].Key)
		assert.Equal(t, "12345", entry.Context[0].String)
		assert.Equal(t, "session_id", entry.Context[1].Key)
		assert.Equal(t, "sess_xyz", entry.Context[1].String)
	})

	t.Run("works with global logger", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		restore := ReplaceGlobalLogger(logger)
		defer restore()

		ctx := context.Background() // No logger in context

		// Create context with operation
		ctx = WithOperation(ctx, "background-task")

		// Log something
		Info(ctx, "task started")

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		assert.Equal(t, "background-task", entry.LoggerName)
	})

	t.Run("nested operations", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		// First operation
		ctx = WithOperation(ctx, "http-handler")
		Info(ctx, "request received")

		// Nested operation
		ctx = WithOperation(ctx, "database")
		Info(ctx, "query executed")

		require.Equal(t, 2, logs.Len())
		assert.Equal(t, "http-handler", logs.All()[0].LoggerName)
		assert.Equal(t, "http-handler.database", logs.All()[1].LoggerName)
	})
}

func TestWithFields(t *testing.T) {
	t.Run("adds fields to logger", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		// Add fields to context
		ctx = WithFields(ctx,
			xfield.String("request_id", "req-123"),
			xfield.String("user_id", "user-456"),
		)

		// Log something
		Info(ctx, "processing request")

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		assert.Equal(t, "processing request", entry.Message)

		// Check fields
		require.Equal(t, 2, len(entry.Context))
		assert.Equal(t, "request_id", entry.Context[0].Key)
		assert.Equal(t, "req-123", entry.Context[0].String)
		assert.Equal(t, "user_id", entry.Context[1].Key)
		assert.Equal(t, "user-456", entry.Context[1].String)
	})

	t.Run("fields persist across multiple log calls", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		// Add fields
		ctx = WithFields(ctx, xfield.String("trace_id", "trace-xyz"))

		// Multiple log calls
		Info(ctx, "step 1")
		Info(ctx, "step 2")
		Info(ctx, "step 3")

		require.Equal(t, 3, logs.Len())
		for i := 0; i < 3; i++ {
			entry := logs.All()[i]
			require.Equal(t, 1, len(entry.Context))
			assert.Equal(t, "trace_id", entry.Context[0].Key)
			assert.Equal(t, "trace-xyz", entry.Context[0].String)
		}
	})

	t.Run("can add fields incrementally", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		// Add first set of fields
		ctx = WithFields(ctx, xfield.String("key1", "value1"))

		// Add more fields
		ctx = WithFields(ctx, xfield.String("key2", "value2"))

		// Log
		Info(ctx, "message")

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]

		// Should have both fields
		require.Equal(t, 2, len(entry.Context))
		assert.Equal(t, "key1", entry.Context[0].Key)
		assert.Equal(t, "value1", entry.Context[0].String)
		assert.Equal(t, "key2", entry.Context[1].Key)
		assert.Equal(t, "value2", entry.Context[1].String)
	})

	t.Run("works with global logger", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		restore := ReplaceGlobalLogger(logger)
		defer restore()

		ctx := context.Background() // No logger in context

		// Add fields
		ctx = WithFields(ctx, xfield.String("field", "value"))

		// Log
		Info(ctx, "test message")

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		require.Equal(t, 1, len(entry.Context))
		assert.Equal(t, "field", entry.Context[0].Key)
	})

	t.Run("supports various field types", func(t *testing.T) {
		logger, logs := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		ctx = WithFields(ctx,
			xfield.String("string", "value"),
			xfield.Int("int", 42),
			xfield.Bool("bool", true),
			xfield.Float64("float", 3.14),
		)

		Info(ctx, "test")

		require.Equal(t, 1, logs.Len())
		entry := logs.All()[0]
		require.Equal(t, 4, len(entry.Context))

		assert.Equal(t, "string", entry.Context[0].Key)
		assert.Equal(t, zapcore.StringType, entry.Context[0].Type)

		assert.Equal(t, "int", entry.Context[1].Key)
		assert.Equal(t, zapcore.Int64Type, entry.Context[1].Type)

		assert.Equal(t, "bool", entry.Context[2].Key)
		assert.Equal(t, zapcore.BoolType, entry.Context[2].Type)

		assert.Equal(t, "float", entry.Context[3].Key)
		assert.Equal(t, zapcore.Float64Type, entry.Context[3].Type)
	})
}

func TestLoggerFromContext_Public(t *testing.T) {
	t.Run("returns logger from context", func(t *testing.T) {
		logger, _ := initTestLogger(t)
		ctx := ContextWithLogger(context.Background(), logger)

		// Use public function
		extractedLogger := LoggerFromContext(ctx)
		require.NotNil(t, extractedLogger)
		assert.Equal(t, logger, extractedLogger)
	})

	t.Run("returns global logger when not in context", func(t *testing.T) {
		logger, _ := initTestLogger(t)
		restore := ReplaceGlobalLogger(logger)
		defer restore()

		ctx := context.Background()

		// Use public function
		extractedLogger := LoggerFromContext(ctx)
		require.NotNil(t, extractedLogger)
		// Should return the global logger we set
		assert.Equal(t, logger, extractedLogger)
	})
}
