package kafka

import (
	"context"
	"encoding/json"
	"log"
	"notification-service/config"
	"notification-service/services"
	"shared"
	"strings"

	"github.com/IBM/sarama"
)

type Consumer struct {
	emailService *services.EmailService
	smsService   *services.SMSService
}

func NewConsumer() *Consumer {
	return &Consumer{
		emailService: services.NewEmailService(),
		smsService:   services.NewSMSService(),
	}
}

// Setup is called when a new consumer session is started
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called when a consumer session is ended
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from Kafka
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Process the message
			if err := c.processMessage(message); err != nil {
				log.Printf("Failed to process message: %v", err)
			} else {
				// Mark message as processed
				session.MarkMessage(message, "")
			}

		case <-session.Context().Done():
			return nil
		}
	}
}

// processMessage processes individual Kafka messages
func (c *Consumer) processMessage(message *sarama.ConsumerMessage) error {
	log.Printf("Received message from topic %s: %s", message.Topic, string(message.Value))

	var event shared.KafkaEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return err
	}

	switch event.EventType {
	case shared.UserRegistered:
		return c.handleUserRegistered(event)
	case shared.PasswordReset:
		return c.handlePasswordReset(event)
	case shared.OTPRequested:
		return c.handleOTPRequested(event)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
		return nil
	}
}

// handleUserRegistered handles user registration events
func (c *Consumer) handleUserRegistered(event shared.KafkaEvent) error {
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		return nil
	}

	user, ok := data["user"].(map[string]interface{})
	if !ok {
		return nil
	}

	email, _ := user["email"].(string)
	firstName, _ := user["first_name"].(string)

	if email == "" {
		return nil
	}

	// Send welcome email
	return c.emailService.SendWelcomeEmail(email, firstName)
}

// handlePasswordReset handles password reset events
func (c *Consumer) handlePasswordReset(event shared.KafkaEvent) error {
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		return nil
	}

	email, _ := data["email"].(string)
	otp, _ := data["otp"].(string)

	if email == "" || otp == "" {
		return nil
	}

	// Send password reset email
	return c.emailService.SendPasswordResetEmail(email, otp)
}

// handleOTPRequested handles OTP request events
func (c *Consumer) handleOTPRequested(event shared.KafkaEvent) error {
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		return nil
	}

	email, _ := data["email"].(string)
	phone, _ := data["phone"].(string)
	otp, _ := data["otp"].(string)
	notificationType, _ := data["type"].(string)

	if otp == "" {
		return nil
	}

	switch notificationType {
	case "email":
		if email != "" {
			return c.emailService.SendOTPEmail(email, otp)
		}
	case "sms":
		if phone != "" {
			return c.smsService.SendOTPSMS(phone, otp)
		}
	}

	return nil
}

// StartConsumer starts the Kafka consumer
func StartConsumer(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer := NewConsumer()
	brokers := strings.Split(config.AppConfig.KafkaBrokers, ",")

	consumerGroup, err := sarama.NewConsumerGroup(brokers, "notification-service", config)
	if err != nil {
		return err
	}

	topics := []string{"user-events", "notification-events"}

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, topics, consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	log.Println("Kafka consumer started successfully")
	return nil
}