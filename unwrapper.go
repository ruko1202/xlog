package xlog

// Unwrapper is an optional interface that a Logger may implement to expose
// the underlying backend logger (e.g. *zap.Logger, *slog.Logger).
type Unwrapper[T any] interface {
	Unwrap() T
}

// Unwrap returns the underlying backend logger of type T.
// Returns (zero, false) if the logger does not implement Unwrapper or the type does not match.
//
// Example:
//
//	zapLogger, ok := xlog.Unwrap[*zap.Logger](logger)
//	slogLogger, ok := xlog.Unwrap[*slog.Logger](logger)
func Unwrap[T any](logger Logger) (T, bool) {
	u, ok := logger.(Unwrapper[T])
	if !ok {
		var zero T
		return zero, false
	}
	return u.Unwrap(), true
}
