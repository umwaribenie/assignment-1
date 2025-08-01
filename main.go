package main

import (
	"generalusermanagement/config"
	"generalusermanagement/database"
	"generalusermanagement/routes"
	"log"
	"os"
)

// @title User Management API
// @version 1.0
// @description A comprehensive user management system with authentication and authorization
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8082
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	config.LoadConfig()

	// Check if we should skip database for demo
	if os.Getenv("SKIP_DB") != "true" {
		// Initialize database
		database.InitDB()
		defer database.CloseDB()
	} else {
		log.Println("⚠️ Running in DEMO mode - Database connections skipped")
	}

	// Setup routes
	r := routes.SetupRoutes()

	// Start server
	log.Printf("🚀 Server starting on %s:%s", config.AppConfig.Host, config.AppConfig.Port)
	if err := r.Run(":" + config.AppConfig.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}