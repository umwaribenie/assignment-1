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
	Port string

	// Kafka
	KafkaBrokers           []string
	KafkaGroupID           string
	KafkaNotificationTopic string

	// Email (SMTP)
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// SMS (Twilio)
	SMSEnabled       bool
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	// Templates
	TemplateDir string
}

var AppConfig *Config

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config := &Config{
		// Application
		Port: getEnv("PORT", "8002"),

		// Kafka
		KafkaGroupID:           getEnv("KAFKA_GROUP_ID", "notification-service"),
		KafkaNotificationTopic: getEnv("KAFKA_NOTIFICATION_TOPIC", "notifications"),

		// Email
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@example.com"),

		// SMS
		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),

		// Templates
		TemplateDir: getEnv("TEMPLATE_DIR", "./templates"),
	}

	// Parse SMTP port
	if port, err := strconv.Atoi(getEnv("SMTP_PORT", "587")); err == nil {
		config.SMTPPort = port
	} else {
		config.SMTPPort = 587
	}

	// Parse SMS enabled
	config.SMSEnabled = getEnv("SMS_ENABLED", "false") == "true"

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