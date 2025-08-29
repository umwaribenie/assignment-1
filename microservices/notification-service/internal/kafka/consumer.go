package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"notification-service/internal/models"
	"notification-service/internal/service"

	"github.com/segmentio/kafka-go"
	sharedKafka "microservices/shared/kafka"
)

// Consumer represents a Kafka consumer
type Consumer struct {
	reader           *kafka.Reader
	notificationService service.NotificationService
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic, groupID string, notificationService service.NotificationService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader:           reader,
		notificationService: notificationService,
	}
}

// Start starts consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) {
	log.Printf("Starting Kafka consumer for topic: %s", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
			return
		default:
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				time.Sleep(time.Second)
				continue
			}

			log.Printf("Received message from topic %s: %s", m.Topic, string(m.Value))
			c.processMessage(ctx, m)
		}
	}
}

// processMessage processes a Kafka message
func (c *Consumer) processMessage(ctx context.Context, m kafka.Message) {
	var baseEvent sharedKafka.BaseEvent
	if err := json.Unmarshal(m.Value, &baseEvent); err != nil {
		log.Printf("Failed to unmarshal base event: %v", err)
		return
	}

	switch baseEvent.Type {
	case sharedKafka.EventUserCreated:
		c.handleUserCreated(ctx, m.Value)
	case sharedKafka.EventUserUpdated:
		c.handleUserUpdated(ctx, m.Value)
	case sharedKafka.EventUserStatusChanged:
		c.handleUserStatusChanged(ctx, m.Value)
	case sharedKafka.EventUserDeleted:
		c.handleUserDeleted(ctx, m.Value)
	default:
		log.Printf("Unknown event type: %s", baseEvent.Type)
	}
}

// handleUserCreated handles user created events
func (c *Consumer) handleUserCreated(ctx context.Context, message []byte) {
	var event sharedKafka.UserCreatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Failed to unmarshal user created event: %v", err)
		return
	}

	// Create welcome email notification
	req := models.CreateNotificationRequest{
		UserID:    event.Data.UserID,
		Type:      models.NotificationTypeEmail,
		Subject:   "Welcome to Our Platform!",
		Message:   fmt.Sprintf("Welcome %s %s! Thank you for joining our platform.", event.Data.FirstName, event.Data.LastName),
		Recipient: event.Data.Email,
		Template:  "welcome_email",
		Variables: map[string]interface{}{
			"firstName": event.Data.FirstName,
			"lastName":  event.Data.LastName,
			"email":     event.Data.Email,
			"username":  event.Data.Username,
		},
	}

	if _, err := c.notificationService.CreateNotification(ctx, req); err != nil {
		log.Printf("Failed to create welcome notification: %v", err)
	}

	// Create welcome SMS notification if phone is available
	if event.Data.Phone != "" {
		smsReq := models.CreateNotificationRequest{
			UserID:    event.Data.UserID,
			Type:      models.NotificationTypeSMS,
			Message:   fmt.Sprintf("Welcome %s! Your account has been created successfully.", event.Data.FirstName),
			Recipient: event.Data.Phone,
			Template:  "welcome_sms",
			Variables: map[string]interface{}{
				"firstName": event.Data.FirstName,
			},
		}

		if _, err := c.notificationService.CreateNotification(ctx, smsReq); err != nil {
			log.Printf("Failed to create welcome SMS notification: %v", err)
		}
	}

	log.Printf("Processed user created event for user: %s", event.Data.UserID)
}

// handleUserUpdated handles user updated events
func (c *Consumer) handleUserUpdated(ctx context.Context, message []byte) {
	var event sharedKafka.UserUpdatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Failed to unmarshal user updated event: %v", err)
		return
	}

	// Create profile update notification
	req := models.CreateNotificationRequest{
		UserID:    event.Data.UserID,
		Type:      models.NotificationTypeEmail,
		Subject:   "Profile Updated",
		Message:   fmt.Sprintf("Hi %s, your profile has been updated successfully.", event.Data.FirstName),
		Recipient: event.Data.Email,
		Template:  "profile_updated",
		Variables: map[string]interface{}{
			"firstName": event.Data.FirstName,
			"changes":   event.Data.Changes,
		},
	}

	if _, err := c.notificationService.CreateNotification(ctx, req); err != nil {
		log.Printf("Failed to create profile update notification: %v", err)
	}

	log.Printf("Processed user updated event for user: %s", event.Data.UserID)
}

// handleUserStatusChanged handles user status change events
func (c *Consumer) handleUserStatusChanged(ctx context.Context, message []byte) {
	var event sharedKafka.UserStatusChangedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Failed to unmarshal user status changed event: %v", err)
		return
	}

	// Create status change notification
	req := models.CreateNotificationRequest{
		UserID:    event.Data.UserID,
		Type:      models.NotificationTypeEmail,
		Subject:   "Account Status Changed",
		Message:   fmt.Sprintf("Hi %s, your account status has been changed from %s to %s.", event.Data.FirstName, event.Data.OldStatus, event.Data.NewStatus),
		Recipient: event.Data.Email,
		Template:  "status_changed",
		Variables: map[string]interface{}{
			"firstName":  event.Data.FirstName,
			"oldStatus":  event.Data.OldStatus,
			"newStatus":  event.Data.NewStatus,
			"changedBy":  event.Data.ChangedBy,
		},
	}

	if _, err := c.notificationService.CreateNotification(ctx, req); err != nil {
		log.Printf("Failed to create status change notification: %v", err)
	}

	log.Printf("Processed user status changed event for user: %s", event.Data.UserID)
}

// handleUserDeleted handles user deleted events
func (c *Consumer) handleUserDeleted(ctx context.Context, message []byte) {
	var event sharedKafka.UserDeletedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Failed to unmarshal user deleted event: %v", err)
		return
	}

	// Create account deletion notification
	req := models.CreateNotificationRequest{
		UserID:    event.Data.UserID,
		Type:      models.NotificationTypeEmail,
		Subject:   "Account Deleted",
		Message:   fmt.Sprintf("Hi %s, your account has been deleted successfully.", event.Data.Username),
		Recipient: event.Data.Email,
		Template:  "account_deleted",
		Variables: map[string]interface{}{
			"username":  event.Data.Username,
			"deletedBy": event.Data.DeletedBy,
		},
	}

	if _, err := c.notificationService.CreateNotification(ctx, req); err != nil {
		log.Printf("Failed to create account deletion notification: %v", err)
	}

	log.Printf("Processed user deleted event for user: %s", event.Data.UserID)
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}