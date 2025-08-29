package kafka

import (
	"time"

	"github.com/google/uuid"
)

// EventType defines the type of events that can be published
type EventType string

const (
	// User events
	EventUserCreated      EventType = "user.created"
	EventUserUpdated      EventType = "user.updated"
	EventUserStatusChanged EventType = "user.status_changed"
	EventUserDeleted      EventType = "user.deleted"
	
	// Notification events
	EventNotificationSent EventType = "notification.sent"
	EventNotificationFailed EventType = "notification.failed"
)

// BaseEvent represents the base structure for all events
type BaseEvent struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	BaseEvent
	Data struct {
		UserID    string `json:"user_id"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	} `json:"data"`
}

// UserUpdatedEvent represents a user update event
type UserUpdatedEvent struct {
	BaseEvent
	Data struct {
		UserID    string `json:"user_id"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		Changes   map[string]interface{} `json:"changes"`
	} `json:"data"`
}

// UserStatusChangedEvent represents a user status change event
type UserStatusChangedEvent struct {
	BaseEvent
	Data struct {
		UserID     string `json:"user_id"`
		Email      string `json:"email"`
		Username   string `json:"username"`
		OldStatus  string `json:"old_status"`
		NewStatus  string `json:"new_status"`
		ChangedBy  string `json:"changed_by"`
	} `json:"data"`
}

// UserDeletedEvent represents a user deletion event
type UserDeletedEvent struct {
	BaseEvent
	Data struct {
		UserID    string `json:"user_id"`
		Email     string `json:"email"`
		Username  string `json:"username"`
		DeletedBy string `json:"deleted_by"`
	} `json:"data"`
}

// NotificationEvent represents a notification event
type NotificationEvent struct {
	BaseEvent
	Data struct {
		UserID    string `json:"user_id"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Type      string `json:"type"` // email, sms, push
		Subject   string `json:"subject"`
		Message   string `json:"message"`
		Template  string `json:"template"`
		Variables map[string]interface{} `json:"variables"`
	} `json:"data"`
}

// Kafka topic constants
const (
	TopicUserEvents         = "user-events"
	TopicNotificationEvents = "notification-events"
	TopicUserNotifications  = "user-notifications"
)

// NewBaseEvent creates a new base event with default values
func NewBaseEvent(eventType EventType, source string) BaseEvent {
	return BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Source:    source,
	}
}