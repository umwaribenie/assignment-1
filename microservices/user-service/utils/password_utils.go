package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"shared"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash checks if a password matches its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateOTP generates a random 6-digit OTP
func GenerateOTP() (string, error) {
	max := big.NewInt(999999)
	min := big.NewInt(100000)
	
	n, err := rand.Int(rand.Reader, max.Sub(max, min).Add(max, big.NewInt(1)))
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%06d", n.Add(n, min).Int64()), nil
}

// ValidateRole checks if a role is valid
func ValidateRole(role shared.UserRole) bool {
	switch role {
	case shared.RoleUser, shared.RoleAdmin, shared.RoleSuperAdmin, shared.RoleTrainer, 
		 shared.RoleInstructor, shared.RoleFrontdesk, shared.RoleFinance, shared.RoleSeler, shared.RoleMember:
		return true
	default:
		return false
	}
}

// ValidateStatus checks if a status is valid
func ValidateStatus(status shared.UserStatus) bool {
	switch status {
	case shared.StatusActive, shared.StatusInactive, shared.StatusSuspended, shared.StatusDeleted:
		return true
	default:
		return false
	}
}