# =====================================
# xlog - zap.Logger wrapper to provide logger via context
# =====================================
# Main Makefile for building, testing, and managing the xlog project.
#
# Common commands:
#   make all        - Run full build pipeline (deps, tools, tests, lint, format)
#   make deps       - Download and tidy Go dependencies
#   make bin-deps   - Install all required binary tools
#   make tloc       - Run all tests
#   make test-cov   - Run tests with coverage report
#   make lint       - Run linter checks
#   make fmt        - Format code with gofmt and goimports
# =====================================

# GOBIN - Directory for installed Go binaries (golangci-lint, etc.)
# Default: ./bin in the project root
GOBIN			?= $(PWD)/bin

# Ensure GOBIN directory exists
$(shell mkdir -p $(GOBIN))

# -------------------------------------
# Default target
# -------------------------------------
# Runs the complete build pipeline: install dependencies, binary tools,
# run tests, format code, and run linter checks
.PHONY: all
all: deps bin-deps tloc fmt lint

# -------------------------------------
# Install deps and tools
# -------------------------------------

# deps - Download and organize Go module dependencies
# Downloads all required Go modules and removes unused ones
.PHONY: deps
deps:
	go mod download
	go mod tidy


# bin-deps - Install all required binary tools
# Installs:
#   - golangci-lint: comprehensive Go linter
.PHONY: bin-deps
bin-deps:
	GOBIN=$(GOBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.2 && \
	GOBIN=$(GOBIN) go install golang.org/x/perf/cmd/benchstat@latest

# -------------------------------------
# Tests
# -------------------------------------

# tloc - Run all tests with race detection disabled
# Options:
#   -p 2: Run tests in parallel with max 2 processes
#   -count 2: Run each test 2 times to catch flaky tests
# Note: Package tests are run from project root
.PHONY: tloc
tloc:
	go test -p 2 -count 2 ./...
	#cd ./test/via_pkg/ && go test -p 2 -count 2 ./...

# test-cov - Run tests with coverage analysis
# Generates coverage report excluding mock files
# Options:
#   -race: Enable data race detection
#   -p 2: Run tests in parallel with max 2 processes
#   -count 2: Run each test 2 times
#   -coverprofile: Output file for coverage data
#   -covermode atomic: Use atomic mode for race detector compatibility
#   --coverpkg: Measure coverage for internal/ packages only
# Output files:
#   coverage.tmp: Raw coverage data
#   coverage.out: Filtered coverage data (no mocks)
#   coverage.report: Human-readable coverage report
.PHONY: test-cov
test-cov:
	go test -race -p 2 -count 2 -coverprofile=coverage.out -covermode atomic ./...
	go tool cover -func=coverage.out | sed 's|github.com/ruko1202/xlog/||' | sed -E 's/\t+/\t/g' | tee coverage.report

# -------------------------------------
# Benchmarking and performance analysis
# -------------------------------------

# BENCHMARK_FILE - Target file for benchmark results
# Default: benchmarks.txt
BENCHMARK_FILE ?= benchmarks.txt

# BASELINE_FILE - Baseline benchmark file for comparison
# Default: benchmarks_baseline.txt
BASELINE_FILE ?= benchmarks_baseline.txt

# BENCH_COMPARE_FILE - benchmark compare result file
# Default: benchmarks_compare.txt
BENCH_COMPARE_FILE ?= benchmarks_compare.txt


# bench - Run full benchmark suite and save results
# Options:
#   -bench=.: Run all benchmarks
#   -benchmem: Include memory allocation statistics
#   -benchtime=3s: Run each benchmark for 3 seconds
#   -run='^$': Don't run any tests (only benchmarks)
# Output: file (default: benchmarks.txt)
# Example: make .bench file=my_benchmarks.txt
.PHONY: .bench
.bench: file=
.bench:
	$(info Running benchmarks and saving to ${file}...)
	@go test -bench=. -benchmem -benchtime=1s -count=10 -run='^$$' | tee ${file}

# bench - Save current benchmarks for future comparisons with baseline
# This creates a baseline file that can be used with bench-compare
# benchmarks.txt - target file for benchmark results
.PHONY: bench
bench:
	@make .bench file=${BENCHMARK_FILE}

# bench-baseline - Save current benchmarks as baseline for future comparisons
# This creates a baseline file that can be used with bench-compare
# benchmarks_baseline.txt - baseline benchmark file for comparison
.PHONY: bench-baseline
bench-baseline:
	@make .bench file=${BASELINE_FILE}

# bench-compare - Compare current benchmarks with baseline
# Requires benchstat to be installed (will auto-install if missing)
# Compares BASELINE_FILE with BENCHMARK_FILE
# Example: make bench-compare
# Example: make bench-compare BASELINE_FILE=old.txt BENCHMARK_FILE=new.txt
.PHONY: bench-compare
bench-compare:
	$(info Comparing $(BASELINE_FILE) with $(BENCHMARK_FILE)...)
	@echo "========================================"
	@echo "Benchmark Comparison"
	@echo "========================================"
	@echo "Baseline: $(BASELINE_FILE)"
	@echo "Current:  $(BENCHMARK_FILE)"
	@echo ""
	@$(GOBIN)/benchstat $(BASELINE_FILE) $(BENCHMARK_FILE) | tee ${BENCH_COMPARE_FILE}

# bench-full - Run benchmarks and compare with baseline in one command
# This is a convenience target that runs both bench and bench-compare
# Requires BASELINE_FILE to exist (create with bench-baseline first)
# Example: make bench-full
.PHONY: bench-full
bench-full: bench bench-compare

# bench-res - Run benchmarks with CPU and memory profiling
# Generates cpu.prof and mem.prof file for analysis with pprof
# Example: make bench-res
# Analyze: go tool pprof -http=:8080 cpu.prof
.PHONY: bench-res
bench-res:
	$(info Running benchmarks with CPU profiling...)
	@go test -bench=. -benchmem -benchtime=3s -cpuprofile=cpu.prof -memprofile=mem.prof -run='^$$'
	@echo ""
	@echo "CPU and memory profile saved to cpu.prof and mem.prof"
	@echo "Analyze with: go tool pprof -http=:8080 cpu.prof / mem.prof"

# -------------------------------------
# Linter and formatter
# -------------------------------------

# lint - Run comprehensive linter checks
# Uses golangci-lint with configuration from .golangci.yml
# Checks for code quality, style, bugs, and best practices
.PHONY: lint
lint:
	$(info $(M) running linter...)
	@$(GOBIN)/golangci-lint run

# fmt - Format code and organize imports
# Uses gofmt for code formatting and goimports for import organization
# Automatically fixes formatting issues
.PHONY: fmt
fmt:
	$(info $(M) fmt project...)
	@# Format code with gofmt and organize imports with goimports
	@$(GOBIN)/golangci-lint fmt -E gofmt -E goimports ./...
