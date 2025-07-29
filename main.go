package main

import (
	"log"
	"generalusermanagement/config"
	"generalusermanagement/database"
	"generalusermanagement/routes"
	_ "generalusermanagement/docs" // This line is important for go-swagger
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
	config.LoadConfig()
	database.InitDB()
	defer database.CloseDB()

	r := routes.SetupRoutes()
	
	log.Printf("Server starting on port %s", config.AppConfig.Port)
	log.Fatal(r.Run(":" + config.AppConfig.Port))
}