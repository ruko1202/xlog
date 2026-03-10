package xlog

// NoopLogger is a logger that does nothing.
// This is used as a fallback when no logger is configured.
type NoopLogger struct{}

// NewNoopLogger creates a new no-op logger.
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

func (l *NoopLogger) Debug(msg string, fields ...Field) {}
func (l *NoopLogger) Info(msg string, fields ...Field)  {}
func (l *NoopLogger) Warn(msg string, fields ...Field)  {}
func (l *NoopLogger) Error(msg string, fields ...Field) {}
func (l *NoopLogger) Fatal(msg string, fields ...Field) {}
func (l *NoopLogger) Panic(msg string, fields ...Field) { panic(msg) }
func (l *NoopLogger) With(fields ...Field) Logger       { return l }
func (l *NoopLogger) Named(name string) Logger          { return l }
func (l *NoopLogger) Sync() error                       { return nil }
