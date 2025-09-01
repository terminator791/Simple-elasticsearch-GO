package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/auth"
	"github.com/terminator791/Simple-elasticsearch-GO/internal/models"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtService *auth.JWTService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_claims", claims)

		c.Next()
	}
}

// RequireRole middleware that requires specific role(s)
func (m *AuthMiddleware) RequireRole(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check authentication
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_claims", claims)

		c.Next()
	}
}

// OptionalAuth middleware that extracts user info if token is provided
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token != "" {
			claims, err := m.jwtService.ValidateToken(token)
			if err == nil {
				// Set user info in context if token is valid
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
				c.Set("user_role", claims.Role)
				c.Set("user_claims", claims)
			}
		}

		c.Next()
	}
}

// RequireAdmin middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(models.RoleAdmin)
}

// RequireVendorOrAdmin middleware that requires vendor or admin role
func (m *AuthMiddleware) RequireVendorOrAdmin() gin.HandlerFunc {
	return m.RequireRole(models.RoleVendor, models.RoleAdmin)
}

// RequireOwnershipOrAdmin middleware that requires resource ownership or admin role
func (m *AuthMiddleware) RequireOwnershipOrAdmin(resourceOwnerID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check authentication
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if user is admin or owns the resource
		if claims.Role != models.RoleAdmin && claims.UserID.String() != resourceOwnerID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_claims", claims)

		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// GetCurrentUserID helper function to get current user ID from context
func GetCurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}

	id, ok := userID.(uuid.UUID)
	return id, ok
}

// GetCurrentUserRole helper function to get current user role from context
func GetCurrentUserRole(c *gin.Context) (models.UserRole, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	role, ok := userRole.(models.UserRole)
	return role, ok
}

// GetCurrentUserClaims helper function to get current user claims from context
func GetCurrentUserClaims(c *gin.Context) (*auth.Claims, bool) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*auth.Claims)
	return userClaims, ok
}

// IsAdmin helper function to check if current user is admin
func IsAdmin(c *gin.Context) bool {
	role, exists := GetCurrentUserRole(c)
	return exists && role == models.RoleAdmin
}

// IsVendor helper function to check if current user is vendor
func IsVendor(c *gin.Context) bool {
	role, exists := GetCurrentUserRole(c)
	return exists && role == models.RoleVendor
}

// IsCustomer helper function to check if current user is customer
func IsCustomer(c *gin.Context) bool {
	role, exists := GetCurrentUserRole(c)
	return exists && role == models.RoleCustomer
}