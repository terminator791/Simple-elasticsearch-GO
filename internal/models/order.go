package models

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// Order represents an order in the e-commerce system
type Order struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	UserID        uuid.UUID   `json:"user_id" db:"user_id"`
	Status        OrderStatus `json:"status" db:"status"`
	TotalAmount   float64     `json:"total_amount" db:"total_amount"`
	ShippingCost  float64     `json:"shipping_cost" db:"shipping_cost"`
	TaxAmount     float64     `json:"tax_amount" db:"tax_amount"`
	DiscountAmount float64    `json:"discount_amount" db:"discount_amount"`
	ShippingAddress Address   `json:"shipping_address" db:"shipping_address"`
	BillingAddress  Address   `json:"billing_address" db:"billing_address"`
	Notes         string      `json:"notes" db:"notes"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
	Items         []OrderItem `json:"items,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OrderID   uuid.UUID `json:"order_id" db:"order_id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	UnitPrice float64   `json:"unit_price" db:"unit_price"`
	TotalPrice float64  `json:"total_price" db:"total_price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Product   *Product  `json:"product,omitempty"`
}

// Address represents shipping and billing addresses
type Address struct {
	Street     string `json:"street" db:"street"`
	City       string `json:"city" db:"city"`
	State      string `json:"state" db:"state"`
	ZipCode    string `json:"zip_code" db:"zip_code"`
	Country    string `json:"country" db:"country"`
	Phone      string `json:"phone" db:"phone"`
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Product   *Product  `json:"product,omitempty"`
}

// Cart represents a user's shopping cart
type Cart struct {
	Items      []CartItem `json:"items"`
	TotalItems int        `json:"total_items"`
	TotalPrice float64    `json:"total_price"`
}

// CreateOrderRequest represents the request for creating an order
type CreateOrderRequest struct {
	ShippingAddress Address `json:"shipping_address" binding:"required"`
	BillingAddress  Address `json:"billing_address" binding:"required"`
	Notes          string  `json:"notes"`
}

// AddToCartRequest represents the request for adding items to cart
type AddToCartRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

// UpdateCartItemRequest represents the request for updating cart items
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}

// OrderSearchRequest represents search parameters for orders
type OrderSearchRequest struct {
	UserID    *uuid.UUID   `form:"user_id"`
	Status    *OrderStatus `form:"status"`
	StartDate *time.Time   `form:"start_date"`
	EndDate   *time.Time   `form:"end_date"`
	Page      int          `form:"page" binding:"min=1"`
	Size      int          `form:"size" binding:"min=1,max=100"`
	SortBy    string       `form:"sort_by"`
	SortOrder string       `form:"sort_order"`
}

// OrderSearchResponse represents search results for orders
type OrderSearchResponse struct {
	Orders []Order `json:"orders"`
	Total  int64   `json:"total"`
	Page   int     `json:"page"`
	Size   int     `json:"size"`
}

// OrderAnalytics represents order analytics data
type OrderAnalytics struct {
	TotalOrders    int64   `json:"total_orders"`
	TotalRevenue   float64 `json:"total_revenue"`
	AverageOrderValue float64 `json:"average_order_value"`
	OrdersByStatus map[OrderStatus]int64 `json:"orders_by_status"`
	RevenueByMonth map[string]float64 `json:"revenue_by_month"`
	TopProducts    []ProductSales `json:"top_products"`
}

// ProductSales represents product sales data
type ProductSales struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	TotalSold   int       `json:"total_sold"`
	TotalRevenue float64  `json:"total_revenue"`
}

// GetGrandTotal calculates the grand total including shipping and tax
func (o *Order) GetGrandTotal() float64 {
	return o.TotalAmount + o.ShippingCost + o.TaxAmount - o.DiscountAmount
}

// CanBeCancelled checks if the order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusConfirmed
}

// CalculateTotal calculates the total price for the cart
func (c *Cart) CalculateTotal() {
	c.TotalItems = 0
	c.TotalPrice = 0
	
	for _, item := range c.Items {
		c.TotalItems += item.Quantity
		if item.Product != nil {
			c.TotalPrice += float64(item.Quantity) * item.Product.Price
		}
	}
}