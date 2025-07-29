package routes

import (
	"generalusermanagement/handlers"
	"generalusermanagement/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all application routes
func SetupRoutes() *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.Default()

	// Apply CORS middleware globally
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "General User Management API is running",
		})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API version 1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := v1.Group("/")
		{
			// Authentication routes
			auth := public.Group("auth")
			{
				auth.POST("/register", handlers.Register)
				auth.POST("/login", handlers.Login)
				auth.POST("/password-reset/request", handlers.RequestPasswordReset)
				auth.POST("/password-reset/confirm", handlers.ConfirmPasswordReset)
			}
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Authentication status check
			auth := protected.Group("auth")
			{
				auth.GET("/check", handlers.CheckAuth)
				auth.POST("/logout", handlers.Logout)
			}

			// User management routes
			users := protected.Group("users")
			{
				// Admin only routes
				adminOnly := users.Group("/")
				adminOnly.Use(middleware.AdminOnly())
				{
					adminOnly.GET("/", handlers.GetAllUsers)
				}

				// User or Admin routes (users can access their own data)
				userOrAdmin := users.Group("/")
				userOrAdmin.Use(middleware.UserOrAdmin())
				{
					userOrAdmin.GET("/:id", handlers.GetUserByID)
					userOrAdmin.PUT("/:id", handlers.UpdateUser)
				}

				// Admin only user management
				adminUserManagement := users.Group("/")
				adminUserManagement.Use(middleware.AdminOnly())
				{
					adminUserManagement.DELETE("/:id", handlers.DeleteUser)
				}

				// User-specific routes (authenticated user only)
				users.PUT("/password", handlers.UpdatePassword)
			}

			// File upload routes
			upload := protected.Group("upload")
			{
				upload.POST("/profile-picture", handlers.UploadProfilePicture)
				upload.DELETE("/profile-picture", handlers.DeleteProfilePicture)
			}
		}
	}

	// File serving routes (public access)
	r.GET("/files/:filename", handlers.ServeFile)

	// Legacy routes for backward compatibility (without /api/v1 prefix)
	legacy := r.Group("/")
	{
		// Public legacy routes
		legacyAuth := legacy.Group("auth")
		{
			legacyAuth.POST("/register", handlers.Register)
			legacyAuth.POST("/login", handlers.Login)
			legacyAuth.POST("/password-reset/request", handlers.RequestPasswordReset)
			legacyAuth.POST("/password-reset/confirm", handlers.ConfirmPasswordReset)
		}

		// Protected legacy routes
		legacyProtected := legacy.Group("/")
		legacyProtected.Use(middleware.AuthMiddleware())
		{
			// Legacy auth routes
			legacyAuthProtected := legacyProtected.Group("auth")
			{
				legacyAuthProtected.GET("/check", handlers.CheckAuth)
				legacyAuthProtected.POST("/logout", handlers.Logout)
			}

			// Legacy user routes
			legacyUsers := legacyProtected.Group("users")
			{
				// Admin only
				legacyAdminOnly := legacyUsers.Group("/")
				legacyAdminOnly.Use(middleware.AdminOnly())
				{
					legacyAdminOnly.GET("/", handlers.GetAllUsers)
					legacyAdminOnly.DELETE("/:id", handlers.DeleteUser)
				}

				// User or Admin
				legacyUserOrAdmin := legacyUsers.Group("/")
				legacyUserOrAdmin.Use(middleware.UserOrAdmin())
				{
					legacyUserOrAdmin.GET("/:id", handlers.GetUserByID)
					legacyUserOrAdmin.PUT("/:id", handlers.UpdateUser)
				}

				// User-specific
				legacyUsers.PUT("/password", handlers.UpdatePassword)
			}

			// Legacy upload routes
			legacyUpload := legacyProtected.Group("upload")
			{
				legacyUpload.POST("/profile-picture", handlers.UploadProfilePicture)
				legacyUpload.DELETE("/profile-picture", handlers.DeleteProfilePicture)
			}
		}
	}

	return r
}