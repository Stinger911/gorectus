.DEFAULT_GOAL := help

help: ## this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m%s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf "\n\n"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=gorectus-server
BINARY_PATH=./bin/$(BINARY_NAME)

build: swagger ## Build the application
	$(GOBUILD) -o $(BINARY_PATH) -v ./cmd/server

test: ## Run tests
	$(GOTEST) -v ./cmd/server

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./cmd/server
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

test-race: ## Run tests with race detection
	$(GOTEST) -v -race ./cmd/server

benchmark: ## Run benchmarks
	$(GOTEST) -v -bench=. ./cmd/server

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_PATH)
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

run: build ## Run the server
	$(BINARY_PATH)

dev:  ## Development run with auto-reload (requires air)
	$(shell go env GOPATH)/bin/air -c .air.toml

fmt: ## Format code
	$(GOCMD) fmt ./...

lint: ## Run linter (requires golangci-lint)
	$(shell go env GOPATH)/bin/golangci-lint run

check: fmt lint test ## Run all checks (test, lint, format)

install-tools: ## Install development tools
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Database and Migration Commands
db-up: ## Start the PostgreSQL database using Docker
	docker-compose up -d postgres

db-down: ## Stop the PostgreSQL database
	docker-compose down

db-logs: ## View PostgreSQL logs
	docker-compose logs -f postgres

# Run database migrations
migrate-up: ## Apply database migrations
	$(GOCMD) run cmd/migrate/main.go -up

migrate-down: ## Rollback database migrations
	$(GOCMD) run cmd/migrate/main.go -down -steps 1

migrate-reset: ## Reset database migrations
	$(GOCMD) run cmd/migrate/main.go -reset

migrate-status: ## Show database migration status
	$(GOCMD) run cmd/migrate/main.go

migrate-force: ## Force database migration to a specific version
	$(GOCMD) run cmd/migrate/main.go -force $(VERSION)

hash: ## Generate password hash
	$(GOCMD) run cmd/migrate/main.go -hash $(PASSWORD)

validate-migrations: ## Validate migration files
	./scripts/validate-migrations.sh

setup: deps validate-migrations db-up ## Setup development environment
	@echo "Waiting for database to be ready..."
	@sleep 5
	$(MAKE) migrate-up
	@echo "Development environment ready!"
	@echo "Admin credentials: admin@gorectus.local / admin123"

start: setup ## Start everything for development
	$(MAKE) run

swagger: ## Generate Swagger documentation
	$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -o docs

install-swagger: ## Install Swagger CLI tool
	$(GOGET) github.com/swaggo/swag/cmd/swag@latest

.PHONY: build test test-coverage test-race benchmark clean deps run dev fmt lint check install-tools db-up db-down db-logs migrate-up migrate-down migrate-reset migrate-status migrate-force hash validate-migrations setup start swagger install-swagger
