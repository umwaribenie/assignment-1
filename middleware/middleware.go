package middleware

import (
	"net/http"
	"strings"
	"time"

	"generalusermanagement/database"
	"generalusermanagement/models"
	"generalusermanagement/utils"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// AuthMiddleware validates JWT token
func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Check if token is blacklisted
		var exists bool
		err := database.DB.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE token = $1 AND expires_at > NOW())
		`, tokenString).Scan(&exists)
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Database error",
			})
			c.Abort()
			return
		}

		if exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Token has been revoked",
			})
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid token",
			})
			c.Abort()
			return
		}

		// Extract user information from claims
		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid token claims",
			})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Verify user still exists and is active
		var userExists bool
		var userStatus string
		var isActive bool
		err = database.DB.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL), 
			       COALESCE(status, ''), COALESCE(is_active, false)
			FROM users WHERE id = $1 AND deleted_at IS NULL
		`, userID).Scan(&userExists, &userStatus, &isActive)

		if err != nil || !userExists || userStatus != string(models.StatusActive) || !isActive {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "User account is not active",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", userID)
		c.Set("user_role", role)
		c.Set("token", tokenString)

		c.Next()
	})
}

// AdminOnly middleware ensures only admin users can access the endpoint
func AdminOnly() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != string(models.RoleAdmin) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Access denied. Admin role required",
			})
			c.Abort()
			return
		}
		c.Next()
	})
}

// UserOrAdmin middleware allows both users and admins to access the endpoint
// Users can only access their own data, admins can access all data
func UserOrAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRole := c.GetString("user_role")
		currentUserID := c.GetString("user_id")
		targetUserID := c.Param("id")

		// Admin can access any user's data
		if userRole == string(models.RoleAdmin) {
			c.Next()
			return
		}

		// Regular users can only access their own data
		if userRole == string(models.RoleUser) && currentUserID == targetUserID {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied",
		})
		c.Abort()
	})
}

// Legacy function for backward compatibility
func BlacklistToken(token string, expiresAt time.Time) error {
	_, err := database.DB.Exec(`
		INSERT INTO token_blacklist (token, expires_at) 
		VALUES ($1, $2)
	`, token, expiresAt)
	return err
}