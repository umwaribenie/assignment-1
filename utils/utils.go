package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"generalusermanagement/config"
	"generalusermanagement/models"
)

// Email related functions
func SendEmail(to, subject, body string) error {
	cfg := config.AppConfig
	
	if cfg.SMTPUser == "" || cfg.SMTPPassword == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", to, subject, body))
	
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	return smtp.SendMail(addr, auth, cfg.SMTPFrom, []string{to}, msg)
}



// String utility functions
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[randomIndex.Int64()]
	}
	return string(b), nil
}

// Validation functions
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func IsValidPhoneNumber(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

func IsValidPassword(password string) bool {
	if len(password) < 6 {
		return false
	}
	
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	
	return hasUpper && hasLower && hasNumber
}

// Formatting functions
func FormatPhoneNumber(phone string) string {
	// Remove all non-digit characters except +
	re := regexp.MustCompile(`[^\d+]`)
	cleaned := re.ReplaceAllString(phone, "")
	
	// If it doesn't start with +, add country code (assuming +1 for US)
	if !strings.HasPrefix(cleaned, "+") {
		cleaned = "+1" + cleaned
	}
	
	return cleaned
}

func SanitizeString(input string) string {
	// Remove potentially dangerous characters
	re := regexp.MustCompile(`[<>\"'&]`)
	return re.ReplaceAllString(strings.TrimSpace(input), "")
}

// Time utilities
func GetCurrentTimestamp() time.Time {
	return time.Now().UTC()
}

func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// File utilities
func GetFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}

func IsValidImageExtension(ext string) bool {
	validExts := []string{"jpg", "jpeg", "png", "gif", "webp"}
	ext = strings.ToLower(ext)
	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// User-related utilities
func GetUserDisplayName(user models.User) string {
	if user.FirstName != "" && user.LastName != "" {
		return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	if user.FirstName != "" {
		return user.FirstName
	}
	return user.Username
}

// Database utilities
func BuildUpdateQuery(tableName string, updates map[string]interface{}, whereClause string) (string, []interface{}) {
	if len(updates) == 0 {
		return "", nil
	}
	
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates))
	argIndex := 1
	
	for column, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", column, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	query := fmt.Sprintf("UPDATE %s SET %s, updated_at = CURRENT_TIMESTAMP WHERE %s", 
		tableName, strings.Join(setParts, ", "), whereClause)
	
	return query, args
}

// Error handling utilities
func HandleDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	
	errStr := err.Error()
	
	// Check for common PostgreSQL errors
	if strings.Contains(errStr, "duplicate key value violates unique constraint") {
		if strings.Contains(errStr, "email") {
			return fmt.Errorf("email already exists")
		}
		if strings.Contains(errStr, "username") {
			return fmt.Errorf("username already exists")
		}
		if strings.Contains(errStr, "client_id") {
			return fmt.Errorf("client ID already exists")
		}
		return fmt.Errorf("duplicate entry")
	}
	
	if strings.Contains(errStr, "no rows in result set") {
		return fmt.Errorf("record not found")
	}
	
	return err
}

// Pagination utilities
func CalculateOffset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}

func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 1
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

// Response utilities
func BuildSuccessResponse(message string, data interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"success": true,
		"message": message,
	}
	
	if data != nil {
		response["data"] = data
	}
	
	return response
}

func BuildErrorResponse(message string, err error) map[string]interface{} {
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}
	
	if err != nil {
		response["error"] = err.Error()
	}
	
	return response
}