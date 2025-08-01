package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"generalusermanagement/config"
	"generalusermanagement/models"
	"log"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type Claims struct {
	UserID   string          `json:"user_id"`
	Email    string          `json:"email"`
	Role     models.UserRole `json:"role"`
	Username string          `json:"username"`
	ClientID string          `json:"client_id"`
	jwt.RegisteredClaims
}

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

func GenerateOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func GenerateClientID() string {
	return fmt.Sprintf("USR_%d_%s", time.Now().Unix(), uuid.New().String()[:8])
}

func GenerateSlug(text string) string {
	slug := strings.ToLower(text)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = fmt.Sprintf("user-%s", uuid.New().String()[:8])
	}

	return slug
}

func GenerateUniqueSlug(baseText string, userID string) string {
	baseSlug := GenerateSlug(baseText)

	if userID != "" {
		baseSlug = fmt.Sprintf("%s-%s", baseSlug, userID[:8])
	} else {
		baseSlug = fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:8])
	}

	return baseSlug
}

func GenerateUsername(firstName, lastName string) string {
	username := strings.ToLower(fmt.Sprintf("%s.%s", firstName, lastName))
	reg := regexp.MustCompile(`[^a-z0-9.]`)
	username = reg.ReplaceAllString(username, "")
	username = fmt.Sprintf("%s.%s", username, uuid.New().String()[:4])

	return username
}

func ValidateRole(role models.UserRole) bool {
	validRoles := []models.UserRole{
		models.RoleUser, models.RoleAdmin, models.RoleSuperAdmin,
	}

	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

func IsAdminRole(role models.UserRole) bool {
	return role == models.RoleAdmin || role == models.RoleSuperAdmin
}

func CanManageUsers(role models.UserRole) bool {
	return role == models.RoleAdmin || role == models.RoleSuperAdmin
}

func SendOTPEmail(email, otp, purpose string) error {
	log.Printf("📧 OTP Email would be sent to %s: %s (Email service not configured)", email, otp)
	return nil
}