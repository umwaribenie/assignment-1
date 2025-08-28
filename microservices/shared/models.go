package shared

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleUser       UserRole = "user"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "super_admin"
	RoleTrainer    UserRole = "trainer"
	RoleInstructor UserRole = "instructor"
	RoleFrontdesk  UserRole = "frontdesk"
	RoleFinance    UserRole = "finance"
	RoleSeler      UserRole = "seler"
	RoleMember     UserRole = "member"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"
)

// User represents a user in the system (shared between services)
type User struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	ClientID   string     `json:"client_id" db:"client_id"`
	Email      string     `json:"email" db:"email"`
	FirstName  string     `json:"first_name" db:"first_name"`
	LastName   string     `json:"last_name" db:"last_name"`
	Phone      string     `json:"phone" db:"phone"`
	Username   string     `json:"username" db:"username"`
	Role       UserRole   `json:"role" db:"role"`
	Status     UserStatus `json:"status" db:"status"`
	Slug       string     `json:"slug" db:"slug"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// Kafka Event Types
type EventType string

const (
	UserRegistered    EventType = "user.registered"
	UserUpdated       EventType = "user.updated"
	PasswordReset     EventType = "password.reset.requested"
	OTPRequested      EventType = "otp.requested"
)

// KafkaEvent represents a Kafka event
type KafkaEvent struct {
	EventType EventType   `json:"event_type"`
	UserID    uuid.UUID   `json:"user_id"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Notification Events
type UserRegisteredEvent struct {
	User User `json:"user"`
}

type PasswordResetEvent struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type OTPRequestEvent struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
	Type  string `json:"type"` // "email" or "sms"
}