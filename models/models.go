package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

// User represents a user in the system
type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ClientID        string     `json:"client_id" db:"client_id"`
	Email           string     `json:"email" db:"email"`
	FirstName       string     `json:"first_name" db:"first_name"`
	LastName        string     `json:"last_name" db:"last_name"`
	NationalID      *string    `json:"national_id,omitempty" db:"national_id"`
	PassportNumber  *string    `json:"passport_number,omitempty" db:"passport_number"`
	Password        string     `json:"-" db:"password"` // Hidden from JSON
	Phone           string     `json:"phone" db:"phone"`
	ProfilePicture  *string    `json:"profile_picture,omitempty" db:"profile_picture"`
	Username        string     `json:"username" db:"username"`
	Role            UserRole   `json:"role" db:"role"`
	Status          UserStatus `json:"status" db:"status"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserResponse represents the user data returned to clients (without sensitive info)
type UserResponse struct {
	ID             uuid.UUID  `json:"id"`
	ClientID       string     `json:"client_id"`
	Email          string     `json:"email"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	NationalID     *string    `json:"national_id,omitempty"`
	PassportNumber *string    `json:"passport_number,omitempty"`
	Phone          string     `json:"phone"`
	ProfilePicture *string    `json:"profile_picture,omitempty"`
	Username       string     `json:"username"`
	Role           UserRole   `json:"role"`
	Status         UserStatus `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email          string     `json:"email" binding:"required,email"`
	FirstName      string     `json:"first_name" binding:"required"`
	LastName       string     `json:"last_name" binding:"required"`
	NationalID     *string    `json:"national_id,omitempty"`
	PassportNumber *string    `json:"passport_number,omitempty"`
	Password       string     `json:"password" binding:"required,min=8"`
	Phone          string     `json:"phone" binding:"required"`
	Username       string     `json:"username" binding:"required"`
	Role           UserRole   `json:"role,omitempty"`
	Status         UserStatus `json:"status,omitempty"`
}

// UpdateUserRequest represents the request to update user information
type UpdateUserRequest struct {
	FirstName      *string    `json:"first_name,omitempty"`
	LastName       *string    `json:"last_name,omitempty"`
	NationalID     *string    `json:"national_id,omitempty"`
	PassportNumber *string    `json:"passport_number,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	Username       *string    `json:"username,omitempty"`
	Role           *UserRole  `json:"role,omitempty"`
	Status         *UserStatus `json:"status,omitempty"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token        string       `json:"token"`
	User         UserResponse `json:"user"`
	ExpiresAt    time.Time    `json:"expires_at"`
	RefreshToken string       `json:"refresh_token,omitempty"`
}

// PasswordResetRequest represents the password reset request
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordUpdateRequest represents the password update request
type PasswordUpdateRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// PasswordResetConfirmRequest represents the password reset confirmation request
type PasswordResetConfirmRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

// UserFilter represents filters for user queries
type UserFilter struct {
	Role     *UserRole   `json:"role,omitempty"`
	Status   *UserStatus `json:"status,omitempty"`
	Search   *string     `json:"search,omitempty"` // Search in email, username, first_name, last_name
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
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