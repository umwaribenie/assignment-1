package models

import "time"

// NotificationEvent represents events received from Kafka
type NotificationEvent struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	Email     string                 `json:"email"`
	Phone     string                 `json:"phone,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Event types
const (
	EventUserRegistered      = "USER_REGISTERED"
	EventPasswordReset       = "PASSWORD_RESET_REQUESTED"
	EventPasswordChanged     = "PASSWORD_CHANGED"
	EventEmailVerification   = "EMAIL_VERIFICATION"
	EventLoginNotification   = "LOGIN_NOTIFICATION"
	EventAccountSuspended    = "ACCOUNT_SUSPENDED"
	EventAccountReactivated  = "ACCOUNT_REACTIVATED"
	EventOTPGenerated        = "OTP_GENERATED"
)

// EmailNotification represents an email to be sent
type EmailNotification struct {
	To          []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments []string
}

// SMSNotification represents an SMS to be sent
type SMSNotification struct {
	To   string
	Body string
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	Name    string
	Subject string
	Body    string
	IsHTML  bool
}