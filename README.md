# xlog

[![CI](https://github.com/ruko1202/xlog/actions/workflows/ci.yml/badge.svg)](https://github.com/ruko1202/xlog/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ruko1202/xlog)](https://goreportcard.com/report/github.com/ruko1202/xlog)
![Coverage](https://img.shields.io/badge/Coverage-97.2%25-brightgreen)
[![Go Reference](https://pkg.go.dev/badge/github.com/ruko1202/xlog.svg)](https://pkg.go.dev/github.com/ruko1202/xlog)

A wrapper around [zap](https://github.com/uber-go/zap) for context-aware logging.

## Features

- Context-aware logging - logger is extracted from context
- Global logger fallback support with thread-safe replacement
- Structured logging via zap.Field
- Printf-style formatting (Debugf, Infof, etc.)
- All logging levels: Debug, Info, Warn, Error, Fatal, Panic
- Zero allocation when logger is not in context (uses zap.NewNop())

## Installation

```bash
go get github.com/ruko1202/xlog
```

## Quick Start

For a complete working example, see [example/http](example/http/main.go).

### Basic Usage

```go
package main

import (
    "context"
    "github.com/ruko1202/xlog"
    "go.uber.org/zap"
)

const (
	appName = "xlogApp"
)

func main() {
    // Create logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Add logger to context
    ctx := context.Background()
    ctx = xlog.ContextWithLogger(ctx, logger)

    // Use logging
    xlog.Infof(ctx, "application `%s` started", appName)
    xlog.Debug(ctx, "debug information", zap.String("version", "1.0.0"))
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
    // Replace global logger
    logger, _ := zap.NewDevelopment()
    xlog.ReplaceGlobal(logger)

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

#### `ReplaceGlobal(logger *zap.Logger) func()`

Replaces the global logger and returns a function to restore the previous logger.
This function is thread-safe and useful for testing or temporarily changing the logger.

```go
logger, _ := zap.NewProduction()
restore := xlog.ReplaceGlobal(logger)
defer restore() // Restore previous logger when done
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

**⚠️ Performance Warning**: This function has significant performance overhead (~4μs and ~59KB allocations per call, see [Benchmarks](#benchmarks)).

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

Alternative using:

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    ctx = xlog.WithOperation(ctx, "handleRequest",
      zap.String("request_id", uuid.NewString()),
      zap.String("method", r.Method),
      zap.String("path", r.URL.Path),
      zap.String("query", r.URL.RawQuery),
    )

    xlog.Info(ctx, "request processing started")
    // ...
}
```

See [example/http](example/http/main.go) for a complete working example.

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

xlog uses zap as the underlying logger, which provides:

- Zero allocation when using structured fields
- High performance (structured logging is ~10x faster than fmt.Printf)
- Minimal overhead when logger is not in context

## License

MIT

## Dependencies

- [go.uber.org/zap](https://github.com/uber-go/zap) - underlying logger
- [github.com/stretchr/testify](https://github.com/stretchr/testify) - for testing