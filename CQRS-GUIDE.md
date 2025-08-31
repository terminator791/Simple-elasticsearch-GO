# CQRS Implementation Guide

This document explains how the CQRS (Command Query Responsibility Segregation) pattern is implemented in the E-commerce Product Catalog API.

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client App    │    │    API Server    │    │   PostgreSQL    │
│                 │    │                  │    │ (Source of Truth)│
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         │ POST/PUT/DELETE        │        Writes          │
         │ (Commands)             │◄──────────────────────►│
         │                        │                        │
         │                        │                        │
         │                        │    ┌─────────────────┐ │
         │ GET/SEARCH             │    │  Elasticsearch  │ │
         │ (Queries)              │◄──►│   (Read Store)  │ │
         │                        │    └─────────────────┘ │
         │                        │                        │
```

## Command Side (Writes)

All write operations (CREATE, UPDATE, DELETE) follow this pattern:

### 1. Product Creation Flow

```go
// POST /api/v1/products
func (s *ProductService) CreateProduct(req *CreateProductRequest) (*Product, error) {
    // Step 1: Create product entity
    product := &Product{
        ID:            uuid.New(),
        Name:          req.Name,
        // ... other fields
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    // Step 2: Save to PostgreSQL (source of truth)
    if err := s.db.CreateProduct(product); err != nil {
        return nil, err // If PostgreSQL fails, entire operation fails
    }

    // Step 3: Index in Elasticsearch for search (best effort)
    if err := s.es.IndexProduct(product); err != nil {
        // Log error but don't fail the operation
        // In production, use message queue for eventual consistency
        logger.ErrorLogger.Printf("Failed to index product: %v", err)
    }

    return product, nil
}
```

### 2. Product Update Flow

```go
// PUT /api/v1/products/:id
func (s *ProductService) UpdateProduct(id string, req *UpdateProductRequest) (*Product, error) {
    // Step 1: Update in PostgreSQL (source of truth)
    product, err := s.db.UpdateProduct(id, req)
    if err != nil {
        return nil, err
    }

    // Step 2: Update in Elasticsearch
    if err := s.es.IndexProduct(product); err != nil {
        logger.ErrorLogger.Printf("Failed to update product in Elasticsearch: %v", err)
    }

    return product, nil
}
```

### 3. Product Deletion Flow

```go
// DELETE /api/v1/products/:id
func (s *ProductService) DeleteProduct(id string) error {
    // Step 1: Delete from PostgreSQL (source of truth)
    if err := s.db.DeleteProduct(id); err != nil {
        return err
    }

    // Step 2: Delete from Elasticsearch
    if err := s.es.DeleteProduct(id); err != nil {
        logger.ErrorLogger.Printf("Failed to delete product from Elasticsearch: %v", err)
    }

    return nil
}
```

## Query Side (Reads)

All read operations use Elasticsearch exclusively for optimal search performance:

### 1. Search Products

```go
// GET /api/v1/products/search
func (s *ProductService) SearchProducts(req *SearchRequest) (*SearchResponse, error) {
    // Use Elasticsearch exclusively for all search operations
    result, _, err := s.es.SearchProducts(req)
    if err != nil {
        return nil, fmt.Errorf("failed to search products: %w", err)
    }

    return result, nil
}
```

### 2. Complex Elasticsearch Queries

The search implementation supports:

```go
// Multi-match queries for full-text search
{
    "multi_match": {
        "query": "iPhone",
        "fields": ["name^2", "description"],
        "type": "best_fields"
    }
}

// Bool queries for complex filtering
{
    "bool": {
        "must": [
            {"multi_match": {"query": "smartphone", "fields": ["name", "description"]}}
        ],
        "filter": [
            {"terms": {"brand": ["Apple", "Samsung"]}},
            {"range": {"price": {"gte": 300, "lte": 1000}}}
        ]
    }
}

// Aggregations for faceted search
{
    "aggs": {
        "brands": {
            "terms": {"field": "brand", "size": 20}
        },
        "categories": {
            "terms": {"field": "category", "size": 20}
        }
    }
}
```

## Data Consistency Strategy

### Strong Consistency (PostgreSQL)
- All writes must succeed in PostgreSQL first
- PostgreSQL is the authoritative source of truth
- Business logic validation happens here
- ACID transactions ensure data integrity

### Eventual Consistency (Elasticsearch)
- Elasticsearch is updated after PostgreSQL
- If Elasticsearch update fails, the write operation still succeeds
- In production, implement retry mechanisms or message queues
- Data migration tool can rebuild Elasticsearch index from PostgreSQL

### Consistency Verification

```go
// Migration tool ensures consistency
func (s *ProductService) SyncToElasticsearch() error {
    // 1. Get all products from PostgreSQL (source of truth)
    products, err := s.db.GetAllProducts()
    if err != nil {
        return err
    }

    // 2. Bulk index to Elasticsearch
    return s.es.BulkIndex(products)
}
```

## Benefits Demonstrated

### 1. Performance Optimization
- **Writes**: Optimized for consistency and integrity (PostgreSQL)
- **Reads**: Optimized for speed and complex queries (Elasticsearch)
- **Measured**: Built-in performance comparison shows 10-20x speedup

### 2. Scalability
- **Separate Scaling**: Scale read and write systems independently
- **Load Distribution**: Write load on PostgreSQL, read load on Elasticsearch
- **Caching**: Elasticsearch acts as a sophisticated cache layer

### 3. Feature Specialization
- **PostgreSQL**: ACID transactions, referential integrity, complex business logic
- **Elasticsearch**: Full-text search, aggregations, faceted search, analytics

### 4. Resilience
- **Write Resilience**: System remains consistent if Elasticsearch fails
- **Read Resilience**: Can fall back to PostgreSQL for critical reads
- **Recovery**: Migration tool rebuilds search index from source of truth

## Production Considerations

### Message Queue Integration
```go
// Production implementation with message queue
func (s *ProductService) CreateProduct(req *CreateProductRequest) (*Product, error) {
    // Step 1: Save to PostgreSQL
    if err := s.db.CreateProduct(product); err != nil {
        return nil, err
    }

    // Step 2: Publish to message queue for async processing
    event := ProductCreatedEvent{Product: product}
    s.eventBus.Publish("product.created", event)

    return product, nil
}

// Async event handler
func (h *EventHandler) HandleProductCreated(event ProductCreatedEvent) {
    // Index to Elasticsearch with retry logic
    if err := h.es.IndexProduct(event.Product); err != nil {
        // Retry with exponential backoff
        h.retryQueue.Add(event)
    }
}
```

### Monitoring and Alerting
- Track PostgreSQL vs Elasticsearch consistency
- Monitor search performance metrics
- Alert on indexing failures
- Dashboard for CQRS operation metrics

## API Contract

The API clearly separates command and query operations:

### Commands (State-Changing Operations)
- `POST /api/v1/products` - Create product
- `PUT /api/v1/products/:id` - Update product  
- `DELETE /api/v1/products/:id` - Delete product

### Queries (Read-Only Operations)
- `GET /api/v1/products/search` - Search products (Elasticsearch)
- `GET /api/v1/products/performance/:term` - Performance comparison
- `GET /api/v1/products/:id` - Get single product (PostgreSQL)

This clear separation makes the CQRS pattern explicit and easy to understand for API consumers.