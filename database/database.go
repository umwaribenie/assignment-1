package database

import (
	"database/sql"
	"fmt"
	"log"

	"generalusermanagement/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	cfg := config.AppConfig
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	log.Println("Successfully connected to database")
	createTables()
}

func createTables() {
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		client_id VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		first_name VARCHAR(255) NOT NULL,
		last_name VARCHAR(255) NOT NULL,
		national_id VARCHAR(255),
		passport_number VARCHAR(255),
		password VARCHAR(255) NOT NULL,
		phone VARCHAR(255) NOT NULL,
		profile_picture VARCHAR(500),
		username VARCHAR(255) UNIQUE NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP
	);
	`

	createTokenBlacklistTable := `
	CREATE TABLE IF NOT EXISTS token_blacklist (
		id SERIAL PRIMARY KEY,
		token VARCHAR(1000) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	createPasswordResetTable := `
	CREATE TABLE IF NOT EXISTS password_resets (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) NOT NULL,
		otp VARCHAR(10) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		used BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	tables := []string{createUsersTable, createTokenBlacklistTable, createPasswordResetTable}

	for _, table := range tables {
		_, err := DB.Exec(table)
		if err != nil {
			log.Fatal("Error creating table:", err)
		}
	}

	log.Println("All tables created successfully")
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}