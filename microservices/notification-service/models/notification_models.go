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
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
	NotificationStatusRetry   NotificationStatus = "retry"
)

// Notification represents a notification record
type Notification struct {
	ID          uuid.UUID          `json:"id"`
	Type        NotificationType   `json:"type"`
	Recipient   string             `json:"recipient"` // Email or phone number
	Subject     string             `json:"subject,omitempty"`
	Content     string             `json:"content"`
	Status      NotificationStatus `json:"status"`
	ErrorMsg    string             `json:"error_msg,omitempty"`
	RetryCount  int                `json:"retry_count"`
	UserID      *uuid.UUID         `json:"user_id,omitempty"`
	TemplateID  string             `json:"template_id,omitempty"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
	ScheduledAt *time.Time         `json:"scheduled_at,omitempty"`
	SentAt      *time.Time         `json:"sent_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
}

// SMSTemplate represents an SMS template
type SMSTemplate struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	To       string            `json:"to" binding:"required,email"`
	Subject  string            `json:"subject" binding:"required"`
	Content  string            `json:"content" binding:"required"`
	IsHTML   bool              `json:"is_html"`
	UserID   *uuid.UUID        `json:"user_id,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SendSMSRequest represents a request to send an SMS
type SendSMSRequest struct {
	To       string            `json:"to" binding:"required"`
	Content  string            `json:"content" binding:"required"`
	UserID   *uuid.UUID        `json:"user_id,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SendTemplateEmailRequest represents a request to send a templated email
type SendTemplateEmailRequest struct {
	To         string            `json:"to" binding:"required,email"`
	TemplateID string            `json:"template_id" binding:"required"`
	Variables  map[string]string `json:"variables,omitempty"`
	UserID     *uuid.UUID        `json:"user_id,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Built-in templates
var EmailTemplates = map[string]EmailTemplate{
	"welcome": {
		ID:       "welcome",
		Name:     "Welcome Email",
		Subject:  "Welcome to {{.AppName}}!",
		HTMLBody: `<h1>Welcome {{.FirstName}}!</h1><p>Thank you for joining {{.AppName}}. We're excited to have you on board!</p>`,
		TextBody: "Welcome {{.FirstName}}! Thank you for joining {{.AppName}}. We're excited to have you on board!",
	},
	"password_reset": {
		ID:       "password_reset",
		Name:     "Password Reset",
		Subject:  "Password Reset - {{.AppName}}",
		HTMLBody: `<h2>Password Reset Request</h2><p>Your OTP code is: <strong>{{.OTP}}</strong></p><p>This code expires in 15 minutes.</p>`,
		TextBody: "Password Reset Request. Your OTP code is: {{.OTP}}. This code expires in 15 minutes.",
	},
	"otp_verification": {
		ID:       "otp_verification",
		Name:     "OTP Verification",
		Subject:  "Verification Code - {{.AppName}}",
		HTMLBody: `<h2>Verification Code</h2><p>Your verification code is: <strong>{{.OTP}}</strong></p><p>This code expires in 10 minutes.</p>`,
		TextBody: "Your verification code is: {{.OTP}}. This code expires in 10 minutes.",
	},
}

var SMSTemplates = map[string]SMSTemplate{
	"otp_sms": {
		ID:      "otp_sms",
		Name:    "OTP SMS",
		Content: "Your {{.AppName}} verification code is: {{.OTP}}. Expires in 10 minutes.",
	},
	"password_reset_sms": {
		ID:      "password_reset_sms",
		Name:    "Password Reset SMS",
		Content: "Your {{.AppName}} password reset code is: {{.OTP}}. Expires in 15 minutes.",
	},
}