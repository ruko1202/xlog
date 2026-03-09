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
	_tracerMu   sync.Mutex
	_tracerName = "github.com/ruko1202/xlog"
)

// ReplaceTracerName sets the global tracer name used when creating new tracers.
// This should be called during application initialization before any tracing operations.
// The default tracer name is "github.com/ruko1202/xlog".
func ReplaceTracerName(tracerName string) {
	_tracerMu.Lock()
	defer _tracerMu.Unlock()

	_tracerName = tracerName
}

// ContextWithTracer returns a new context with the provided tracer attached.
// The tracer can be retrieved later using TracerFromContext.
func ContextWithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerCtxKey, tracer)
}

// TracerFromContext extracts a tracer from the context.
// If no tracer is found, returns the global tracer from otel.GetTracerProvider().
func TracerFromContext(ctx context.Context) trace.Tracer {
	return tracerFromContext(ctx)
}

func tracerFromContext(ctx context.Context) trace.Tracer {
	tracer, ok := ctx.Value(loggerCtxKey).(trace.Tracer)
	if !ok {
		return otel.GetTracerProvider().Tracer(_tracerName)
	}

	return tracer
}
