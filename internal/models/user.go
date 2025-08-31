package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents user roles in the system
type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleVendor   UserRole = "vendor"
	RoleAdmin    UserRole = "admin"
)

// User represents a user in the e-commerce system
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"` // Hidden from JSON responses
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Role      UserRole  `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserRegistrationRequest represents the request for user registration
type UserRegistrationRequest struct {
	Email     string   `json:"email" binding:"required,email"`
	Password  string   `json:"password" binding:"required,min=6"`
	FirstName string   `json:"first_name" binding:"required"`
	LastName  string   `json:"last_name" binding:"required"`
	Role      UserRole `json:"role" binding:"omitempty,oneof=customer vendor"`
}

// UserLoginRequest represents the request for user login
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserUpdateRequest represents the request for updating user profile
type UserUpdateRequest struct {
	FirstName *string   `json:"first_name,omitempty"`
	LastName  *string   `json:"last_name,omitempty"`
	Role      *UserRole `json:"role,omitempty"`
	IsActive  *bool     `json:"is_active,omitempty"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

// UserProfile represents user profile with additional information
type UserProfile struct {
	User         *User `json:"user"`
	OrderCount   int   `json:"order_count"`
	ReviewCount  int   `json:"review_count"`
	TotalSpent   float64 `json:"total_spent,omitempty"` // Only for customers
	ProductCount int   `json:"product_count,omitempty"` // Only for vendors
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsVendor checks if the user has vendor role
func (u *User) IsVendor() bool {
	return u.Role == RoleVendor
}

// IsCustomer checks if the user has customer role
func (u *User) IsCustomer() bool {
	return u.Role == RoleCustomer
}