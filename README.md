# xlog

[![CI](https://github.com/ruko1202/xlog/actions/workflows/ci.yml/badge.svg)](https://github.com/ruko1202/xlog/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ruko1202/xlog)](https://goreportcard.com/report/github.com/ruko1202/xlog)
![Coverage](https://img.shields.io/badge/Coverage-75.9%25-brightgreen)
[![Go Reference](https://pkg.go.dev/badge/github.com/ruko1202/xlog.svg)](https://pkg.go.dev/github.com/ruko1202/xlog)

A wrapper around [zap](https://github.com/uber-go/zap) for context-aware logging.

## Features

- Context-aware logging - logger is extracted from context
- Global logger fallback support with thread-safe replacement
- **Structured logging via `field` package** - clear separation between logging and field constructors
- Backend-agnostic field types - works with Zap, slog, or custom loggers
- Printf-style formatting (Debugf, Infof, etc.)
- All logging levels: Debug, Info, Warn, Error, Fatal, Panic
- Zero allocation when logger is not in context (uses noop adapter)
- **OpenTelemetry integration** - span management with automatic logger enrichment
- Distributed tracing support with span creation, events, and attributes

## Installation

```bash
go get github.com/ruko1202/xlog
```

## Quick Start

For a complete working example, see [example/app](example/app/main.go).

### Basic Usage

```go
package main

import (
	"context"
	"github.com/ruko1202/xlog"
	"github.com/ruko1202/xlog/xfield"
	"go.uber.org/zap"
)

const (
	appName = "xlogApp"
)

func main() {
	// Create logger and wrap it in adapter
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Add logger to context
	ctx := context.Background()
	ctx = xlog.ContextWithLogger(ctx, xlog.NewZapAdapter(logger))

	// Use logging
	xlog.Infof(ctx, "application `%s` started", appName)
	xlog.Debug(ctx, "debug information", field.String("version", "1.0.0"))
}
```

### Using Global Logger
It is convenient if you have background jobs that carry their own context.Context

```go
package main

import (
    "context"
    "github.com/ruko1202/xlog"
    "go.uber.org/zap"
)

func main() {
    // Replace global logger using zap's native function
    logger, _ := zap.NewDevelopment()
    zap.ReplaceGlobals(logger)

    // Use without adding logger to context
    ctx := context.Background()
    xlog.Info(ctx, "using global logger")
}
```

## API

### Context Functions

#### `ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context`

Adds a logger to the context.

```go
ctx := xlog.ContextWithLogger(context.Background(), logger)
```

#### `LoggerFromContext(ctx context.Context) *zap.Logger`

Extracts logger from context. If logger is not found, returns the global logger.

```go
logger := xlog.LoggerFromContext(ctx)
logger.Info("direct zap usage")
```

#### Global Logger Management

Use zap's native `ReplaceGlobals()` function to replace the global logger. This is useful for testing or temporarily changing the logger.

```go
logger, _ := zap.NewProduction()
undo := zap.ReplaceGlobals(logger)
defer undo() // Restore previous logger when done
```

#### `WithOperation(ctx context.Context, operation string, fields ...zap.Field) context.Context`

Creates a new context with a named logger for a specific operation.

```go
ctx = xlog.WithOperation(ctx, "payment-processing",
    zap.String("user_id", "12345"),
    zap.String("payment_id", "pay_xyz"),
)
xlog.Info(ctx, "processing payment")
```

**⚠️ Performance Warning**: This function has performance overhead due to creating new logger instances (~4μs and allocations per call).

**Recommended alternative** - use `WithFields` instead:

```go
// Instead of WithOperation - use WithFields for better performance
ctx = xlog.WithFields(ctx,
    zap.String("operation", "payment-processing"),
    zap.String("user_id", "12345"),
    zap.String("payment_id", "pay_xyz"),
)
xlog.Info(ctx, "processing payment")
```

Or use fields directly in log calls for single statements:

```go
xlog.Info(ctx, "processing payment",
    zap.String("operation", "payment-processing"),
    zap.String("user_id", "12345"),
    zap.String("payment_id", "pay_xyz"),
)
```

Use `WithOperation` only when:
- You need the operation name in the logger namespace (affects log structure)
- Performance is not critical
- The operation context will be used for many log statements

#### `WithFields(ctx context.Context, fields ...zap.Field) context.Context`

Creates a new context with additional fields added to the logger. This is the recommended way to add persistent fields to a logger context.

```go
// Add user context to logger
ctx = xlog.WithFields(ctx,
    zap.String("user_id", "12345"),
    zap.String("session_id", "sess_xyz"),
)

// All subsequent logs will include these fields
xlog.Info(ctx, "user action performed")
xlog.Debug(ctx, "processing request")
```

**Performance**: Better than `WithOperation` but still creates new logger instances. For single log statements, passing fields directly is most efficient.

### Span Management Functions

xlog provides integration with OpenTelemetry for distributed tracing. These functions help manage spans alongside logging.

#### `ReplaceTracerName(name string) func()`

Sets the global tracer name for creating spans and returns a function to restore the previous name. Call this once during application initialization. This function is thread-safe.

```go
restore := xlog.ReplaceTracerName("my-service")
defer restore() // Restore previous tracer name when done
```

#### `ContextWithTracer(ctx context.Context, tracer trace.Tracer) context.Context`

Adds a tracer to the context. If tracer is nil, the global tracer is used.

```go
tracer := otel.GetTracerProvider().Tracer("my-service")
ctx = xlog.ContextWithTracer(ctx, tracer)
```

#### `TracerFromContext(ctx context.Context) trace.Tracer`

Extracts tracer from context. If no tracer is found, returns the global tracer.

```go
tracer := xlog.TracerFromContext(ctx)
ctx, span := tracer.Start(ctx, "my-operation")
defer span.End()
```

#### `SpanFromContext(ctx context.Context) trace.Span`

Extracts the current span from context. If no span is found, returns a NoopSpan (safe to use).

```go
span := xlog.SpanFromContext(ctx)
span.SetAttributes(attribute.String("key", "value"))
```

#### `WithOperationSpan(ctx context.Context, operation string, fields ...zap.Field) (context.Context, trace.Span)`

Creates a new span for an operation and adds it to the context along with an enriched logger. The logger automatically includes the operation name and any additional fields.

```go
ctx, span := xlog.WithOperationSpan(ctx, "process-payment",
    zap.String("user_id", "12345"),
    zap.String("payment_id", "pay_xyz"),
)
defer span.End()

xlog.Info(ctx, "processing payment") // Includes operation and fields
```

**Returns:**
- New context with span and enriched logger
- Span instance that should be ended with `defer span.End()`

**Usage pattern:**
```go
func handleRequest(ctx context.Context) error {
    ctx, span := xlog.WithOperationSpan(ctx, "handleRequest")
    defer span.End()

    // All logs in this context will include the operation name
    xlog.Info(ctx, "request started")

    // ...

    return nil
}
```

#### `AddSpanEvent(ctx context.Context, message string)`

Adds an event to the current span (if present in context). Useful for marking important moments in trace execution.

```go
xlog.AddSpanEvent(ctx, "database query started")
// ... perform database query ...
xlog.AddSpanEvent(ctx, "database query completed")
```

#### `SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue)`

Sets attributes on the current span (if present in context). Use OpenTelemetry attribute types.

```go
import "go.opentelemetry.io/otel/attribute"

xlog.SetSpanAttributes(ctx,
    attribute.String("user.id", "12345"),
    attribute.Int("user.age", 25),
    attribute.Bool("user.premium", true),
)
```

#### `RecordSpanError(ctx context.Context, err error, options ...trace.EventOption)`

Records an error on the current span and sets its status to Error. Safe to call even when no span is present in context.

```go
if err := doSomething(); err != nil {
    xlog.RecordSpanError(ctx, err, trace.WithStackTrace(true))
    return err
}
```

**Note:** All span functions work safely even when no span is present in context (no-op behavior).

### Logging Functions

All logging functions require `context.Context` as the first argument.

#### Structured Logging

Functions with structured fields (zap.Field):

- `Debug(ctx context.Context, msg string, fields ...zap.Field)`
- `Info(ctx context.Context, msg string, fields ...zap.Field)`
- `Warn(ctx context.Context, msg string, fields ...zap.Field)`
- `Error(ctx context.Context, msg string, fields ...zap.Field)`
- `Fatal(ctx context.Context, msg string, fields ...zap.Field)`
- `Panic(ctx context.Context, msg string, fields ...zap.Field)`

```go
xlog.Info(ctx, "user logged in",
    zap.String("user_id", "12345"),
    zap.String("ip", "192.168.1.1"),
    zap.Duration("login_time", time.Second*2),
)
```

#### Printf-style Logging

Functions with string formatting:

- `Debugf(ctx context.Context, template string, args ...any)`
- `Infof(ctx context.Context, template string, args ...any)`
- `Warnf(ctx context.Context, template string, args ...any)`
- `Errorf(ctx context.Context, template string, args ...any)`
- `Fatalf(ctx context.Context, template string, args ...any)`
- `Panicf(ctx context.Context, template string, args ...any)`

```go
userID := "12345"
xlog.Infof(ctx, "user %s logged in", userID)
xlog.Errorf(ctx, "request processing error: code %d", 500)
```

## Complete Example

The [example/app](example/app/) directory contains a complete working application demonstrating xlog integration with OpenTelemetry, distributed tracing, and metrics:

**Structure:**
- `main.go` - Application setup, Echo server, OpenTelemetry initialization
- `http_handler.go` - HTTP handlers with span creation and logging
- `otel.go` - OpenTelemetry configuration (traces via gRPC, metrics via Prometheus)
- `worker.go` - Background worker simulating HTTP requests
- `compose.yml` - Docker Compose setup with Jaeger, Prometheus, and Grafana
- `otel-collector-config.yaml` - OpenTelemetry Collector configuration

**Features demonstrated:**
- Context-aware logging with automatic trace ID injection
- Integration with Echo web framework
- OpenTelemetry spans with `xlog.WithOperationSpan()`
- Distributed tracing with Jaeger
- Prometheus metrics collection
- Trace ID propagation in HTTP headers (X-Request-ID)
- Background workers with proper context handling

**Running the example:**

```bash
cd example

# Start infrastructure (Jaeger, Prometheus, Grafana, OTel Collector)
docker compose up -d

# Run the application
go run app/*.go

# Test the API
curl http://localhost:8080/api/work?user_id=123
curl http://localhost:8080/api/work?user_id=456&fail=true

# View traces: http://localhost:16686 (Jaeger UI)
# View metrics: http://localhost:9090 (Prometheus)
# View dashboards: http://localhost:3000 (Grafana, admin/admin)
```

## Usage Examples

### HTTP Handler

**Recommended approach** using `WithFields`:

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Add request context using WithFields (performance-friendly)
    ctx = xlog.WithFields(ctx,
        zap.String("request_id", uuid.NewString()),
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
        zap.String("query", r.URL.RawQuery),
    )

    xlog.Info(ctx, "request processing started")

    if err := processRequest(ctx, r); err != nil {
        xlog.Error(ctx, "request processing error", zap.Error(err))
        http.Error(w, "Internal Server Error", 500)
        return
    }

    xlog.Info(ctx, "request successfully processed")
}

func processRequest(ctx context.Context, r *http.Request) error {
    // For single log - pass fields directly (most efficient)
    userID := r.URL.Query().Get("userId")
    xlog.Debug(ctx, "processing user request",
        zap.String("operation", "process-request"),
        zap.String("user_id", userID),
    )

    // ... business logic ...

    return nil
}
```

**Alternative (with operation namespace):**

If you need the operation name in logger namespace (which affects log structure):

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    ctx = xlog.WithOperation(ctx, "handleRequest",
        zap.String("request_id", uuid.NewString()),
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
    )

    xlog.Info(ctx, "request processing started")
    // ...
}
```

**With distributed tracing:**

For complete observability with OpenTelemetry spans:

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    ctx, span := xlog.WithOperationSpan(ctx, "handleRequest",
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
    )
    defer span.End()

    xlog.Info(ctx, "request processing started")
    // Logs will include trace_id and span_id automatically
    // ...
}
```

See [example/app](example/app/main.go) for a complete working example with Echo framework and OpenTelemetry.

### Database Operations

```go
func getUserByID(ctx context.Context, userID string) (*User, error) {
    xlog.Debug(ctx, "fetching user from database", zap.String("user_id", userID))

    user, err := db.Query(ctx, "SELECT * FROM users WHERE id = ?", userID)
    if err != nil {
        xlog.Error(ctx, "database query error",
            zap.String("user_id", userID),
            zap.Error(err),
        )
        return nil, err
    }

    xlog.Info(ctx, "user found", zap.String("user_id", userID))
    return user, nil
}
```

### Background Tasks

```go
func backgroundWorker(ctx context.Context) {
    // Add worker context using WithFields
    ctx = xlog.WithFields(ctx,
        zap.String("worker", "background-processor"),
    )

    xlog.Info(ctx, "starting background processor")

    for {
        select {
        case <-ctx.Done():
            xlog.Info(ctx, "stopping processor")
            return
        case task := <-taskQueue:
            // Performance-friendly: pass fields directly for one-off logs
            xlog.Debug(ctx, "processing task",
                zap.String("task_id", task.ID),
            )
            if err := processTask(ctx, task); err != nil {
                xlog.Error(ctx, "task processing error",
                    zap.String("task_id", task.ID),
                    zap.Error(err),
                )
            }
        }
    }
}
```

## Logging Levels

- **Debug** - Detailed debug information
- **Info** - Informational messages about normal operation
- **Warn** - Warnings about potential issues
- **Error** - Errors that need to be handled
- **Fatal** - Critical errors that terminate the application (calls os.Exit(1))
- **Panic** - Critical errors that cause panic

## Development

### Testing

The package includes a comprehensive test suite. To run tests:

```bash
# Run all tests
make tloc

# Run tests with coverage
make test-cov
```

Or using go commands directly:

```bash
# Run tests
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Code Quality

```bash
# Install linter
make bin-deps

# Run linter
make lint

# Format code
make fmt
```

### CI/CD

The project uses GitHub Actions for continuous integration:

- **CI Pipeline** - Runs on every push and pull request:
  - Linting with golangci-lint v2.3.0
  - Tests on Go 1.23 and 1.24 with race detector
  - Coverage reporting in pull requests
  - Automatic coverage badge updates on main branch

- **Release Pipeline** - Triggered on version tags (v*):
  - Automated releases using GoReleaser
  - Changelog generation

See [.github/workflows/](.github/workflows/) for workflow configurations.

## Performance

xlog provides flexible logger abstraction through adapters (ZapAdapter, SlogAdapter), allowing you to switch between different logging backends while maintaining a consistent API.

**📊 Benchmarking:** See [BENCHMARKING.md](./BENCHMARKING.md) for detailed performance guidelines and baseline metrics.

### Performance Characteristics

**Adapter Overhead:**
The adapter layer introduces minimal overhead:
- **Basic logging**: ~2ns/op, 0 allocations
- **Context reuse**: ~2ns/op, 0 allocations (13% faster than baseline!)
- **Context creation in loop**: +2-3 allocations per call (avoid this pattern)

**Key Performance Metrics:**
- Logging with fields directly: ~2ns, 0 allocs
- WithFields (reused context): ~2ns, 0 allocs
- WithOperation (reused context): ~2ns, 0 allocs
- WithOperationSpan: ~670ns, 7 allocs

### Performance Best Practices

Follow these patterns to achieve optimal performance with zero allocations in hot paths:

#### 1. ✅ CRITICAL: Always Reuse Context

**The most important performance rule.** Creating context in a loop causes significant overhead:

```go
// ❌ ANTI-PATTERN: Creates 4-7 allocations per iteration!
for i := 0; i < 1000; i++ {
    ctx := xlog.WithOperation(baseCtx, "process-item")  // 4 allocs, 64ns
    processItem(ctx, items[i])
}

// ✅ BEST PRACTICE: 0 allocations, 32x faster
ctx := xlog.WithOperation(baseCtx, "process-batch")  // One-time cost
for i := 0; i < 1000; i++ {
    processItem(ctx, items[i])  // 0 allocs, 2ns
}
```

**Impact**: Reusing context is **32x faster** and produces **zero allocations** vs creating in loop.

#### 2. ✅ Pass Fields Directly for Single Logs

For one-off log statements, pass fields directly (most efficient):

```go
// ✅ BEST: 0 allocations
xlog.Info(ctx, "processing order",
    field.String("order_id", orderID),
    field.String("status", "pending"),
)
```

#### 3. ✅ Use WithFields for Multiple Logs

When logging multiple times with the same fields, create enriched context once:

```go
// ✅ GOOD: Create context once, reuse many times
ctx = xlog.WithFields(ctx,
    field.String("user_id", userID),
    field.String("session_id", sessionID),
)

// All subsequent logs include these fields with 0 allocations
xlog.Info(ctx, "user authenticated")  // 0 allocs
xlog.Debug(ctx, "loading preferences")  // 0 allocs
xlog.Info(ctx, "session started")  // 0 allocs
```

#### 4. ⚠️ Use WithOperation Sparingly

`WithOperation` creates a named logger (affects log namespace). Use only when needed:

```go
// ✅ WHEN TO USE: You need operation in logger namespace
ctx = xlog.WithOperation(ctx, "payment-processor",
    field.String("user_id", userID),
)

// ⚠️ ALTERNATIVE: Use WithFields + operation field (same performance)
ctx = xlog.WithFields(ctx,
    field.String("operation", "payment-processor"),
    field.String("user_id", userID),
)
```

#### 5. ✅ Appropriate Span Granularity

Spans have moderate cost (~670ns, 7 allocs). Create them at appropriate granularity:

```go
// ❌ BAD: Span per row (thousands of allocations)
for row := range rows {
    ctx, span := xlog.WithOperationSpan(ctx, "validate-row")
    validate(row)
    span.End()
}

// ✅ GOOD: One span for entire batch
ctx, span := xlog.WithOperationSpan(ctx, "validate-batch")
defer span.End()
for row := range rows {
    validate(row)  // 0 allocations
}
```

#### 6. ✅ Prefer Structured Logging

Structured fields are more efficient and enable better log analysis:

```go
// ⚠️ Less efficient and harder to query
xlog.Infof(ctx, "user %s logged in from IP %s", userID, ip)

// ✅ Efficient and queryable
xlog.Info(ctx, "user logged in",
    field.String("user_id", userID),
    field.String("ip", ip),
)
```

### Performance Anti-Patterns to Avoid

| Anti-Pattern | Impact | Fix |
|-------------|--------|-----|
| `WithOperation/WithFields in loop` | +4-7 allocs/iter, 32x slower | Create context once before loop |
| `WithOperationSpan per item` | +7 allocs/iter, 670ns | Create span for batch |
| `Printf-style logging` | String allocations | Use structured fields |
| `Unused enriched context` | Wasted allocations | Pass fields directly if logging once |

### Measured Performance (Apple M4 Pro)

Based on comprehensive benchmarking with statistical significance (n=10):

| Operation | Time | Allocations | Use Case |
|-----------|------|-------------|----------|
| Basic logging | ~2ns | 0 | Hot path logging |
| Context reuse | ~2ns | 0 | Multiple logs with same context |
| Context creation (WithOperation) | ~61ns | 4 | Setup phase only |
| Context creation (WithFields) | ~99ns | 5 | Setup phase only |
| WithOperationSpan | ~670ns | 7 | Request/operation boundaries |

**Key Insight**: The adapter overhead is negligible (~2ns) when using proper patterns (context reuse). Anti-patterns (create in loop) show 100% allocation increase.

## License

MIT

## Dependencies

- [go.uber.org/zap](https://github.com/uber-go/zap) - underlying logger
- [go.opentelemetry.io/otel](https://github.com/open-telemetry/opentelemetry-go) - OpenTelemetry tracing
- [github.com/stretchr/testify](https://github.com/stretchr/testify) - for testing