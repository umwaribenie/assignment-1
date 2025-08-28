package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	RoleSeller     UserRole = "seller"
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
	ID                 uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email              string         `gorm:"uniqueIndex;not null" json:"email"`
	Username           string         `gorm:"uniqueIndex;not null" json:"username"`
	Password           string         `gorm:"not null" json:"-"`
	FirstName          string         `json:"first_name"`
	LastName           string         `json:"last_name"`
	PhoneNumber        string         `json:"phone_number"`
	ProfilePicture     string         `json:"profile_picture"`
	Bio                string         `json:"bio"`
	DateOfBirth        *time.Time     `json:"date_of_birth"`
	Gender             string         `json:"gender"`
	Address            string         `json:"address"`
	City               string         `json:"city"`
	State              string         `json:"state"`
	Country            string         `json:"country"`
	ZipCode            string         `json:"zip_code"`
	Role               UserRole       `gorm:"type:varchar(50);default:'user'" json:"role"`
	Status             UserStatus     `gorm:"type:varchar(50);default:'active'" json:"status"`
	EmailVerified      bool           `gorm:"default:false" json:"email_verified"`
	EmailVerifiedAt    *time.Time     `json:"email_verified_at"`
	PhoneVerified      bool           `gorm:"default:false" json:"phone_verified"`
	PhoneVerifiedAt    *time.Time     `json:"phone_verified_at"`
	TwoFactorEnabled   bool           `gorm:"default:false" json:"two_factor_enabled"`
	LastLoginAt        *time.Time     `json:"last_login_at"`
	LastLoginIP        string         `json:"last_login_ip"`
	LoginCount         int            `gorm:"default:0" json:"login_count"`
	FailedLoginCount   int            `gorm:"default:0" json:"failed_login_count"`
	PasswordChangedAt  *time.Time     `json:"password_changed_at"`
	PasswordResetToken string         `json:"-"`
	PasswordResetExpiry *time.Time    `json:"-"`
	Preferences        map[string]interface{} `gorm:"type:jsonb" json:"preferences"`
	Metadata           map[string]interface{} `gorm:"type:jsonb" json:"metadata"`
	ClientID           string         `json:"client_id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserFilter represents filters for querying users
type UserFilter struct {
	PageNumber int
	PageSize   int
	Role       *UserRole
	Status     *UserStatus
	Search     string
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data         interface{} `json:"data"`
	TotalRecords int64       `json:"total_records"`
	TotalPages   int         `json:"total_pages"`
	CurrentPage  int         `json:"current_page"`
	PageSize     int         `json:"page_size"`
}