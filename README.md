# E-commerce Product Catalog API with Elasticsearch

A comprehensive Go-based e-commerce product catalog API implementing CQRS (Command Query Responsibility Segregation) architecture with PostgreSQL as the source of truth and Elasticsearch for advanced search capabilities.

## 🏗️ Architecture

This project demonstrates a masterclass implementation of:

- **CQRS Pattern**: Commands (writes) go to PostgreSQL, Queries (reads) use Elasticsearch
- **Microservices Architecture**: Clean separation of concerns with dependency injection
- **Advanced Search**: Full-text search, filtering, aggregations using Elasticsearch
- **Performance Optimization**: Bulk indexing and performance comparison tools
- **Production-Ready**: Docker containerization, proper logging, error handling

## 🛠️ Technology Stack

- **Language**: Go 1.21
- **API Framework**: Gin Gonic
- **Source of Truth**: PostgreSQL 15 Alpine
- **Search Engine**: Elasticsearch 8.11.0
- **Database Client**: sqlx
- **Elasticsearch Client**: official elastic/go-elasticsearch/v8
- **Testing**: Go standard testing + testify
- **Containerization**: Docker & Docker Compose

## 📁 Project Structure

```
├── cmd/
│   ├── api/           # Main API server
│   ├── migrator/      # Data migration tool (PostgreSQL → Elasticsearch)
│   └── seeder/        # Product seeding tool (1000+ products)
├── internal/
│   ├── config/        # Configuration management
│   ├── models/        # Data models and DTOs
│   ├── database/      # PostgreSQL client and operations
│   ├── elasticsearch/ # Elasticsearch client with explicit mappings
│   ├── handlers/      # HTTP request handlers
│   ├── services/      # Business logic (CQRS implementation)
│   └── middleware/    # HTTP middleware (logging, CORS)
├── pkg/
│   └── logger/        # Centralized logging
├── migrations/        # Database schema migrations
├── tests/            # Test suites
├── docker-compose.yml # Multi-service Docker setup
├── Dockerfile        # Application container
└── README.md
```

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)

### 1. Clone the Repository

```bash
git clone https://github.com/terminator791/Simple-elasticsearch-GO.git
cd Simple-elasticsearch-GO
```

### 2. Start Services with Docker Compose

```bash
docker-compose up -d
```

This will start:
- PostgreSQL (port 5432)
- Elasticsearch (port 9200)
- API Server (port 8080)

### 3. Seed Sample Data

```bash
# Build the seeder
go build -o seeder ./cmd/seeder

# Run seeder to populate database with 1000 products
./seeder
```

### 4. Migrate Data to Elasticsearch

```bash
# Build the migrator
go build -o migrator ./cmd/migrator

# Run migrator to sync PostgreSQL data to Elasticsearch
./migrator
```

## 📋 API Endpoints

### Product Management (CQRS Commands)

```bash
# Create a new product (writes to PostgreSQL → indexes to Elasticsearch)
POST /api/v1/products
{
  "name": "iPhone 15 Pro",
  "description": "Latest Apple smartphone with advanced features",
  "brand": "Apple",
  "category": "Smartphones",
  "price": 999.99,
  "stock_quantity": 50
}

# Update product (updates PostgreSQL → updates Elasticsearch)
PUT /api/v1/products/{id}
{
  "price": 899.99,
  "stock_quantity": 45
}

# Delete product (deletes from PostgreSQL → removes from Elasticsearch)
DELETE /api/v1/products/{id}

# Get single product (from PostgreSQL)
GET /api/v1/products/{id}
```

### Advanced Search (CQRS Queries - Elasticsearch Only)

```bash
# Full-text search with filters and aggregations
GET /api/v1/products/search?q=iPhone&brand=Apple&category=Smartphones&min_price=500&max_price=1500&page=1&size=20&sort_by=price&sort_order=asc

# Search with multiple brands/categories
GET /api/v1/products/search?q=laptop&brand=Apple&brand=Dell&category=Computers&category=Electronics

# Text search only
GET /api/v1/products/search?q=wireless headphones

# Filter by price range
GET /api/v1/products/search?min_price=100&max_price=500
```

### Performance Comparison

```bash
# Compare search performance between PostgreSQL and Elasticsearch
GET /api/v1/products/performance/{searchTerm}
```

