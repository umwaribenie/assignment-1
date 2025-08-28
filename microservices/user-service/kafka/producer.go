package kafka

import (
	"encoding/json"
	"log"
	"shared"
	"strings"
	"user-service/config"

	"github.com/IBM/sarama"
)

var Producer sarama.SyncProducer

// InitKafkaProducer initializes Kafka producer
func InitKafkaProducer() {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	brokers := strings.Split(config.AppConfig.KafkaBrokers, ",")
	
	var err error
	Producer, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatal("Failed to create Kafka producer:", err)
	}

	log.Println("Kafka producer initialized successfully")
}

// CloseKafkaProducer closes Kafka producer
func CloseKafkaProducer() {
	if Producer != nil {
		Producer.Close()
	}
}

// PublishEvent publishes an event to Kafka
func PublishEvent(topic string, event shared.KafkaEvent) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(eventBytes),
	}

	partition, offset, err := Producer.SendMessage(message)
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
		return err
	}

	log.Printf("Event published to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// PublishUserRegisteredEvent publishes user registered event
func PublishUserRegisteredEvent(user shared.User) error {
	event := shared.KafkaEvent{
		EventType: shared.UserRegistered,
		UserID:    user.ID,
		Data: shared.UserRegisteredEvent{
			User: user,
		},
		Timestamp: user.CreatedAt,
	}

	return PublishEvent("user-events", event)
}

// PublishPasswordResetEvent publishes password reset event
func PublishPasswordResetEvent(email, otp string) error {
	event := shared.KafkaEvent{
		EventType: shared.PasswordReset,
		Data: shared.PasswordResetEvent{
			Email: email,
			OTP:   otp,
		},
	}

	return PublishEvent("notification-events", event)
}