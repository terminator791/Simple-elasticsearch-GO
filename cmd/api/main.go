package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/config"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/database"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/elasticsearch"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/handlers"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/middleware"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/services"
	"github.com/terminator791/Simple-elasticsearch-GO/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewClient(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Elasticsearch
	es, err := elasticsearch.NewClient(&cfg.Elasticsearch)
	if err != nil {
		log.Fatal("Failed to connect to Elasticsearch:", err)
	}

	// Initialize services
	productService := services.NewProductService(db, es)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productService)

	// Setup router
	router := setupRouter(productHandler)

	// Start server
	logger.InfoLogger.Printf("Starting server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRouter(productHandler *handlers.ProductHandler) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", productHandler.HealthCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("", productHandler.CreateProduct)
			products.GET("/search", productHandler.SearchProducts)
			products.GET("/performance/:searchTerm", productHandler.ComparePerformance)
			products.GET("/:id", productHandler.GetProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}
	}

	return router
}