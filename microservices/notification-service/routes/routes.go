package routes

import (
	"notification-service/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Initialize notification handler
	notificationHandler := handlers.NewNotificationHandler()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "notification-service"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Notification routes
		notifications := v1.Group("/notifications")
		{
			// Send notifications
			notifications.POST("/email", notificationHandler.SendEmail)
			notifications.POST("/sms", notificationHandler.SendSMS)
			notifications.POST("/email/template", notificationHandler.SendTemplateEmail)
			
			// Get notification details
			notifications.GET("/:id", notificationHandler.GetNotification)
			
			// Template management
			notifications.GET("/templates/email", notificationHandler.GetEmailTemplates)
			notifications.GET("/templates/sms", notificationHandler.GetSMSTemplates)
		}
	}

	return r
}