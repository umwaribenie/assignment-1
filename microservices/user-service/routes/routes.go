package routes

import (
	"user-service/handlers"
	"user-service/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "user-service"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/password-reset/request", handlers.RequestPasswordReset)
			auth.POST("/password-reset/confirm", handlers.ConfirmPasswordReset)
		}

		users := v1.Group("/users")
		{
			// Public user creation
			users.POST("/", handlers.CreateUser)
			
			// Protected routes
			protected := users.Group("/")
			protected.Use(middleware.AuthMiddleware())
			{
				protected.GET("/", handlers.GetAllUsers)
				protected.GET("/:id", handlers.GetUserByID)
				protected.POST("/logout", handlers.Logout)
				
				// Admin only routes
				admin := protected.Group("/")
				admin.Use(middleware.AdminMiddleware())
				{
					// Add admin-specific user management routes here
				}
			}
		}
	}

	return r
}