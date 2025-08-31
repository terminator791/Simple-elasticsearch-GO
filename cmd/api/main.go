package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/auth"
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

	// Initialize authentication services
	jwtService := auth.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.Issuer)
	passwordService := auth.NewPasswordService(nil)

	// Initialize services
	productService := services.NewProductService(db, es)
	userService := services.NewUserService(db, jwtService, passwordService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productService)
	userHandler := handlers.NewUserHandler(userService)

	// Setup router
	router := setupRouter(productHandler, userHandler, authMiddleware)

	// Start server
	logger.InfoLogger.Printf("Starting server on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRouter(productHandler *handlers.ProductHandler, userHandler *handlers.UserHandler, authMiddleware *middleware.AuthMiddleware) *gin.Engine {
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
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// User routes
		users := v1.Group("/users")
		{
			// Protected routes
			users.GET("/profile", authMiddleware.RequireAuth(), userHandler.GetProfile)
			users.PUT("/profile", authMiddleware.RequireAuth(), userHandler.UpdateProfile)

			// Admin only routes
			users.GET("", authMiddleware.RequireAdmin(), userHandler.GetUsersByRole)
			users.GET("/:id", authMiddleware.RequireAdmin(), userHandler.GetUser)
			users.PUT("/:id", authMiddleware.RequireAdmin(), userHandler.UpdateUser)
			users.POST("/:id/deactivate", authMiddleware.RequireAdmin(), userHandler.DeactivateUser)
			users.POST("/:id/activate", authMiddleware.RequireAdmin(), userHandler.ActivateUser)
		}

		// Product routes
		products := v1.Group("/products")
		{
			// Public routes
			products.GET("", productHandler.GetAllProducts)
			products.GET("/search", productHandler.SearchProducts)
			products.GET("/performance/batch", productHandler.BatchComparePerformance)
			products.GET("/performance/:searchTerm", productHandler.ComparePerformance)
			products.GET("/:id", productHandler.GetProduct)

			// Protected routes (require authentication)
			products.POST("", authMiddleware.RequireVendorOrAdmin(), productHandler.CreateProduct)
			products.PUT("/:id", authMiddleware.RequireVendorOrAdmin(), productHandler.UpdateProduct)
			products.DELETE("/:id", authMiddleware.RequireVendorOrAdmin(), productHandler.DeleteProduct)
		}
	}

	return router
}
