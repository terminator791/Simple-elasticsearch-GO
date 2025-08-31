package main

import (
	"fmt"
	"log"
	"time"

	"github.com/terminator791/Simple-elasticsearch-GO/internal/config"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/database"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/elasticsearch"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/services"
	"github.com/terminator791/Simple-elasticsearch-GO/pkg/logger"
)

func main() {
	logger.InfoLogger.Println("Running performance benchmark...")

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

	// Test different search scenarios
	searchTerms := []string{"iPhone", "laptop", "headphones", "Nike", "Apple"}

	fmt.Println("Search Term\t\tPostgreSQL\tElasticsearch\tSpeedup")
	fmt.Println("---------------------------------------------------------------")

	for _, term := range searchTerms {
		metrics, err := productService.CompareSearchPerformance(term)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to run performance test for '%s': %v", term, err)
			continue
		}

		fmt.Printf("%s\t\t%v\t\t%v\t\t%.2fx\n",
			term,
			metrics.PostgreSQLTime.Truncate(time.Millisecond),
			metrics.ElasticsearchTime.Truncate(time.Millisecond),
			metrics.SpeedupFactor,
		)
	}

	// Test complex search scenarios
	fmt.Println("\nComplex Search Performance:")
	fmt.Println("Scenario\t\tPostgreSQL\tElasticsearch\tSpeedup")
	fmt.Println("---------------------------------------------------------------")

	// Test with filters
	searchReq := &models.SearchRequest{
		Query:    "smartphone",
		Brand:    []string{"Apple", "Samsung"},
		MinPrice: &[]float64{300}[0],
		MaxPrice: &[]float64{1000}[0],
		Page:     1,
		Size:     20,
	}

	start := time.Now()
	_, _, pgErr := db.SearchProducts("smartphone")
	pgDuration := time.Since(start)

	start = time.Now()
	_, esDuration, esErr := es.SearchProducts(searchReq)

	if pgErr == nil && esErr == nil {
		speedup := float64(pgDuration.Nanoseconds()) / float64(esDuration.Nanoseconds())
		fmt.Printf("Complex Filter\t\t%v\t\t%v\t\t%.2fx\n",
			pgDuration.Truncate(time.Millisecond),
			esDuration.Truncate(time.Millisecond),
			speedup,
		)
	}

	logger.InfoLogger.Println("Performance benchmark completed")
}
