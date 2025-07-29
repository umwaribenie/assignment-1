package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"generalusermanagement/config"
	"generalusermanagement/models"
	"math/big"
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
	UserID   string           `json:"user_id"`
	Email    string           `json:"email"`
	Role     models.UserRole  `json:"role"`
	Username string           `json:"username"`
	ClientID string           `json:"client_id"`
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