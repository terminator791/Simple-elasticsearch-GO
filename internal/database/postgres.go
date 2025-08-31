package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
		INSERT INTO products (id, vendor_id, name, description, brand, category, price, stock_quantity, sku, is_active, weight, dimensions, image_urls, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := c.db.Exec(query, 
		product.ID, product.VendorID, product.Name, product.Description, product.Brand, product.Category, 
		product.Price, product.StockQuantity, product.SKU, product.IsActive, product.Weight, product.Dimensions, 
		pq.Array(product.ImageURLs), pq.Array(product.Tags), product.CreatedAt, product.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetProductByID retrieves a product by ID
func (c *Client) GetProductByID(id string) (*models.Product, error) {
	var product models.Product
	query := `SELECT id, vendor_id, name, description, brand, category, price, stock_quantity, sku, is_active, weight, dimensions, image_urls, tags, created_at, updated_at 
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

// User-related database operations

// CreateUser creates a new user in the database
func (c *Client) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := c.db.Exec(query, user.ID, user.Email, user.Password, user.FirstName, user.LastName, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (c *Client) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at 
			  FROM users WHERE email = $1`

	err := c.db.Get(&user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (c *Client) GetUserByID(id string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at 
			  FROM users WHERE id = $1`

	err := c.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates a user in the database
func (c *Client) UpdateUser(id string, updates *models.UserUpdateRequest) (*models.User, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *updates.FirstName)
		argIndex++
	}
	if updates.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *updates.LastName)
		argIndex++
	}
	if updates.Role != nil {
		setParts = append(setParts, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, *updates.Role)
		argIndex++
	}
	if updates.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *updates.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		return c.GetUserByID(id)
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf(`
		UPDATE users 
		SET %s, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $%d
		RETURNING id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at`,
		setClause, argIndex)

	args = append(args, id)

	var user models.User
	err := c.db.Get(&user, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// GetUsersByRole retrieves users by role with pagination
func (c *Client) GetUsersByRole(role models.UserRole, page, size int) ([]models.User, int64, error) {
	offset := (page - 1) * size

	// Get users
	var users []models.User
	query := `SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at 
			  FROM users WHERE role = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	err := c.db.Select(&users, query, role, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM users WHERE role = $1`
	err = c.db.Get(&total, countQuery, role)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return users, total, nil
}

// GetCustomerStats retrieves order count and total spent for a customer
func (c *Client) GetCustomerStats(userID uuid.UUID) (int, float64, error) {
	var orderCount int
	var totalSpent float64

	query := `SELECT COALESCE(COUNT(*), 0), COALESCE(SUM(total_amount), 0) 
			  FROM orders WHERE user_id = $1 AND status NOT IN ('cancelled', 'refunded')`

	err := c.db.QueryRow(query, userID).Scan(&orderCount, &totalSpent)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get customer stats: %w", err)
	}

	return orderCount, totalSpent, nil
}

// GetUserReviewCount retrieves the number of reviews written by a user
func (c *Client) GetUserReviewCount(userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM reviews WHERE user_id = $1`

	err := c.db.Get(&count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user review count: %w", err)
	}

	return count, nil
}

// GetVendorProductCount retrieves the number of products owned by a vendor
func (c *Client) GetVendorProductCount(userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM products WHERE vendor_id = $1`

	err := c.db.Get(&count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get vendor product count: %w", err)
	}

	return count, nil
}

// GetVendorReviewCount retrieves the number of reviews for a vendor's products
func (c *Client) GetVendorReviewCount(userID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM reviews r 
		JOIN products p ON r.product_id = p.id 
		WHERE p.vendor_id = $1`

	err := c.db.Get(&count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get vendor review count: %w", err)
	}

	return count, nil
}
