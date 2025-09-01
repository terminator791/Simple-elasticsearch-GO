package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/handlers"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
)

// MockProductService is a mock implementation of ProductService
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(req *models.CreateProductRequest) (*models.Product, error) {
	args := m.Called(req)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) GetProductByID(id string) (*models.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) UpdateProduct(id string, req *models.UpdateProductRequest) (*models.Product, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductService) DeleteProduct(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductService) SearchProducts(req *models.SearchRequest) (*models.SearchResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*models.SearchResponse), args.Error(1)
}

func (m *MockProductService) GetAllProducts(page, size int) (*models.SearchResponse, error) {
	args := m.Called(page, size)
	return args.Get(0).(*models.SearchResponse), args.Error(1)
}

func (m *MockProductService) CompareSearchPerformance(searchTerm string) (*models.PerformanceMetrics, error) {
	args := m.Called(searchTerm)
	return args.Get(0).(*models.PerformanceMetrics), args.Error(1)
}

func TestProductHandler_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockProductService)
	handler := handlers.NewProductHandler(mockService)

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "ecommerce-api", response["service"])
}

func TestProductHandler_CreateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockProductService)
	handler := handlers.NewProductHandler(mockService)

	// Test successful creation
	t.Run("successful creation", func(t *testing.T) {
		expectedProduct := &models.Product{
			Name:        "Test Product",
			Description: "Test Description",
			Brand:       "Test Brand",
			Category:    "Test Category",
			Price:       99.99,
		}

		mockService.On("CreateProduct", mock.AnythingOfType("*models.CreateProductRequest")).Return(expectedProduct, nil).Once()

		createReq := models.CreateProductRequest{
			Name:          "Test Product",
			Description:   "Test Description",
			Brand:         "Test Brand",
			Category:      "Test Category",
			Price:         99.99,
			StockQuantity: 10,
			SKU:           "TEST-SKU-001",
		}

		reqBody, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/products", handler.CreateProduct)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	// Test invalid request
	t.Run("invalid request", func(t *testing.T) {
		invalidReq := map[string]interface{}{
			"name": "", // missing required field
		}

		reqBody, _ := json.Marshal(invalidReq)
		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/products", handler.CreateProduct)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestProductHandler_GetAllProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockProductService)
	handler := handlers.NewProductHandler(mockService)

	// Test successful retrieval
	t.Run("successful retrieval", func(t *testing.T) {
		expectedResponse := &models.SearchResponse{
			Products: []models.Product{
				{
					Name:        "iPhone 15 Pro",
					Description: "Latest Apple smartphone",
					Brand:       "Apple",
					Category:    "Smartphones",
					Price:       999.99,
				},
				{
					Name:        "Samsung Galaxy S24",
					Description: "Android flagship smartphone",
					Brand:       "Samsung",
					Category:    "Smartphones",
					Price:       899.99,
				},
			},
			Total: 2,
			Page:  1,
			Size:  50,
			Aggregations: map[string]models.Aggregation{
				"brands": {
					Buckets: []models.Bucket{
						{Key: "Apple", Count: 1},
						{Key: "Samsung", Count: 1},
					},
				},
			},
		}

		mockService.On("GetAllProducts", 1, 50).Return(expectedResponse, nil).Once()

		req, _ := http.NewRequest("GET", "/products", nil)
		w := httptest.NewRecorder()
		router := gin.New()
		router.GET("/products", handler.GetAllProducts)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify response structure
		assert.Contains(t, response, "products")
		assert.Contains(t, response, "total")
		assert.Contains(t, response, "aggregations")
		assert.Contains(t, response, "meta")

		// Verify meta information highlighting Elasticsearch
		meta := response["meta"].(map[string]interface{})
		assert.Equal(t, "Elasticsearch", meta["powered_by"])
		assert.Contains(t, meta["message"], "Elasticsearch")

		mockService.AssertExpectations(t)
	})

	// Test with pagination parameters
	t.Run("with pagination", func(t *testing.T) {
		expectedResponse := &models.SearchResponse{
			Products: []models.Product{},
			Total:    100,
			Page:     2,
			Size:     25,
		}

		mockService.On("GetAllProducts", 2, 25).Return(expectedResponse, nil).Once()

		req, _ := http.NewRequest("GET", "/products?page=2&size=25", nil)
		w := httptest.NewRecorder()
		router := gin.New()
		router.GET("/products", handler.GetAllProducts)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
