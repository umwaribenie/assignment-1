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

	// API routes matching swagger basePath: /api/v1/users
	api := r.Group("/api/v1/users")
	{
		// Public authentication routes
		api.POST("/login", handlers.Login)
		api.POST("/register", handlers.Register)
		api.POST("/password-reset", handlers.RequestPasswordReset)
		api.POST("/reset-password/email", handlers.ResetPasswordEmail)
		api.POST("/confirm-password-reset-otp", handlers.ConfirmPasswordReset)
		api.POST("/verify-login-otp", handlers.VerifyLoginOTP)

		// Protected routes (require authentication)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Authentication status and logout
			protected.GET("/check", handlers.CheckAuth)
			protected.POST("/logout", handlers.Logout)
			protected.POST("/update-password", handlers.UpdatePassword)

			// Admin-only routes
			adminOnly := protected.Group("/")
			adminOnly.Use(middleware.AdminOnly())
			{
				// Get all users (admin only)
				adminOnly.GET("/", handlers.GetAllUsers)
				
				// Admin user creation
				adminOnly.POST("/registerusersbyadmin", handlers.RegisterUserByAdmin)
				
				// Admin password update
				adminOnly.POST("/update-password/admin", handlers.UpdatePasswordByAdmin)
			}

			// User management routes with access control
			userRoutes := protected.Group("/")
			userRoutes.Use(middleware.UserOrAdmin())
			{
				// Get user by ID (user can access own data, admin can access all)
				userRoutes.GET("/:id", handlers.GetUserByID)
				
				// Update user (user can update own data, admin can update all)
				userRoutes.PATCH("/:id", handlers.UpdateUser)
			}

			// Admin-only user management
			adminUserManagement := protected.Group("/")
			adminUserManagement.Use(middleware.AdminOnly())
			{
				// Delete user (admin only)
				adminUserManagement.DELETE("/:id", handlers.DeleteUser)
			}

			// Find user by slug (protected)
			protected.GET("/slug/:slug", handlers.GetUserBySlug)

			// File upload routes
			upload := protected.Group("/upload")
			{
				upload.POST("/profile-picture", handlers.UploadProfilePicture)
				upload.DELETE("/profile-picture", handlers.DeleteProfilePicture)
			}
		}
	}

	// Legacy routes for backward compatibility (without /api/v1/users prefix)
	legacy := r.Group("/")
	{
		// Legacy authentication routes
		legacyAuth := legacy.Group("auth")
		{
			legacyAuth.POST("/register", handlers.Register)
			legacyAuth.POST("/login", handlers.Login)
			legacyAuth.POST("/password-reset/request", handlers.RequestPasswordReset)
			legacyAuth.POST("/password-reset/confirm", handlers.ConfirmPasswordReset)
		}

		// Legacy protected routes
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