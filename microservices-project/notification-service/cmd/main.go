package main

import (
	"context"
	"log"
	"net/http"
	"notification-service/config"
	"notification-service/internal/kafka"
	"notification-service/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize services
	emailService := service.NewEmailService(config.AppConfig)
	smsService := service.NewSMSService(config.AppConfig)

	// Initialize Kafka consumer
	consumer, err := kafka.NewNotificationConsumer(
		config.AppConfig.KafkaBrokers,
		config.AppConfig.KafkaGroupID,
		[]string{config.AppConfig.KafkaNotificationTopic},
		emailService,
		smsService,
	)
	if err != nil {
		log.Fatal("Failed to create Kafka consumer:", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Kafka consumer in a goroutine
	go func() {
		log.Println("Starting Kafka consumer...")
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}()

	// Setup HTTP server for health checks
	router := gin.Default()
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "notification-service",
			"version": "1.0.0",
		})
	})

	// Metrics endpoint (you can add prometheus metrics here)
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"emails_sent": 0, // TODO: Implement actual metrics
			"sms_sent": 0,
		})
	})

	// Create server
	srv := &http.Server{
		Addr:    ":" + config.AppConfig.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Notification Service starting on port %s", config.AppConfig.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down notification service...")

	// Cancel context to stop Kafka consumer
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Stop Kafka consumer
	if err := consumer.Stop(); err != nil {
		log.Printf("Error stopping Kafka consumer: %v", err)
	}

	log.Println("Notification service exited")
}