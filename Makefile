# Claude Agent SDK Go Makefile

.PHONY: all build test fmt lint clean coverage help examples

# Default target
all: fmt lint build test

# Build the SDK
build:
	@echo "Building Claude Agent SDK..."
	go build ./...

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linters
lint:
	@echo "Running linters..."
	@if [ -f "$(shell go env GOPATH)/bin/golangci-lint" ]; then \
		$(shell go env GOPATH)/bin/golangci-lint run ./...; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f coverage.out coverage.html

# Build examples
examples:
	@echo "Building examples..."
	@if [ -d "examples" ]; then \
		for example in examples/*/; do \
			if [ -f "$$example/main.go" ]; then \
				echo "Building $$example"; \
				go build -o "$${example%/}" "$$example/main.go"; \
			fi; \
		done; \
	else \
		echo "No examples directory found"; \
	fi

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod verify

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Run tests in watch mode
watch-test:
	@echo "Running tests in watch mode..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -r go test ./...; \
	else \
		echo "entr not installed. Install with: brew install entr"; \
	fi

# Generate documentation (if godoc is available)
docs:
	@echo "Starting documentation server..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Documentation available at http://localhost:6060/pkg/github.com/anthropics/claude-agent-sdk-go/"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Format, lint, build, and test"
	@echo "  build        - Build the SDK"
	@echo "  test         - Run all tests"
	@echo "  coverage     - Run tests with coverage report"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linters"
	@echo "  clean        - Clean build artifacts"
	@echo "  examples     - Build examples"
	@echo "  bench        - Run benchmarks"
	@echo "  deps         - Install dependencies"
	@echo "  update-deps  - Update dependencies"
	@echo "  watch-test   - Run tests in watch mode"
	@echo "  docs         - Start documentation server"
	@echo "  help         - Show this help message"