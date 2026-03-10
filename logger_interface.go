package xlog

// Logger is the interface that wraps the basic logging methods.
// This interface allows xlog to work with any logging backend (zap, slog, logrus, etc).
type Logger interface {
	// Debug logs a debug-level message with structured fields.
	Debug(msg string, fields ...Field)

	// Info logs an info-level message with structured fields.
	Info(msg string, fields ...Field)

	// Warn logs a warning-level message with structured fields.
	Warn(msg string, fields ...Field)

	// Error logs an error-level message with structured fields.
	Error(msg string, fields ...Field)

	// Fatal logs a fatal-level message with structured fields and terminates the program.
	Fatal(msg string, fields ...Field)

	// Panic logs a panic-level message with structured fields and panics.
	Panic(msg string, fields ...Field)

	// With creates a child logger with the given fields pre-attached.
	// All subsequent logs from this logger will include these fields.
	With(fields ...Field) Logger

	// Named creates a child logger with the given name appended.
	// This is useful for adding operation or component names to logs.
	Named(name string) Logger

	// Sync flushes any buffered log entries.
	// Applications should call Sync before exiting to ensure all logs are written.
	Sync() error
}
