package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the user service
type User struct {
	ID                     uuid.UUID            `json:"id" db:"id"`
	ClientID               string               `json:"client_id" db:"client_id"`
	Email                  string               `json:"email" db:"email"`
	FirstName              string               `json:"first_name" db:"first_name"`
	LastName               string               `json:"last_name" db:"last_name"`
	NationalID             *string              `json:"national_id,omitempty" db:"national_id"`
	PassportNumber         *string              `json:"passport_number,omitempty" db:"passport_number"`
	Password               string               `json:"-" db:"password"` // Hidden from JSON
	Phone                  string               `json:"phone" db:"phone"`
	ProfilePicture         *string              `json:"profile_picture,omitempty" db:"profile_picture"`
	Username               string               `json:"username" db:"username"`
	Role                   string               `json:"role" db:"role"`
	Status                 string               `json:"status" db:"status"`
	Slug                   string               `json:"slug" db:"slug"`
	ReferralCode           *string              `json:"referral_code,omitempty" db:"referral_code"`
	HasActiveSubscription  bool                 `json:"has_active_subscription" db:"has_active_subscription"`
	IsActive               bool                 `json:"is_active" db:"is_active"`
	OTPRequired            bool                 `json:"otp_required" db:"otp_required"`
	CreatedBy              *string              `json:"created_by,omitempty" db:"created_by"`
	CreatedAt              time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at" db:"updated_at"`
	DeletedAt              *time.Time           `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	ClientID      string  `json:"client_id" binding:"required"`
	Email         string  `json:"email" binding:"required,email"`
	FirstName     string  `json:"first_name" binding:"required"`
	LastName      string  `json:"last_name" binding:"required"`
	NationalID    *string `json:"national_id,omitempty"`
	PassportNumber *string `json:"passport_number,omitempty"`
	Password      string  `json:"password" binding:"required,min=6"`
	Phone         string  `json:"phone" binding:"required"`
	ProfilePicture *string `json:"profile_picture,omitempty"`
	Username      string  `json:"username" binding:"required"`
	Role          string  `json:"role" binding:"required"`
	ReferralCode  *string `json:"referral_code,omitempty"`
	CreatedBy     *string `json:"created_by,omitempty"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	FirstName     *string `json:"first_name,omitempty"`
	LastName      *string `json:"last_name,omitempty"`
	NationalID    *string `json:"national_id,omitempty"`
	PassportNumber *string `json:"passport_number,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	ProfilePicture *string `json:"profile_picture,omitempty"`
	Username      *string `json:"username,omitempty"`
	Role          *string `json:"role,omitempty"`
	Status        *string `json:"status,omitempty"`
	ReferralCode  *string `json:"referral_code,omitempty"`
}

// UserFilter represents filters for user queries
type UserFilter struct {
	PageNumber int     `json:"page_number"`
	PageSize   int     `json:"page_size"`
	Role       *string `json:"role,omitempty"`
	Status     *string `json:"status,omitempty"`
	Search     *string `json:"search,omitempty"`
}

// UserResponse represents the user data returned to clients
type UserResponse struct {
	ID                    uuid.UUID `json:"id"`
	ClientID              string    `json:"client_id"`
	Email                 string    `json:"email"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	NationalID            *string   `json:"national_id,omitempty"`
	PassportNumber        *string   `json:"passport_number,omitempty"`
	Phone                 string    `json:"phone"`
	ProfilePicture        *string   `json:"profile_picture,omitempty"`
	Username              string    `json:"username"`
	Role                  string    `json:"role"`
	Status                string    `json:"status"`
	Slug                  string    `json:"slug"`
	ReferralCode          *string   `json:"referral_code,omitempty"`
	HasActiveSubscription bool      `json:"has_active_subscription"`
	IsActive              bool      `json:"is_active"`
	OTPRequired           bool      `json:"otp_required"`
	CreatedBy             *string   `json:"created_by,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
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