package handlers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
)

// TestProductModels tests the product model structures
func TestProductModels(t *testing.T) {
	// Test Product model
	product := models.Product{
		ID:            uuid.New(),
		Name:          "Test Product",
		Description:   "Test Description",
		Brand:         "Test Brand",
		Category:      "Test Category",
		Price:         99.99,
		StockQuantity: 10,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Test JSON marshaling
	productJSON, err := json.Marshal(product)
	assert.NoError(t, err)
	assert.Contains(t, string(productJSON), "Test Product")

	// Test JSON unmarshaling
	var unmarshaledProduct models.Product
	err = json.Unmarshal(productJSON, &unmarshaledProduct)
	assert.NoError(t, err)
	assert.Equal(t, product.Name, unmarshaledProduct.Name)
	assert.Equal(t, product.Price, unmarshaledProduct.Price)
}

// TestCreateProductRequest tests the create product request validation
func TestCreateProductRequest(t *testing.T) {
	req := models.CreateProductRequest{
		Name:          "iPhone 15 Pro",
		Description:   "Latest Apple smartphone",
		Brand:         "Apple",
		Category:      "Smartphones",
		Price:         999.99,
		StockQuantity: 50,
	}

	// Test JSON marshaling
	reqJSON, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(reqJSON), "iPhone 15 Pro")

	// Test that all required fields are present
	assert.NotEmpty(t, req.Name)
	assert.NotEmpty(t, req.Description)
	assert.NotEmpty(t, req.Brand)
	assert.NotEmpty(t, req.Category)
	assert.Greater(t, req.Price, 0.0)
	assert.GreaterOrEqual(t, req.StockQuantity, 0)
}

// TestSearchRequest tests the search request structure
func TestSearchRequest(t *testing.T) {
	minPrice := 100.0
	maxPrice := 500.0

	req := models.SearchRequest{
		Query:     "iPhone",
		Brand:     []string{"Apple", "Samsung"},
		Category:  []string{"Smartphones"},
		MinPrice:  &minPrice,
		MaxPrice:  &maxPrice,
		Page:      1,
		Size:      20,
		SortBy:    "price",
		SortOrder: "asc",
	}

	// Test that filters are properly set
	assert.Equal(t, "iPhone", req.Query)
	assert.Contains(t, req.Brand, "Apple")
	assert.Contains(t, req.Brand, "Samsung")
	assert.Equal(t, &minPrice, req.MinPrice)
	assert.Equal(t, &maxPrice, req.MaxPrice)
	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 20, req.Size)
}

// TestSearchResponse tests the search response structure
func TestSearchResponse(t *testing.T) {
	products := []models.Product{
		{
			ID:            uuid.New(),
			Name:          "iPhone 15",
			Description:   "Apple smartphone",
			Brand:         "Apple",
			Category:      "Smartphones",
			Price:         799.99,
			StockQuantity: 25,
		},
		{
			ID:            uuid.New(),
			Name:          "Galaxy S24",
			Description:   "Samsung smartphone",
			Brand:         "Samsung",
			Category:      "Smartphones",
			Price:         699.99,
			StockQuantity: 30,
		},
	}

	aggregations := map[string]models.Aggregation{
		"brands": {
			Buckets: []models.Bucket{
				{Key: "Apple", Count: 10},
				{Key: "Samsung", Count: 8},
			},
		},
	}

	response := models.SearchResponse{
		Products:     products,
		Total:        25,
		Page:         1,
		Size:         20,
		Aggregations: aggregations,
	}

	// Test response structure
	assert.Len(t, response.Products, 2)
	assert.Equal(t, int64(25), response.Total)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 20, response.Size)

	// Test aggregations
	brandAgg, exists := response.Aggregations["brands"]
	assert.True(t, exists)
	assert.Len(t, brandAgg.Buckets, 2)
	assert.Equal(t, "Apple", brandAgg.Buckets[0].Key)
	assert.Equal(t, int64(10), brandAgg.Buckets[0].Count)
}

// TestPerformanceMetrics tests the performance comparison structure
func TestPerformanceMetrics(t *testing.T) {
	metrics := models.PerformanceMetrics{
		PostgreSQLTime:    150 * time.Millisecond,
		ElasticsearchTime: 12 * time.Millisecond,
		SpeedupFactor:     12.5,
		Query:             "iPhone",
	}

	// Test metrics calculation
	assert.Equal(t, 150*time.Millisecond, metrics.PostgreSQLTime)
	assert.Equal(t, 12*time.Millisecond, metrics.ElasticsearchTime)
	assert.Equal(t, 12.5, metrics.SpeedupFactor)
	assert.Equal(t, "iPhone", metrics.Query)

	// Test JSON marshaling
	metricsJSON, err := json.Marshal(metrics)
	assert.NoError(t, err)
	assert.Contains(t, string(metricsJSON), "speedup_factor")
	assert.Contains(t, string(metricsJSON), "postgresql_time")
	assert.Contains(t, string(metricsJSON), "elasticsearch_time")
}
