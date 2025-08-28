package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypePush  NotificationType = "push"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusDelivered NotificationStatus = "delivered"
)

// Notification represents a notification in the system
type Notification struct {
	ID          uuid.UUID           `json:"id" db:"id"`
	UserID      uuid.UUID           `json:"user_id" db:"user_id"`
	Type        NotificationType    `json:"type" db:"type"`
	Subject     string              `json:"subject" db:"subject"`
	Message     string              `json:"message" db:"message"`
	Template    *string             `json:"template,omitempty" db:"template"`
	Variables   map[string]string   `json:"variables,omitempty" db:"variables"`
	Status      NotificationStatus  `json:"status" db:"status"`
	RetryCount  int                 `json:"retry_count" db:"retry_count"`
	MaxRetries  int                 `json:"max_retries" db:"max_retries"`
	SentAt      *time.Time          `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt *time.Time          `json:"delivered_at,omitempty" db:"delivered_at"`
	Error       *string             `json:"error,omitempty" db:"error"`
	CreatedAt   time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at" db:"updated_at"`
}

// CreateNotificationRequest represents the request to create a notification
type CreateNotificationRequest struct {
	UserID    uuid.UUID           `json:"user_id" binding:"required"`
	Type      NotificationType    `json:"type" binding:"required"`
	Subject   string              `json:"subject" binding:"required"`
	Message   string              `json:"message" binding:"required"`
	Template  *string             `json:"template,omitempty"`
	Variables map[string]string   `json:"variables,omitempty"`
	MaxRetries int                `json:"max_retries"`
}

// NotificationFilter represents filters for notification queries
type NotificationFilter struct {
	PageNumber int                 `json:"page_number"`
	PageSize   int                 `json:"page_size"`
	UserID     *uuid.UUID          `json:"user_id,omitempty"`
	Type       *NotificationType   `json:"type,omitempty"`
	Status     *NotificationStatus `json:"status,omitempty"`
}

// NotificationResponse represents the notification data returned to clients
type NotificationResponse struct {
	ID          uuid.UUID           `json:"id"`
	UserID      uuid.UUID           `json:"user_id"`
	Type        NotificationType    `json:"type"`
	Subject     string              `json:"subject"`
	Message     string              `json:"message"`
	Template    *string             `json:"template,omitempty"`
	Variables   map[string]string   `json:"variables,omitempty"`
	Status      NotificationStatus  `json:"status"`
	RetryCount  int                 `json:"retry_count"`
	MaxRetries  int                 `json:"max_retries"`
	SentAt      *time.Time          `json:"sent_at,omitempty"`
	DeliveredAt *time.Time          `json:"delivered_at,omitempty"`
	Error       *string             `json:"error,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	PageNumber int         `json:"page_number"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// EmailNotification represents an email notification
type EmailNotification struct {
	To      string            `json:"to"`
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	HTML    bool              `json:"html"`
	Headers map[string]string `json:"headers,omitempty"`
}

// SMSNotification represents an SMS notification
type SMSNotification struct {
	To      string            `json:"to"`
	Message string            `json:"message"`
	From    *string           `json:"from,omitempty"`
}

// PushNotification represents a push notification
type PushNotification struct {
	UserID  uuid.UUID         `json:"user_id"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Data    map[string]string `json:"data,omitempty"`
	Token   *string           `json:"token,omitempty"`
}