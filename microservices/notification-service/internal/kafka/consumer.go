package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"notification-service/internal/models"
	"notification-service/internal/service"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// Consumer represents a Kafka consumer
type Consumer struct {
	reader *kafka.Reader
	notificationService *service.NotificationService
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic, groupID string, notificationService *service.NotificationService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader: reader,
		notificationService: notificationService,
	}
}

// Start starts consuming messages
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
				continue
			}

			log.Printf("Received message from topic %s: %s", m.Topic, string(m.Value))

			// Process the message
			if err := c.processMessage(ctx, m); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}
}

// processMessage processes a Kafka message
func (c *Consumer) processMessage(ctx context.Context, m kafka.Message) error {
	// Parse the event
	var event map[string]interface{}
	if err := json.Unmarshal(m.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	eventType, ok := event["type"].(string)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	switch eventType {
	case "user.created":
		return c.handleUserCreated(ctx, event)
	case "user.updated":
		return c.handleUserUpdated(ctx, event)
	case "user.status_changed":
		return c.handleUserStatusChanged(ctx, event)
	default:
		log.Printf("Unknown event type: %s", eventType)
		return nil
	}
}

// handleUserCreated handles user created events
func (c *Consumer) handleUserCreated(ctx context.Context, event map[string]interface{}) error {
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event data")
	}

	userIDStr, ok := data["user_id"].(string)
	if !ok {
		return fmt.Errorf("invalid user_id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user_id format: %w", err)
	}

	email, ok := data["email"].(string)
	if !ok {
		return fmt.Errorf("invalid email")
	}

	firstName, ok := data["first_name"].(string)
	if !ok {
		return fmt.Errorf("invalid first_name")
	}

	lastName, ok := data["last_name"].(string)
	if !ok {
		return fmt.Errorf("invalid last_name")
	}

	// Create welcome notification
	req := models.CreateNotificationRequest{
		UserID:    userID,
		Type:      models.NotificationTypeEmail,
		Subject:   "Welcome to Our Platform!",
		Message:   fmt.Sprintf("Hello %s %s! Welcome to our platform. We're excited to have you on board.", firstName, lastName),
		MaxRetries: 3,
	}

	if err := c.notificationService.CreateNotification(ctx, req); err != nil {
		return fmt.Errorf("failed to create welcome notification: %w", err)
	}

	log.Printf("Created welcome notification for user: %s", userID)
	return nil
}

// handleUserUpdated handles user updated events
func (c *Consumer) handleUserUpdated(ctx context.Context, event map[string]interface{}) error {
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event data")
	}

	userIDStr, ok := data["user_id"].(string)
	if !ok {
		return fmt.Errorf("invalid user_id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user_id format: %w", err)
	}

	email, ok := data["email"].(string)
	if !ok {
		return fmt.Errorf("invalid email")
	}

	// Create profile update notification
	req := models.CreateNotificationRequest{
		UserID:    userID,
		Type:      models.NotificationTypeEmail,
		Subject:   "Profile Updated",
		Message:   "Your profile has been successfully updated. If you didn't make this change, please contact support immediately.",
		MaxRetries: 3,
	}

	if err := c.notificationService.CreateNotification(ctx, req); err != nil {
		return fmt.Errorf("failed to create profile update notification: %w", err)
	}

	log.Printf("Created profile update notification for user: %s", userID)
	return nil
}

// handleUserStatusChanged handles user status changed events
func (c *Consumer) handleUserStatusChanged(ctx context.Context, event map[string]interface{}) error {
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event data")
	}

	userIDStr, ok := data["user_id"].(string)
	if !ok {
		return fmt.Errorf("invalid user_id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user_id format: %w", err)
	}

	email, ok := data["email"].(string)
	if !ok {
		return fmt.Errorf("invalid email")
	}

	oldStatus, ok := data["old_status"].(string)
	if !ok {
		return fmt.Errorf("invalid old_status")
	}

	newStatus, ok := data["new_status"].(string)
	if !ok {
		return fmt.Errorf("invalid new_status")
	}

	// Create status change notification
	req := models.CreateNotificationRequest{
		UserID:    userID,
		Type:      models.NotificationTypeEmail,
		Subject:   "Account Status Changed",
		Message:   fmt.Sprintf("Your account status has been changed from %s to %s. If you have any questions, please contact support.", oldStatus, newStatus),
		MaxRetries: 3,
	}

	if err := c.notificationService.CreateNotification(ctx, req); err != nil {
		return fmt.Errorf("failed to create status change notification: %w", err)
	}

	log.Printf("Created status change notification for user: %s", userID)
	return nil
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}