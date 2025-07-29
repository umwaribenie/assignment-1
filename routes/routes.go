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

	// File serving routes (public access)
	r.GET("/files/:filename", handlers.ServeFile)

	// API version 1 routes with base path /api/v1/users
	v1 := r.Group("/api/v1/users")
	{
		// Public routes (no authentication required)
		
		// User Registration
		v1.POST("/register", handlers.RegisterUser)
		
		// Authentication routes
		v1.POST("/login", handlers.Login)
		v1.POST("/verify-login-otp", handlers.VerifyLoginOTP)
		v1.POST("/password-reset", handlers.RequestPasswordReset)
		v1.POST("/reset-password/email", handlers.ResetPasswordWithEmail)
		v1.POST("/confirm-password-reset-otp", handlers.ConfirmPasswordReset)

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Authentication status check
			protected.GET("/check", handlers.CheckAuth)
			
			// Password update routes
			protected.POST("/update-password", handlers.UpdatePassword)

			// User lookup routes
			protected.GET("/slug/:slug", handlers.GetUserBySlug)
			protected.GET("/:id", handlers.GetUserByID)
			
			// User management routes that require proper authorization
			userOrAdmin := protected.Group("/")
			userOrAdmin.Use(middleware.UserOrAdmin())
			{
				userOrAdmin.PATCH("/:id", handlers.UpdateUser)
			}

			// Admin-only routes
			adminOnly := protected.Group("/")
			adminOnly.Use(middleware.AdminOnly())
			{
				// Get all users with advanced filtering
				adminOnly.GET("/", handlers.GetAllUsers)
				
				// Admin user creation
				adminOnly.POST("/registerusersbyadmin", handlers.CreateUserByAdmin)
				
				// Admin password update
				adminOnly.POST("/update-password/admin", handlers.UpdatePasswordByAdmin)
				
				// User deletion
				adminOnly.DELETE("/:id", handlers.DeleteUser)
			}
		}
	}

	// Legacy routes for backward compatibility (original structure)
	legacy := r.Group("/")
	{
		// Public legacy auth routes
		legacyAuth := legacy.Group("auth")
		{
			legacyAuth.POST("/register", handlers.RegisterUser) // Updated to use new handler
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

	// Additional routes to match the exact Swagger API structure
	apiV1 := r.Group("/api/v1")
	{
		// File upload routes (keeping the existing structure)
		protectedApi := apiV1.Group("/")
		protectedApi.Use(middleware.AuthMiddleware())
		{
			upload := protectedApi.Group("upload")
			{
				upload.POST("/profile-picture", handlers.UploadProfilePicture)
				upload.DELETE("/profile-picture", handlers.DeleteProfilePicture)
			}
		}
	}

	return r
}