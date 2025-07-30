package middleware

import (
	"generalusermanagement/database"
	"generalusermanagement/models"
	"generalusermanagement/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header required",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		if isTokenBlacklisted(token) {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Token is invalid or expired",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid token",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_username", claims.Username)
		c.Set("client_id", claims.ClientID)
		c.Set("token", token)

		c.Next()
	}
}

// AdminOnly middleware ensures only admin or super_admin users can access the endpoint
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "User role not found in token",
			})
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if !utils.IsAdminRole(role) {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserManagementAccess middleware allows users with user management capabilities
func UserManagementAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "User role not found in token",
			})
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		if !utils.CanManageUsers(role) {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "User management access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserOrAdmin middleware allows both users and admins to access the endpoint
// But users can only access their own data
func UserOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "User role not found in token",
			})
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		
		// If admin or super admin, allow access to everything
		if utils.IsAdminRole(role) {
			c.Next()
			return
		}

		// If frontdesk or other user management roles, allow limited access
		if utils.CanManageUsers(role) {
			c.Next()
			return
		}

		// If regular user, check if they're accessing their own data
		userID, _ := c.Get("user_id")
		requestedUserID := c.Param("id")
		
		if requestedUserID != "" && userID != requestedUserID {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "Access denied: You can only access your own data",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuperAdminOnly middleware ensures only super_admin users can access the endpoint
func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "User role not found in token",
			})
			c.Abort()
			return
		}

		if userRole != models.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "Super admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleBasedAccess middleware allows access based on specific roles
func RoleBasedAccess(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "User role not found in token",
			})
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		
		// Check if user role is in allowed roles
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Insufficient permissions for this operation",
		})
		c.Abort()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isTokenBlacklisted checks if a token is in the blacklist
func isTokenBlacklisted(token string) bool {
	query := `SELECT COUNT(*) FROM token_blacklist WHERE token = $1 AND expires_at > NOW()`
	var count int
	
	if err := database.DB.QueryRow(query, token).Scan(&count); err != nil {
		return false
	}
	
	return count > 0
}

// BlacklistToken adds a token to the blacklist
func BlacklistToken(token string, expiresAt time.Time) error {
	query := `INSERT INTO token_blacklist (token, expires_at) VALUES ($1, $2)`
	_, err := database.DB.Exec(query, token, expiresAt)
	return err
}