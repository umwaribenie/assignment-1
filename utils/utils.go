package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"generalusermanagement/config"
	"generalusermanagement/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// CheckPasswordHash compares password with hash (backward compatibility)
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Claims represents the JWT claims
type Claims struct {
	UserID   string           `json:"user_id"`
	Email    string           `json:"email"`
	Role     models.UserRole  `json:"role"`
	Username string           `json:"username"`
	ClientID string           `json:"client_id"`
	jwt.RegisteredClaims
}

// GenerateOTP generates a random 6-digit OTP
func GenerateOTP() string {
	otp := ""
	for i := 0; i < 6; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		otp += fmt.Sprintf("%d", num)
	}
	return otp
}

// GenerateClientID generates a unique client ID
func GenerateClientID() string {
	return fmt.Sprintf("USR_%d_%s", time.Now().Unix(), uuid.New().String()[:8])
}

// GenerateSlug creates a URL-friendly slug from a string
func GenerateSlug(text string) string {
	// Convert to lowercase
	slug := strings.ToLower(text)
	
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	
	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")
	
	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Unix()
	slug = fmt.Sprintf("%s-%d", slug, timestamp)
	
	return slug
}

// GenerateUniqueSlug generates a unique slug with a random suffix
func GenerateUniqueSlug(text string) string {
	baseSlug := GenerateSlug(text)
	
	// Add random suffix for uniqueness
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("%s-%d", baseSlug, randomNum)
}

// GenerateReferralCode generates a random referral code
func GenerateReferralCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := ""
	for i := 0; i < 8; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code += string(charset[num.Int64()])
	}
	return code
}

// IsValidEmail checks if email format is valid
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// SanitizeString removes potentially harmful characters from strings
func SanitizeString(input string) string {
	// Remove HTML tags and scripts
	reg := regexp.MustCompile(`<[^>]*>`)
	sanitized := reg.ReplaceAllString(input, "")
	
	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)
	
	return sanitized
}

// ParseDateFilter parses date filter strings (YYYY-MM-DD format)
func ParseDateFilter(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}
	
	parsedTime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	
	return &parsedTime, nil
}

// CalculatePagination calculates pagination values
func CalculatePagination(page, pageSize, total int) (int, *int, *int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	
	totalPages := (total + pageSize - 1) / pageSize
	
	var nextPage, prevPage *int
	
	if page < totalPages {
		next := page + 1
		nextPage = &next
	}
	
	if page > 1 {
		prev := page - 1
		prevPage = &prev
	}
	
	return totalPages, nextPage, prevPage
}

// JWT Functions

// GetJWTSecret returns the JWT secret from config
func GetJWTSecret() string {
	return config.AppConfig.JWTSecret
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(userID, role string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(config.AppConfig.JWTExpiryHours) * time.Hour)
	
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns claims
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// Legacy JWT functions for backward compatibility

// GenerateJWTLegacy generates JWT token in legacy format
func GenerateJWTLegacy(user models.User) (string, time.Time, error) {
	expirationTime := time.Now().Add(time.Duration(config.AppConfig.JWTExpiryHours) * time.Hour)
	
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    string(user.Role),
		"exp":     expirationTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}