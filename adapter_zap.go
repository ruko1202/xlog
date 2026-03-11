package xlog

import (
	"time"

	"go.uber.org/zap"

	"github.com/ruko1202/xlog/xfield"
)

// ZapAdapter adapts a zap.Logger to the xlog.Logger interface.
type ZapAdapter struct {
	logger *zap.Logger
}

// NewZapAdapter creates a new ZapAdapter wrapping the given zap.Logger.
func NewZapAdapter(logger *zap.Logger) Logger {
	if logger == nil {
		logger = zap.L()
	}
	return &ZapAdapter{logger: logger}
}

// Debug logs a debug-level message.
func (z *ZapAdapter) Debug(msg string, fields ...xfield.Field) {
	z.logger.Debug(msg, fieldsToZapFields(fields)...)
}

// Info logs an info-level message.
func (z *ZapAdapter) Info(msg string, fields ...xfield.Field) {
	z.logger.Info(msg, fieldsToZapFields(fields)...)
}

// Warn logs a warning-level message.
func (z *ZapAdapter) Warn(msg string, fields ...xfield.Field) {
	z.logger.Warn(msg, fieldsToZapFields(fields)...)
}

// Error logs an error-level message.
func (z *ZapAdapter) Error(msg string, fields ...xfield.Field) {
	z.logger.Error(msg, fieldsToZapFields(fields)...)
}

// Fatal logs a fatal-level message and terminates the program.
func (z *ZapAdapter) Fatal(msg string, fields ...xfield.Field) {
	z.logger.Fatal(msg, fieldsToZapFields(fields)...)
}

// Panic logs a panic-level message and panics.
func (z *ZapAdapter) Panic(msg string, fields ...xfield.Field) {
	z.logger.Panic(msg, fieldsToZapFields(fields)...)
}

// With creates a child logger with pre-attached fields.
func (z *ZapAdapter) With(fields ...xfield.Field) Logger {
	return &ZapAdapter{
		logger: z.logger.With(fieldsToZapFields(fields)...),
	}
}

// Named creates a child logger with the given name.
func (z *ZapAdapter) Named(name string) Logger {
	return &ZapAdapter{
		logger: z.logger.Named(name),
	}
}

// Sync flushes any buffered log entries.
func (z *ZapAdapter) Sync() error {
	return z.logger.Sync()
}

// Unwrap returns the underlying zap.Logger.
// This is useful for cases where you need direct access to zap-specific features.
func (z *ZapAdapter) Unwrap() *zap.Logger {
	return z.logger
}

// fieldsToZapFields converts xlog.Field slice to zap.Field slice.
func fieldsToZapFields(fields []xfield.Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, fieldToZapField(f))
	}

	return zapFields
}

// fieldToZapField converts a single xlog.Field to zap.Field.
//
//nolint:gocyclo,funlen // switch on field types requires many cases
func fieldToZapField(f xfield.Field) zap.Field {
	switch f.Type {
	case xfield.StringType:
		return zap.String(f.Key, f.String)

	case xfield.Int64Type:
		return zap.Int64(f.Key, f.Integer)

	case xfield.Uint64Type:
		// #nosec G115 - safe conversion as Uint64 values are stored as int64
		return zap.Uint64(f.Key, uint64(f.Integer))

	case xfield.Float64Type:
		return zap.Float64(f.Key, f.Float)

	case xfield.BoolType:
		return zap.Bool(f.Key, f.Integer == 1)

	case xfield.TimeType:
		if t, ok := f.Interface.(time.Time); ok {
			return zap.Time(f.Key, t)
		}
		// Fallback: use nanoseconds stored in Integer
		return zap.Time(f.Key, time.Unix(0, f.Integer))

	case xfield.DurationType:
		return zap.Duration(f.Key, time.Duration(f.Integer))

	case xfield.ErrorType:
		if err, ok := f.Interface.(error); ok && err != nil {
			return zap.Error(err)
		}
		// Fallback for nil errors
		return zap.Skip()

	case xfield.ArrayType:
		// Handle common array types
		switch v := f.Interface.(type) {
		case []string:
			return zap.Strings(f.Key, v)
		case []int:
			return zap.Ints(f.Key, v)
		case []int32:
			return zap.Int32s(f.Key, v)
		case []int64:
			return zap.Int64s(f.Key, v)
		case []uint:
			return zap.Uints(f.Key, v)
		case []uint32:
			return zap.Uint32s(f.Key, v)
		case []uint64:
			return zap.Uint64s(f.Key, v)
		case []float32:
			return zap.Float32s(f.Key, v)
		case []float64:
			return zap.Float64s(f.Key, v)
		case []bool:
			return zap.Bools(f.Key, v)
		case []time.Duration:
			return zap.Durations(f.Key, v)
		default:
			// Fallback to Any for unsupported array types
			return zap.Any(f.Key, v)
		}

	case xfield.BinaryType:
		if b, ok := f.Interface.([]byte); ok {
			return zap.Binary(f.Key, b)
		}
		return zap.Any(f.Key, f.Interface)

	case xfield.ObjectType:
		return zap.Any(f.Key, f.Interface)

	case xfield.AnyType:
		return zap.Any(f.Key, f.Interface)

	default:
		// Unknown type: use Any as fallback
		return zap.Any(f.Key, f.Interface)
	}
}
