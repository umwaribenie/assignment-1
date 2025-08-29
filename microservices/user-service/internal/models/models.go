package models

import (
	"time"

	"gorm.io/gorm"
	sharedModels "microservices/shared/models"
)

// User represents the user database model
type User struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ClientID       string         `gorm:"unique" json:"clientId"`
	Email          string         `gorm:"uniqueIndex" json:"email"`
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName"`
	NationalID     *string        `gorm:"unique" json:"nationalId,omitempty"`
	PassportNumber *string        `gorm:"unique" json:"passportNumber,omitempty"`
	Password       string         `json:"-"`
	Phone          string         `gorm:"unique" json:"phone"`
	ProfilePicture *string        `json:"profilePicture,omitempty"`
	Username       string         `gorm:"uniqueIndex" json:"username"`
	Slug           string         `gorm:"uniqueIndex" json:"slug"`
	Role           sharedModels.UserRole       `gorm:"type:varchar(50);default:'user'" json:"role"`
	Status         sharedModels.UserStatus     `gorm:"type:varchar(50);default:'active'" json:"status"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// CreateUserRequest represents the request for creating a user
type CreateUserRequest struct {
	ClientID       string  `json:"clientId" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	FirstName      string  `json:"firstName" binding:"required"`
	LastName       string  `json:"lastName" binding:"required"`
	NationalID     *string `json:"nationalId"`
	PassportNumber *string `json:"passportNumber"`
	Password       string  `json:"password" binding:"required,min=6"`
	Phone          string  `json:"phone" binding:"required"`
	ProfilePicture *string `json:"profilePicture"`
	Username       string  `json:"username" binding:"required"`
}

// CreateUserByAdminRequest represents the request for creating a user by admin
type CreateUserByAdminRequest struct {
	Email          string                `json:"email" binding:"required,email"`
	FirstName      string                `json:"firstName" binding:"required"`
	LastName       string                `json:"lastName" binding:"required"`
	NationalID     *string               `json:"nationalId"`
	PassportNumber *string               `json:"passportNumber"`
	Password       string                `json:"password" binding:"required,min=6"`
	Phone          string                `json:"phone" binding:"required"`
	ProfilePicture *string               `json:"profilePicture"`
	Role           sharedModels.UserRole `json:"role" binding:"required,oneof=user admin"`
	Username       string                `json:"username" binding:"required"`
}

// UpdateUserRequest represents the request for updating a user
type UpdateUserRequest struct {
	Email          *string                `json:"email,omitempty"`
	FirstName      *string                `json:"firstName,omitempty"`
	LastName       *string                `json:"lastName,omitempty"`
	NationalID     *string                `json:"nationalId,omitempty"`
	PassportNumber *string                `json:"passportNumber,omitempty"`
	Phone          *string                `json:"phone,omitempty"`
	ProfilePicture *string                `json:"profilePicture,omitempty"`
	Role           *sharedModels.UserRole `json:"role,omitempty" binding:"omitempty,oneof=user admin"`
	Username       *string                `json:"username,omitempty"`
}

// UserFilter represents filtering options for user queries
type UserFilter struct {
	PageNumber int    `json:"pageNumber"`
	PageSize   int    `json:"pageSize"`
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Search     string `json:"search,omitempty"`
	Role       string `json:"role,omitempty"`
	Status     string `json:"status,omitempty"`
}

// UserResponse represents the user response structure
type UserResponse struct {
	ID             string                `json:"id"`
	ClientID       string                `json:"clientId"`
	Email          string                `json:"email"`
	FirstName      string                `json:"firstName"`
	LastName       string                `json:"lastName"`
	NationalID     *string               `json:"nationalId,omitempty"`
	PassportNumber *string               `json:"passportNumber,omitempty"`
	Phone          string                `json:"phone"`
	ProfilePicture *string               `json:"profilePicture,omitempty"`
	Username       string                `json:"username"`
	Slug           string                `json:"slug"`
	Role           sharedModels.UserRole `json:"role"`
	Status         sharedModels.UserStatus `json:"status"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	CurrentPage  int         `json:"currentPage"`
	LastPage     int         `json:"lastPage"`
	List         interface{} `json:"list"`
	NextPage     *int        `json:"nextPage,omitempty"`
	PreviousPage *int        `json:"previousPage,omitempty"`
	Status       string      `json:"status"`
	Total        int64       `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}