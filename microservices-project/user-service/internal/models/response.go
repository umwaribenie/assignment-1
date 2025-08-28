package models

import "time"

// LoginResponse represents login response
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// UserResponse represents user response
type UserResponse struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	Username        string     `json:"username"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	PhoneNumber     string     `json:"phone_number,omitempty"`
	ProfilePicture  string     `json:"profile_picture,omitempty"`
	Role            UserRole   `json:"role"`
	Status          UserStatus `json:"status"`
	EmailVerified   bool       `json:"email_verified"`
	PhoneVerified   bool       `json:"phone_verified"`
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NotificationEvent represents events sent to notification service
type NotificationEvent struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	Email     string                 `json:"email"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Event types for notification service
const (
	EventUserRegistered      = "USER_REGISTERED"
	EventPasswordReset       = "PASSWORD_RESET_REQUESTED"
	EventPasswordChanged     = "PASSWORD_CHANGED"
	EventEmailVerification   = "EMAIL_VERIFICATION"
	EventLoginNotification   = "LOGIN_NOTIFICATION"
	EventAccountSuspended    = "ACCOUNT_SUSPENDED"
	EventAccountReactivated  = "ACCOUNT_REACTIVATED"
)