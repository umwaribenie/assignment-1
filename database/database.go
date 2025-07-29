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
	// Create clients table
	createClientsTable := `
	CREATE TABLE IF NOT EXISTS clients (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(255) UNIQUE NOT NULL,
		logo VARCHAR(500),
		bg_image VARCHAR(500),
		address TEXT,
		description TEXT,
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Create institutions table
	createInstitutionsTable := `
	CREATE TABLE IF NOT EXISTS institutions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255),
		phone_number VARCHAR(255),
		address TEXT,
		institution_no INTEGER,
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Updated users table with new fields
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		client_id VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		first_name VARCHAR(255) NOT NULL,
		last_name VARCHAR(255) NOT NULL,
		national_id VARCHAR(255),
		passport_number VARCHAR(255),
		password VARCHAR(255) NOT NULL,
		phone VARCHAR(255),
		profile_picture VARCHAR(500),
		username VARCHAR(255) UNIQUE NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		slug VARCHAR(255) UNIQUE NOT NULL,
		notes TEXT,
		institution_id VARCHAR(255),
		subscription_status VARCHAR(50),
		has_active_subscription BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT TRUE,
		otp_required BOOLEAN DEFAULT FALSE,
		referral_code VARCHAR(255),
		created_by VARCHAR(255),
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

	// New login OTP table for OTP-based authentication
	createLoginOTPTable := `
	CREATE TABLE IF NOT EXISTS login_otps (
		id SERIAL PRIMARY KEY,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		otp VARCHAR(10) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		used BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	tables := []string{
		createClientsTable,
		createInstitutionsTable,
		createUsersTable,
		createTokenBlacklistTable,
		createPasswordResetTable,
		createLoginOTPTable,
	}

	for _, table := range tables {
		_, err := DB.Exec(table)
		if err != nil {
			log.Fatal("Error creating table:", err)
		}
	}

	// Create indexes for better performance
	createIndexes()

	log.Println("All tables created successfully")
}

func createIndexes() {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_client_id ON users(client_id);",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_users_slug ON users(slug);",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);",
		"CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);",
		"CREATE INDEX IF NOT EXISTS idx_users_subscription_status ON users(subscription_status);",
		"CREATE INDEX IF NOT EXISTS idx_users_institution_id ON users(institution_id);",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_institutions_client_id ON institutions(client_id);",
		"CREATE INDEX IF NOT EXISTS idx_institutions_slug ON institutions(slug);",
		"CREATE INDEX IF NOT EXISTS idx_clients_slug ON clients(slug);",
		"CREATE INDEX IF NOT EXISTS idx_login_otps_user_id ON login_otps(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_login_otps_expires_at ON login_otps(expires_at);",
		"CREATE INDEX IF NOT EXISTS idx_password_resets_email ON password_resets(email);",
		"CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at);",
	}

	for _, index := range indexes {
		_, err := DB.Exec(index)
		if err != nil {
			log.Printf("Warning: Error creating index: %v", err)
		}
	}

	log.Println("Database indexes created successfully")
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}