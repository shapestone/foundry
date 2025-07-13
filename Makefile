# Makefile for Foundry

.PHONY: build test test-unit test-integration test-coverage test-race test-foundry test-autowire test-middleware install clean test-clean deps
.PHONY: test-scaffolder test-scaffolder-unit test-scaffolder-integration test-scaffolder-coverage test-scaffolder-race bench-scaffolder dev-test-scaffolder
.PHONY: test-parser test-parser-unit test-parser-integration test-parser-coverage test-parser-race bench-parser dev-test-parser

# Ensure dependencies are up to date
deps:
	go mod tidy
	go mod download

# Build the binary
build: deps
	go build -o foundry ./cmd/foundry

# Run tests
test: deps
	go test -v ./...

# Unit tests only
test-unit: deps
	go test -v ./cmd/... ./internal/...

# Integration tests only
test-integration: deps
	go test -v ./test/integration/...

# Test with coverage
test-coverage: deps
	go test -v -cover ./...

# Test with coverage and output to file
test-coverage-file: deps
	go test -coverpkg=./cmd/foundry/cmd,./internal/... -coverprofile=coverage.out ./test/...
	go tool cover -func=coverage.out

# Test with race detector
test-race: deps
	go test -v -race ./...

# Test foundry commands (original test)
test-foundry: build
	@chmod +x scripts/test-foundry.sh
	@./scripts/test-foundry.sh

# Test auto-wire functionality
test-autowire: build
	@chmod +x scripts/test-autowire.sh
	@./scripts/test-autowire.sh

# Test middleware functionality
test-middleware: build
	@chmod +x scripts/test-middleware.sh
	@./scripts/test-middleware.sh

# ==========================================
# SCAFFOLDER TEST COMMANDS
# ==========================================

# Run all scaffolder tests
test-scaffolder: deps test-scaffolder-unit test-scaffolder-integration

# Run only scaffolder unit tests (fast)
test-scaffolder-unit: deps
	@echo "🧪 Running scaffolder unit tests..."
	go test -v ./test/scaffolder/... -short

# Run only scaffolder integration tests (slower)
test-scaffolder-integration: deps
	@echo "🔗 Running scaffolder integration tests..."
	go test -v ./test/scaffolder/... -run Integration

# Run scaffolder tests with coverage
test-scaffolder-coverage: deps
	@echo "📊 Running scaffolder tests with coverage..."
	go test -v ./test/scaffolder/... -coverprofile=coverage-scaffolder.out
	go tool cover -html=coverage-scaffolder.out -o coverage-scaffolder.html
	@echo "✅ Coverage report generated: coverage-scaffolder.html"

# Run scaffolder tests with race detection
test-scaffolder-race: deps
	@echo "🏃 Running scaffolder tests with race detection..."
	go test -v ./test/scaffolder/... -race

# Benchmark scaffolder performance
bench-scaffolder: deps
	@echo "⚡ Running scaffolder benchmarks..."
	go test -v ./test/scaffolder/... -bench=. -benchmem

# Development workflow - watch for changes
dev-test-scaffolder: deps
	@echo "👀 Running scaffolder tests in watch mode..."
	@while true; do \
		echo "Running tests..."; \
		go test -v ./test/scaffolder/... -short; \
		echo "Watching for changes... (Ctrl+C to exit)"; \
		sleep 2; \
	done

# ==========================================
# PARSER TEST COMMANDS
# ==========================================

# Run all parser tests
test-parser: deps test-parser-unit test-parser-integration

# Run only parser unit tests (fast)
test-parser-unit: deps
	@echo "🧪 Running parser unit tests..."
	go test -v ./test/parser/... -short

# Run only parser integration tests (slower)
test-parser-integration: deps
	@echo "🔗 Running parser integration tests..."
	go test -v ./test/parser/... -run Integration

# Run parser tests with coverage
test-parser-coverage: deps
	@echo "📊 Running parser tests with coverage..."
	go test -v ./test/parser/... -coverprofile=coverage-parser.out
	go tool cover -html=coverage-parser.out -o coverage-parser.html
	@echo "✅ Coverage report generated: coverage-parser.html"

