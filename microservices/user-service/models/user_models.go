package models

import (
	"time"
	"shared"

	"github.com/google/uuid"
)

// Extending shared User model with additional fields specific to user service
type User struct {
	shared.User
	NationalID             *string              `json:"national_id,omitempty" db:"national_id"`
	PassportNumber         *string              `json:"passport_number,omitempty" db:"passport_number"`
	Password               string               `json:"-" db:"password"` // Hidden from JSON
	ProfilePicture         *string              `json:"profile_picture,omitempty" db:"profile_picture"`
	SubscriptionStatus     *SubscriptionStatus  `json:"subscription_status,omitempty" db:"subscription_status"`
	InstitutionID          *string              `json:"institution_id,omitempty" db:"institution_id"`
	CommissionPercentage   *int                 `json:"commission_percentage,omitempty" db:"commission_percentage"`
	SelerType              *SelerType           `json:"seler_type,omitempty" db:"seler_type"`
	Specialization         *string              `json:"specialization,omitempty" db:"specialization"`
	Notes                  *string              `json:"notes,omitempty" db:"notes"`
	ReferralCode           *string              `json:"referral_code,omitempty" db:"referral_code"`
	HasActiveSubscription  bool                 `json:"has_active_subscription" db:"has_active_subscription"`
	IsActive               bool                 `json:"is_active" db:"is_active"`
	OTPRequired            bool                 `json:"otp_required" db:"otp_required"`
	CreatedBy              *string              `json:"created_by,omitempty" db:"created_by"`
	DeletedAt              *time.Time           `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SubscriptionStatus represents the subscription status of a user
type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "active"
	SubscriptionInactive SubscriptionStatus = "inactive"
	SubscriptionExpired  SubscriptionStatus = "expired"
	SubscriptionOnHold   SubscriptionStatus = "onhold"
	SubscriptionPaused   SubscriptionStatus = "paused"
	SubscriptionCanceled SubscriptionStatus = "canceled"
)

// SelerType represents the type of seller
type SelerType string

const (
	SelerTypeSeler    SelerType = "seler"
	SelerTypePromoter SelerType = "promoter"
)

// UserResponse represents the user data returned to clients (without sensitive info)
type UserResponse struct {
	ID                    uuid.UUID           `json:"id"`
	ClientID              string              `json:"client_id"`
	Email                 string              `json:"email"`
	FirstName             string              `json:"first_name"`
	LastName              string              `json:"last_name"`
	NationalID            *string             `json:"national_id,omitempty"`
	PassportNumber        *string             `json:"passport_number,omitempty"`
	Phone                 string              `json:"phone"`
	ProfilePicture        *string             `json:"profile_picture,omitempty"`
	Username              string              `json:"username"`
	Role                  shared.UserRole     `json:"role"`
	Status                shared.UserStatus   `json:"status"`
	SubscriptionStatus    *SubscriptionStatus `json:"subscription_status,omitempty"`
	InstitutionID         *string             `json:"institution_id,omitempty"`
	CommissionPercentage  *int                `json:"commission_percentage,omitempty"`
	SelerType             *SelerType          `json:"seler_type,omitempty"`
	Specialization        *string             `json:"specialization,omitempty"`
	Notes                 *string             `json:"notes,omitempty"`
	Slug                  string              `json:"slug"`
	ReferralCode          *string             `json:"referral_code,omitempty"`
	HasActiveSubscription bool                `json:"has_active_subscription"`
	IsActive              bool                `json:"is_active"`
	OTPRequired           bool                `json:"otp_required"`
	CreatedBy             *string             `json:"created_by,omitempty"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	ClientID       string           `json:"client_id" binding:"required"`
	Email          string           `json:"email" binding:"required,email"`
	FirstName      string           `json:"first_name" binding:"required"`
	LastName       string           `json:"last_name" binding:"required"`
	NationalID     *string          `json:"national_id,omitempty"`
	PassportNumber *string          `json:"passport_number,omitempty"`
	Password       string           `json:"password" binding:"required,min=5,max=16"`
	Phone          string           `json:"phone" binding:"required"`
	Username       string           `json:"username,omitempty"`
	Role           shared.UserRole  `json:"role,omitempty"`
	Status         shared.UserStatus `json:"status,omitempty"`
}

// UpdateUserRequest represents the request to update user information
type UpdateUserRequest struct {
	Email            *string          `json:"email,omitempty"`
	FirstName        *string          `json:"first_name,omitempty"`
	LastName         *string          `json:"last_name,omitempty"`
	NationalID       *string          `json:"national_id,omitempty"`
	PassportNumber   *string          `json:"passport_number,omitempty"`
	Phone            *string          `json:"phone,omitempty"`
	Username         *string          `json:"username,omitempty"`
	Role             *shared.UserRole `json:"role,omitempty"`
	Status           *shared.UserStatus `json:"status,omitempty"`
	Notes            *string          `json:"notes,omitempty"`
	Specialization   *string          `json:"specialization,omitempty"`
	ProfilePicture   *string          `json:"profile_picture,omitempty"`
	Password         *string          `json:"password,omitempty" binding:"omitempty,min=5,max=16"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	ClientID *string `json:"client_id,omitempty"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
	ExpiresAt   time.Time    `json:"expires_at"`
}

// PasswordResetRequest represents the password reset request
type PasswordResetRequest struct {
	Username string `json:"username" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
}

// PasswordResetConfirmRequest represents the password reset confirmation request
type PasswordResetConfirmRequest struct {
	OTP      string `json:"otp" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	List        interface{} `json:"list"`
	CurrentPage int         `json:"current_page"`
	LastPage    int         `json:"last_page"`
	NextPage    *int        `json:"next_page,omitempty"`
	PreviousPage *int       `json:"previous_page,omitempty"`
	Total       int64       `json:"total"`
	Status      string      `json:"status"`
}

// UserFilter represents filters for user queries
type UserFilter struct {
	Role               *shared.UserRole    `json:"role,omitempty"`
	Status             *shared.UserStatus  `json:"status,omitempty"`
	SubscriptionStatus *SubscriptionStatus `json:"subscription_status,omitempty"`
	InstitutionID      *string             `json:"institution_id,omitempty"`
	Search             *string             `json:"search,omitempty"` // Search in email, username, first_name, last_name
	From               *string             `json:"from,omitempty"`   // Date filter from
	To                 *string             `json:"to,omitempty"`     // Date filter to
	PageNumber         int                 `json:"page_number"`
	PageSize           int                 `json:"page_size"`
}