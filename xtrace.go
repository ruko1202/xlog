package xlog

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type xTraceCtxKey int

const (
	tracerCtxKey xTraceCtxKey = iota
)

var (
	_tracerMu   sync.RWMutex
	_tracerName = "github.com/ruko1202/xlog"
)

// ReplaceTracerName sets the global tracer name used when creating new tracers.
// This should be called during application initialization before any tracing operations.
// The default tracer name is "github.com/ruko1202/xlog".
// This function is thread-safe and can be called concurrently.
// Returns the previous tracer name.
func ReplaceTracerName(name string) func() {
	if name == "" {
		return func() {}
	}

	_tracerMu.Lock()
	prev := _tracerName
	_tracerName = name
	_tracerMu.Unlock()
	return func() { ReplaceTracerName(prev) }
}

func getTracerName() string {
	_tracerMu.RLock()
	defer _tracerMu.RUnlock()
	return _tracerName
}

// ContextWithTracer returns a new context with the provided tracer attached.
// The tracer can be retrieved later using TracerFromContext.
func ContextWithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	if tracer == nil {
		tracer = otel.GetTracerProvider().Tracer(getTracerName())
	}
	return context.WithValue(ctx, tracerCtxKey, tracer)
}

// TracerFromContext extracts a tracer from the context.
// If no tracer is found, returns the global tracer from otel.GetTracerProvider().
//
// Example:
//
//	tracer := xlog.TracerFromContext(ctx)
//	ctx, span := tracer.Start(ctx, "my-operation")
//	defer span.End()
func TracerFromContext(ctx context.Context) trace.Tracer {
	return tracerFromContext(ctx)
}

func tracerFromContext(ctx context.Context) trace.Tracer {
	tracer, ok := ctx.Value(tracerCtxKey).(trace.Tracer)
	if !ok {
		tracer = otel.GetTracerProvider().Tracer(getTracerName())
	}

	return tracer
}
