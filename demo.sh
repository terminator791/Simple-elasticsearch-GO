#!/bin/bash

# Demo script to showcase the 5000 product seeder and GetAllProducts functionality
# This script demonstrates Elasticsearch's power for fast search and aggregations

echo "🚀 Elasticsearch Product Catalog Demo"
echo "======================================"
echo

# Build all components
echo "📦 Building all components..."
make build
echo

# Check if Docker services are running (for actual demo)
echo "🔍 Checking Docker services..."
if ! docker ps | grep -q elasticsearch; then
    echo "⚠️  Note: Elasticsearch container not running. For full demo, run 'make docker-up' first."
    echo "    This demo will show the code structure and capabilities."
else
    echo "✅ Elasticsearch is running"
fi
echo

# Show seeder capabilities
echo "🌱 Seeder Capabilities:"
echo "   - Generates 5000 realistic products"
echo "   - 20 different brands (Apple, Samsung, Sony, Dell, HP, etc.)"
echo "   - Multiple categories (Smartphones, Laptops, Gaming, Sports, etc.)"
echo "   - Realistic pricing based on product categories"
echo "   - Batch processing with progress logging"
echo "   - Uses CQRS pattern: PostgreSQL → Elasticsearch"
echo

# Show API endpoints
echo "🔗 Available API Endpoints:"
echo "   GET  /api/v1/products              - Get all products (Elasticsearch powered)"
echo "   GET  /api/v1/products/search       - Advanced search with filters"
echo "   GET  /api/v1/products/performance  - Compare PostgreSQL vs Elasticsearch"
echo "   POST /api/v1/products              - Create new product"
echo "   GET  /api/v1/products/{id}         - Get single product"
echo

# Show Elasticsearch features
echo "⚡ Elasticsearch Features Demonstrated:"
echo "   ✓ Fast full-text search across 5000 products"
echo "   ✓ Real-time aggregations (brands, categories)"
echo "   ✓ Efficient pagination"
echo "   ✓ Advanced filtering capabilities"
echo "   ✓ Performance comparison with PostgreSQL"
echo

# Show example API calls
echo "📞 Example API Calls:"
echo "   # Get all products with aggregations:"
echo "   curl 'http://localhost:8080/api/v1/products?page=1&size=10'"
echo
echo "   # Search for iPhones:"
echo "   curl 'http://localhost:8080/api/v1/products/search?q=iPhone'"
echo
echo "   # Compare search performance:"
echo "   curl 'http://localhost:8080/api/v1/products/performance/iPhone'"
echo

# Show file structure
echo "📁 Implementation Files:"
echo "   ├── cmd/seeder/main.go     - 5000 product seeder"
echo "   ├── cmd/migrator/main.go   - PostgreSQL → Elasticsearch migrator"
echo "   ├── cmd/api/main.go        - Main API server with new routes"
echo "   ├── internal/services/     - GetAllProducts method using Elasticsearch"
echo "   └── internal/handlers/     - GetAllProducts API handler"
echo

echo "🎯 To run the full demo:"
echo "   1. Start services: make docker-up"
echo "   2. Seed data:     make seed"
echo "   3. Start API:     make run"
echo "   4. Test endpoints with the curl commands above"
echo

echo "✨ This demonstrates Elasticsearch's power for:"
echo "   • Fast search across large datasets (5000 products)"
echo "   • Real-time aggregations and faceted search"
echo "   • Performance benefits over traditional SQL queries"
echo "   • Scalable pagination and filtering"
echo

echo "Demo script completed! The implementation showcases Elasticsearch's capabilities."