# Run parser tests with race detection
test-parser-race: deps
	@echo "🏃 Running parser tests with race detection..."
	go test -v ./test/parser/... -race

# Benchmark parser performance
bench-parser: deps
	@echo "⚡ Running parser benchmarks..."
	go test -v ./test/parser/... -bench=. -benchmem

# Development workflow - watch for changes
dev-test-parser: deps
	@echo "👀 Running parser tests in watch mode..."
	@while true; do \
		echo "Running tests..."; \
		go test -v ./test/parser/... -short; \
		echo "Watching for changes... (Ctrl+C to exit)"; \
		sleep 2; \
	done

# ==========================================
# COMBINED TEST COMMANDS
# ==========================================

# Install foundry to GOPATH/bin
install: deps
	go install ./cmd/foundry

# Clean build artifacts and test directories
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f foundry foundry-test
	@echo "🧹 Cleaning test output directory..."
	@rm -rf test-output/
	@rm -rf tmp/
	@echo "🧹 Cleaning example projects..."
	@rm -rf myapp testapp
	@echo "🧹 Cleaning scaffolder test artifacts..."
	@rm -f coverage-scaffolder.out coverage-scaffolder.html
	@echo "🧹 Cleaning parser test artifacts..."
	@rm -f coverage-parser.out coverage-parser.html
	@rm -rf /tmp/foundry-test-*
	@echo "🧹 Cleaning Go test cache..."
	@go clean -testcache
	@echo "✅ Clean complete"

# Clean and then run tests
test-clean: clean test

# Run all tests including scaffolder and parser
test-all: test test-scaffolder test-parser test-foundry test-autowire test-middleware

# Development helpers
.PHONY: fmt lint

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Show what commands are available
help:
	@echo "📋 Available commands:"
	@echo ""
	@echo "🔧 Build & Install:"
	@echo "  make deps               - Update and download dependencies"
	@echo "  make build              - Build the foundry binary"
	@echo "  make install            - Install foundry to GOPATH/bin"
	@echo ""
	@echo "🧪 Basic Testing:"
	@echo "  make test               - Run unit tests"
	@echo "  make test-unit          - Run unit tests only"
	@echo "  make test-integration   - Run integration tests only"
	@echo "  make test-coverage      - Run tests with coverage"
	@echo "  make test-coverage-file - Run tests with coverage report"
	@echo "  make test-race          - Run tests with race detector"
	@echo ""
	@echo "🏗️  Scaffolder Testing:"
	@echo "  make test-scaffolder           - Run all scaffolder tests"
	@echo "  make test-scaffolder-unit      - Run scaffolder unit tests (fast)"
	@echo "  make test-scaffolder-integration - Run scaffolder integration tests"
	@echo "  make test-scaffolder-coverage  - Run scaffolder tests with coverage"
	@echo "  make test-scaffolder-race      - Run scaffolder tests with race detection"
	@echo "  make bench-scaffolder          - Benchmark scaffolder performance"
	@echo "  make dev-test-scaffolder       - Watch mode for scaffolder tests"
	@echo ""
	@echo "🔍 Parser Testing:"
	@echo "  make test-parser              - Run all parser tests"
	@echo "  make test-parser-unit         - Run parser unit tests (fast)"
	@echo "  make test-parser-integration  - Run parser integration tests"
	@echo "  make test-parser-coverage     - Run parser tests with coverage"
	@echo "  make test-parser-race         - Run parser tests with race detection"
	@echo "  make bench-parser             - Benchmark parser performance"
	@echo "  make dev-test-parser          - Watch mode for parser tests"
	@echo ""
	@echo "🚀 End-to-End Testing:"
	@echo "  make test-foundry       - Test foundry commands"
	@echo "  make test-autowire      - Test auto-wire functionality"
	@echo "  make test-middleware    - Test middleware functionality"
	@echo "  make test-all           - Run all tests (unit, scaffolder, parser, e2e)"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  make test-clean         - Clean and run tests"
	@echo "  make clean              - Clean build artifacts and test output"
	@echo "  make fmt                - Format Go code"
	@echo "  make lint               - Run linter (requires golangci-lint)"