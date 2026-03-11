package xlog

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	"github.com/ruko1202/xlog/xfield"
)

func TestFieldToZapField(t *testing.T) {
	t.Run("converts all basic types", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			field xfield.Field
		}{
			{
				name:  "String",
				field: xfield.String("key", "value"),
			}, {
				name:  "Int64",
				field: xfield.Int64("key", 123),
			}, {
				name:  "Uint64",
				field: xfield.Uint64("key", 456),
			}, {
				name:  "Float64",
				field: xfield.Float64("key", 3.14),
			}, {
				name:  "Bool",
				field: xfield.Bool("key", true),
			}, {
				name:  "Time",
				field: xfield.Time("key", time.Now()),
			}, {
				name:  "Duration",
				field: xfield.Duration("key", 5*time.Second),
			}, {
				name:  "Error",
				field: xfield.Error(errors.New("test")),
			}, {
				name:  "Binary",
				field: xfield.Binary("key", []byte("a")),
			}, {
				name:  "Any",
				field: xfield.Any("key", []byte("a")),
			}, {
				name:  "Any",
				field: xfield.Object("key", []byte("a")),
			}, {
				name:  "Strings",
				field: xfield.Strings("key", []string{"a", "b", "c"}),
			}, {
				name:  "Ints",
				field: xfield.Ints("key", []int{1, 2, 3}),
			}, {
				name:  "Int32s",
				field: xfield.Int32s("key", []int32{1, 2, 3}),
			}, {
				name:  "Int64s",
				field: xfield.Int64s("key", []int64{1, 2, 3}),
			}, {
				name:  "UInts",
				field: xfield.UInts("key", []uint{1, 2, 3}),
			}, {
				name:  "UInt32s",
				field: xfield.UInt32s("key", []uint32{1, 2, 3}),
			}, {
				name:  "UInt64s",
				field: xfield.UInt64s("key", []uint64{1, 2, 3}),
			}, {
				name:  "Float32s",
				field: xfield.Float32s("key", []float32{1.1, 2.2, 3.3}),
			}, {
				name:  "Float64s",
				field: xfield.Float64s("key", []float64{1.1, 2.2, 3.3}),
			}, {
				name:  "Bools",
				field: xfield.Bools("key", []bool{true, false}),
			}, {
				name:  "Bools",
				field: xfield.Durations("key", []time.Duration{time.Second, time.Millisecond}),
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				zapField := fieldToZapField(tc.field)
				assert.Equal(t, tc.field.Key, zapField.Key)
				assert.Equal(t, tc.field.String, zapField.String)
			})
		}
	})
}

func initZapAdapter(t *testing.T) (Logger, logObserver) {
	t.Helper()
	logger, zapObservedLogs := initZapTestLogger(t)

	return NewZapAdapter(logger), func() []*logEntry {
		zapEntries := zapObservedLogs.All()

		entries := make([]*logEntry, 0, len(zapEntries))
		for _, e := range zapEntries {
			entries = append(entries, &logEntry{
				Level:      mapZapLevel(e.Level),
				Message:    e.Message,
				Time:       e.Time,
				LoggerName: e.LoggerName,
				ContextMap: e.ContextMap(),
			})
		}

		return entries
	}
}

func mapZapLevel(level zapcore.Level) int {
	switch level {
	case zapcore.DebugLevel:
		return debugLevel
	case zapcore.InfoLevel:
		return infoLevel
	case zapcore.WarnLevel:
		return warnLevel
	case zapcore.ErrorLevel:
		return errorLevel
	case zapcore.PanicLevel:
		return panicLevel
	case zapcore.FatalLevel:
		return fatalLevel
	default:
		return infoLevel
	}
}
