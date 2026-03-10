#!/bin/bash

# Script to compare benchmark results before and after changes
# Usage: ./scripts/compare_benchmarks.sh [after_benchmarks.txt]

set -e

BASELINE="benchmarks_baseline_before_adapters.txt"
AFTER="${1:-benchmarks_after_adapters.txt}"

if [ ! -f "$BASELINE" ]; then
    echo "Error: Baseline file not found: $BASELINE"
    echo "Please run benchmarks first: go test -bench=. -benchmem -benchtime=3s -run=^$ > $BASELINE"
    exit 1
fi

if [ ! -f "$AFTER" ]; then
    echo "Running new benchmarks..."
    go test -bench=. -benchmem -benchtime=3s -run='^$' | tee "$AFTER"
fi

echo ""
echo "========================================"
echo "Benchmark Comparison"
echo "========================================"
echo ""
echo "Baseline: $BASELINE"
echo "After:    $AFTER"
echo ""

# Check if benchstat is installed
if ! command -v benchstat &> /dev/null; then
    echo "Installing benchstat..."
    go install golang.org/x/perf/cmd/benchstat@latest
fi

echo ""
echo "Detailed comparison:"
echo "--------------------"
benchstat "$BASELINE" "$AFTER"

echo ""
echo "========================================"
echo "Key metrics to watch:"
echo "========================================"
echo "1. Basic logging overhead (target: <5 ns/op increase)"
echo "2. Context operations (target: <10 ns/op increase)"
echo "3. Memory allocations (should remain unchanged)"
echo "4. Field conversion (should remain similar)"
