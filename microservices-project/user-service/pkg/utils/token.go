package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateResetToken generates a random reset token
func GenerateResetToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based token
		return hex.EncodeToString([]byte(string(time.Now().Unix())))
	}
	return hex.EncodeToString(bytes)
}

// GenerateVerificationToken generates a random verification token
func GenerateVerificationToken() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based token
		return hex.EncodeToString([]byte(string(time.Now().Unix())))
	}
	return hex.EncodeToString(bytes)
}