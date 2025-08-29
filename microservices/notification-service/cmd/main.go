package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/handlers"
	"notification-service/internal/kafka"
	"notification-service/internal/models"
	"notification-service/internal/repository"
	"notification-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Notification Service API
// @version 1.0
// @description This is the Notification Service API for the microservices architecture.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8082
// @BasePath /api/v1
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repository
	notificationRepo := repository.NewNotificationRepository(db)

	// Initialize service
	notificationService := service.NewNotificationService(notificationRepo)

	// Initialize Kafka consumer
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	kafkaTopic := getEnv("KAFKA_TOPIC_USER_EVENTS", "user-events")
	kafkaGroupID := getEnv("KAFKA_GROUP_ID", "notification-service")

	consumer := kafka.NewConsumer(kafkaBrokers, kafkaTopic, kafkaGroupID, notificationService)

	// Start Kafka consumer in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Printf("Starting Kafka consumer for topic: %s", kafkaTopic)
		consumer.Start(ctx)
	}()

	// Initialize handlers
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Notification routes
		notifications := api.Group("/notifications")
		{
			notifications.POST("/", notificationHandler.CreateNotification)
			notifications.GET("/", notificationHandler.GetAllNotifications)
			notifications.GET("/user/:userId", notificationHandler.GetNotificationsByUserID)
			notifications.GET("/:id", notificationHandler.GetNotificationByID)
			notifications.POST("/retry", notificationHandler.RetryFailedNotifications)
		}

		// Health check
		api.GET("/health", notificationHandler.HealthCheck)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := getEnv("PORT", "8082")
	serverAddr := fmt.Sprintf(":%s", port)

	// Graceful shutdown
	go func() {
		log.Printf("Notification Service starting on port %s", port)
		if err := router.Run(serverAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Notification Service...")

	// Cancel context to stop Kafka consumer
	cancel()

	// Give outstanding requests a deadline for completion
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Close Kafka consumer
	if err := consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	log.Println("Notification Service stopped")
}

// initDatabase initializes the database connection
func initDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "12345"),
		getEnv("DB_NAME", "notification_service"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&models.Notification{}); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}