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
	StatusDeleted  UserStatus = "deleted"
)

// SubscriptionStatus represents subscription status
type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "active"
	SubscriptionInactive SubscriptionStatus = "inactive"
	SubscriptionExpired  SubscriptionStatus = "expired"
	SubscriptionOnHold   SubscriptionStatus = "onhold"
	SubscriptionPaused   SubscriptionStatus = "paused"
	SubscriptionCanceled SubscriptionStatus = "canceled"
)

// Client represents a client/organization
type Client struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Logo        *string   `json:"logo,omitempty" db:"logo"`
	BgImage     *string   `json:"bgImage,omitempty" db:"bg_image"`
	Address     *string   `json:"address,omitempty" db:"address"`
	Description *string   `json:"description,omitempty" db:"description"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Institution represents an institution
type Institution struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ClientID      uuid.UUID `json:"clientId" db:"client_id"`
	Name          string    `json:"name" db:"name"`
	Slug          string    `json:"slug" db:"slug"`
	Email         *string   `json:"email,omitempty" db:"email"`
	PhoneNumber   *string   `json:"phoneNumber,omitempty" db:"phone_number"`
	Address       *string   `json:"address,omitempty" db:"address"`
	InstitutionNo *int      `json:"institution_no,omitempty" db:"institution_no"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID                     uuid.UUID           `json:"id" db:"id"`
	ClientID               string              `json:"clientId" db:"client_id"`
	Email                  string              `json:"email" db:"email"`
	FirstName              string              `json:"firstname" db:"first_name"`
	LastName               string              `json:"lastname" db:"last_name"`
	NationalID             *string             `json:"nationalId,omitempty" db:"national_id"`
	PassportNumber         *string             `json:"passportNumber,omitempty" db:"passport_number"`
	Password               string              `json:"-" db:"password"` // Hidden from JSON
	Phone                  *string             `json:"phone,omitempty" db:"phone"`
	ProfilePicture         *string             `json:"profilePicture,omitempty" db:"profile_picture"`
	Username               string              `json:"username" db:"username"`
	Role                   UserRole            `json:"role" db:"role"`
	Status                 UserStatus          `json:"status" db:"status"`
	Slug                   string              `json:"slug" db:"slug"`
	Notes                  *string             `json:"notes,omitempty" db:"notes"`
	InstitutionID          *string             `json:"institutionId,omitempty" db:"institution_id"`
	SubscriptionStatus     *SubscriptionStatus `json:"subscriptionStatus,omitempty" db:"subscription_status"`
	HasActiveSubscription  bool                `json:"hasActiveSubscription" db:"has_active_subscription"`
	IsActive              bool                `json:"is_active" db:"is_active"`
	OTPRequired           bool                `json:"otpRequired" db:"otp_required"`
	ReferralCode          *string             `json:"referralCode,omitempty" db:"referral_code"`
	CreatedBy             *string             `json:"createdBy,omitempty" db:"created_by"`
	CreatedAt             time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at" db:"updated_at"`
	DeletedAt             *time.Time          `json:"deleted_at,omitempty" db:"deleted_at"`
}

