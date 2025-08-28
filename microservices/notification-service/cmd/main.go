package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/handlers"
	"notification-service/internal/kafka"
	"notification-service/internal/repository"
	"notification-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
)

// @title Notification Service API
// @version 1.0
// @description Notification management microservice with Kafka event processing
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

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
	defer db.Close()

	// Initialize repositories
	notificationRepo := repository.NewNotificationRepository(db)

	// Initialize services
	notificationService := service.NewNotificationService(notificationRepo)

	// Initialize Kafka consumer
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	consumer := kafka.NewConsumer(
		kafkaBrokers,
		getEnv("KAFKA_TOPIC_USER_EVENTS", "user-events"),
		getEnv("KAFKA_GROUP_ID", "notification-service"),
		notificationService,
	)
	defer consumer.Close()

	// Start Kafka consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		consumer.Start(ctx)
	}()

	// Initialize handlers
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Setup Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
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
	api := r.Group("/api/v1")
	{
		// Notification routes
		notifications := api.Group("/notifications")
		{
			notifications.POST("", notificationHandler.CreateNotification)
			notifications.GET("", notificationHandler.GetAllNotifications)
			notifications.GET("/:id", notificationHandler.GetNotificationByID)
			notifications.GET("/user/:user_id", notificationHandler.GetNotificationsByUserID)
		}

		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"service":   "notification-service",
				"timestamp": time.Now().Unix(),
			})
		})
	}

	// Swagger documentation
	docs.SwaggerInfo.Title = "Notification Service API"
	docs.SwaggerInfo.Description = "Notification management microservice with Kafka event processing"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = getEnv("HOST", "localhost:8082")
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := getEnv("PORT", "8082")
	log.Printf("Notification service starting on port %s", port)

	// Graceful shutdown
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down notification service...")

	// Cancel context to stop Kafka consumer
	cancel()

	// Give outstanding requests a deadline for completion
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Close connections gracefully
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	if err := consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	log.Println("Notification service stopped")
}

// initDatabase initializes the database connection
func initDatabase() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "notification_service"),
		getEnv("DB_SSLMODE", "disable"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connected successfully")
	return db, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}