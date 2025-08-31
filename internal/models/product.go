package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the e-commerce catalog
type Product struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	Brand         string    `json:"brand" db:"brand"`
	Category      string    `json:"category" db:"category"`
	Price         float64   `json:"price" db:"price"`
	StockQuantity int       `json:"stock_quantity" db:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// CreateProductRequest represents the request for creating a product
type CreateProductRequest struct {
	Name          string  `json:"name" binding:"required"`
	Description   string  `json:"description" binding:"required"`
	Brand         string  `json:"brand" binding:"required"`
	Category      string  `json:"category" binding:"required"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	StockQuantity int     `json:"stock_quantity" binding:"required,gte=0"`
}

// UpdateProductRequest represents the request for updating a product
type UpdateProductRequest struct {
	Name          *string  `json:"name,omitempty"`
	Description   *string  `json:"description,omitempty"`
	Brand         *string  `json:"brand,omitempty"`
	Category      *string  `json:"category,omitempty"`
	Price         *float64 `json:"price,omitempty"`
	StockQuantity *int     `json:"stock_quantity,omitempty"`
}

// SearchRequest represents search parameters
type SearchRequest struct {
	Query     string   `form:"q"`
	Brand     []string `form:"brand"`
	Category  []string `form:"category"`
	MinPrice  *float64 `form:"min_price"`
	MaxPrice  *float64 `form:"max_price"`
	Page      int      `form:"page" binding:"min=1"`
	Size      int      `form:"size" binding:"min=1,max=100"`
	SortBy    string   `form:"sort_by"`
	SortOrder string   `form:"sort_order"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Products     []Product              `json:"products"`
	Total        int64                  `json:"total"`
	Page         int                    `json:"page"`
	Size         int                    `json:"size"`
	Aggregations map[string]Aggregation `json:"aggregations"`
}

// Aggregation represents aggregation results
type Aggregation struct {
	Buckets []Bucket `json:"buckets"`
}

// Bucket represents an aggregation bucket
type Bucket struct {
	Key   string `json:"key"`
	Count int64  `json:"doc_count"`
}

// PerformanceMetrics represents performance comparison metrics
type PerformanceMetrics struct {
	PostgreSQLTime    time.Duration `json:"postgresql_time"`
	ElasticsearchTime time.Duration `json:"elasticsearch_time"`
	SpeedupFactor     float64       `json:"speedup_factor"`
	Query             string        `json:"query"`
}

// MarshalJSON implements custom JSON marshalling to format durations as human-readable strings
func (p PerformanceMetrics) MarshalJSON() ([]byte, error) {
	type perfAlias struct {
		PostgreSQLTime    string  `json:"postgresql_time"`
		ElasticsearchTime string  `json:"elasticsearch_time"`
		SpeedupFactor     float64 `json:"speedup_factor"`
		Query             string  `json:"query"`
	}

	a := perfAlias{
		PostgreSQLTime:    p.PostgreSQLTime.String(),
		ElasticsearchTime: p.ElasticsearchTime.String(),
		SpeedupFactor:     p.SpeedupFactor,
		Query:             p.Query,
	}

	return jsonMarshal(a)
}

// jsonMarshal is a tiny wrapper to avoid importing encoding/json at top-level twice in generated patches
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
