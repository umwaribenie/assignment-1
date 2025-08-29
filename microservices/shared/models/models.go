package models

import (
	"time"
)

// UserRole defines the type for user roles which can be user or admin.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// UserStatus defines the type for user statuses (active, inactive and deleted).
type UserStatus string

const (
	ActiveStatus   UserStatus = "active"
	InactiveStatus UserStatus = "inactive"
	DeletedStatus  UserStatus = "deleted"
)

// User represents the user data structure shared across services
type User struct {
	ID             string         `json:"id"`
	ClientID       string         `json:"clientId"`
	Email          string         `json:"email"`
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName"`
	NationalID     *string        `json:"nationalId,omitempty"`
	PassportNumber *string        `json:"passportNumber,omitempty"`
	Phone          string         `json:"phone"`
	ProfilePicture *string        `json:"profilePicture,omitempty"`
	Username       string         `json:"username"`
	Slug           string         `json:"slug"`
	Role           UserRole       `json:"role"`
	Status         UserStatus     `json:"status"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// UserResponse represents the user response structure
type UserResponse struct {
	ID             string         `json:"id"`
	ClientID       string         `json:"clientId"`
	Email          string         `json:"email"`
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName"`
	NationalID     *string        `json:"nationalId,omitempty"`
	PassportNumber *string        `json:"passportNumber,omitempty"`
	Phone          string         `json:"phone"`
	ProfilePicture *string        `json:"profilePicture,omitempty"`
	Username       string         `json:"username"`
	Slug           string         `json:"slug"`
	Role           UserRole       `json:"role"`
	Status         UserStatus     `json:"status"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
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