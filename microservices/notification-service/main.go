package main

import (
	"context"
	"log"
	"notification-service/cache"
	"notification-service/config"
	"notification-service/kafka"
	"notification-service/routes"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize Redis cache
	cache.InitRedis()
	defer cache.CloseRedis()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Start Kafka consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := kafka.StartConsumer(ctx); err != nil {
			log.Printf("Failed to start Kafka consumer: %v", err)
		}
	}()

	// Setup HTTP routes
	r := routes.SetupRoutes()
	
	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Notification Service starting on port %s", config.AppConfig.Port)
		if err := r.Run(":" + config.AppConfig.Port); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down notification service...")
	cancel() // Cancel context to stop all goroutines
	wg.Wait() // Wait for all goroutines to finish
	log.Println("Notification service stopped")
}