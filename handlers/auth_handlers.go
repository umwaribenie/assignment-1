package handlers

import (
	"database/sql"
	"generalusermanagement/database"
	"generalusermanagement/middleware"
	"generalusermanagement/models"
	"generalusermanagement/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Login godoc
// @Summary User login
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.APIResponse{data=models.LoginResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Find user by email
	var user models.User
	query := `SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
	          password, phone, profile_picture, username, role, status, created_at, updated_at, deleted_at 
	          FROM users WHERE email = $1 AND deleted_at IS NULL`
	
	row := database.DB.QueryRow(query, req.Email)
	err := row.Scan(&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Password, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	// Check if user is active
	if user.Status != models.StatusActive {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Account is not active",
		})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	// Generate JWT token
	token, expiresAt, err := utils.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate token",
			Error:   err.Error(),
		})
		return
	}

	// Convert user to response format
	userResponse := models.UserResponse{
		ID:             user.ID,
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       user.Username,
		Role:           user.Role,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	loginResponse := models.LoginResponse{
		Token:     token,
		User:      userResponse,
		ExpiresAt: expiresAt,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    loginResponse,
	})
}

// Register godoc
// @Summary User registration
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User data"
// @Success 201 {object} models.APIResponse{data=models.UserResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Check if user already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)`
	err := database.DB.QueryRow(checkQuery, req.Email, req.Username).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Message: "User with this email or username already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to hash password",
			Error:   err.Error(),
		})
		return
	}

	// Set default values
	if req.Role == "" {
		req.Role = models.RoleUser
	}
	if req.Status == "" {
		req.Status = models.StatusActive
	}

	// Generate client ID
	clientID := utils.GenerateClientID()

	// Insert user
	var user models.User
	insertQuery := `INSERT INTO users (client_id, email, first_name, last_name, national_id, passport_number, 
	                password, phone, username, role, status) 
	                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
	                RETURNING id, client_id, email, first_name, last_name, national_id, passport_number, 
	                phone, profile_picture, username, role, status, created_at, updated_at`

	err = database.DB.QueryRow(insertQuery, clientID, req.Email, req.FirstName, req.LastName,
		req.NationalID, req.PassportNumber, hashedPassword, req.Phone, req.Username, req.Role, req.Status).
		Scan(&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
			&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
			&user.Username, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	// Convert to response format
	userResponse := models.UserResponse{
		ID:             user.ID,
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       user.Username,
		Role:           user.Role,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    userResponse,
	})
}

// Logout godoc
// @Summary User logout
// @Description Logout user and blacklist token
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	token, _ := c.Get("token")
	tokenString := token.(string)

	// Get token expiration from JWT claims
	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid token",
			Error:   err.Error(),
		})
		return
	}

	// Blacklist the token
	err = middleware.BlacklistToken(tokenString, claims.ExpiresAt.Time)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to logout",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Logout successful",
	})
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Request a password reset OTP via email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.PasswordResetRequest true "Password reset request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /auth/password-reset/request [post]
func RequestPasswordReset(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Check if user exists
	var userExists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	err := database.DB.QueryRow(checkQuery, req.Email).Scan(&userExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	if !userExists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Generate OTP
	otp, err := utils.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate OTP",
			Error:   err.Error(),
		})
		return
	}

	// Store OTP in database (expires in 15 minutes)
	expiresAt := time.Now().Add(15 * time.Minute)
	insertQuery := `INSERT INTO password_resets (email, otp, expires_at) VALUES ($1, $2, $3)`
	_, err = database.DB.Exec(insertQuery, req.Email, otp, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to save OTP",
			Error:   err.Error(),
		})
		return
	}

	// TODO: Send OTP via email (implement email service)
	// For now, we'll just return success
	// In production, you would integrate with an email service like SendGrid, AWS SES, etc.

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password reset OTP sent to your email",
		Data: map[string]interface{}{
			"otp": otp, // Remove this in production
		},
	})
}

// ConfirmPasswordReset godoc
// @Summary Confirm password reset
// @Description Reset password using OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.PasswordResetConfirmRequest true "Password reset confirmation"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /auth/password-reset/confirm [post]
func ConfirmPasswordReset(c *gin.Context) {
	var req models.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Verify OTP
	var resetID int
	checkQuery := `SELECT id FROM password_resets 
	               WHERE email = $1 AND otp = $2 AND expires_at > NOW() AND used = FALSE`
	err := database.DB.QueryRow(checkQuery, req.Email, req.OTP).Scan(&resetID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Invalid or expired OTP",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to hash password",
			Error:   err.Error(),
		})
		return
	}

	// Update user password
	updateQuery := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE email = $2`
	_, err = database.DB.Exec(updateQuery, hashedPassword, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update password",
			Error:   err.Error(),
		})
		return
	}

	// Mark OTP as used
	markUsedQuery := `UPDATE password_resets SET used = TRUE WHERE id = $1`
	_, err = database.DB.Exec(markUsedQuery, resetID)
	if err != nil {
		// Log error but don't fail the request
		// The password was already updated successfully
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}

// CheckAuth godoc
// @Summary Check authentication status
// @Description Check if the current token is valid and return user info
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 401 {object} models.APIResponse
// @Router /auth/check [get]
func CheckAuth(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	// Get user from database
	var user models.User
	query := `SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
	          phone, profile_picture, username, role, status, created_at, updated_at 
	          FROM users WHERE id = $1 AND deleted_at IS NULL`
	
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	row := database.DB.QueryRow(query, userUUID)
	err = row.Scan(&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	// Convert to response format
	userResponse := models.UserResponse{
		ID:             user.ID,
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       user.Username,
		Role:           user.Role,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Authentication valid",
		Data:    userResponse,
	})
}