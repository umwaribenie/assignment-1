package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database Configuration
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	DBMaxOpenConns      int
	DBMaxIdleConns      int
	DBConnMaxLifetime   int
	DBConnMaxIdleTime   int

	// JWT Configuration
	JWTSecret      string
	JWTExpiryHours int

	// Server Configuration
	Port string
	Host string

	// File Upload Configuration
	UploadPath     string
	MaxFileSize    int64

	// Environment
	Environment string

	// Email Configuration
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		// Database Configuration
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnv("DB_PORT", "5432"),
		DBUser:              getEnv("DB_USER", "postgres"),
		DBPassword:          getEnv("DB_PASSWORD", ""),
		DBName:              getEnv("DB_NAME", "generalusermanagement"),
		DBSSLMode:           getEnv("DB_SSL_MODE", "disable"),
		DBMaxOpenConns:      getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:      getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxLifetime:   getEnvAsInt("DB_CONN_MAX_LIFETIME", 5),
		DBConnMaxIdleTime:   getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 5),

		// JWT Configuration
		JWTSecret:      getEnv("JWT_SECRET", "your_super_secret_jwt_key_here_make_it_long_and_random"),
		JWTExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),

		// Server Configuration
		Port: getEnv("SERVER_PORT", "8082"),
		Host: getEnv("SERVER_HOST", "localhost"),

		// File Upload Configuration
		UploadPath:     getEnv("UPLOAD_PATH", "./uploads"),
		MaxFileSize:    getEnvAsInt64("MAX_FILE_SIZE", 5242880), // 5MB

		// Environment
		Environment: getEnv("ENV", "development"),

		// Email Configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
	}

	log.Println("Configuration loaded successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}