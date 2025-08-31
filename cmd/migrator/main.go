package main

import (
	"log"

	"github.com/terminator791/Simple-elasticsearch-GO/internal/config"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/database"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/elasticsearch"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/services"
	"github.com/terminator791/Simple-elasticsearch-GO/pkg/logger"
)

func main() {
	logger.InfoLogger.Println("Starting data migration from PostgreSQL to Elasticsearch...")

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

	// Initialize service
	productService := services.NewProductService(db, es)

	// Get all products from PostgreSQL
	logger.InfoLogger.Println("Fetching all products from PostgreSQL...")
	products, err := productService.GetAllProductsFromDB()
	if err != nil {
		log.Fatal("Failed to get products from database:", err)
	}

	logger.InfoLogger.Printf("Found %d products to migrate", len(products))

	if len(products) == 0 {
		logger.InfoLogger.Println("No products found to migrate")
		return
	}

	// Bulk index to Elasticsearch
	logger.InfoLogger.Println("Bulk indexing products to Elasticsearch...")
	if err := productService.BulkIndexProducts(products); err != nil {
		log.Fatal("Failed to bulk index products:", err)
	}

	logger.InfoLogger.Printf("Successfully migrated %d products to Elasticsearch", len(products))
	logger.InfoLogger.Println("Migration completed!")
}