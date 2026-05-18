# Makefile for videoforge microservices
# Provides common development tasks

# Variables
BINARY_NAME=videoforge
BUILD_DIR=./bin
GO=go
GOFLAGS=-v
MIGRATE_BIN=migrate
MIGRATE_SOURCE=file://./migrations

# Default target
.PHONY: all
all: build

# Build all services and packages
.PHONY: build
build:
	@echo "Building all services..."
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR) ./...

# Build specific service
.PHONY: build-service
build-service:
	@if [ -z "$(SERVICE)" ]; then echo "Usage: make build-service SERVICE=services/auth"; exit 1; fi
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(SERVICE) ./$(SERVICE)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out -covermode=atomic ./...

# Run tests with verbose output
.PHONY: test-verbose
test-verbose:
	$(GO) test $(GOFLAGS) -v -race ./...

# Run tests for specific package
.PHONY: test-pkg
test-pkg:
	@if [ -z "$(PKG)" ]; then echo "Usage: make test-pkg PKG=./pkg/logger"; exit 1; fi
	$(GO) test $(GOFLAGS) -v $(PKG)

# Run linting
.PHONY: lint
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null 2>&1 || echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || $(GO) vet ./...

# Run gofmt check
.PHONY: fmt-check
fmt-check:
	@echo "Checking format..."
	@$(GO)fmt -l .

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@$(GO)fmt -w .

# Run database migrations - up
.PHONY: migrate-up
migrate-up:
	@echo "Running migrations up..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL not set"; \
		exit 1; \
	fi
	@which $(MIGRATE_BIN) > /dev/null 2>&1 || echo "Please install migrate: https://github.com/golang-migrate/migrate"
	@which $(MIGRATE_BIN) > /dev/null 2>&1 && $(MIGRATE_BIN) -path ./migrations -database "$(DATABASE_URL)" up

# Run database migrations - down
.PHONY: migrate-down
migrate-down:
	@echo "Running migrations down..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL not set"; \
		exit 1; \
	fi
	@which $(MIGRATE_BIN) > /dev/null 2>&1 || echo "Please install migrate: https://github.com/golang-migrate/migrate"
	@which $(MIGRATE_BIN) > /dev/null 2>&1 && $(MIGRATE_BIN) -path ./migrations -database "$(DATABASE_URL)" down

# Create new migration
.PHONY: migrate-new
migrate-new:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-new NAME=create_users_table"; exit 1; fi
	@mkdir -p ./migrations
	@which $(MIGRATE_BIN) > /dev/null 2>&1 && $(MIGRATE_BIN) create -seq -dir ./migrations $(NAME) || echo "Please install migrate"

# Run the application
.PHONY: run
run:
	@echo "Running application..."
	$(GO) run ./cmd/app

# Run docker-compose services
.PHONY: docker-up
docker-up:
	docker-compose up -d

# Stop docker-compose services
.PHONY: docker-down
docker-down:
	docker-compose down

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

# Install dependencies
.PHONY: deps
deps:
	$(GO) mod download
	$(GO) mod tidy

# Tidy dependencies
.PHONY: tidy
tidy:
	$(GO) mod tidy
	cd pkg && $(GO) mod tidy

# Setup local development environment
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	@cp -n .env.example .env 2>/dev/null || true
	@docker-compose up -d
	@echo "Development environment ready."
	@echo "Copy .env.example to .env and configure if needed."

# Show help
.PHONY: help
help:
	@echo "Videoforge Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build all services and packages"
	@echo "  build-service - Build specific service (SERVICE=services/auth)"
	@echo "  test           - Run tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-pkg       - Run tests for specific package (PKG=./pkg/logger)"
	@echo "  lint           - Run linters"
	@echo "  fmt            - Format code"
	@echo "  fmt-check     - Check code format"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  migrate-new   - Create new migration (NAME=create_users_table)"
	@echo "  run           - Run the application"
	@echo "  docker-up     - Start docker-compose services"
	@echo "  docker-down  - Stop docker-compose services"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  tidy          - Tidy dependencies"
	@echo "  dev-setup     - Setup local development environment"
	@echo "  help          - Show this help"