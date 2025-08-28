package middleware

import (
	"database/sql"
	"net/http"
	"shared"
	"user-service/database"
	"user-service/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := utils.ExtractTokenFromHeader(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Error: "Authorization header required"})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		var count int
		err = database.DB.QueryRow(`SELECT COUNT(*) FROM token_blacklist WHERE token = $1 AND expires_at > NOW()`, token).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Database error"})
			c.Abort()
			return
		}

		if count > 0 {
			c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Error: "Token has been revoked"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Error: "Invalid token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("username", claims.Username)
		c.Set("client_id", claims.ClientID)

		c.Next()
	}
}

// AdminMiddleware checks if user has admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Error: "User role not found"})
			c.Abort()
			return
		}

		userRole := role.(shared.UserRole)
		if userRole != shared.RoleAdmin && userRole != shared.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, shared.ErrorResponse{Error: "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}