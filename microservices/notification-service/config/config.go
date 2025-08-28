package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	RedisURL     string
	RedisPassword string
	RedisDB      int
	KafkaBrokers string
	
	// Email configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	
	// SMS configuration (example with Twilio)
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string
}

var AppConfig Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = Config{
		Port:         getEnv("NOTIFICATION_SERVICE_PORT", "8082"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:      getEnvAsInt("REDIS_DB", 1), // Different DB from user service
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		
		// Email configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", "your-email@gmail.com"),
		SMTPPassword: getEnv("SMTP_PASSWORD", "your-app-password"),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@yourdomain.com"),
		FromName:     getEnv("FROM_NAME", "Your App"),
		
		// SMS configuration
		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
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