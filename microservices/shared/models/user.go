package models

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

// User represents a user in the system
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
	Role                   UserRole             `json:"role" db:"role"`
	Status                 UserStatus           `json:"status" db:"status"`
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
	Role                  UserRole            `json:"role"`
	Status                UserStatus          `json:"status"`
	Slug                  string              `json:"slug"`
	ReferralCode          *string             `json:"referral_code,omitempty"`
	HasActiveSubscription bool                `json:"has_active_subscription"`
	IsActive              bool                `json:"is_active"`
	OTPRequired           bool                `json:"otp_required"`
	CreatedBy             *string             `json:"created_by,omitempty"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at"`
}

// UserFilter represents filters for user queries
type UserFilter struct {
	PageNumber int       `json:"page_number"`
	PageSize   int       `json:"page_size"`
	Role       *UserRole `json:"role,omitempty"`
	Status     *UserStatus `json:"status,omitempty"`
	Search     *string   `json:"search,omitempty"`
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