package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/database"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
	"github.com/terminator791/Simple-elasticsearch-GO/pkg/logger"
)

// OrderService handles order-related operations
type OrderService struct {
	db *database.Client
}

// NewOrderService creates a new order service
func NewOrderService(db *database.Client) *OrderService {
	return &OrderService{
		db: db,
	}
}

// AddToCart adds a product to the user's cart
func (s *OrderService) AddToCart(userID uuid.UUID, req *models.AddToCartRequest) error {
	// Check if product exists and is available
	product, err := s.db.GetProductByID(req.ProductID.String())
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return fmt.Errorf("product not found")
	}
	if !product.IsAvailable() {
		return fmt.Errorf("product is not available")
	}
	if product.StockQuantity < req.Quantity {
		return fmt.Errorf("insufficient stock: only %d available", product.StockQuantity)
	}

	// Check if item already exists in cart
	existingItem, err := s.db.GetCartItem(userID, req.ProductID)
	if err != nil && err.Error() != "cart item not found" {
		return fmt.Errorf("failed to check existing cart item: %w", err)
	}

	if existingItem != nil {
		// Update quantity
		newQuantity := existingItem.Quantity + req.Quantity
		if newQuantity > product.StockQuantity {
			return fmt.Errorf("insufficient stock: only %d available", product.StockQuantity)
		}
		return s.db.UpdateCartItemQuantity(existingItem.ID, newQuantity)
	}

	// Add new item to cart
	cartItem := &models.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.db.CreateCartItem(cartItem)
	if err != nil {
		return fmt.Errorf("failed to add item to cart: %w", err)
	}

	logger.InfoLogger.Printf("Item added to cart: user=%s, product=%s, quantity=%d", userID, req.ProductID, req.Quantity)
	return nil
}

