# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=gorectus-server
BINARY_PATH=./bin/$(BINARY_NAME)

# Build the application
build:
	$(GOBUILD) -o $(BINARY_PATH) -v ./cmd/server

# Run tests
test:
	$(GOTEST) -v ./cmd/server

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./cmd/server
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
test-race:
	$(GOTEST) -v -race ./cmd/server

# Run benchmarks
benchmark:
	$(GOTEST) -v -bench=. ./cmd/server

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run the server
run:
	$(GOBUILD) -o $(BINARY_PATH) -v ./cmd/server
	$(BINARY_PATH)

# Development run with auto-reload (requires air)
dev:
	air -c .air.toml

# Format code
fmt:
	$(GOCMD) fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Run all checks (test, lint, format)
check: fmt lint test

# Install development tools
install-tools:
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: build test test-coverage test-race benchmark clean deps run dev fmt lint check install-tools
