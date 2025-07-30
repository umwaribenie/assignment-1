package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"generalusermanagement/config"
	"generalusermanagement/models"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

// Claims represents the JWT claims
type Claims struct {
	UserID   string          `json:"user_id"`
	Email    string          `json:"email"`
	Role     models.UserRole `json:"role"`
	Username string          `json:"username"`
	ClientID string          `json:"client_id"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(user models.User) (string, time.Time, error) {
	expirationTime := time.Now().Add(time.Duration(config.AppConfig.JWTExpiryHours) * time.Hour)

	claims := &Claims{
		UserID:   user.ID.String(),
		Email:    user.Email,
		Role:     user.Role,
		Username: user.Username,
		ClientID: user.ClientID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "generalusermanagement",
			Subject:   user.ID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))

	return tokenString, expirationTime, err
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}

// GenerateOTP generates a random 6-digit OTP
func GenerateOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

// GenerateClientID generates a unique client ID
func GenerateClientID() string {
	return fmt.Sprintf("USR_%d_%s", time.Now().Unix(), uuid.New().String()[:8])
}

// GenerateSlug generates a URL-friendly slug from a string
func GenerateSlug(text string) string {
	// Convert to lowercase
	slug := strings.ToLower(text)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// If slug is empty, generate a random one
	if slug == "" {
		slug = fmt.Sprintf("user-%s", uuid.New().String()[:8])
	}

	return slug
}

// GenerateUniqueSlug generates a unique slug by appending a random string if needed
func GenerateUniqueSlug(baseText string, userID string) string {
	baseSlug := GenerateSlug(baseText)

	// Append first 8 characters of user ID to ensure uniqueness
	if userID != "" {
		baseSlug = fmt.Sprintf("%s-%s", baseSlug, userID[:8])
	} else {
		baseSlug = fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:8])
	}

	return baseSlug
}

// GenerateReferralCode generates a unique referral code
func GenerateReferralCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// ValidateRole checks if a role is valid
func ValidateRole(role models.UserRole) bool {
	validRoles := []models.UserRole{
		models.RoleUser,
		models.RoleAdmin,
		models.RoleSuperAdmin,
		models.RoleTrainer,
		models.RoleInstructor,
		models.RoleFrontdesk,
		models.RoleFinance,
		models.RoleSeler,
		models.RoleMember,
	}

	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// ValidateSubscriptionStatus checks if a subscription status is valid
func ValidateSubscriptionStatus(status models.SubscriptionStatus) bool {
	validStatuses := []models.SubscriptionStatus{
		models.SubscriptionActive,
		models.SubscriptionInactive,
		models.SubscriptionExpired,
		models.SubscriptionOnHold,
		models.SubscriptionPaused,
		models.SubscriptionCanceled,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// ValidateSelerType checks if a seller type is valid
func ValidateSelerType(selerType models.SelerType) bool {
	return selerType == models.SelerTypeSeler || selerType == models.SelerTypePromoter
}

// IsAdminRole checks if a role has admin privileges
func IsAdminRole(role models.UserRole) bool {
	return role == models.RoleAdmin || role == models.RoleSuperAdmin
}

// CanManageUsers checks if a role can manage other users
func CanManageUsers(role models.UserRole) bool {
	return role == models.RoleAdmin || role == models.RoleSuperAdmin || role == models.RoleFrontdesk
}

// GenerateUsername generates a username from first and last name
func GenerateUsername(firstName, lastName string) string {
	username := strings.ToLower(fmt.Sprintf("%s.%s", firstName, lastName))

	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9.]`)
	username = reg.ReplaceAllString(username, "")

	// Add random suffix to ensure uniqueness
	username = fmt.Sprintf("%s.%s", username, uuid.New().String()[:4])

	return username
}
