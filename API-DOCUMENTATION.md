# Enterprise E-commerce API Documentation

## Overview

This is a comprehensive enterprise-level e-commerce API built with Go, featuring CQRS architecture, JWT authentication, and advanced Elasticsearch integration. The system supports multi-vendor marketplaces with role-based access control.

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <token>
```

### User Roles

- **Customer**: Can browse products, manage cart, place orders, write reviews
- **Vendor**: Can manage their own products, view their orders and analytics
- **Admin**: Full system access, user management, analytics

## API Endpoints

### Authentication

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe",
  "role": "customer"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

### User Management

#### Get Profile
```http
GET /api/v1/users/profile
Authorization: Bearer <token>
```

#### Update Profile
```http
PUT /api/v1/users/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "first_name": "Jane",
  "last_name": "Smith"
}
```

#### Get Users by Role (Admin only)
```http
GET /api/v1/users?role=vendor&page=1&size=20
Authorization: Bearer <admin_token>
```

### Product Management

#### Create Product (Vendor/Admin only)
```http
POST /api/v1/products
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Wireless Headphones",
  "description": "High-quality wireless headphones with noise cancellation",
  "brand": "TechCorp",
  "category": "Electronics",
  "price": 199.99,
  "stock_quantity": 50,
  "sku": "TC-WH-001",
  "weight": 0.5,
  "dimensions": "20x15x8cm",
  "image_urls": ["https://example.com/image1.jpg"],
  "tags": ["wireless", "noise-cancellation", "bluetooth"]
}
```

#### Search Products
```http
GET /api/v1/products/search?q=headphones&brand=TechCorp&min_price=100&max_price=300&page=1&size=20
```

#### Get All Products
```http
GET /api/v1/products?page=1&size=20
```

#### Get Product by ID
```http
GET /api/v1/products/{product_id}
```

#### Update Product (Vendor/Admin only)
```http
PUT /api/v1/products/{product_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "price": 179.99,
  "stock_quantity": 30
}
```

#### Delete Product (Vendor/Admin only)
```http
DELETE /api/v1/products/{product_id}
Authorization: Bearer <token>
```

### Shopping Cart

#### Get Cart
```http
GET /api/v1/cart
Authorization: Bearer <token>
```

#### Add Item to Cart
```http
POST /api/v1/cart/items
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": "123e4567-e89b-12d3-a456-426614174000",
  "quantity": 2
}
```

#### Update Cart Item
```http
PUT /api/v1/cart/items/{item_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "quantity": 3
}
```

#### Remove Item from Cart
```http
DELETE /api/v1/cart/items/{item_id}
Authorization: Bearer <token>
```

#### Clear Cart
```http
DELETE /api/v1/cart
Authorization: Bearer <token>
```

### Order Management

#### Create Order
```http
POST /api/v1/orders
Authorization: Bearer <token>
Content-Type: application/json

