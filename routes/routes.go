package routes

import (
	"generalusermanagement/handlers"
	"generalusermanagement/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "API is running"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// NEW API ROUTES
	v1 := r.Group("/api/v1/users")
	{
		v1.POST("/register", handlers.RegisterNewUser)
		v1.POST("/login", handlers.Login)
		v1.POST("/password-reset", handlers.RequestPasswordReset)
		v1.POST("/confirm-password-reset-otp", handlers.ConfirmPasswordReset)

		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/check", handlers.CheckAuth)
			protected.GET("/", handlers.GetAllUsers)
			protected.GET("/:id", handlers.GetUserByID)
			protected.GET("/slug/:slug", handlers.GetUserBySlug)
			protected.PATCH("/:id", handlers.UpdateUser)
			protected.DELETE("/:id", handlers.DeleteUser)
			protected.POST("/registerusersbyadmin", handlers.CreateUserByAdmin)
			protected.POST("/update-password", handlers.UpdatePassword)
		}
	}

	r.POST("/reset-password/email", handlers.ResetPasswordWithEmail)

	authGroup := r.Group("/")
	authGroup.Use(middleware.AuthMiddleware())
	{
		authGroup.POST("/update-password", handlers.AuthUpdatePassword) // ✅ NEW SECTION ADDED
	}

	legacy := r.Group("/")
	{
		auth := legacy.Group("auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/logout", handlers.Logout)
		}
	}

	return r
}