// UserResponse represents the user data returned to clients (without sensitive info)
type UserResponse struct {
	ID                    uuid.UUID           `json:"id"`
	ClientID              string              `json:"clientId"`
	Email                 string              `json:"email"`
	FirstName             string              `json:"firstname"`
	LastName              string              `json:"lastname"`
	NationalID            *string             `json:"nationalId,omitempty"`
	PassportNumber        *string             `json:"passportNumber,omitempty"`
	Phone                 *string             `json:"phone,omitempty"`
	ProfilePicture        *string             `json:"profilePicture,omitempty"`
	Username              string              `json:"username"`
	Role                  UserRole            `json:"role"`
	Status                UserStatus          `json:"status"`
	Slug                  string              `json:"slug"`
	Notes                 *string             `json:"notes,omitempty"`
	InstitutionID         *string             `json:"institutionId,omitempty"`
	SubscriptionStatus    *SubscriptionStatus `json:"subscriptionStatus,omitempty"`
	HasActiveSubscription bool                `json:"hasActiveSubscription"`
	IsActive             bool                `json:"is_active"`
	OTPRequired          bool                `json:"otpRequired"`
	ReferralCode         *string             `json:"referralCode,omitempty"`
	CreatedBy            *string             `json:"createdBy,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

// DTO Models matching swagger definitions

// UserDto for registration
type UserDto struct {
	ClientID       string  `json:"clientId" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	FirstName      string  `json:"firstName" binding:"required"`
	LastName       string  `json:"lastName" binding:"required"`
	NationalID     *string `json:"nationalId,omitempty"`
	PassportNumber *string `json:"passportNumber,omitempty"`
	Password       string  `json:"password" binding:"required,min=5,max=16"`
	Phone          *string `json:"phone,omitempty"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Username       *string `json:"username,omitempty"`
}

// CreateUserByAdminDto for admin user creation
type CreateUserByAdminDto struct {
	Email          string   `json:"email" binding:"required"`
	FirstName      string   `json:"firstName" binding:"required"`
	LastName       string   `json:"lastName" binding:"required"`
	Password       string   `json:"password" binding:"required,min=5,max=16"`
	Role           UserRole `json:"role" binding:"required"`
	NationalID     *string  `json:"nationalId,omitempty"`
	PassportNumber *string  `json:"passportNumber,omitempty"`
	Phone          *string  `json:"phone,omitempty"`
	ProfilePicture *string  `json:"profilePicture,omitempty"`
	Username       *string  `json:"username,omitempty"`
	InstitutionID  *string  `json:"institutionId,omitempty"`
}

// UpdateUserDto for user updates
type UpdateUserDto struct {
	Email          *string   `json:"email,omitempty"`
	FirstName      *string   `json:"firstName,omitempty"`
	LastName       *string   `json:"lastName,omitempty"`
	NationalID     *string   `json:"nationalId,omitempty"`
	PassportNumber *string   `json:"passportNumber,omitempty"`
	Password       *string   `json:"password,omitempty,min=5,max=16"`
	Phone          *string   `json:"phone,omitempty"`
	ProfilePicture *string   `json:"profilePicture,omitempty"`
	Username       *string   `json:"username,omitempty"`
	Role           *UserRole `json:"role,omitempty"`
	Notes          *string   `json:"notes,omitempty"`
}

// LoginData for login request
type LoginData struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	ClientID *string `json:"clientId,omitempty"`
}

// LoginSuccessResponse for successful login
type LoginSuccessResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

// PasswordResetRequest for password reset
type PasswordResetRequest struct {
	Username string `json:"username" binding:"required"`
	ClientID string `json:"clientId" binding:"required"`
}

// ConfirmOTPRequest for OTP confirmation
type ConfirmOTPRequest struct {
	OTP      string `json:"otp" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// ConfirmLoginOTPRequest for login OTP verification
type ConfirmLoginOTPRequest struct {
	OTP string `json:"otp" binding:"required"`
}

// UpdatePasswordRequest for password updates
type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// Response DTOs
type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// PaginationResponse for paginated results
type PaginationResponse struct {
	List        interface{} `json:"list"`
	Total       int         `json:"total"`
	CurrentPage int         `json:"currentPage"`
	LastPage    int         `json:"lastPage"`
	NextPage    *int        `json:"nextPage,omitempty"`
	PreviousPage *int       `json:"previousPage,omitempty"`
	Status      string      `json:"status"`
}

// Response wrapper
type Response struct {
	Data       interface{} `json:"data,omitempty"`
	Message    interface{} `json:"message,omitempty"`
	Error      *string     `json:"error,omitempty"`
	Status     string      `json:"status"`
	StatusCode int         `json:"statusCode"`
}

// UserFilter represents filters for user queries with new fields
type UserFilter struct {
	Role               *UserRole            `json:"role,omitempty"`
	Status             *UserStatus          `json:"status,omitempty"`
	SubscriptionStatus *SubscriptionStatus  `json:"subscriptionStatus,omitempty"`
	Search             *string              `json:"search,omitempty"`
	From               *string              `json:"from,omitempty"`
	To                 *string              `json:"to,omitempty"`
	InstitutionID      *string              `json:"institutionId,omitempty"`
	PageNumber         int                  `json:"pageNumber"`
	PageSize           int                  `json:"pageSize"`
}

// APIResponse represents a standard API response (keeping for backward compatibility)
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

// LoginOTP represents a login OTP record
type LoginOTP struct {
	ID        int       `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	OTP       string    `json:"otp" db:"otp"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}