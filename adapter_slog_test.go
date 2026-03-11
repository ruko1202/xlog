package xlog

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ruko1202/xlog/xfield"
)

func TestFieldToSlogAttr(t *testing.T) {
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
				name:  "Object",
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
				name:  "Durations",
				field: xfield.Durations("key", []time.Duration{time.Second, time.Millisecond}),
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				slogAttr := fieldToSlogAttr(tc.field)
				assert.Equal(t, tc.field.Key, slogAttr.Key)
			})
		}
	})
}

func TestFieldToSlogAttr_TypeConversions(t *testing.T) {
	t.Run("String type", func(t *testing.T) {
		field := xfield.String("name", "test")
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "name", attr.Key)
		assert.Equal(t, "test", attr.Value.String())
	})

	t.Run("Int64 type", func(t *testing.T) {
		field := xfield.Int64("count", 42)
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "count", attr.Key)
		assert.Equal(t, int64(42), attr.Value.Int64())
	})

	t.Run("Uint64 type", func(t *testing.T) {
		field := xfield.Uint64("id", 100)
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "id", attr.Key)
		assert.Equal(t, uint64(100), attr.Value.Uint64())
	})

	t.Run("Float64 type", func(t *testing.T) {
		field := xfield.Float64("pi", 3.14159)
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "pi", attr.Key)
		assert.InDelta(t, 3.14159, attr.Value.Any().(float64), 0.00001)
	})

	t.Run("Bool type", func(t *testing.T) {
		field := xfield.Bool("active", true)
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "active", attr.Key)
		assert.True(t, attr.Value.Bool())
	})

	t.Run("Time type", func(t *testing.T) {
		now := time.Now()
		field := xfield.Time("timestamp", now)
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "timestamp", attr.Key)
		assert.Equal(t, now.UnixNano(), attr.Value.Time().UnixNano())
	})

	t.Run("Duration type", func(t *testing.T) {
		dur := 5 * time.Second
		field := xfield.Duration("elapsed", dur)
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "elapsed", attr.Key)
		assert.Equal(t, dur, attr.Value.Duration())
	})

	t.Run("Error type with valid error", func(t *testing.T) {
		err := errors.New("test error")
		field := xfield.Error(err)
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "error", attr.Key)
		assert.Equal(t, err, attr.Value.Any())
	})

	t.Run("Error type with nil error", func(t *testing.T) {
		field := xfield.Error(nil)
		attr := fieldToSlogAttr(field)

		// Should return empty attr for nil errors
		assert.Empty(t, attr.Key)
	})

	t.Run("Binary type", func(t *testing.T) {
		data := []byte("hello")
		field := xfield.Binary("data", data)
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "data", attr.Key)
		assert.Equal(t, data, attr.Value.Any())
	})

	t.Run("Any type", func(t *testing.T) {
		value := map[string]string{"key": "value"}
		field := xfield.Any("metadata", value)
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "metadata", attr.Key)
		assert.Equal(t, value, attr.Value.Any())
	})

	t.Run("Array types", func(t *testing.T) {
		field := xfield.Strings("tags", []string{"go", "logging", "slog"})
		attr := fieldToSlogAttr(field)

		assert.Equal(t, "tags", attr.Key)
		assert.Equal(t, []string{"go", "logging", "slog"}, attr.Value.Any())
	})
}

