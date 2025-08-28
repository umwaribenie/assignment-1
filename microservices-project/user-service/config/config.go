package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Application
	Port           string
	JWTSecret      string
	JWTExpiryHours int

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Kafka
	KafkaBrokers         []string
	KafkaGroupID         string
	KafkaNotificationTopic string

	// Services
	NotificationServiceURL string
}

var AppConfig *Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config := &Config{
		// Application
		Port:      getEnv("PORT", "8001"),
		JWTSecret: getEnv("JWT_SECRET", "default-secret-key"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "user"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "userdb"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// Kafka
		KafkaGroupID:         getEnv("KAFKA_GROUP_ID", "user-service"),
		KafkaNotificationTopic: getEnv("KAFKA_NOTIFICATION_TOPIC", "notifications"),

		// Services
		NotificationServiceURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8002"),
	}

	// Parse JWT expiry hours
	if hours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24")); err == nil {
		config.JWTExpiryHours = hours
	} else {
		config.JWTExpiryHours = 24
	}

	// Parse Redis DB
	if db, err := strconv.Atoi(getEnv("REDIS_DB", "0")); err == nil {
		config.RedisDB = db
	} else {
		config.RedisDB = 0
	}

	// Parse Kafka brokers
	brokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	config.KafkaBrokers = strings.Split(brokers, ",")

	AppConfig = config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetDatabaseURL() string {
	return "host=" + AppConfig.DBHost +
		" port=" + AppConfig.DBPort +
		" user=" + AppConfig.DBUser +
		" password=" + AppConfig.DBPassword +
		" dbname=" + AppConfig.DBName +
		" sslmode=" + AppConfig.DBSSLMode
}