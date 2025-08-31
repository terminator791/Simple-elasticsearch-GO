package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
)

// ProductServiceInterface defines the interface for product service operations
type ProductServiceInterface interface {
	CreateProduct(req *models.CreateProductRequest) (*models.Product, error)
	GetProductByID(id string) (*models.Product, error)
	UpdateProduct(id string, req *models.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(id string) error
	SearchProducts(req *models.SearchRequest) (*models.SearchResponse, error)
	GetAllProducts(page, size int) (*models.SearchResponse, error)
	CompareSearchPerformance(searchTerm string) (*models.PerformanceMetrics, error)
}

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	productService ProductServiceInterface
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService ProductServiceInterface) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProduct handles POST /products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.CreateProduct(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetProduct handles GET /products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")

	product, err := h.productService.GetProductByID(id)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct handles PUT /products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.UpdateProduct(id, &req)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct handles DELETE /products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	err := h.productService.DeleteProduct(id)
	if err != nil {
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetAllProducts handles GET /products
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	size := 50

	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := c.Query("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 {
			size = s
		}
	}

	// Get all products using Elasticsearch (showcasing its power)
	result, err := h.productService.GetAllProducts(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add metadata to showcase Elasticsearch capabilities
	response := gin.H{
		"products":     result.Products,
		"total":        result.Total,
		"page":         result.Page,
		"size":         result.Size,
		"aggregations": result.Aggregations,
		"meta": gin.H{
			"message":    "Data retrieved using Elasticsearch for fast search and aggregations",
			"powered_by": "Elasticsearch",
			"features": []string{
				"Fast full-text search",
				"Real-time aggregations",
				"Scalable pagination",
				"Advanced filtering capabilities",
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// SearchProducts handles GET /products/search
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	result, err := h.productService.SearchProducts(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ComparePerformance handles GET /products/performance/:searchTerm
func (h *ProductHandler) ComparePerformance(c *gin.Context) {
	searchTerm := c.Param("searchTerm")
	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search term is required"})
		return
	}

	metrics, err := h.productService.CompareSearchPerformance(searchTerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// HealthCheck handles GET /health
func (h *ProductHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "ecommerce-api",
	})
}