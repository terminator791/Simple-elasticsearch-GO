package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/auth"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/database"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
	"github.com/terminator791/Simple-elasticsearch-GO/pkg/logger"
)

// UserService handles user-related operations
type UserService struct {
	db       *database.Client
	jwtService *auth.JWTService
	passwordService *auth.PasswordService
}

// NewUserService creates a new user service
func NewUserService(db *database.Client, jwtService *auth.JWTService, passwordService *auth.PasswordService) *UserService {
	return &UserService{
		db:              db,
		jwtService:      jwtService,
		passwordService: passwordService,
	}
}

// RegisterUser registers a new user
func (s *UserService) RegisterUser(req *models.UserRegistrationRequest) (*models.LoginResponse, error) {
	// Validate password strength
	if err := s.passwordService.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.db.GetUserByEmail(req.Email)
	if err != nil && err.Error() != "user not found" {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = models.RoleCustomer
	}

	// Create user
	user := &models.User{
		ID:        uuid.New(),
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := s.db.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	logger.InfoLogger.Printf("User registered successfully: %s", user.Email)

	return &models.LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

// LoginUser authenticates a user and returns a token
func (s *UserService) LoginUser(req *models.UserLoginRequest) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.db.GetUserByEmail(req.Email)
	if err != nil {
		if err.Error() == "user not found" {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Verify password
	isValid, err := s.passwordService.VerifyPassword(req.Password, user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !isValid {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	logger.InfoLogger.Printf("User logged in successfully: %s", user.Email)

	return &models.LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

// GetUserProfile gets a user's profile with additional information
func (s *UserService) GetUserProfile(userID uuid.UUID) (*models.UserProfile, error) {
	// Get user
	user, err := s.db.GetUserByID(userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	profile := &models.UserProfile{
		User: user,
	}

	// Get additional statistics based on user role
	switch user.Role {
	case models.RoleCustomer:
		// Get order count and total spent
		orderCount, totalSpent, err := s.db.GetCustomerStats(userID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to get customer stats: %v", err)
		} else {
			profile.OrderCount = orderCount
			profile.TotalSpent = totalSpent
		}

		// Get review count
		reviewCount, err := s.db.GetUserReviewCount(userID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to get review count: %v", err)
		} else {
			profile.ReviewCount = reviewCount
		}

	case models.RoleVendor:
		// Get product count
		productCount, err := s.db.GetVendorProductCount(userID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to get vendor product count: %v", err)
		} else {
			profile.ProductCount = productCount
		}

		// Get review count for vendor's products
		reviewCount, err := s.db.GetVendorReviewCount(userID)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to get vendor review count: %v", err)
		} else {
			profile.ReviewCount = reviewCount
		}
	}

	return profile, nil
}

// UpdateUserProfile updates a user's profile
func (s *UserService) UpdateUserProfile(userID uuid.UUID, req *models.UserUpdateRequest) (*models.User, error) {
	// Update user in database
	updatedUser, err := s.db.UpdateUser(userID.String(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.InfoLogger.Printf("User profile updated successfully: %s", updatedUser.Email)

	return updatedUser, nil
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	return s.db.GetUserByID(userID)
}

// GetUsersByRole gets users by role with pagination
func (s *UserService) GetUsersByRole(role models.UserRole, page, size int) ([]models.User, int64, error) {
	return s.db.GetUsersByRole(role, page, size)
}

// DeactivateUser deactivates a user account
func (s *UserService) DeactivateUser(userID uuid.UUID) error {
	req := &models.UserUpdateRequest{
		IsActive: &[]bool{false}[0],
	}

	_, err := s.db.UpdateUser(userID.String(), req)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	logger.InfoLogger.Printf("User deactivated: %s", userID)
	return nil
}

// ActivateUser activates a user account
func (s *UserService) ActivateUser(userID uuid.UUID) error {
	req := &models.UserUpdateRequest{
		IsActive: &[]bool{true}[0],
	}

	_, err := s.db.UpdateUser(userID.String(), req)
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	logger.InfoLogger.Printf("User activated: %s", userID)
	return nil
}