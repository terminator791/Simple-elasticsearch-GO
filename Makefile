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

# Run seeder inside docker network (avoids host-local Postgres conflicts)
.PHONY: docker-seed
docker-seed:
	# Run seeder using golang container attached to compose network so it connects to services by name
	docker run --rm --network simple-elasticsearch-go_default -v $(PWD):/app -w /app \
		-e DB_HOST=postgres -e DB_PORT=5432 -e DB_USER=postgres -e DB_PASSWORD=postgres -e DB_NAME=ecommerce \
		-e ES_HOST=http://elasticsearch:9200 golang:1.25-alpine sh -c "apk add --no-cache git ca-certificates && go mod download && go run ./cmd/seeder"

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