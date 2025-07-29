package main

import (
	"fmt"
	"log"

	"generalusermanagement/config"
	"generalusermanagement/database"
	"generalusermanagement/routes"

	_ "generalusermanagement/docs"
)

// @title General User Management API
// @version 1.0
// @description A comprehensive user management system with authentication, authorization, and file upload capabilities.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8082
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize database
	database.InitDB()
	defer database.CloseDB()

	// Setup routes
	r := routes.SetupRoutes()

	// Start server
	serverAddr := fmt.Sprintf("%s:%s", config.AppConfig.ServerHost, config.AppConfig.ServerPort)
	log.Printf("Server starting on %s", serverAddr)
	log.Printf("Swagger documentation available at http://%s/swagger/index.html", serverAddr)

	if err := r.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}