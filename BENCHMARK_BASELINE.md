# Baseline Benchmarks (Before Adapter Implementation)

**Date:** 2026-03-10
**CPU:** Apple M4 Pro (arm64)
**Benchmark time:** 3 seconds per test

## Summary

This document captures the performance baseline before implementing the adapter pattern for different logging backends. These metrics will be used to measure the performance impact of the adapter layer.

## Key Metrics

### Basic Logging Performance

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| zap (direct) | 0.35 | 0 | 0 |
| xlog (wrapper) | 2.33 | 0 | 0 |

**Note:** xlog is ~6.6x slower than direct zap (overhead: ~2 ns/op)

### WithOperation Performance

#### Creating logger in loop
| Scenario | ns/op | B/op | allocs/op |
|----------|-------|------|-----------|
| zap - with operation name | 37.33 | 128 | 1 |
| zap - with operation + fields | 93.28 | 320 | 3 |
| xlog - with operation name | 59.02 | 176 | 2 |
| xlog - with operation + fields | 116.7 | 368 | 4 |

#### Reusing logger/context
| Scenario | ns/op | B/op | allocs/op |
|----------|-------|------|-----------|
| zap - with operation name | 0.35 | 0 | 0 |
| zap - with operation + fields | 0.50 | 0 | 0 |
| xlog - with operation name | 2.17 | 0 | 0 |
| xlog - with operation + fields | 2.31 | 0 | 0 |

**Key insight:** Reusing contexts/loggers is ~20-50x faster than recreating them

### WithFields Performance

#### Creating logger in loop
| Scenario | ns/op | B/op | allocs/op |
|----------|-------|------|-----------|
| zap | 56.96 | 192 | 2 |
| xlog | 76.94 | 240 | 3 |

#### Reusing logger/context
| Scenario | ns/op | B/op | allocs/op |
|----------|-------|------|-----------|
| zap | 0.38 | 0 | 0 |
| xlog | 2.31 | 0 | 0 |

### Tracing Performance

#### WithOperationSpan
| Scenario | ns/op | B/op | allocs/op |
|----------|-------|------|-----------|
| without fields | 973.0 | 1288 | 5 |
| with 3 fields | 956.6 | 1754 | 8 |
| with 10 fields | 2588 | 3112 | 11 |

#### SetSpanAttributes
| Scenario | ns/op | B/op | allocs/op |
|----------|-------|------|-----------|
| single attribute | 77.11 | 115 | 1 |
| 3 attributes | 130.3 | 156 | 0 |
| 10 attributes | 349.4 | 547 | 0 |
| no span in context | 27.45 | 64 | 1 |

#### Other tracing operations
| Operation | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| AddSpanEvent (simple) | 144.2 | 0 | 0 |
| AddSpanEvent (no span) | 3.68 | 0 | 0 |
| SpanFromContext (with span) | 4.83 | 0 | 0 |
| SpanFromContext (without span) | 1.59 | 0 | 0 |

### Field Conversion Performance

| Scenario | ns/op | B/op | allocs/op |
|----------|-------|------|-----------|
| empty fields | 1.32 | 0 | 0 |
| 3 string fields | 75.95 | 192 | 1 |
| 3 mixed fields | 75.31 | 192 | 1 |
| 10 mixed fields | 252.8 | 704 | 1 |
| with unsupported types | 54.90 | 128 | 1 |

### Span Creation Comparison

| Method | ns/op | B/op | allocs/op |
|--------|-------|------|-----------|
| xlog.WithOperationSpan | 911.2 | 1791 | 8 |
| manual span + logger | 2062 | 1852 | 8 |

**Key insight:** xlog.WithOperationSpan is ~2.3x faster than manual approach

## Expected Impact of Adapter Layer

When implementing adapters, we expect:

1. **Small overhead for adapter dispatch** (~1-5 ns/op for interface method calls)
2. **No additional allocations** if properly implemented
3. **Field conversion overhead** already measured above

## Monitoring Strategy

After adapter implementation, compare:

1. Basic logging overhead (target: <5 ns/op increase)
2. Context operations overhead (target: <10 ns/op increase)
3. Memory allocations should remain unchanged
4. Field conversion performance should remain similar

## Full Results

See `benchmarks_baseline_before_adapters.txt` for complete benchmark output.
