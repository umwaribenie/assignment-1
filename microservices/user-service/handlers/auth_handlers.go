package handlers

import (
	"database/sql"
	"net/http"
	"shared"
	"time"
	"user-service/cache"
	"user-service/database"
	"user-service/kafka"
	"user-service/models"
	"user-service/utils"

	"github.com/gin-gonic/gin"
)

// Login authenticates a user and returns JWT token
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Error: "Invalid request data"})
		return
	}

	// Try cache first
	cacheKey := cache.UserEmailCacheKey(req.Username)
	var cachedUser models.User
	var user models.User
	var err error

	if cacheErr := cache.GetCache(cacheKey, &cachedUser); cacheErr == nil {
		user = cachedUser
	} else {
		// Query database
		query := `SELECT id, client_id, email, first_name, last_name, password, phone, username, role, status, slug, created_at, updated_at FROM users WHERE (username = $1 OR email = $1) AND deleted_at IS NULL`
		
		err = database.DB.QueryRow(query, req.Username).Scan(
			&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName, &user.Password, &user.Phone, &user.Username, &user.Role, &user.Status, &user.Slug, &user.CreatedAt, &user.UpdatedAt)

		if err != nil {
			c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Error: "Invalid username or password"})
			return
		}

		// Cache user for 5 minutes
		cache.SetCache(cacheKey, user, 5*time.Minute)
	}

	if user.Status != shared.StatusActive {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Error: "Account is not active"})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Error: "Invalid username or password"})
		return
	}

	// Generate JWT token
	token, expiresAt, err := utils.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	userResponse := models.UserResponse{
		ID: user.ID, ClientID: user.ClientID, Email: user.Email, FirstName: user.FirstName,
		LastName: user.LastName, Phone: user.Phone, Username: user.Username, Role: user.Role,
		Status: user.Status, Slug: user.Slug, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}

	response := models.LoginResponse{
		AccessToken: token,
		User:        userResponse,
		ExpiresAt:   expiresAt,
	}

	c.JSON(http.StatusOK, shared.APIResponse{Success: true, Data: response})
}

// RequestPasswordReset requests a password reset OTP
func RequestPasswordReset(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Error: "Invalid request data"})
		return
	}

	// Check if user exists
	var userEmail string
	err := database.DB.QueryRow(`SELECT email FROM users WHERE (username = $1 OR email = $1) AND client_id = $2 AND deleted_at IS NULL`, req.Username, req.ClientID).Scan(&userEmail)
	if err != nil {
		// Don't reveal if user exists or not
		c.JSON(http.StatusOK, shared.SuccessResponse{Message: "If the user exists, a password reset OTP has been sent"})
		return
	}

	// Generate OTP
	otp, _ := utils.GenerateOTP()
	expiresAt := time.Now().Add(15 * time.Minute) // OTP expires in 15 minutes

	// Store OTP in database
	database.DB.Exec(`INSERT INTO password_resets (email, otp, expires_at, created_at) VALUES ($1, $2, $3, $4)`, userEmail, otp, expiresAt, time.Now())

	// Publish password reset event to Kafka (notification service will handle sending email)
	if err := kafka.PublishPasswordResetEvent(userEmail, otp); err != nil {
		// Log error but still return success to user
		c.JSON(http.StatusOK, shared.SuccessResponse{Message: "Password reset OTP has been sent to your email"})
		return
	}

	c.JSON(http.StatusOK, shared.SuccessResponse{Message: "Password reset OTP has been sent to your email"})
}

// ConfirmPasswordReset confirms password reset with OTP
func ConfirmPasswordReset(c *gin.Context) {
	var req models.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Error: "Invalid request data"})
		return
	}

	// Verify OTP
	var email string
	err := database.DB.QueryRow(`SELECT email FROM password_resets WHERE otp = $1 AND expires_at > NOW() AND used = FALSE ORDER BY created_at DESC LIMIT 1`, req.OTP).Scan(&email)
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Error: "Invalid or expired OTP"})
		return
	}

	// Hash new password
	hashedPassword, _ := utils.HashPassword(req.Password)

	// Update password
	database.DB.Exec(`UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE email = $2 AND deleted_at IS NULL`, hashedPassword, email)
	database.DB.Exec(`UPDATE password_resets SET used = TRUE WHERE otp = $1`, req.OTP)

	// Clear user cache
	cache.DeleteCache(cache.UserEmailCacheKey(email))

	c.JSON(http.StatusOK, shared.SuccessResponse{Message: "Password reset successful"})
}

// Logout handles user logout (adds token to blacklist)
func Logout(c *gin.Context) {
	token, err := utils.ExtractTokenFromHeader(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Error: "Invalid authorization header"})
		return
	}

	// Validate token to get expiration time
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, shared.ErrorResponse{Error: "Invalid token"})
		return
	}

	// Add token to blacklist
	database.DB.Exec(`INSERT INTO token_blacklist (token, expires_at, created_at) VALUES ($1, $2, $3)`,
		token, claims.ExpiresAt.Time, time.Now())

	c.JSON(http.StatusOK, shared.SuccessResponse{Message: "Logged out successfully"})
}