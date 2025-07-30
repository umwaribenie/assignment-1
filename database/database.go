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
	migrateExistingTables()
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
		subscription_status VARCHAR(50),
		institution_id VARCHAR(255),
		commission_percentage INTEGER CHECK (commission_percentage >= 1 AND commission_percentage <= 100),
		seler_type VARCHAR(50),
		specialization VARCHAR(255),
		notes TEXT,
		slug VARCHAR(255) UNIQUE,
		referral_code VARCHAR(255),
		has_active_subscription BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT TRUE,
		otp_required BOOLEAN DEFAULT FALSE,
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

	createLoginOTPTable := `
	CREATE TABLE IF NOT EXISTS login_otps (
		id SERIAL PRIMARY KEY,
		user_id UUID NOT NULL,
		otp VARCHAR(10) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		used BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`

	// Create indexes for better performance
	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_users_client_id ON users(client_id);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_slug ON users(slug);
	CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
	CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
	CREATE INDEX IF NOT EXISTS idx_users_subscription_status ON users(subscription_status);
	CREATE INDEX IF NOT EXISTS idx_users_institution_id ON users(institution_id);
	CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
	CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
	CREATE INDEX IF NOT EXISTS idx_login_otps_user_id ON login_otps(user_id);
	CREATE INDEX IF NOT EXISTS idx_login_otps_expires_at ON login_otps(expires_at);
	`

	tables := []string{
		createUsersTable,
		createTokenBlacklistTable,
		createPasswordResetTable,
		createLoginOTPTable,
		createIndexes,
	}

	for _, table := range tables {
		_, err := DB.Exec(table)
		if err != nil {
			log.Printf("Error creating table/index: %v", err)
			// Don't fail completely, some errors might be expected (like index already exists)
		}
	}

	log.Println("All tables created successfully")
}

// migrateExistingTables adds new columns to existing tables if they don't exist
func migrateExistingTables() {
	migrations := []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS subscription_status VARCHAR(50)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS institution_id VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS commission_percentage INTEGER CHECK (commission_percentage >= 1 AND commission_percentage <= 100)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS seler_type VARCHAR(50)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS specialization VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS notes TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS slug VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS referral_code VARCHAR(255)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS has_active_subscription BOOLEAN DEFAULT FALSE",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS otp_required BOOLEAN DEFAULT FALSE",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS created_by VARCHAR(255)",
	}

	for _, migration := range migrations {
		_, err := DB.Exec(migration)
		if err != nil {
			log.Printf("Migration error (may be expected): %v", err)
		}
	}

	// Update existing users to have slugs if they don't have them
	updateSlugsQuery := `
		UPDATE users 
		SET slug = LOWER(CONCAT(username, '-', SUBSTRING(id::text, 1, 8)))
		WHERE slug IS NULL OR slug = ''
	`
	_, err := DB.Exec(updateSlugsQuery)
	if err != nil {
		log.Printf("Error updating slugs: %v", err)
	}

	// Add unique constraint to slug if it doesn't exist
	_, err = DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS users_slug_unique ON users(slug)")
	if err != nil {
		log.Printf("Error creating unique index on slug: %v", err)
	}

	log.Println("Database migrations completed")
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
