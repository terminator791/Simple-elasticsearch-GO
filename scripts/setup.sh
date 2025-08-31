#!/bin/bash

set -e

echo "🚀 Setting up E-commerce Product Catalog API..."

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Build the project
echo "📦 Building Go binaries..."
go mod download
make build

# Start infrastructure
echo "🐳 Starting Docker services..."
docker-compose up -d postgres elasticsearch

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check PostgreSQL
echo "🗄️  Checking PostgreSQL connection..."
until docker-compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done

# Check Elasticsearch
echo "🔍 Checking Elasticsearch connection..."
until curl -s http://localhost:9200/_cluster/health >/dev/null 2>&1; do
    echo "Waiting for Elasticsearch..."
    sleep 2
done

# Seed sample data
echo "🌱 Seeding sample data..."
./bin/seeder

# Migrate to Elasticsearch
echo "📊 Migrating data to Elasticsearch..."
./bin/migrator

# Start API server
echo "🌐 Starting API server..."
docker-compose up -d api

echo "✅ Setup complete!"
echo ""
echo "🎉 Your E-commerce API is now running!"
echo "📍 API URL: http://localhost:8080"
echo "📍 Health Check: http://localhost:8080/health"
echo "📍 PostgreSQL: localhost:5432"
echo "📍 Elasticsearch: http://localhost:9200"
echo ""
echo "🔗 Try these endpoints:"
echo "  GET  http://localhost:8080/api/v1/products/search?q=iPhone"
echo "  GET  http://localhost:8080/api/v1/products/search?brand=Apple"
echo "  GET  http://localhost:8080/api/v1/products/performance/laptop"
echo ""
echo "📚 Check README.md for complete API documentation"