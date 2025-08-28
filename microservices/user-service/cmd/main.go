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

	"user-service/internal/handlers"
	"user-service/internal/repository"
	"user-service/internal/service"
	"user-service/pkg/kafka"
	"user-service/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
)

// @title User Service API
// @version 1.0
// @description User management microservice with Redis caching and Kafka events
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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

	// Initialize Redis cache
	cache := utils.NewCacheClient(
		getEnv("REDIS_ADDR", "localhost:6379"),
		getEnv("REDIS_PASSWORD", ""),
		0, // default DB
	)
	defer cache.Close()

	// Initialize Kafka producer
	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	producer := kafka.NewProducer(kafkaBrokers)
	defer producer.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, cache)

	// Initialize services
	userService := service.NewUserService(userRepo, producer)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)

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
		// User routes
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.GetAllUsers)
			users.GET("/email", userHandler.GetUserByEmail)
			users.GET("/:id", userHandler.GetUserByID)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"service":   "user-service",
				"timestamp": time.Now().Unix(),
			})
		})
	}

	// Swagger documentation
	docs.SwaggerInfo.Title = "User Service API"
	docs.SwaggerInfo.Description = "User management microservice with Redis caching and Kafka events"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = getEnv("HOST", "localhost:8081")
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := getEnv("PORT", "8081")
	log.Printf("User service starting on port %s", port)

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

	log.Println("Shutting down user service...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close connections gracefully
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	if err := cache.Close(); err != nil {
		log.Printf("Error closing cache: %v", err)
	}

	if err := producer.Close(); err != nil {
		log.Printf("Error closing Kafka producer: %v", err)
	}

	log.Println("User service stopped")
}

// initDatabase initializes the database connection
func initDatabase() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "user_service"),
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