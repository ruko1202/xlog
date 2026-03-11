package xlog

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/ruko1202/xlog/xfield"
)

// SlogOption is a function that configures a SlogAdapter.
type SlogOption func(*SlogAdapter)

// WithExitFunc sets a custom exit function (for testing).
func WithExitFunc(fn func()) SlogOption {
	return func(s *SlogAdapter) {
		s.exitFunc = fn
	}
}

// WithPanicFunc sets a custom panic function (for testing).
func WithPanicFunc(fn func(string)) SlogOption {
	return func(s *SlogAdapter) {
		s.panicFunc = fn
	}
}

// SlogAdapter adapts a slog.Logger to the xlog.Logger interface.
type SlogAdapter struct {
	logger    *slog.Logger
	ctx       context.Context // context for slog operations
	exitFunc  func()          // function to call instead of os.Exit (for testing)
	panicFunc func(string)    // function to call instead of panic (for testing)
}

// NewSlogAdapter creates a new SlogAdapter wrapping the given slog.Logger.
func NewSlogAdapter(logger *slog.Logger, options ...SlogOption) Logger {
	return NewSlogAdapterWithContext(context.Background(), logger, options...)
}

// NewSlogAdapterWithContext creates a new SlogAdapter with a context.
// If logger is nil, it uses slog.Default().
// The context is used for all logging operations.
func NewSlogAdapterWithContext(ctx context.Context, logger *slog.Logger, options ...SlogOption) Logger {
	if logger == nil {
		logger = slog.Default()
	}
	adapter := &SlogAdapter{
		logger: logger,
		ctx:    ctx,
		exitFunc: func() {
			os.Exit(1)
		},
		panicFunc: func(msg string) {
			panic(msg)
		},
	}

	for _, opt := range options {
		opt(adapter)
	}

	return adapter
}

// Debug logs a debug-level message.
func (s *SlogAdapter) Debug(msg string, fields ...xfield.Field) {
	s.logger.DebugContext(s.ctx, msg, fieldsToSlogAttrs(fields)...)
}

// Info logs an info-level message.
func (s *SlogAdapter) Info(msg string, fields ...xfield.Field) {
	s.logger.InfoContext(s.ctx, msg, fieldsToSlogAttrs(fields)...)
}

// Warn logs a warning-level message.
func (s *SlogAdapter) Warn(msg string, fields ...xfield.Field) {
	s.logger.WarnContext(s.ctx, msg, fieldsToSlogAttrs(fields)...)
}

// Error logs an error-level message.
func (s *SlogAdapter) Error(msg string, fields ...xfield.Field) {
	s.logger.ErrorContext(s.ctx, msg, fieldsToSlogAttrs(fields)...)
}

// Fatal logs a fatal-level message and terminates the program.
// Note: slog doesn't have a Fatal level, so we log as Error with a special marker and exit.
func (s *SlogAdapter) Fatal(msg string, fields ...xfield.Field) {
	attrs := fieldsToSlogAttrs(fields)
	attrs = append(attrs, slog.String("_level", "fatal"))
	s.logger.ErrorContext(s.ctx, msg, attrs...)
	s.exitFunc()
}

// Panic logs a panic-level message and panics.
// Note: slog doesn't have a Panic level, so we log as Error with a special marker and panic.
func (s *SlogAdapter) Panic(msg string, fields ...xfield.Field) {
	attrs := fieldsToSlogAttrs(fields)
	attrs = append(attrs, slog.String("_level", "panic"))
	s.logger.ErrorContext(s.ctx, msg, attrs...)
	s.panicFunc(msg)
}

// With creates a child logger with pre-attached fields.
func (s *SlogAdapter) With(fields ...xfield.Field) Logger {
	return s.WithContext(s.ctx, fields...)
}

// Named creates a child logger with the given name.
// In slog, this is implemented by adding a "logger" field with the name.
func (s *SlogAdapter) Named(name string) Logger {
	return &SlogAdapter{
		logger:    s.logger.With(slog.String("logger", name)),
		ctx:       s.ctx,
		exitFunc:  s.exitFunc,
		panicFunc: s.panicFunc,
	}
}

// Sync flushes any buffered log entries.
// Note: slog doesn't have a Sync method, so this is a no-op.
func (s *SlogAdapter) Sync() error {
	return nil
}

// Unwrap returns the underlying slog.Logger.
// This is useful for cases where you need direct access to slog-specific features.
func (s *SlogAdapter) Unwrap() *slog.Logger {
	return s.logger
}

// WithContext returns a new adapter with the given context.
func (s *SlogAdapter) WithContext(ctx context.Context, fields ...xfield.Field) Logger {
	return &SlogAdapter{
		logger:    s.logger.With(fieldsToSlogAttrs(fields)...),
		ctx:       ctx,
		exitFunc:  s.exitFunc,
		panicFunc: s.panicFunc,
	}
}

// fieldsToSlogAttrs converts xlog.Field slice to slog.Attr slice.
func fieldsToSlogAttrs(fields []xfield.Field) []any {
	if len(fields) == 0 {
		return nil
	}

	// slog.Logger methods accept ...any (alternating keys and values)
	// But we'll convert to slog.Attr for better type safety
	attrs := make([]any, 0, len(fields))
	for _, f := range fields {
		attr := fieldToSlogAttr(f)
		if attr.Key != "" { // Skip empty attributes
			attrs = append(attrs, attr)
		}
	}

	return attrs
}

// fieldToSlogAttr converts a single xlog.Field to slog.Attr.
func fieldToSlogAttr(f xfield.Field) slog.Attr {
	switch f.Type {
	case xfield.StringType:
		return slog.String(f.Key, f.String)

	case xfield.Int64Type:
		return slog.Int64(f.Key, f.Integer)

	case xfield.Uint64Type:
		// #nosec G115 - safe conversion as Uint64 values are stored as int64
		return slog.Uint64(f.Key, uint64(f.Integer))

	case xfield.Float64Type:
		return slog.Float64(f.Key, f.Float)

	case xfield.BoolType:
		return slog.Bool(f.Key, f.Integer == 1)

	case xfield.TimeType:
		if t, ok := f.Interface.(time.Time); ok {
			return slog.Time(f.Key, t)
		}
		// Fallback: use nanoseconds stored in Integer
		return slog.Time(f.Key, time.Unix(0, f.Integer))

	case xfield.DurationType:
		return slog.Duration(f.Key, time.Duration(f.Integer))

	case xfield.ErrorType:
		if err, ok := f.Interface.(error); ok && err != nil {
			return slog.Any(f.Key, err)
		}
		// Fallback for nil errors - skip
		return slog.Attr{}

	case xfield.ArrayType, xfield.BinaryType, xfield.ObjectType, xfield.AnyType:
		return slog.Any(f.Key, f.Interface)

	default:
		// Unknown type: use Any as fallback
		return slog.Any(f.Key, f.Interface)
	}
}