{
  "shipping_address": {
    "street": "123 Main St",
    "city": "Anytown",
    "state": "CA",
    "zip_code": "12345",
    "country": "USA",
    "phone": "+1-555-123-4567"
  },
  "billing_address": {
    "street": "123 Main St",
    "city": "Anytown",
    "state": "CA",
    "zip_code": "12345",
    "country": "USA",
    "phone": "+1-555-123-4567"
  },
  "notes": "Please deliver after 5 PM"
}
```

#### Get User Orders
```http
GET /api/v1/orders?page=1&size=10
Authorization: Bearer <token>
```

#### Get Order by ID
```http
GET /api/v1/orders/{order_id}
Authorization: Bearer <token>
```

#### Cancel Order
```http
POST /api/v1/orders/{order_id}/cancel
Authorization: Bearer <token>
```

#### Update Order Status (Vendor/Admin only)
```http
PUT /api/v1/orders/{order_id}/status
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "shipped"
}
```

### Analytics (Admin only)

#### Get Order Analytics
```http
GET /api/v1/admin/orders/analytics
Authorization: Bearer <admin_token>
```

## Error Responses

All endpoints return errors in the following format:

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required or invalid token
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Data Models

### User
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "customer",
  "is_active": true,
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### Product
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "vendor_id": "123e4567-e89b-12d3-a456-426614174001",
  "name": "Wireless Headphones",
  "description": "High-quality wireless headphones",
  "brand": "TechCorp",
  "category": "Electronics",
  "price": 199.99,
  "stock_quantity": 50,
  "sku": "TC-WH-001",
  "is_active": true,
  "weight": 0.5,
  "dimensions": "20x15x8cm",
  "image_urls": ["https://example.com/image1.jpg"],
  "tags": ["wireless", "bluetooth"],
  "average_rating": 4.5,
  "review_count": 25,
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### Order
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "123e4567-e89b-12d3-a456-426614174001",
  "status": "pending",
  "total_amount": 199.99,
  "shipping_cost": 9.99,
  "tax_amount": 16.00,
  "discount_amount": 0.00,
  "shipping_address": {
    "street": "123 Main St",
    "city": "Anytown",
    "state": "CA",
    "zip_code": "12345",
    "country": "USA",
    "phone": "+1-555-123-4567"
  },
  "billing_address": {
    "street": "123 Main St",
    "city": "Anytown",
    "state": "CA",
    "zip_code": "12345",
    "country": "USA",
    "phone": "+1-555-123-4567"
  },
  "items": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174002",
      "product_id": "123e4567-e89b-12d3-a456-426614174003",
      "quantity": 1,
      "unit_price": 199.99,
      "total_price": 199.99,
      "product": {
        "name": "Wireless Headphones",
        "brand": "TechCorp"
      }
    }
  ],
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

### Cart
```json
{
  "items": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "product_id": "123e4567-e89b-12d3-a456-426614174001",
      "quantity": 2,
      "product": {
        "id": "123e4567-e89b-12d3-a456-426614174001",
        "name": "Wireless Headphones",
        "price": 199.99,
        "stock_quantity": 50
      },
      "created_at": "2023-01-01T00:00:00Z"
    }
  ],
  "total_items": 2,
  "total_price": 399.98
}
```

## Business Logic

### Order Status Flow
1. `pending` - Order created, payment pending
2. `confirmed` - Payment confirmed
3. `processing` - Order being prepared
4. `shipped` - Order shipped to customer
5. `delivered` - Order delivered successfully
6. `cancelled` - Order cancelled (only from pending/confirmed)
7. `refunded` - Order refunded

### Pricing Calculation
- **Subtotal**: Sum of all item prices
- **Shipping**: Free for orders over $100, otherwise $9.99
- **Tax**: 8% of subtotal
- **Total**: Subtotal + Shipping + Tax - Discounts

### Stock Management
- Stock is automatically reduced when orders are created
- Stock is restored when orders are cancelled
- Out-of-stock products cannot be added to cart
- Stock validation occurs at checkout

## Rate Limiting

All endpoints are subject to rate limiting:
- **Authentication endpoints**: 5 requests per minute per IP
- **General API**: 100 requests per minute per authenticated user
- **Search endpoints**: 20 requests per minute per IP

## Performance Features

### Elasticsearch Integration
- Full-text search across products
- Real-time aggregations for faceted search
- Performance comparison tools
- Advanced filtering and sorting

### CQRS Architecture
- Writes go to PostgreSQL (source of truth)
- Reads use Elasticsearch for optimal performance
- Eventual consistency with automatic synchronization

### Caching
- Product catalog cached in Elasticsearch
- User sessions cached with JWT
- Database connection pooling

## Security Features

### Authentication & Authorization
- JWT tokens with configurable expiration
- Argon2id password hashing
- Role-based access control
- Session management

### Data Protection
- Input validation and sanitization
- SQL injection prevention with parameterized queries
- XSS protection with proper encoding
- CORS configuration

### API Security
- Request rate limiting
- Authentication required for sensitive operations
- Audit logging for all mutations
- Secure password policies