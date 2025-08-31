package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/config"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
	"github.com/terminator791/Simple-elasticsearch-GO/pkg/logger"
)

// Client wraps the database connection
type Client struct {
	db *sqlx.DB
}

// NewClient creates a new database client
func NewClient(cfg *config.DatabaseConfig) (*Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.InfoLogger.Println("Successfully connected to PostgreSQL")

	return &Client{db: db}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// CreateProduct creates a new product in the database
func (c *Client) CreateProduct(product *models.Product) error {
	query := `
		INSERT INTO products (id, name, description, brand, category, price, stock_quantity, created_at, updated_at)
		VALUES (:id, :name, :description, :brand, :category, :price, :stock_quantity, :created_at, :updated_at)`

	_, err := c.db.NamedExec(query, product)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetProductByID retrieves a product by ID
func (c *Client) GetProductByID(id string) (*models.Product, error) {
	var product models.Product
	query := `SELECT id, name, description, brand, category, price, stock_quantity, created_at, updated_at 
			  FROM products WHERE id = $1`

	err := c.db.Get(&product, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// UpdateProduct updates a product in the database
func (c *Client) UpdateProduct(id string, updates *models.UpdateProductRequest) (*models.Product, error) {
	// Build dynamic update query
	setParts := []string{}
	args := map[string]interface{}{"id": id}

	if updates.Name != nil {
		setParts = append(setParts, "name = :name")
		args["name"] = *updates.Name
	}
	if updates.Description != nil {
		setParts = append(setParts, "description = :description")
		args["description"] = *updates.Description
	}
	if updates.Brand != nil {
		setParts = append(setParts, "brand = :brand")
		args["brand"] = *updates.Brand
	}
	if updates.Category != nil {
		setParts = append(setParts, "category = :category")
		args["category"] = *updates.Category
	}
	if updates.Price != nil {
		setParts = append(setParts, "price = :price")
		args["price"] = *updates.Price
	}
	if updates.StockQuantity != nil {
		setParts = append(setParts, "stock_quantity = :stock_quantity")
		args["stock_quantity"] = *updates.StockQuantity
	}

	if len(setParts) == 0 {
		return c.GetProductByID(id)
	}

	query := fmt.Sprintf(`
		UPDATE products 
		SET %s, updated_at = CURRENT_TIMESTAMP 
		WHERE id = :id
		RETURNING id, name, description, brand, category, price, stock_quantity, created_at, updated_at`,
		fmt.Sprintf("%v", setParts))

	query = fmt.Sprintf(`
		UPDATE products 
		SET %s, updated_at = CURRENT_TIMESTAMP 
		WHERE id = :id
		RETURNING id, name, description, brand, category, price, stock_quantity, created_at, updated_at`,
		fmt.Sprintf("%s", setParts[0]))

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf(`
		UPDATE products 
		SET %s, %s, updated_at = CURRENT_TIMESTAMP 
		WHERE id = :id
		RETURNING id, name, description, brand, category, price, stock_quantity, created_at, updated_at`,
			setParts[0], setParts[i])
	}

	// Rebuild query properly
	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	query = fmt.Sprintf(`
		UPDATE products 
		SET %s, updated_at = CURRENT_TIMESTAMP 
		WHERE id = :id
		RETURNING id, name, description, brand, category, price, stock_quantity, created_at, updated_at`,
		setClause)

	rows, err := c.db.NamedQuery(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var product models.Product
	err = rows.StructScan(&product)
	if err != nil {
		return nil, fmt.Errorf("failed to scan updated product: %w", err)
	}

	return &product, nil
}

// DeleteProduct deletes a product from the database
func (c *Client) DeleteProduct(id string) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := c.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// GetAllProducts retrieves all products from the database
func (c *Client) GetAllProducts() ([]models.Product, error) {
	var products []models.Product
	query := `SELECT id, name, description, brand, category, price, stock_quantity, created_at, updated_at 
			  FROM products ORDER BY created_at DESC`

	err := c.db.Select(&products, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all products: %w", err)
	}

	return products, nil
}

// SearchProducts performs a simple text search on PostgreSQL (for performance comparison)
func (c *Client) SearchProducts(searchTerm string) ([]models.Product, time.Duration, error) {
	start := time.Now()

	var products []models.Product
	query := `
		SELECT id, name, description, brand, category, price, stock_quantity, created_at, updated_at 
		FROM products 
		WHERE name ILIKE $1 OR description ILIKE $1 OR brand ILIKE $1 OR category ILIKE $1
		ORDER BY created_at DESC`

	searchPattern := "%" + searchTerm + "%"
	err := c.db.Select(&products, query, searchPattern)
	
	duration := time.Since(start)
	
	if err != nil {
		return nil, duration, fmt.Errorf("failed to search products: %w", err)
	}

	return products, duration, nil
}