func TestFieldsToSlogAttrs(t *testing.T) {
	t.Run("converts multiple fields", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.String("name", "test"),
			xfield.Int64("count", 42),
			xfield.Bool("active", true),
		}

		attrs := fieldsToSlogAttrs(fields)

		assert.Len(t, attrs, 3)

		// Check first attr
		attr0 := attrs[0].(slog.Attr)
		assert.Equal(t, "name", attr0.Key)
		assert.Equal(t, "test", attr0.Value.String())

		// Check second attr
		attr1 := attrs[1].(slog.Attr)
		assert.Equal(t, "count", attr1.Key)
		assert.Equal(t, int64(42), attr1.Value.Int64())

		// Check third attr
		attr2 := attrs[2].(slog.Attr)
		assert.Equal(t, "active", attr2.Key)
		assert.True(t, attr2.Value.Bool())
	})

	t.Run("returns nil for empty fields", func(t *testing.T) {
		attrs := fieldsToSlogAttrs(nil)
		assert.Nil(t, attrs)
	})

	t.Run("skips fields with empty keys", func(t *testing.T) {
		fields := []xfield.Field{
			xfield.String("valid", "test"),
			xfield.Error(nil), // This will produce empty key
		}

		attrs := fieldsToSlogAttrs(fields)

		// Only one valid attr should be present
		assert.Len(t, attrs, 1)
		attr0 := attrs[0].(slog.Attr)
		assert.Equal(t, "valid", attr0.Key)
	})
}

func initSlogAdapter(t *testing.T) (Logger, logObserver) {
	t.Helper()

	logs := make([]*logEntry, 0)
	handler := &testSlogHandler{
		level: slog.LevelDebug,
		logs:  &logs,
		attrs: make([]slog.Attr, 0),
	}

	logger := slog.New(handler)

	adapter := NewSlogAdapter(logger,
		WithExitFunc(func() {
			// Don't actually exit in tests
		}),
		WithPanicFunc(func(_ string) {
			// Don't actually panic in tests
		}),
	)

	return adapter, func() []*logEntry {
		return handler.entries()
	}
}

// testSlogHandler is a custom slog.Handler that captures log entries for testing.
type testSlogHandler struct {
	level slog.Level
	logs  *[]*logEntry
	attrs []slog.Attr
	group string
}

func (h *testSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *testSlogHandler) Handle(_ context.Context, r slog.Record) error {
	contextMap := make(map[string]interface{})

	// Add pre-attached attributes
	for _, attr := range h.attrs {
		addAttrToMap(contextMap, attr)
	}

	// Add record attributes
	r.Attrs(func(attr slog.Attr) bool {
		addAttrToMap(contextMap, attr)
		return true
	})

	// Extract logger name from "logger" field if present
	loggerName := ""
	if name, ok := contextMap["logger"].(string); ok {
		loggerName = name
		delete(contextMap, "logger") // Remove it from context map to match zap behavior
	}

	// Check for special _level marker to distinguish Fatal/Panic from Error
	zapLevel := mapSlogLevel(r.Level)
	if levelMarker, ok := contextMap["_level"].(string); ok {
		switch levelMarker {
		case "fatal":
			zapLevel = fatalLevel
		case "panic":
			zapLevel = panicLevel
		}
		delete(contextMap, "_level") // Remove marker from context map
	}

	entry := &logEntry{
		Level:      zapLevel,
		Time:       r.Time,
		LoggerName: loggerName,
		Message:    r.Message,
		ContextMap: contextMap,
	}

	*h.logs = append(*h.logs, entry)
	return nil
}

func (h *testSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	// Return a new handler that shares the same logs slice
	return &testSlogHandler{
		level: h.level,
		logs:  h.logs, // Share the same slice
		attrs: newAttrs,
		group: h.group,
	}
}

func (h *testSlogHandler) WithGroup(name string) slog.Handler {
	// Return a new handler that shares the same logs slice
	return &testSlogHandler{
		level: h.level,
		logs:  h.logs, // Share the same slice
		attrs: h.attrs,
		group: name,
	}
}

func (h *testSlogHandler) entries() []*logEntry {
	return *h.logs
}

func mapSlogLevel(level slog.Level) int {
	switch level {
	case slog.LevelDebug:
		return debugLevel
	case slog.LevelInfo:
		return infoLevel
	case slog.LevelWarn:
		return warnLevel
	case slog.LevelError:
		return errorLevel
	default:
		return infoLevel
	}
}

func addAttrToMap(m map[string]interface{}, attr slog.Attr) {
	key := attr.Key
	val := attr.Value.Any()

	if err, ok := val.(error); ok {
		m[key] = err.Error()
	} else {
		m[key] = val
	}
}
