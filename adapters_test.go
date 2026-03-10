package xlog

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	debugLevel = iota - 1
	infoLevel  = iota
	warnLevel  = iota
	errorLevel = iota
	panicLevel
	fatalLevel
)

type logEntry struct {
	Level      int
	Time       time.Time
	LoggerName string
	Message    string
	ContextMap map[string]interface{}
}

type logObserver func() []*logEntry

func TestAdapters(t *testing.T) {
	t.Run("zap", func(t *testing.T) {
		testAdapter(t, initZapAdapter)

		t.Run("Unwrap returns underlying logger", func(t *testing.T) {
			logger := zap.NewNop()
			adapter := NewZapAdapter(logger).(*ZapAdapter)

			unwrapped := adapter.Unwrap()
			assert.Equal(t, logger, unwrapped)
		})
	})

	t.Run("slog", func(t *testing.T) {
		testAdapter(t, initSlogAdapter)

		t.Run("Unwrap returns underlying logger", func(t *testing.T) {
			logger := slog.Default()
			adapter := NewSlogAdapter(logger).(*SlogAdapter)

			unwrapped := adapter.Unwrap()
			assert.Equal(t, logger, unwrapped)
		})
	})
}

func testAdapter(t *testing.T, initAdapter func(t *testing.T) (Logger, logObserver)) {
	t.Run("NewZapAdapter with nil logger", func(t *testing.T) {
		adapter := NewZapAdapter(nil)
		assert.NotNil(t, adapter)
		// Should not panic when logging
		assert.NotPanics(t, func() {
			adapter.Info("test message")
		})
	})

	t.Run("Debug level", func(t *testing.T) {
		adapter, getLogsFunc := initAdapter(t)

		adapter.Debug("debug message", String("key", "value"))

		entries := getLogsFunc()
		require.Len(t, entries, 1)
		assert.Equal(t, "debug message", entries[0].Message)
		assert.EqualValues(t, debugLevel, entries[0].Level)
		assert.Equal(t, "value", entries[0].ContextMap["key"])
	})

	t.Run("Info level", func(t *testing.T) {
		adapter, getLogsFunc := initAdapter(t)

		adapter.Info("info message", Int("count", 42))

		entries := getLogsFunc()
		require.Len(t, entries, 1)
		assert.Equal(t, "info message", entries[0].Message)
		assert.Equal(t, int64(42), entries[0].ContextMap["count"])
	})

	t.Run("Warn level", func(t *testing.T) {
		adapter, getLogsFunc := initAdapter(t)

		adapter.Warn("warning message", Bool("flag", true))

		entries := getLogsFunc()
		require.Len(t, entries, 1)
		assert.EqualValues(t, warnLevel, entries[0].Level)
		assert.True(t, entries[0].ContextMap["flag"].(bool))
	})

	t.Run("Error level", func(t *testing.T) {
		adapter, getLogsFunc := initAdapter(t)

		testErr := errors.New("test error")
		adapter.Error("error message", Err(testErr))

		entries := getLogsFunc()
		require.Len(t, entries, 1)
		assert.EqualValues(t, errorLevel, entries[0].Level)
		// zap.Error stores error as string in observer's ContextMap
		assert.Equal(t, "test error", entries[0].ContextMap["error"])
	})

	t.Run("Fatal level", func(t *testing.T) {
		adapter, getLogsFunc := initAdapter(t)

		testErr := errors.New("test error")
		adapter.Fatal("error message", Err(testErr))

		entries := getLogsFunc()
		require.Len(t, entries, 1)
		assert.EqualValues(t, fatalLevel, entries[0].Level)
		// zap.Error stores error as string in observer's ContextMap
		assert.Equal(t, "test error", entries[0].ContextMap["error"])
	})

	t.Run("Panic level", func(t *testing.T) {
		adapter, getLogsFunc := initAdapter(t)

		testErr := errors.New("test error")
		adapter.Panic("error message", Err(testErr))

		entries := getLogsFunc()
		require.Len(t, entries, 1)
		assert.EqualValues(t, panicLevel, entries[0].Level)
		// zap.Error stores error as string in observer's ContextMap
		assert.Equal(t, "test error", entries[0].ContextMap["error"])
	})

	t.Run("With creates child logger", func(t *testing.T) {
		adapter, getLogsFunc := initAdapter(t)

		childAdapter := adapter.With(String("service", "test"))
		childAdapter.Info("test message")

		entries := getLogsFunc()
		require.Len(t, entries, 1)
		assert.Equal(t, "test", entries[0].ContextMap["service"])
	})

	t.Run("Named creates named logger", func(t *testing.T) {
		adapter, getLogsFunc := initAdapter(t)

		namedAdapter := adapter.Named("myservice")
		namedAdapter.Info("test message")

		entries := getLogsFunc()
		require.Len(t, entries, 1)
		assert.Equal(t, "myservice", entries[0].LoggerName)
	})

	t.Run("Multiple field types", func(t *testing.T) {
		adapter, getLogsFunc := initAdapter(t)

		now := time.Now()
		adapter.Info("complex message",
			String("str", "value"),
			Int64("int", 123),
			Float64("float", 3.14),
			Bool("bool", true),
			Time("time", now),
			Duration("duration", 5*time.Second),
		)

		entries := getLogsFunc()
		require.Len(t, entries, 1)
		ctx := entries[0].ContextMap
		assert.Equal(t, "value", ctx["str"])
		assert.Equal(t, int64(123), ctx["int"])
		assert.InDelta(t, 3.14, ctx["float"], 0.001)
		assert.True(t, ctx["bool"].(bool))
		assert.Equal(t, now.UnixNano(), ctx["time"].(time.Time).UnixNano())
		assert.Equal(t, 5*time.Second, ctx["duration"])
	})
}
