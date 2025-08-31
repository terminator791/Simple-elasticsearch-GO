package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/terminator791/Simple-elasticsearch-GO/internal/config"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/database"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/elasticsearch"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/services"
	"github.com/terminator791/Simple-elasticsearch-GO/pkg/logger"
)

func main() {
	logger.InfoLogger.Println("Starting product seeder...")

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

	// Seed 5000 products
	const totalProducts = 5000
	logger.InfoLogger.Printf("Seeding %d products...", totalProducts)

	// Product data templates
	brands := []string{"Apple", "Samsung", "Sony", "Dell", "HP", "Lenovo", "Microsoft", "Google", "Amazon", "Nike", "Adidas", "Canon", "Nikon", "LG", "Panasonic", "Philips", "Bosch", "Siemens", "Toyota", "Honda"}
	
	productTemplates := []struct {
		namePrefix  string
		description string
		category    string
	}{
		{"iPhone", "Premium smartphone with advanced features", "Smartphones"},
		{"Galaxy", "Android smartphone with great display", "Smartphones"},
		{"MacBook", "Professional laptop for creative work", "Laptops"},
		{"ThinkPad", "Business laptop with reliability", "Laptops"},
		{"iPad", "Versatile tablet for work and entertainment", "Tablets"},
		{"Surface", "2-in-1 device for productivity", "Tablets"},
		{"AirPods", "Wireless earbuds with premium sound", "Headphones"},
		{"Headphones", "Over-ear headphones for audiophiles", "Headphones"},
		{"Camera", "Digital camera for photography enthusiasts", "Cameras"},
		{"Speaker", "Bluetooth speaker with rich sound", "Speakers"},
		{"Gaming Mouse", "Precision gaming mouse for esports", "Gaming"},
		{"Keyboard", "Mechanical keyboard for typing comfort", "Gaming"},
		{"Running Shoes", "Lightweight shoes for running", "Sports"},
		{"Fitness Tracker", "Wearable device for health monitoring", "Fitness"},
		{"Smart TV", "4K smart television with streaming", "Electronics"},
		{"Coffee Maker", "Automatic coffee brewing machine", "Kitchen"},
		{"Blender", "High-performance blender for smoothies", "Kitchen"},
		{"Drill", "Cordless drill for home projects", "Tools"},
		{"Vacuum", "Robotic vacuum cleaner", "Home"},
		{"Air Purifier", "HEPA air purifier for clean air", "Home"},
	}

	// Seed products in batches for better performance
	batchSize := 100
	successCount := 0
	
	for i := 0; i < totalProducts; i++ {
		// Select random template and brand
		template := productTemplates[rand.Intn(len(productTemplates))]
		brand := brands[rand.Intn(len(brands))]
		
		// Generate product name with variation
		productName := fmt.Sprintf("%s %s %d", brand, template.namePrefix, rand.Intn(100)+1)
		
		// Generate price based on category
		var basePrice float64
		switch template.category {
		case "Smartphones", "Laptops":
			basePrice = 500 + rand.Float64()*1500 // $500-$2000
		case "Tablets":
			basePrice = 200 + rand.Float64()*800  // $200-$1000
		case "Headphones", "Speakers":
			basePrice = 50 + rand.Float64()*450   // $50-$500
		case "Cameras":
			basePrice = 300 + rand.Float64()*1700 // $300-$2000
		case "Gaming":
			basePrice = 30 + rand.Float64()*170   // $30-$200
		case "Sports", "Fitness":
			basePrice = 25 + rand.Float64()*275   // $25-$300
		default:
			basePrice = 20 + rand.Float64()*480   // $20-$500
		}
		
		// Round to 2 decimal places
		price := float64(int(basePrice*100)) / 100
		
		// Generate stock quantity
		stockQuantity := rand.Intn(100) + 1 // 1-100
		
		// Create product request
		req := &models.CreateProductRequest{
			Name:          productName,
			Description:   fmt.Sprintf("%s - %s", template.description, productName),
			Brand:         brand,
			Category:      template.category,
			Price:         price,
			StockQuantity: stockQuantity,
		}

		// Create product
		product, err := productService.CreateProduct(req)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to create product %d: %v", i+1, err)
			continue
		}

		successCount++
		
		// Log progress every 500 products
		if (i+1)%500 == 0 {
			logger.InfoLogger.Printf("Progress: %d/%d products created", i+1, totalProducts)
		}

		// Add small delay to avoid overwhelming the system
		if i%batchSize == 0 {
			time.Sleep(100 * time.Millisecond)
		}

		// Store the product ID for logging (optional)
		_ = product
	}

	logger.InfoLogger.Printf("Seeding completed! Successfully created %d out of %d products", successCount, totalProducts)
	
	// Log some statistics
	logger.InfoLogger.Println("Seeder finished. You can now:")
	logger.InfoLogger.Println("- Use the API to search products: GET /api/v1/products/search")
	logger.InfoLogger.Println("- Get all products: GET /api/v1/products")
	logger.InfoLogger.Println("- Compare performance: GET /api/v1/products/performance/{searchTerm}")
}

func init() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
}