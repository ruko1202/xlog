package xlog

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type loggerCalls struct {
	log  func(ctx context.Context, msg string, fields ...Field)
	logf func(ctx context.Context, template string, args ...any)
}

var loggers = map[zapcore.Level]*loggerCalls{
	zapcore.DebugLevel: {log: Debug, logf: Debugf},
	zapcore.InfoLevel:  {log: Info, logf: Infof},
	zapcore.WarnLevel:  {log: Warn, logf: Warnf},
	zapcore.ErrorLevel: {log: Error, logf: Errorf},
	zapcore.FatalLevel: {log: Fatal, logf: Fatalf},
	zapcore.PanicLevel: {log: Panic, logf: Panicf},
}

func TestLogger(t *testing.T) {
	for level, calls := range loggers {
		testLogger(t, level, calls)
	}
}

func testLogger(t *testing.T, level zapcore.Level, calls *loggerCalls) {
	t.Helper()
	ctx := context.Background()

	t.Run(level.String(), func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			logger, logs := initTestLogger(t)
			ctx = ContextWithLogger(ctx, logger)

			message := fmt.Sprintf("test %s message", level)

			calls.log(ctx, message, StringField("key", "value"))

			require.Equal(t, 1, logs.Len())
			entry := logs.All()[0]
			assert.Equal(t, level, entry.Level)
			assert.Equal(t, message, entry.Message)
			assert.Equal(t, 1, len(entry.Context))
			assert.Equal(t, "key", entry.Context[0].Key)
			assert.Equal(t, "value", entry.Context[0].String)
		})
		t.Run("with F", func(t *testing.T) {
			logger, logs := initTestLogger(t)
			ctx = ContextWithLogger(ctx, logger)

			messageTemplate := fmt.Sprintf("test %s message", level) + " %s %d"
			args := []any{"formatted", 42}

			calls.logf(ctx, messageTemplate, args...)

			require.Equal(t, 1, logs.Len())
			entry := logs.All()[0]
			assert.Equal(t, level, entry.Level)
			assert.Equal(t, fmt.Sprintf(messageTemplate, args...), entry.Message)
		})
	})
}

func TestUseGlobalLogger(t *testing.T) {
	logger, logs := initTestLogger(t)
	returnToPrev := ReplaceGlobal(logger)
	t.Cleanup(returnToPrev)

	ctx := context.Background()

	Info(ctx, "info message")

	require.Equal(t, 1, logs.Len())
}
