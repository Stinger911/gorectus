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

# Database and Migration Commands
db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down

db-logs:
	docker-compose logs -f postgres

# Run database migrations
migrate-up:
	$(GOCMD) run cmd/migrate/main.go -up

migrate-down:
	$(GOCMD) run cmd/migrate/main.go -down -steps 1

migrate-reset:
	$(GOCMD) run cmd/migrate/main.go -reset

migrate-status:
	$(GOCMD) run cmd/migrate/main.go

migrate-force:
	$(GOCMD) run cmd/migrate/main.go -force $(VERSION)

# Generate password hash
hash:
	$(GOCMD) run cmd/migrate/main.go -hash $(PASSWORD)

# Validate migration files
validate-migrations:
	./scripts/validate-migrations.sh

# Setup development environment
setup: deps validate-migrations db-up
	@echo "Waiting for database to be ready..."
	@sleep 5
	$(MAKE) migrate-up
	@echo "Development environment ready!"
	@echo "Admin credentials: admin@gorectus.local / admin123"

# Start everything for development
start: setup
	$(MAKE) run

.PHONY: build test test-coverage test-race benchmark clean deps run dev fmt lint check install-tools db-up db-down db-logs migrate-up migrate-down migrate-reset migrate-status migrate-force hash validate-migrations setup start
