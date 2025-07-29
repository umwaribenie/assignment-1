package main

import (
	"fmt"
	"log"

	"generalusermanagement/config"
	"generalusermanagement/database"
	"generalusermanagement/routes"

	_ "generalusermanagement/docs"
)

// @title User mngt
// @version 1.2
// @description General user management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url 
// @contact.email umwaribenie5@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 
// @BasePath /api/v1/users

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