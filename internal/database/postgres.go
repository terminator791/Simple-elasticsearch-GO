package database

import (
	"database/sql"
	"encoding/json"
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

	// Handle nil slices
	var imageURLs, tags pq.StringArray
	if product.ImageURLs != nil {
		imageURLs = pq.StringArray(*product.ImageURLs)
	}
	if product.Tags != nil {
		tags = pq.StringArray(*product.Tags)
	}

	_, err := c.db.Exec(query, 
		product.ID, product.VendorID, product.Name, product.Description, product.Brand, product.Category, 
		product.Price, product.StockQuantity, product.SKU, product.IsActive, product.Weight, product.Dimensions, 
		imageURLs, tags, product.CreatedAt, product.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetProductByID retrieves a product by ID
func (c *Client) GetProductByID(id string) (*models.Product, error) {
	var product models.Product
	var imageURLs, tags pq.StringArray
	
	query := `SELECT id, vendor_id, name, description, brand, category, price, stock_quantity, sku, is_active, weight, dimensions, image_urls, tags, created_at, updated_at 
			  FROM products WHERE id = $1`

	err := c.db.QueryRow(query, id).Scan(
		&product.ID, &product.VendorID, &product.Name, &product.Description, &product.Brand, &product.Category, 
		&product.Price, &product.StockQuantity, &product.SKU, &product.IsActive, &product.Weight, &product.Dimensions, 
		&imageURLs, &tags, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Convert pq.StringArray to *[]string
	if len(imageURLs) > 0 {
		imgURLs := []string(imageURLs)
		product.ImageURLs = &imgURLs
	}
	if len(tags) > 0 {
		tagSlice := []string(tags)
		product.Tags = &tagSlice
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
	if updates.SKU != nil {
		setParts = append(setParts, "sku = :sku")
		args["sku"] = *updates.SKU
	}
	if updates.IsActive != nil {
		setParts = append(setParts, "is_active = :is_active")
		args["is_active"] = *updates.IsActive
	}
	if updates.Weight != nil {
		setParts = append(setParts, "weight = :weight")
		args["weight"] = *updates.Weight
	}
	if updates.Dimensions != nil {
		setParts = append(setParts, "dimensions = :dimensions")
		args["dimensions"] = *updates.Dimensions
	}
	if updates.ImageURLs != nil {
		setParts = append(setParts, "image_urls = :image_urls")
		args["image_urls"] = pq.StringArray(*updates.ImageURLs)
	}
	if updates.Tags != nil {
		setParts = append(setParts, "tags = :tags")
		args["tags"] = pq.StringArray(*updates.Tags)
	}

	if len(setParts) == 0 {
		return c.GetProductByID(id)
	}

	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	query := fmt.Sprintf(`
		UPDATE products 
		SET %s, updated_at = CURRENT_TIMESTAMP 
		WHERE id = :id
		RETURNING id, vendor_id, name, description, brand, category, price, stock_quantity, sku, is_active, weight, dimensions, image_urls, tags, created_at, updated_at`,
		setClause)

	var product models.Product
	var imageURLs, tags pq.StringArray
	
	rows, err := c.db.NamedQuery(query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	err = rows.Scan(
		&product.ID, &product.VendorID, &product.Name, &product.Description, &product.Brand, &product.Category,
		&product.Price, &product.StockQuantity, &product.SKU, &product.IsActive, &product.Weight, &product.Dimensions,
		&imageURLs, &tags, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan updated product: %w", err)
	}

	// Convert pq.StringArray to *[]string
	if len(imageURLs) > 0 {
		imgURLs := []string(imageURLs)
		product.ImageURLs = &imgURLs
	}
	if len(tags) > 0 {
		tagSlice := []string(tags)
		product.Tags = &tagSlice
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
		SELECT id, vendor_id, name, description, brand, category, price, stock_quantity, sku, is_active, weight, dimensions, image_urls, tags, created_at, updated_at 
		FROM products 
		WHERE name ILIKE $1 OR description ILIKE $1 OR brand ILIKE $1 OR category ILIKE $1
		ORDER BY created_at DESC`

	searchPattern := "%" + searchTerm + "%"
	rows, err := c.db.Query(query, searchPattern)

	duration := time.Since(start)

	if err != nil {
		return nil, duration, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		var imageURLs, tags pq.StringArray
		
		err := rows.Scan(
			&product.ID, &product.VendorID, &product.Name, &product.Description, &product.Brand, &product.Category,
			&product.Price, &product.StockQuantity, &product.SKU, &product.IsActive, &product.Weight, &product.Dimensions,
			&imageURLs, &tags, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, duration, fmt.Errorf("failed to scan product: %w", err)
		}

		// Convert pq.StringArray to *[]string
		if len(imageURLs) > 0 {
			imgURLs := []string(imageURLs)
			product.ImageURLs = &imgURLs
		}
		if len(tags) > 0 {
			tagSlice := []string(tags)
			product.Tags = &tagSlice
		}

		products = append(products, product)
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

// Order and Cart related database operations

// CreateCartItem adds an item to the cart
func (c *Client) CreateCartItem(item *models.CartItem) error {
	query := `
		INSERT INTO cart_items (id, user_id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, product_id) 
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = EXCLUDED.updated_at`

	_, err := c.db.Exec(query, item.ID, item.UserID, item.ProductID, item.Quantity, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create cart item: %w", err)
	}

	return nil
}

// GetCartItem retrieves a specific cart item
func (c *Client) GetCartItem(userID, productID uuid.UUID) (*models.CartItem, error) {
	var item models.CartItem
	query := `SELECT id, user_id, product_id, quantity, created_at, updated_at 
			  FROM cart_items WHERE user_id = $1 AND product_id = $2`

	err := c.db.Get(&item, query, userID, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("cart item not found")
		}
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}

	return &item, nil
}

// GetCartItemByID retrieves a cart item by ID
func (c *Client) GetCartItemByID(itemID uuid.UUID) (*models.CartItem, error) {
	var item models.CartItem
	query := `SELECT id, user_id, product_id, quantity, created_at, updated_at 
			  FROM cart_items WHERE id = $1`

	err := c.db.Get(&item, query, itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("cart item not found")
		}
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}

	return &item, nil
}

// GetCartItems retrieves all cart items for a user with product details
func (c *Client) GetCartItems(userID uuid.UUID) ([]models.CartItem, error) {
	var items []models.CartItem
	query := `
		SELECT ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			   p.id as "product.id", p.name as "product.name", p.description as "product.description",
			   p.brand as "product.brand", p.category as "product.category", p.price as "product.price",
			   p.stock_quantity as "product.stock_quantity", p.sku as "product.sku", 
			   p.is_active as "product.is_active", p.weight as "product.weight", 
			   p.dimensions as "product.dimensions", p.image_urls as "product.image_urls", 
			   p.tags as "product.tags", p.created_at as "product.created_at", 
			   p.updated_at as "product.updated_at"
		FROM cart_items ci 
		JOIN products p ON ci.product_id = p.id 
		WHERE ci.user_id = $1 
		ORDER BY ci.created_at DESC`

	rows, err := c.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.CartItem
		var product models.Product
		var imageURLs, tags pq.StringArray
		
		err := rows.Scan(
			&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt,
			&product.ID, &product.Name, &product.Description, &product.Brand, &product.Category,
			&product.Price, &product.StockQuantity, &product.SKU, &product.IsActive, &product.Weight,
			&product.Dimensions, &imageURLs, &tags, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		
		// Convert pq.StringArray to *[]string
		if len(imageURLs) > 0 {
			imgURLs := []string(imageURLs)
			product.ImageURLs = &imgURLs
		}
		if len(tags) > 0 {
			tagSlice := []string(tags)
			product.Tags = &tagSlice
		}
		
		item.Product = &product
		items = append(items, item)
	}

	return items, nil
}

// UpdateCartItemQuantity updates the quantity of a cart item
func (c *Client) UpdateCartItemQuantity(itemID uuid.UUID, quantity int) error {
	query := `UPDATE cart_items SET quantity = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	
	result, err := c.db.Exec(query, quantity, itemID)
	if err != nil {
		return fmt.Errorf("failed to update cart item quantity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cart item not found")
	}

	return nil
}

// DeleteCartItem removes an item from the cart
func (c *Client) DeleteCartItem(itemID uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE id = $1`
	
	result, err := c.db.Exec(query, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete cart item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cart item not found")
	}

	return nil
}

// ClearCart removes all items from a user's cart
func (c *Client) ClearCart(userID uuid.UUID) error {
	query := `DELETE FROM cart_items WHERE user_id = $1`
	
	_, err := c.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

// CreateOrder creates a new order
func (c *Client) CreateOrder(order *models.Order) error {
	query := `
		INSERT INTO orders (id, user_id, status, total_amount, shipping_cost, tax_amount, discount_amount, 
						   shipping_address, billing_address, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	shippingAddrJSON, _ := json.Marshal(order.ShippingAddress)
	billingAddrJSON, _ := json.Marshal(order.BillingAddress)

	_, err := c.db.Exec(query, order.ID, order.UserID, order.Status, order.TotalAmount, order.ShippingCost,
		order.TaxAmount, order.DiscountAmount, shippingAddrJSON, billingAddrJSON, order.Notes,
		order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// CreateOrderItem creates an order item
func (c *Client) CreateOrderItem(item *models.OrderItem) error {
	query := `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, total_price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := c.db.Exec(query, item.ID, item.OrderID, item.ProductID, item.Quantity,
		item.UnitPrice, item.TotalPrice, item.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}

	return nil
}

// UpdateProductStock updates the stock quantity of a product
func (c *Client) UpdateProductStock(productID uuid.UUID, newStock int) error {
	query := `UPDATE products SET stock_quantity = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	
	result, err := c.db.Exec(query, newStock, productID)
	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
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

// GetOrderByID retrieves an order by ID
func (c *Client) GetOrderByID(orderID uuid.UUID) (*models.Order, error) {
	var order models.Order
	var shippingAddrJSON, billingAddrJSON []byte
	
	query := `SELECT id, user_id, status, total_amount, shipping_cost, tax_amount, discount_amount,
					 shipping_address, billing_address, notes, created_at, updated_at
			  FROM orders WHERE id = $1`

	err := c.db.QueryRow(query, orderID).Scan(
		&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.ShippingCost,
		&order.TaxAmount, &order.DiscountAmount, &shippingAddrJSON, &billingAddrJSON,
		&order.Notes, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Unmarshal addresses
	jsonUnmarshal(shippingAddrJSON, &order.ShippingAddress)
	jsonUnmarshal(billingAddrJSON, &order.BillingAddress)

	return &order, nil
}

// GetOrderItems retrieves order items for an order
func (c *Client) GetOrderItems(orderID uuid.UUID) ([]models.OrderItem, error) {
	var items []models.OrderItem
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.unit_price, oi.total_price, oi.created_at,
			   p.name as "product.name", p.brand as "product.brand", p.category as "product.category"
		FROM order_items oi 
		JOIN products p ON oi.product_id = p.id 
		WHERE oi.order_id = $1 
		ORDER BY oi.created_at`

	rows, err := c.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		var product models.Product
		
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.CreatedAt,
			&product.Name, &product.Brand, &product.Category,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		
		product.ID = item.ProductID
		item.Product = &product
		items = append(items, item)
	}

	return items, nil
}

// GetUserOrders retrieves orders for a user with pagination
func (c *Client) GetUserOrders(userID uuid.UUID, page, size int) ([]models.Order, int64, error) {
	offset := (page - 1) * size

	// Get orders
	var orders []models.Order
	query := `SELECT id, user_id, status, total_amount, shipping_cost, tax_amount, discount_amount,
					 shipping_address, billing_address, notes, created_at, updated_at
			  FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := c.db.Query(query, userID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		var shippingAddrJSON, billingAddrJSON []byte
		
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.ShippingCost,
			&order.TaxAmount, &order.DiscountAmount, &shippingAddrJSON, &billingAddrJSON,
			&order.Notes, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}

		// Unmarshal addresses
		jsonUnmarshal(shippingAddrJSON, &order.ShippingAddress)
		jsonUnmarshal(billingAddrJSON, &order.BillingAddress)
		
		orders = append(orders, order)
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM orders WHERE user_id = $1`
	err = c.db.Get(&total, countQuery, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order count: %w", err)
	}

	return orders, total, nil
}

// UpdateOrderStatus updates the status of an order
func (c *Client) UpdateOrderStatus(orderID uuid.UUID, status models.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	
	result, err := c.db.Exec(query, status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// GetOrderAnalytics retrieves order analytics
func (c *Client) GetOrderAnalytics() (*models.OrderAnalytics, error) {
	analytics := &models.OrderAnalytics{
		OrdersByStatus: make(map[models.OrderStatus]int64),
		RevenueByMonth: make(map[string]float64),
		TopProducts:    []models.ProductSales{},
	}

	// Get total orders and revenue
	query := `SELECT COUNT(*), COALESCE(SUM(total_amount), 0) FROM orders WHERE status NOT IN ('cancelled', 'refunded')`
	err := c.db.QueryRow(query).Scan(&analytics.TotalOrders, &analytics.TotalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic analytics: %w", err)
	}

	// Calculate average order value
	if analytics.TotalOrders > 0 {
		analytics.AverageOrderValue = analytics.TotalRevenue / float64(analytics.TotalOrders)
	}

	// Get orders by status
	statusQuery := `SELECT status, COUNT(*) FROM orders GROUP BY status`
	rows, err := c.db.Query(statusQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status models.OrderStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		analytics.OrdersByStatus[status] = count
	}

	// Get revenue by month (last 12 months)
	monthQuery := `
		SELECT DATE_TRUNC('month', created_at) as month, SUM(total_amount) 
		FROM orders 
		WHERE created_at >= CURRENT_DATE - INTERVAL '12 months' 
		  AND status NOT IN ('cancelled', 'refunded')
		GROUP BY month 
		ORDER BY month`
	
	rows, err = c.db.Query(monthQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue by month: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var month time.Time
		var revenue float64
		if err := rows.Scan(&month, &revenue); err != nil {
			continue
		}
		analytics.RevenueByMonth[month.Format("2006-01")] = revenue
	}

	// Get top products
	topProductsQuery := `
		SELECT p.id, p.name, SUM(oi.quantity) as total_sold, SUM(oi.total_price) as total_revenue
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		JOIN orders o ON oi.order_id = o.id
		WHERE o.status NOT IN ('cancelled', 'refunded')
		GROUP BY p.id, p.name
		ORDER BY total_revenue DESC
		LIMIT 10`
	
	rows, err = c.db.Query(topProductsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product models.ProductSales
		if err := rows.Scan(&product.ProductID, &product.ProductName, &product.TotalSold, &product.TotalRevenue); err != nil {
			continue
		}
		analytics.TopProducts = append(analytics.TopProducts, product)
	}

	return analytics, nil
}

// Helper function for JSON unmarshalling
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
