package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// NotificationType defines the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypePush  NotificationType = "push"
)

// NotificationStatus defines the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// Notification represents the notification database model
type Notification struct {
	ID          string             `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID      string             `gorm:"not null" json:"userId"`
	Type        NotificationType   `gorm:"type:varchar(20);not null" json:"type"`
	Subject     string             `gorm:"type:text" json:"subject"`
	Message     string             `gorm:"type:text;not null" json:"message"`
	Recipient   string             `gorm:"not null" json:"recipient"` // email, phone, or device token
	Status      NotificationStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Template    string             `gorm:"type:varchar(100)" json:"template"`
	Variables   JSONB              `gorm:"type:jsonb" json:"variables"`
	RetryCount  int                `gorm:"default:0" json:"retryCount"`
	MaxRetries  int                `gorm:"default:3" json:"maxRetries"`
	SentAt      *time.Time         `json:"sentAt"`
	FailedAt    *time.Time         `json:"failedAt"`
	ErrorMsg    string             `gorm:"type:text" json:"errorMsg"`
	CreatedAt   time.Time          `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt     `gorm:"index" json:"-"`
}

// JSONB is a custom type for handling JSONB fields in PostgreSQL
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// CreateNotificationRequest represents the request for creating a notification
type CreateNotificationRequest struct {
	UserID    string             `json:"userId" binding:"required"`
	Type      NotificationType   `json:"type" binding:"required,oneof=email sms push"`
	Subject   string             `json:"subject"`
	Message   string             `json:"message" binding:"required"`
	Recipient string             `json:"recipient" binding:"required"`
	Template  string             `json:"template"`
	Variables map[string]interface{} `json:"variables"`
}

// NotificationFilter represents filtering options for notification queries
type NotificationFilter struct {
	PageNumber int                `json:"pageNumber"`
	PageSize   int                `json:"pageSize"`
	UserID     string             `json:"userId"`
	Type       NotificationType   `json:"type"`
	Status     NotificationStatus `json:"status"`
	From       string             `json:"from"`
	To         string             `json:"to"`
}

// NotificationResponse represents the notification response structure
type NotificationResponse struct {
	ID          string             `json:"id"`
	UserID      string             `json:"userId"`
	Type        NotificationType   `json:"type"`
	Subject     string             `json:"subject"`
	Message     string             `json:"message"`
	Recipient   string             `json:"recipient"`
	Status      NotificationStatus `json:"status"`
	Template    string             `json:"template"`
	Variables   map[string]interface{} `json:"variables"`
	RetryCount  int                `json:"retryCount"`
	MaxRetries  int                `json:"maxRetries"`
	SentAt      *time.Time         `json:"sentAt"`
	FailedAt    *time.Time         `json:"failedAt"`
	ErrorMsg    string             `json:"errorMsg"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	CurrentPage  int         `json:"currentPage"`
	LastPage     int         `json:"lastPage"`
	List         interface{} `json:"list"`
	NextPage     *int        `json:"nextPage,omitempty"`
	PreviousPage *int        `json:"previousPage,omitempty"`
	Status       string      `json:"status"`
	Total        int64       `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// EmailNotification represents email-specific notification data
type EmailNotification struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	HTML    bool   `json:"html"`
}

// SMSNotification represents SMS-specific notification data
type SMSNotification struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

// PushNotification represents push notification data
type PushNotification struct {
	Token   string `json:"token"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Data    map[string]interface{} `json:"data"`
}