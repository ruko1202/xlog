package xlog

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

type checkWriteHookFunc func(*zapcore.CheckedEntry, []zap.Field)

func (f checkWriteHookFunc) OnWrite(entry *zapcore.CheckedEntry, fields []zap.Field) {
	f(entry, fields)
}

func initTestLogger(t *testing.T) (*zap.Logger, *observer.ObservedLogs) {
	t.Helper()

	observerCore, logs := observer.New(zapcore.DebugLevel)
	logger := zaptest.NewLogger(t,
		zaptest.WrapOptions(
			zap.WrapCore(func(_ zapcore.Core) zapcore.Core {
				return observerCore
			}),
			zap.WithPanicHook(checkWriteHookFunc(func(entry *zapcore.CheckedEntry, fields []zap.Field) {
				t.Logf("%#v, %#v", entry, fields)
			})),
			zap.WithFatalHook(checkWriteHookFunc(func(entry *zapcore.CheckedEntry, fields []zap.Field) {
				t.Logf("%#v, %#v", entry, fields)
			})),
		),
	)

	return logger, logs
}
