package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/middleware"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
)

// UserServiceInterface defines the interface for user service operations
type UserServiceInterface interface {
	RegisterUser(req *models.UserRegistrationRequest) (*models.LoginResponse, error)
	LoginUser(req *models.UserLoginRequest) (*models.LoginResponse, error)
	GetUserProfile(userID uuid.UUID) (*models.UserProfile, error)
	UpdateUserProfile(userID uuid.UUID, req *models.UserUpdateRequest) (*models.User, error)
	GetUserByID(userID string) (*models.User, error)
	GetUsersByRole(role models.UserRole, page, size int) ([]models.User, int64, error)
	DeactivateUser(userID uuid.UUID) error
	ActivateUser(userID uuid.UUID) error
}

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userService UserServiceInterface
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register handles POST /auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req models.UserRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.userService.RegisterUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles POST /auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.userService.LoginUser(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProfile handles GET /users/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	profile, err := h.userService.GetUserProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile handles PUT /users/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Regular users can only update their own profile's basic info
	if !middleware.IsAdmin(c) {
		req.Role = nil
		req.IsActive = nil
	}

	user, err := h.userService.UpdateUserProfile(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUser handles GET /users/:id (Admin only)
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUsersByRole handles GET /users?role=vendor (Admin only)
func (h *UserHandler) GetUsersByRole(c *gin.Context) {
	roleStr := c.Query("role")
	if roleStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role parameter is required"})
		return
	}

	role := models.UserRole(roleStr)
	if role != models.RoleCustomer && role != models.RoleVendor && role != models.RoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	// Parse pagination parameters
	page := 1
	size := 20

	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := c.Query("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	users, total, err := h.userService.GetUsersByRole(role, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"size":  size,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser handles PUT /users/:id (Admin only)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUserProfile(userID, &req)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeactivateUser handles POST /users/:id/deactivate (Admin only)
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.userService.DeactivateUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deactivated successfully"})
}

// ActivateUser handles POST /users/:id/activate (Admin only)
func (h *UserHandler) ActivateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.userService.ActivateUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User activated successfully"})
}