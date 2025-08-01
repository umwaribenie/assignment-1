package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser       UserRole = "user"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "super_admin"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"
)

type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ClientID       string     `json:"client_id" db:"client_id"`
	Email          string     `json:"email" db:"email"`
	FirstName      string     `json:"first_name" db:"first_name"`
	LastName       string     `json:"last_name" db:"last_name"`
	NationalID     *string    `json:"national_id,omitempty" db:"national_id"`
	PassportNumber *string    `json:"passport_number,omitempty" db:"passport_number"`
	Password       string     `json:"-" db:"password"`
	Phone          string     `json:"phone" db:"phone"`
	ProfilePicture *string    `json:"profile_picture,omitempty" db:"profile_picture"`
	Username       string     `json:"username" db:"username"`
	Role           UserRole   `json:"role" db:"role"`
	Status         UserStatus `json:"status" db:"status"`

	Slug string `json:"slug" db:"slug"`

	OTPRequired bool       `json:"otp_required" db:"otp_required"`
	CreatedBy   *string    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type UserResponse struct {
	ClientID       string  `json:"client_id"`
	Email          string  `json:"email"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Phone          string  `json:"phone"`
	Username       string  `json:"username"`
	NationalID     *string `json:"nationalId,omitempty"`
	PassportNumber *string `json:"passportNumber,omitempty"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
}

type UserListResponse struct {
	ClientID  string     `json:"clientId"`
	Email     string     `json:"email"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Phone     string     `json:"phone"`
	Username  string     `json:"username"`
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type CreateUserRequest struct {
	ClientID       string  `json:"client_id" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	FirstName      string  `json:"firstname" binding:"required"`
	LastName       string  `json:"lastname" binding:"required"`
	NationalID     *string `json:"nationalId,omitempty"`
	PassportNumber *string `json:"passportNumber,omitempty"`
	Password       string  `json:"password" binding:"required,min=5,max=16"`
	Phone          string  `json:"phone" binding:"required"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Username       string  `json:"username,omitempty"`
}

type RegisterResponse struct {
	ClientID       string  `json:"clientId"`
	Email          string  `json:"email"`
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	NationalID     *string `json:"nationalId,omitempty"`
	PassportNumber *string `json:"passportNumber,omitempty"`
	Password       string  `json:"password"`
	Phone          string  `json:"phone"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Username       string  `json:"username"`
}

type CreateUserByAdminRequest struct {
	Email     string   `json:"email" binding:"required,email"`
	FirstName string   `json:"first_name" binding:"required"`
	LastName  string   `json:"last_name" binding:"required"`
	Password  string   `json:"password" binding:"required,min=5,max=16"`
	Phone     *string  `json:"phone,omitempty"`
	Username  *string  `json:"username,omitempty"`
	Role      UserRole `json:"role" binding:"required"`
}

type AdminRegisterRequest struct {
	Email          string  `json:"email" binding:"required,email"`
	FirstName      string  `json:"firstName" binding:"required"`
	LastName       string  `json:"lastName" binding:"required"`
	NationalID     *string `json:"nationalId,omitempty"`
	PassportNumber *string `json:"passportNumber,omitempty"`
	Password       string  `json:"password" binding:"required,min=5,max=16"`
	Phone          string  `json:"phone" binding:"required"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Username       string  `json:"username,omitempty"`
}

type AdminRegisterResponse struct {
	Email          string  `json:"email"`
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	NationalID     *string `json:"nationalId,omitempty"`
	PassportNumber *string `json:"passportNumber,omitempty"`
	Password       string  `json:"password"`
	Phone          string  `json:"phone"`
	ProfilePicture *string `json:"profilePicture,omitempty"`
	Username       string  `json:"username"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Username  *string `json:"username,omitempty"`
}

type UserUpdateRequest struct {
	Email          *string   `json:"email,omitempty"`
	FirstName      *string   `json:"firstName,omitempty"`
	LastName       *string   `json:"lastName,omitempty"`
	NationalID     *string   `json:"nationalId,omitempty"`
	PassportNumber *string   `json:"passportNumber,omitempty"`
	Password       *string   `json:"password,omitempty"`
	Phone          *string   `json:"phone,omitempty"`
	ProfilePicture *string   `json:"profilePicture,omitempty"`
	Role           *UserRole `json:"role,omitempty"`
	Username       *string   `json:"username,omitempty"`
}

type UserUpdateResponse struct {
	Email          string   `json:"email"`
	FirstName      string   `json:"firstName"`
	LastName       string   `json:"lastName"`
	NationalID     *string  `json:"nationalId,omitempty"`
	PassportNumber *string  `json:"passportNumber,omitempty"`
	Password       string   `json:"password"`
	Phone          string   `json:"phone"`
	ProfilePicture *string  `json:"profilePicture,omitempty"`
	Role           UserRole `json:"role"`
	Username       string   `json:"username"`
}

type LoginRequest struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	ClientID *string `json:"client_id,omitempty"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
	ExpiresAt   time.Time    `json:"expires_at"`
}

type UserLoginRequest struct {
	ClientID string `json:"clientId" binding:"required"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type UserLoginResponse struct {
	Message string `json:"message"`
}

type OTPLoginRequest struct {
	OTP string `json:"otp" binding:"required"`
}

type PasswordResetRequest struct {
	Username string `json:"username" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
}

type PasswordUpdateRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type PasswordResetConfirmRequest struct {
	OTP      string `json:"otp" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type PaginatedResponse struct {
	List         interface{} `json:"list"`
	CurrentPage  int         `json:"currentPage"`
	LastPage     int         `json:"lastPage"`
	NextPage     *int        `json:"nextPage,omitempty"`
	PreviousPage *int        `json:"previousPage,omitempty"`
	Total        int64       `json:"total"`
	Status       string      `json:"status"`
}

type UserFilter struct {
	Role       *UserRole   `json:"role,omitempty"`
	Status     *UserStatus `json:"status,omitempty"`
	Search     *string     `json:"search,omitempty"`
	PageNumber int         `json:"page_number"`
	PageSize   int         `json:"page_size"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TokenBlacklist struct {
	ID        int       `json:"id" db:"id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type PasswordReset struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	OTP       string    `json:"otp" db:"otp"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type LoginOTP struct {
	ID        int       `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	OTP       string    `json:"otp" db:"otp"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ResetPasswordEmailRequest struct {
	ClientID string `json:"clientId" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type ResetPasswordEmailResponse struct {
	Message string `json:"message"`
}

type AuthUpdatePasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required,min=6"`
	OldPassword string `json:"oldPassword" binding:"required"`
}

type AuthUpdatePasswordResponse struct {
	Message string `json:"message"`
}

type VerifyLoginOTPRequest struct {
	OTP string `json:"otp" binding:"required" example:"123456"`
}

