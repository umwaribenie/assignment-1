package main

import (
	"log"
	"user-service/cache"
	"user-service/config"
	"user-service/database"
	"user-service/kafka"
	"user-service/routes"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize database
	database.InitDB()
	defer database.CloseDB()

	// Initialize Redis cache
	cache.InitRedis()
	defer cache.CloseRedis()

	// Initialize Kafka producer
	kafka.InitKafkaProducer()
	defer kafka.CloseKafkaProducer()

	// Setup routes
	r := routes.SetupRoutes()
	
	log.Printf("User Service starting on port %s", config.AppConfig.Port)
	log.Fatal(r.Run(":" + config.AppConfig.Port))
}