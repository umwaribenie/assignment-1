package kafka

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event
type EventType string

const (
	// User Events
	EventUserCreated     EventType = "user.created"
	EventUserUpdated     EventType = "user.updated"
	EventUserDeleted     EventType = "user.deleted"
	EventUserStatusChanged EventType = "user.status_changed"
	
	// Notification Events
	EventNotificationSent EventType = "notification.sent"
	EventNotificationFailed EventType = "notification.failed"
)

// BaseEvent represents the base structure for all events
type BaseEvent struct {
	ID        uuid.UUID `json:"id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // Service that produced the event
	Version   string    `json:"version"`
}

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	BaseEvent
	Data UserCreatedData `json:"data"`
}

type UserCreatedData struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
}

// UserUpdatedEvent represents a user update event
type UserUpdatedEvent struct {
	BaseEvent
	Data UserUpdatedData `json:"data"`
}

type UserUpdatedData struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
}

// UserStatusChangedEvent represents a user status change event
type UserStatusChangedEvent struct {
	BaseEvent
	Data UserStatusChangedData `json:"data"`
}

type UserStatusChangedData struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	OldStatus  string    `json:"old_status"`
	NewStatus  string    `json:"new_status"`
	ChangedBy  string    `json:"changed_by"`
	Reason     string    `json:"reason,omitempty"`
}

// NotificationEvent represents a notification event
type NotificationEvent struct {
	BaseEvent
	Data NotificationData `json:"data"`
}

type NotificationData struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Type      string    `json:"type"` // email, sms, push
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Template  string    `json:"template,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Kafka Topics
const (
	TopicUserEvents       = "user-events"
	TopicNotificationEvents = "notification-events"
	TopicUserNotifications = "user-notifications"
)