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
	SubscriptionStatus     *SubscriptionStatus  `json:"subscription_status,omitempty" db:"subscription_status"`
	InstitutionID          *string              `json:"institution_id,omitempty" db:"institution_id"`
	CommissionPercentage   *int                 `json:"commission_percentage,omitempty" db:"commission_percentage"`
	SelerType              *SelerType           `json:"seler_type,omitempty" db:"seler_type"`
	Specialization         *string              `json:"specialization,omitempty" db:"specialization"`
	Notes                  *string              `json:"notes,omitempty" db:"notes"`
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
	ClientID       string     `json:"client_id" binding:"required"`
	Email          string     `json:"email" binding:"required,email"`
	FirstName      string     `json:"first_name" binding:"required"`
	LastName       string     `json:"last_name" binding:"required"`
	NationalID     *string    `json:"national_id,omitempty"`
	PassportNumber *string    `json:"passport_number,omitempty"`
	Password       string     `json:"password" binding:"required,min=5,max=16"`
	Phone          string     `json:"phone" binding:"required"`
	Username       string     `json:"username,omitempty"`
	Role           UserRole   `json:"role,omitempty"`
	Status         UserStatus `json:"status,omitempty"`
}

// CreateUserByAdminRequest represents the request to create a new user by admin
type CreateUserByAdminRequest struct {
	Email                string    `json:"email" binding:"required,email"`
	FirstName            string    `json:"first_name" binding:"required"`
	LastName             string    `json:"last_name" binding:"required"`
	NationalID           *string   `json:"national_id,omitempty"`
	PassportNumber       *string   `json:"passport_number,omitempty"`
	Password             string    `json:"password" binding:"required,min=5,max=16"`
	Phone                *string   `json:"phone,omitempty"`
	Username             *string   `json:"username,omitempty"`
	Role                 UserRole  `json:"role" binding:"required,oneof=member admin trainer instructor frontdesk finance seler"`
	InstitutionID        *string   `json:"institution_id,omitempty"`
	CommissionPercentage *int      `json:"commission_percentage,omitempty" binding:"omitempty,min=1,max=100"`
	SelerType            *SelerType `json:"seler_type,omitempty" binding:"omitempty,oneof=seler promoter"`
	Specialization       *string   `json:"specialization,omitempty"`
	ProfilePicture       *string   `json:"profile_picture,omitempty"`
}

// UpdateUserRequest represents the request to update user information
type UpdateUserRequest struct {
	Email            *string   `json:"email,omitempty"`
	FirstName        *string   `json:"first_name,omitempty"`
	LastName         *string   `json:"last_name,omitempty"`
	NationalID       *string   `json:"national_id,omitempty"`
	PassportNumber   *string   `json:"passport_number,omitempty"`
	Phone            *string   `json:"phone,omitempty"`
	Username         *string   `json:"username,omitempty"`
	Role             *UserRole `json:"role,omitempty"`
	Status           *UserStatus `json:"status,omitempty"`
	Notes            *string   `json:"notes,omitempty"`
	Specialization   *string   `json:"specialization,omitempty"`
	ProfilePicture   *string   `json:"profile_picture,omitempty"`
	Password         *string   `json:"password,omitempty" binding:"omitempty,min=5,max=16"`
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

// OTPLoginRequest represents the OTP login verification request
type OTPLoginRequest struct {
	OTP string `json:"otp" binding:"required"`
}

// PasswordResetRequest represents the password reset request
type PasswordResetRequest struct {
	Username string `json:"username" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
}

// PasswordUpdateRequest represents the password update request
type PasswordUpdateRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
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
	Role               *UserRole           `json:"role,omitempty"`
	Status             *UserStatus         `json:"status,omitempty"`
	SubscriptionStatus *SubscriptionStatus `json:"subscription_status,omitempty"`
	InstitutionID      *string             `json:"institution_id,omitempty"`
	Search             *string             `json:"search,omitempty"` // Search in email, username, first_name, last_name
	From               *string             `json:"from,omitempty"`   // Date filter from
	To                 *string             `json:"to,omitempty"`     // Date filter to
	PageNumber         int                 `json:"page_number"`
	PageSize           int                 `json:"page_size"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Response represents a generic response
type Response struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    interface{} `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// TokenBlacklist represents a blacklisted token
type TokenBlacklist struct {
	ID        int       `json:"id" db:"id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// PasswordReset represents a password reset record
type PasswordReset struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	OTP       string    `json:"otp" db:"otp"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// LoginOTP represents a login OTP record
type LoginOTP struct {
	ID        int       `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	OTP       string    `json:"otp" db:"otp"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

