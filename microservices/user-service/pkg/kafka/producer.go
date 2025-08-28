package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// Producer represents a Kafka producer
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
	}
}

// PublishEvent publishes an event to a Kafka topic
func (p *Producer) PublishEvent(ctx context.Context, topic string, event interface{}) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Topic: topic,
		Key:   []byte(uuid.New().String()),
		Value: eventBytes,
		Time:  time.Now(),
	}

	err = p.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	log.Printf("Event published to topic %s: %+v", topic, event)
	return nil
}

// PublishUserCreated publishes a user created event
func (p *Producer) PublishUserCreated(ctx context.Context, userID uuid.UUID, email, firstName, lastName, phone, role string) error {
	event := UserCreatedEvent{
		BaseEvent: BaseEvent{
			ID:        uuid.New(),
			Type:      EventUserCreated,
			Timestamp: time.Now(),
			Source:    "user-service",
			Version:   "1.0",
		},
		Data: UserCreatedData{
			UserID:    userID,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			Phone:     phone,
			Role:      role,
		},
	}

	return p.PublishEvent(ctx, TopicUserEvents, event)
}

// PublishUserUpdated publishes a user updated event
func (p *Producer) PublishUserUpdated(ctx context.Context, userID uuid.UUID, email, firstName, lastName, phone, role, status string) error {
	event := UserUpdatedEvent{
		BaseEvent: BaseEvent{
			ID:        uuid.New(),
			Type:      EventUserUpdated,
			Timestamp: time.Now(),
			Source:    "user-service",
			Version:   "1.0",
		},
		Data: UserUpdatedData{
			UserID:    userID,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			Phone:     phone,
			Role:      role,
			Status:    status,
		},
	}

	return p.PublishEvent(ctx, TopicUserEvents, event)
}

// PublishUserStatusChanged publishes a user status changed event
func (p *Producer) PublishUserStatusChanged(ctx context.Context, userID uuid.UUID, email, oldStatus, newStatus, changedBy, reason string) error {
	event := UserStatusChangedEvent{
		BaseEvent: BaseEvent{
			ID:        uuid.New(),
			Type:      EventUserStatusChanged,
			Timestamp: time.Now(),
			Source:    "user-service",
			Version:   "1.0",
		},
		Data: UserStatusChangedData{
			UserID:     userID,
			Email:      email,
			OldStatus:  oldStatus,
			NewStatus:  newStatus,
			ChangedBy:  changedBy,
			Reason:     reason,
		},
	}

	return p.PublishEvent(ctx, TopicUserEvents, event)
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	return p.writer.Close()
}