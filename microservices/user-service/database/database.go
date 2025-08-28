package database

import (
	"database/sql"
	"log"
	"user-service/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("postgres", config.AppConfig.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to User Service database successfully")
	createTables()
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

func createTables() {
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		client_id VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		first_name VARCHAR(255) NOT NULL,
		last_name VARCHAR(255) NOT NULL,
		national_id VARCHAR(50),
		passport_number VARCHAR(50),
		password VARCHAR(255) NOT NULL,
		phone VARCHAR(20),
		profile_picture TEXT,
		username VARCHAR(255) UNIQUE NOT NULL,
		role VARCHAR(50) DEFAULT 'user',
		status VARCHAR(50) DEFAULT 'active',
		subscription_status VARCHAR(50),
		institution_id VARCHAR(255),
		commission_percentage INTEGER,
		seler_type VARCHAR(50),
		specialization TEXT,
		notes TEXT,
		slug VARCHAR(255) UNIQUE NOT NULL,
		referral_code VARCHAR(50),
		has_active_subscription BOOLEAN DEFAULT FALSE,
		is_active BOOLEAN DEFAULT TRUE,
		otp_required BOOLEAN DEFAULT FALSE,
		created_by VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP
	);`

	createTokenBlacklistTable := `
	CREATE TABLE IF NOT EXISTS token_blacklist (
		id SERIAL PRIMARY KEY,
		token TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	createPasswordResetTable := `
	CREATE TABLE IF NOT EXISTS password_resets (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) NOT NULL,
		otp VARCHAR(10) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		used BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create indexes for better performance
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_users_slug ON users(slug);",
		"CREATE INDEX IF NOT EXISTS idx_users_client_id ON users(client_id);",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);",
		"CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);",
		"CREATE INDEX IF NOT EXISTS idx_token_blacklist_token ON token_blacklist(token);",
		"CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires_at ON token_blacklist(expires_at);",
		"CREATE INDEX IF NOT EXISTS idx_password_resets_email ON password_resets(email);",
		"CREATE INDEX IF NOT EXISTS idx_password_resets_otp ON password_resets(otp);",
		"CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at);",
	}

	tables := []string{
		createUsersTable,
		createTokenBlacklistTable,
		createPasswordResetTable,
	}

	for _, table := range tables {
		if _, err := DB.Exec(table); err != nil {
			log.Fatal("Failed to create table:", err)
		}
	}

	for _, index := range createIndexes {
		if _, err := DB.Exec(index); err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}

	log.Println("Database tables created successfully")
}