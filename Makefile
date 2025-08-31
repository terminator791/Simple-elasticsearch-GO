.PHONY: build test clean run seed migrate docker-up docker-down

# Build the application
build:
	go build -o bin/api ./cmd/api
	go build -o bin/seeder ./cmd/seeder
	go build -o bin/migrator ./cmd/migrator

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run the API server locally
run: build
	./bin/api

# Seed sample data
seed: build
	./bin/seeder

# Migrate data to Elasticsearch
migrate: build
	./bin/migrator

# Start Docker services
docker-up:
	docker-compose up -d

# Stop Docker services
docker-down:
	docker-compose down

# Start infrastructure only (for local development)
docker-dev:
	docker-compose up -d postgres elasticsearch

# Build and run everything
all: docker-up seed migrate

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run development server with hot reload (requires air)
dev:
	air -c .air.toml

# Show help
help:
	@echo "Available commands:"
	@echo "  build         - Build all binaries"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Run API server locally"
	@echo "  seed          - Seed sample data"
	@echo "  migrate       - Migrate data to Elasticsearch"
	@echo "  docker-up     - Start all Docker services"
	@echo "  docker-down   - Stop all Docker services"
	@echo "  docker-dev    - Start infrastructure only"
	@echo "  all           - Build and run everything"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  deps          - Install/update dependencies"
	@echo "  dev           - Run with hot reload"
	@echo "  help          - Show this help"