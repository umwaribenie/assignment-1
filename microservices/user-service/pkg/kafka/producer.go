package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	sharedKafka "microservices/shared/kafka"
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
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("%T", event)),
		Value: eventJSON,
	})

	if err != nil {
		return fmt.Errorf("failed to publish event to topic %s: %w", topic, err)
	}

	log.Printf("Published event to topic %s: %s", topic, string(eventJSON))
	return nil
}

// PublishUserCreated publishes a user created event
func (p *Producer) PublishUserCreated(ctx context.Context, userID, email, username, firstName, lastName, phone string) error {
	event := sharedKafka.UserCreatedEvent{
		BaseEvent: sharedKafka.NewBaseEvent(sharedKafka.EventUserCreated, "user-service"),
	}
	event.Data.UserID = userID
	event.Data.Email = email
	event.Data.Username = username
	event.Data.FirstName = firstName
	event.Data.LastName = lastName
	event.Data.Phone = phone

	return p.PublishEvent(ctx, sharedKafka.TopicUserEvents, event)
}

// PublishUserUpdated publishes a user updated event
func (p *Producer) PublishUserUpdated(ctx context.Context, userID, email, username, firstName, lastName, phone string, changes map[string]interface{}) error {
	event := sharedKafka.UserUpdatedEvent{
		BaseEvent: sharedKafka.NewBaseEvent(sharedKafka.EventUserUpdated, "user-service"),
	}
	event.Data.UserID = userID
	event.Data.Email = email
	event.Data.Username = username
	event.Data.FirstName = firstName
	event.Data.LastName = lastName
	event.Data.Phone = phone
	event.Data.Changes = changes

	return p.PublishEvent(ctx, sharedKafka.TopicUserEvents, event)
}

// PublishUserStatusChanged publishes a user status changed event
func (p *Producer) PublishUserStatusChanged(ctx context.Context, userID, email, username, oldStatus, newStatus, changedBy string) error {
	event := sharedKafka.UserStatusChangedEvent{
		BaseEvent: sharedKafka.NewBaseEvent(sharedKafka.EventUserStatusChanged, "user-service"),
	}
	event.Data.UserID = userID
	event.Data.Email = email
	event.Data.Username = username
	event.Data.OldStatus = oldStatus
	event.Data.NewStatus = newStatus
	event.Data.ChangedBy = changedBy

	return p.PublishEvent(ctx, sharedKafka.TopicUserEvents, event)
}

// PublishUserDeleted publishes a user deleted event
func (p *Producer) PublishUserDeleted(ctx context.Context, userID, email, username, deletedBy string) error {
	event := sharedKafka.UserDeletedEvent{
		BaseEvent: sharedKafka.NewBaseEvent(sharedKafka.EventUserDeleted, "user-service"),
	}
	event.Data.UserID = userID
	event.Data.Email = email
	event.Data.Username = username
	event.Data.DeletedBy = deletedBy

	return p.PublishEvent(ctx, sharedKafka.TopicUserEvents, event)
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	return p.writer.Close()
}