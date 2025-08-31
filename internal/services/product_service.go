package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/database"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/elasticsearch"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
	"github.com/terminator791/Simple-elasticsearch-GO/pkg/logger"
)

// ProductService implements CQRS pattern for product operations
type ProductService struct {
	db *database.Client
	es *elasticsearch.Client
}

// NewProductService creates a new product service
func NewProductService(db *database.Client, es *elasticsearch.Client) *ProductService {
	return &ProductService{
		db: db,
		es: es,
	}
}

// CreateProduct handles product creation (Command side)
// Writes to PostgreSQL first, then indexes in Elasticsearch
func (s *ProductService) CreateProduct(req *models.CreateProductRequest) (*models.Product, error) {
	// Create product model
	product := &models.Product{
		ID:            uuid.New(),
		Name:          req.Name,
		Description:   req.Description,
		Brand:         req.Brand,
		Category:      req.Category,
		Price:         req.Price,
		StockQuantity: req.StockQuantity,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Step 1: Save to PostgreSQL (source of truth)
	if err := s.db.CreateProduct(product); err != nil {
		return nil, fmt.Errorf("failed to create product in database: %w", err)
	}

	// Step 2: Index in Elasticsearch for search
	if err := s.es.IndexProduct(product); err != nil {
		logger.ErrorLogger.Printf("Failed to index product in Elasticsearch: %v", err)
		// Note: In production, you might want to implement a retry mechanism
		// or use a message queue for eventual consistency
	}

	logger.InfoLogger.Printf("Product created successfully: %s", product.ID)
	return product, nil
}

// UpdateProduct handles product updates (Command side)
// Updates PostgreSQL first, then updates Elasticsearch index
func (s *ProductService) UpdateProduct(id string, req *models.UpdateProductRequest) (*models.Product, error) {
	// Step 1: Update in PostgreSQL (source of truth)
	product, err := s.db.UpdateProduct(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update product in database: %w", err)
	}

	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Step 2: Update in Elasticsearch
	if err := s.es.IndexProduct(product); err != nil {
		logger.ErrorLogger.Printf("Failed to update product in Elasticsearch: %v", err)
	}

	logger.InfoLogger.Printf("Product updated successfully: %s", product.ID)
	return product, nil
}

// DeleteProduct handles product deletion (Command side)
// Deletes from PostgreSQL first, then removes from Elasticsearch
func (s *ProductService) DeleteProduct(id string) error {
	// Step 1: Delete from PostgreSQL (source of truth)
	if err := s.db.DeleteProduct(id); err != nil {
		return fmt.Errorf("failed to delete product from database: %w", err)
	}

	// Step 2: Delete from Elasticsearch
	if err := s.es.DeleteProduct(id); err != nil {
		logger.ErrorLogger.Printf("Failed to delete product from Elasticsearch: %v", err)
	}

	logger.InfoLogger.Printf("Product deleted successfully: %s", id)
	return nil
}

// SearchProducts handles product search (Query side)
// Uses Elasticsearch exclusively for all read operations
func (s *ProductService) SearchProducts(req *models.SearchRequest) (*models.SearchResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 100 {
		req.Size = 100
	}

	// Use Elasticsearch for all search operations
	result, _, err := s.es.SearchProducts(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return result, nil
}

// GetProductByID retrieves a single product (Query side)
// Uses PostgreSQL for now, but could be moved to Elasticsearch
func (s *ProductService) GetProductByID(id string) (*models.Product, error) {
	product, err := s.db.GetProductByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	return product, nil
}

// CompareSearchPerformance compares search performance between PostgreSQL and Elasticsearch
func (s *ProductService) CompareSearchPerformance(searchTerm string) (*models.PerformanceMetrics, error) {
	// Test PostgreSQL search
	_, pgDuration, err := s.db.SearchProducts(searchTerm)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL search failed: %w", err)
	}

	// Test Elasticsearch search
	searchReq := &models.SearchRequest{
		Query: searchTerm,
		Page:  1,
		Size:  20,
	}
	_, esDuration, err := s.es.SearchProducts(searchReq)
	if err != nil {
		return nil, fmt.Errorf("Elasticsearch search failed: %w", err)
	}

	speedupFactor := float64(pgDuration.Nanoseconds()) / float64(esDuration.Nanoseconds())

	metrics := &models.PerformanceMetrics{
		PostgreSQLTime:    pgDuration,
		ElasticsearchTime: esDuration,
		SpeedupFactor:     speedupFactor,
		Query:             searchTerm,
	}

	logger.InfoLogger.Printf("Performance comparison - PostgreSQL: %v, Elasticsearch: %v, Speedup: %.2fx",
		pgDuration, esDuration, speedupFactor)

	return metrics, nil
}

// GetAllProductsFromDB retrieves all products from PostgreSQL (for migration purposes)
func (s *ProductService) GetAllProductsFromDB() ([]models.Product, error) {
	return s.db.GetAllProducts()
}

// GetAllProducts retrieves all products using Elasticsearch (Query side)
// This demonstrates Elasticsearch's power for fast retrieval with aggregations
func (s *ProductService) GetAllProducts(page, size int) (*models.SearchResponse, error) {
	// Set defaults if not provided
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 50 // Default to 50 for all products view
	}
	if size > 1000 {
		size = 1000 // Cap at 1000 for performance
	}

	// Create a search request to get all products with aggregations
	req := &models.SearchRequest{
		Query:     "", // Empty query to match all products
		Page:      page,
		Size:      size,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// Use Elasticsearch to get all products with powerful aggregations
	result, _, err := s.es.SearchProducts(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get all products: %w", err)
	}

	return result, nil
}

// BulkIndexProducts performs bulk indexing to Elasticsearch
func (s *ProductService) BulkIndexProducts(products []models.Product) error {
	return s.es.BulkIndex(products)
}