// GetCart retrieves the user's cart with product details
func (s *OrderService) GetCart(userID uuid.UUID) (*models.Cart, error) {
	cartItems, err := s.db.GetCartItems(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	cart := &models.Cart{
		Items: cartItems,
	}
	cart.CalculateTotal()

	return cart, nil
}

// UpdateCartItem updates the quantity of a cart item
func (s *OrderService) UpdateCartItem(userID uuid.UUID, itemID uuid.UUID, req *models.UpdateCartItemRequest) error {
	// Verify the cart item belongs to the user
	cartItem, err := s.db.GetCartItemByID(itemID)
	if err != nil {
		return fmt.Errorf("failed to get cart item: %w", err)
	}
	if cartItem.UserID != userID {
		return fmt.Errorf("cart item not found")
	}

	// If quantity is 0, remove the item
	if req.Quantity == 0 {
		return s.db.DeleteCartItem(itemID)
	}

	// Check stock availability
	product, err := s.db.GetProductByID(cartItem.ProductID.String())
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil || !product.IsAvailable() {
		return fmt.Errorf("product is no longer available")
	}
	if product.StockQuantity < req.Quantity {
		return fmt.Errorf("insufficient stock: only %d available", product.StockQuantity)
	}

	// Update quantity
	err = s.db.UpdateCartItemQuantity(itemID, req.Quantity)
	if err != nil {
		return fmt.Errorf("failed to update cart item: %w", err)
	}

	logger.InfoLogger.Printf("Cart item updated: user=%s, item=%s, quantity=%d", userID, itemID, req.Quantity)
	return nil
}

// RemoveFromCart removes an item from the cart
func (s *OrderService) RemoveFromCart(userID uuid.UUID, itemID uuid.UUID) error {
	// Verify the cart item belongs to the user
	cartItem, err := s.db.GetCartItemByID(itemID)
	if err != nil {
		return fmt.Errorf("failed to get cart item: %w", err)
	}
	if cartItem.UserID != userID {
		return fmt.Errorf("cart item not found")
	}

	// Remove the item
	err = s.db.DeleteCartItem(itemID)
	if err != nil {
		return fmt.Errorf("failed to remove item from cart: %w", err)
	}

	logger.InfoLogger.Printf("Item removed from cart: user=%s, item=%s", userID, itemID)
	return nil
}

// ClearCart removes all items from the user's cart
func (s *OrderService) ClearCart(userID uuid.UUID) error {
	err := s.db.ClearCart(userID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	logger.InfoLogger.Printf("Cart cleared for user: %s", userID)
	return nil
}

// CreateOrder creates an order from the user's cart
func (s *OrderService) CreateOrder(userID uuid.UUID, req *models.CreateOrderRequest) (*models.Order, error) {
	// Get cart items
	cart, err := s.GetCart(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// Validate stock availability for all items
	for _, item := range cart.Items {
		if item.Product == nil {
			return nil, fmt.Errorf("product information not available for item")
		}
		if !item.Product.IsAvailable() {
			return nil, fmt.Errorf("product %s is no longer available", item.Product.Name)
		}
		if item.Product.StockQuantity < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for %s: only %d available", item.Product.Name, item.Product.StockQuantity)
		}
	}

	// Calculate totals
	subtotal := cart.TotalPrice
	shippingCost := calculateShippingCost(subtotal)
	taxAmount := calculateTaxAmount(subtotal)
	
	// Create order
	order := &models.Order{
		ID:             uuid.New(),
		UserID:         userID,
		Status:         models.OrderStatusPending,
		TotalAmount:    subtotal,
		ShippingCost:   shippingCost,
		TaxAmount:      taxAmount,
		DiscountAmount: 0, // TODO: Implement discount logic
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		Notes:          req.Notes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Create order in database
	err = s.db.CreateOrder(order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items and update stock
	for _, cartItem := range cart.Items {
		orderItem := &models.OrderItem{
			ID:         uuid.New(),
			OrderID:    order.ID,
			ProductID:  cartItem.ProductID,
			Quantity:   cartItem.Quantity,
			UnitPrice:  cartItem.Product.Price,
			TotalPrice: float64(cartItem.Quantity) * cartItem.Product.Price,
			CreatedAt:  time.Now(),
		}

		err = s.db.CreateOrderItem(orderItem)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		// Update product stock
		newStock := cartItem.Product.StockQuantity - cartItem.Quantity
		err = s.db.UpdateProductStock(cartItem.ProductID, newStock)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to update stock for product %s: %v", cartItem.ProductID, err)
		}
	}

	// Clear the cart
	err = s.ClearCart(userID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to clear cart after order creation: %v", err)
	}

	// Get the complete order with items
	completeOrder, err := s.GetOrderByID(order.ID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get complete order after creation: %v", err)
		return order, nil
	}

	logger.InfoLogger.Printf("Order created successfully: %s for user %s", order.ID, userID)
	return completeOrder, nil
}

// GetOrderByID retrieves an order by ID with all details
func (s *OrderService) GetOrderByID(orderID uuid.UUID) (*models.Order, error) {
	order, err := s.db.GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get order items
	orderItems, err := s.db.GetOrderItems(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}

	order.Items = orderItems
	return order, nil
}

// GetUserOrders retrieves orders for a user with pagination
func (s *OrderService) GetUserOrders(userID uuid.UUID, page, size int) (*models.OrderSearchResponse, error) {
	orders, total, err := s.db.GetUserOrders(userID, page, size)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	return &models.OrderSearchResponse{
		Orders: orders,
		Total:  total,
		Page:   page,
		Size:   size,
	}, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(orderID uuid.UUID, status models.OrderStatus) error {
	err := s.db.UpdateOrderStatus(orderID, status)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	logger.InfoLogger.Printf("Order status updated: %s -> %s", orderID, status)
	return nil
}

// CancelOrder cancels an order and restores stock
func (s *OrderService) CancelOrder(orderID uuid.UUID) error {
	// Get order
	order, err := s.GetOrderByID(orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can be cancelled
	if !order.CanBeCancelled() {
		return fmt.Errorf("order cannot be cancelled (current status: %s)", order.Status)
	}

	// Restore stock for all items
	for _, item := range order.Items {
		product, err := s.db.GetProductByID(item.ProductID.String())
		if err != nil {
			logger.ErrorLogger.Printf("Failed to get product for stock restoration: %v", err)
			continue
		}
		if product != nil {
			newStock := product.StockQuantity + item.Quantity
			err = s.db.UpdateProductStock(item.ProductID, newStock)
			if err != nil {
				logger.ErrorLogger.Printf("Failed to restore stock for product %s: %v", item.ProductID, err)
			}
		}
	}

	// Update order status
	err = s.UpdateOrderStatus(orderID, models.OrderStatusCancelled)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	logger.InfoLogger.Printf("Order cancelled and stock restored: %s", orderID)
	return nil
}

// GetOrderAnalytics retrieves order analytics
func (s *OrderService) GetOrderAnalytics() (*models.OrderAnalytics, error) {
	analytics, err := s.db.GetOrderAnalytics()
	if err != nil {
		return nil, fmt.Errorf("failed to get order analytics: %w", err)
	}

	return analytics, nil
}

// Helper functions

func calculateShippingCost(subtotal float64) float64 {
	if subtotal >= 100 {
		return 0 // Free shipping for orders over $100
	}
	return 9.99
}

func calculateTaxAmount(subtotal float64) float64 {
	return subtotal * 0.08 // 8% tax rate
}