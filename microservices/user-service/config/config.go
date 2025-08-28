package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	JWTExpiryHours int
	RedisURL       string
	RedisPassword  string
	RedisDB        int
	KafkaBrokers   string
}

var AppConfig Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = Config{
		Port:           getEnv("USER_SERVICE_PORT", "8081"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:password@localhost/userdb?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		RedisURL:       getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvAsInt("REDIS_DB", 0),
		KafkaBrokers:   getEnv("KAFKA_BROKERS", "localhost:9092"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}