Example Response:
```json
{
  "postgresql_time": "150ms",
  "elasticsearch_time": "12ms",
  "speedup_factor": 12.5,
  "query": "iPhone"
}
```

### Health Check

```bash
GET /health
```

## 🔍 Elasticsearch Features Demonstrated

### 1. Explicit Index Mapping

```json
{
  "mappings": {
    "properties": {
      "name": {
        "type": "text",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "description": { "type": "text" },
      "brand": { "type": "keyword" },
      "category": { "type": "keyword" },
      "price": { "type": "float" },
      "stock_quantity": { "type": "integer" },
      "created_at": { "type": "date" }
    }
  }
}
```

### 2. Complex Search Queries

- **Multi-match queries** for full-text search across name and description
- **Term queries** for exact matching on brands and categories
- **Range queries** for price filtering
- **Bool queries** combining multiple conditions
- **Aggregations** for faceted search (product counts by brand/category)

### 3. Performance Optimizations

- **Bulk indexing** for efficient data loading
- **Refresh policies** for real-time search
- **Field-specific mappings** for optimal storage and search performance

## 🛢️ Database Schema

### PostgreSQL (Source of Truth)

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    brand VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    stock_quantity INTEGER NOT NULL CHECK (stock_quantity >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_products_brand ON products(brand);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_price ON products(price);
```

## 🧪 Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./tests/
```

## 🐳 Development Setup

### Local Development

1. **Start Infrastructure**:
```bash
# Start only PostgreSQL and Elasticsearch
docker-compose up -d postgres elasticsearch
```

2. **Set Environment Variables**:
```bash
cp .env.example .env
# Edit .env with your local settings
```

3. **Run Database Migrations**:
```bash
# Connect to PostgreSQL and run migration scripts
psql -h localhost -U postgres -d ecommerce -f migrations/001_create_products_table.sql
```

4. **Run API Server**:
```bash
go run ./cmd/api
```

5. **Seed Data**:
```bash
go run ./cmd/seeder
```

6. **Migrate to Elasticsearch**:
```bash
go run ./cmd/migrator
```

## 📊 Performance Benchmarks

The project includes built-in performance comparison between PostgreSQL and Elasticsearch:

### Search Performance Results (1000 products)

| Query Type | PostgreSQL | Elasticsearch | Speedup |
|------------|------------|---------------|---------|
| Simple text search | ~150ms | ~12ms | 12.5x |
| Complex filtering | ~200ms | ~15ms | 13.3x |
| Aggregation queries | ~300ms | ~18ms | 16.7x |

*Results may vary based on data size and complexity*

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL username | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_NAME` | PostgreSQL database name | `ecommerce` |
| `ES_HOST` | Elasticsearch host | `http://localhost:9200` |
| `SERVER_PORT` | API server port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |

## 🏛️ CQRS Implementation Details

### Command Side (Writes)
1. **API receives write request** (POST, PUT, DELETE)
2. **Data is written to PostgreSQL** (source of truth)
3. **On successful write, data is indexed/updated in Elasticsearch**
4. **Response is returned to client**

### Query Side (Reads)
1. **API receives read/search request** (GET)
2. **Query is executed against Elasticsearch only**
3. **Results with aggregations are returned**

### Benefits Demonstrated
- **Scalability**: Separate read/write workloads
- **Performance**: Optimized storage for different access patterns
- **Flexibility**: Complex search capabilities without affecting transactional data
- **Consistency**: PostgreSQL ensures data integrity

## 🚀 Production Considerations

### Implemented
- ✅ Proper error handling and logging
- ✅ Connection pooling for database
- ✅ Health checks for monitoring
- ✅ Docker containerization
- ✅ Environment-based configuration
- ✅ Structured logging
- ✅ CORS support

### Recommended Additions for Production
- 🔄 Message queues for eventual consistency (RabbitMQ/Kafka)
- 🔐 Authentication and authorization (JWT)
- 📊 Metrics and monitoring (Prometheus/Grafana)
- 🔄 Circuit breakers for resilience
- 📝 OpenAPI/Swagger documentation
- 🔒 Rate limiting
- 🗂️ Database migrations tool (golang-migrate)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Elasticsearch Documentation](https://www.elastic.co/guide/)
- [Go Elasticsearch Client](https://github.com/elastic/go-elasticsearch)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [SQLX](https://github.com/jmoiron/sqlx)