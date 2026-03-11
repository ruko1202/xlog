# Benchmarking Guide

## Overview

This project uses benchmarks to track performance at all stages of development. It is especially important to monitor performance when introducing new abstraction layers, such as adapters for different loggers.

## Baseline Metrics

Baseline performance metrics are recorded in the [`BENCHMARK_BASELINE.md`](./BENCHMARK_BASELINE.md) file before implementing the adapter layer.

### Key Baseline Metrics

- **Direct zap**: 0.35 ns/op, 0 allocs
- **xlog wrapper**: 2.33 ns/op, 0 allocs (overhead ~2 ns)
- **Context reuse**: 2.17-2.31 ns/op vs 59-117 ns/op when created in a loop

## Running Benchmarks

### Full Benchmark Suite

```bash
# Run all benchmarks with 3-second execution time
go test -bench=. -benchmem -benchtime=3s -run='^$'
```

### Specific Benchmarks

```bash
# Basic loggers only
go test -bench=BenchmarkLogger -benchmem -benchtime=3s

# Context operations only
go test -bench=BenchmarkAdvance -benchmem -benchtime=3s

# Tracing only
go test -bench=BenchmarkWithOperationSpan -benchmem -benchtime=3s
```

### Saving Results

```bash
# Save results to a file for later comparison
go test -bench=. -benchmem -benchtime=3s -run='^$' | tee my_benchmarks.txt
```

## Performance Comparison

### Automatic Comparison

Use the script to automatically compare with baseline:

```bash
# Compare with baseline (will create new file benchmarks_after_adapters.txt)
./scripts/compare_benchmarks.sh

# Compare with specific file
./scripts/compare_benchmarks.sh my_benchmarks.txt
```

### Manual Comparison with benchstat

```bash
# Install benchstat (if not installed)
go install golang.org/x/perf/cmd/benchstat@latest

# Compare two files
benchstat benchmarks_baseline_before_adapters.txt benchmarks_after_adapters.txt
```

### Interpreting benchstat Results

```
name                    old time/op  new time/op  delta
Logger/zap-14           0.35ns ± 0%  0.37ns ± 0%  +5.71%
Logger/xlog-14          2.33ns ± 0%  2.98ns ± 0%  +27.90%
```

- `~` (tilde) - no statistically significant difference
- `+X%` - performance **degraded** by X%
- `-X%` - performance **improved** by X%
- `± Y%` - measurement error margin

## Acceptance Criteria After Adapter Implementation

After implementing the adapter layer we expect:

### ✅ Acceptable

- Basic logging: **< 5 ns/op** additional overhead
- Context operations: **< 10 ns/op** additional overhead
- Memory allocations: **no changes**
- Field conversion: **< 10% slowdown**

### ⚠️ Requires Attention

- Basic logging: **5-10 ns/op** overhead
- Context operations: **10-20 ns/op** overhead
- Additional allocations: **+1 alloc/op**

### ❌ Unacceptable

- Basic logging: **> 10 ns/op** overhead
- Context operations: **> 20 ns/op** overhead
- Significant allocation increase: **+2 or more alloc/op**

## Common Bottlenecks

### 1. Interface Dispatch Overhead

```go
// Problem: too many abstraction levels
type Logger interface {
    Log(msg string)
}

type Adapter struct {
    logger Logger  // another level
}

// Solution: minimize nesting levels
```

### 2. Allocation in Hot Path

```go
// Problem: creating slices on every call
func (a *Adapter) Log(msg string, fields ...Field) {
    attrs := make([]slog.Attr, len(fields))  // allocation!
    // ...
}

// Solution: reuse or pre-allocate
```

### 3. Excessive Type Conversion

```go
// Problem: multiple conversions
func convert(f Field) SlogAttr {
    switch f.Type {
    case IntType:
        return slog.Int64(f.Key, int64(f.Value))  // conversion
    }
}

// Solution: store in optimal format from the start
```

## Profiling

### CPU Profiling

```bash
# Create CPU profile
go test -bench=BenchmarkLogger -cpuprofile=cpu.prof

# Analyze profile
go tool pprof -http=:8080 cpu.prof
```

### Memory Profiling

```bash
# Create memory profile
go test -bench=BenchmarkLogger -memprofile=mem.prof

# Analyze profile
go tool pprof -http=:8080 mem.prof
```

### Trace Profiling

```bash
# Create trace
go test -bench=BenchmarkLogger -trace=trace.out

# Analyze trace
go tool trace trace.out
```

## Continuous Monitoring

### In CI/CD

Add performance checks to CI:

```yaml
# Example for GitHub Actions
- name: Run benchmarks
  run: |
    go test -bench=. -benchmem -benchtime=1s -run='^$' | tee new.txt
    benchstat benchmarks_baseline_before_adapters.txt new.txt
```

### Local Development

Before each major change:

```bash
# 1. Record current state
go test -bench=. -benchmem -benchtime=3s -run='^$' > before.txt

# 2. Make changes

# 3. Compare
go test -bench=. -benchmem -benchtime=3s -run='^$' > after.txt
benchstat before.txt after.txt
```

## Additional Resources

- [Go Benchmarking Guide](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [benchstat documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Go Performance Wiki](https://github.com/golang/go/wiki/Performance)
- [Project baseline metrics](./BENCHMARK_BASELINE.md)
- [Comparison scripts](./scripts/README.md)
