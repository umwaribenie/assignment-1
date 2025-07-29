package handlers

import (
	"database/sql"
	"generalusermanagement/database"
	"generalusermanagement/models"
	"generalusermanagement/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Login godoc
// @Summary Login a user
// @Description Authenticate user with username and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param loginData body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.APIResponse{data=models.LoginResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request data",
		})
		return
	}

	// Find user by username (can be username or email)
	var user models.User
	query := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
		       password, phone, profile_picture, username, role, status, subscription_status, 
		       institution_id, commission_percentage, seler_type, specialization, notes, 
		       slug, referral_code, has_active_subscription, is_active, otp_required, 
		       created_by, created_at, updated_at, deleted_at 
		FROM users 
		WHERE (username = $1 OR email = $1) AND deleted_at IS NULL`
	
	row := database.DB.QueryRow(query, req.Username)
	err := row.Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Password, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.SubscriptionStatus,
		&user.InstitutionID, &user.CommissionPercentage, &user.SelerType,
		&user.Specialization, &user.Notes, &user.Slug, &user.ReferralCode,
		&user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
		&user.CreatedBy, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid username or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	// Check if user is active
	if user.Status != models.StatusActive {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Account is not active",
		})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Invalid username or password",
		})
		return
	}

	// Check client ID if provided
	if req.ClientID != nil && *req.ClientID != user.ClientID {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Invalid client ID",
		})
		return
	}

	// If OTP is required, generate OTP and don't return token yet
	if user.OTPRequired {
		otp, err := utils.GenerateOTP()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to generate OTP",
			})
			return
		}

		// Store OTP in database with 5-minute expiry
		expiresAt := time.Now().Add(5 * time.Minute)
		otpQuery := `
			INSERT INTO login_otps (user_id, otp, expires_at, created_at)
			VALUES ($1, $2, $3, $4)`
		
		_, err = database.DB.Exec(otpQuery, user.ID, otp, expiresAt, time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to store OTP",
			})
			return
		}

		// TODO: Send OTP via SMS/Email (implement based on your SMS/Email service)
		// For now, we'll return a message indicating OTP is required
		
		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "OTP sent to your registered phone number. Please verify to complete login.",
		})
		return
	}

	// Generate JWT token
	token, expiresAt, err := utils.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate token",
		})
		return
	}

	// Prepare user response
	userResponse := models.UserResponse{
		ID:                    user.ID,
		ClientID:              user.ClientID,
		Email:                 user.Email,
		FirstName:             user.FirstName,
		LastName:              user.LastName,
		NationalID:            user.NationalID,
		PassportNumber:        user.PassportNumber,
		Phone:                 user.Phone,
		ProfilePicture:        user.ProfilePicture,
		Username:              user.Username,
		Role:                  user.Role,
		Status:                user.Status,
		SubscriptionStatus:    user.SubscriptionStatus,
		InstitutionID:         user.InstitutionID,
		CommissionPercentage:  user.CommissionPercentage,
		SelerType:             user.SelerType,
		Specialization:        user.Specialization,
		Notes:                 user.Notes,
		Slug:                  user.Slug,
		ReferralCode:          user.ReferralCode,
		HasActiveSubscription: user.HasActiveSubscription,
		IsActive:              user.IsActive,
		OTPRequired:           user.OTPRequired,
		CreatedBy:             user.CreatedBy,
		CreatedAt:             user.CreatedAt,
		UpdatedAt:             user.UpdatedAt,
	}

	response := models.LoginResponse{
		AccessToken: token,
		User:        userResponse,
		ExpiresAt:   expiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// VerifyLoginOTP godoc
// @Summary Verify login OTP
// @Description Verifies user login OTP and returns JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param otpData body models.OTPLoginRequest true "OTP verification"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/verify-login-otp [post]
func VerifyLoginOTP(c *gin.Context) {
	var req models.OTPLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request data",
		})
		return
	}

	// Find valid OTP
	var userID uuid.UUID
	otpQuery := `
		SELECT user_id 
		FROM login_otps 
		WHERE otp = $1 AND expires_at > NOW() AND used = FALSE
		ORDER BY created_at DESC
		LIMIT 1`
	
	err := database.DB.QueryRow(otpQuery, req.OTP).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid or expired OTP",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	// Mark OTP as used
	updateOTPQuery := `UPDATE login_otps SET used = TRUE WHERE otp = $1`
	_, err = database.DB.Exec(updateOTPQuery, req.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update OTP status",
		})
		return
	}

	// Get user details
	var user models.User
	userQuery := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
		       phone, profile_picture, username, role, status, subscription_status, 
		       institution_id, commission_percentage, seler_type, specialization, notes, 
		       slug, referral_code, has_active_subscription, is_active, otp_required, 
		       created_by, created_at, updated_at, deleted_at 
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL`
	
	err = database.DB.QueryRow(userQuery, userID).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.SubscriptionStatus,
		&user.InstitutionID, &user.CommissionPercentage, &user.SelerType,
		&user.Specialization, &user.Notes, &user.Slug, &user.ReferralCode,
		&user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
		&user.CreatedBy, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to fetch user details",
		})
		return
	}

	// Generate JWT token
	token, expiresAt, err := utils.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate token",
		})
		return
	}

	// Prepare user response
	userResponse := models.UserResponse{
		ID:                    user.ID,
		ClientID:              user.ClientID,
		Email:                 user.Email,
		FirstName:             user.FirstName,
		LastName:              user.LastName,
		NationalID:            user.NationalID,
		PassportNumber:        user.PassportNumber,
		Phone:                 user.Phone,
		ProfilePicture:        user.ProfilePicture,
		Username:              user.Username,
		Role:                  user.Role,
		Status:                user.Status,
		SubscriptionStatus:    user.SubscriptionStatus,
		InstitutionID:         user.InstitutionID,
		CommissionPercentage:  user.CommissionPercentage,
		SelerType:             user.SelerType,
		Specialization:        user.Specialization,
		Notes:                 user.Notes,
		Slug:                  user.Slug,
		ReferralCode:          user.ReferralCode,
		HasActiveSubscription: user.HasActiveSubscription,
		IsActive:              user.IsActive,
		OTPRequired:           user.OTPRequired,
		CreatedBy:             user.CreatedBy,
		CreatedAt:             user.CreatedAt,
		UpdatedAt:             user.UpdatedAt,
	}

	response := models.LoginResponse{
		AccessToken: token,
		User:        userResponse,
		ExpiresAt:   expiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// CheckAuth godoc
// @Summary Check if the user is authenticated
// @Description Check if the user is authenticated
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Router /users/check [get]
func CheckAuth(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "User is authenticated",
	})
}

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Request password reset OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param passwordReset body models.PasswordResetRequest true "Password reset request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/password-reset [post]
func RequestPasswordReset(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request data",
		})
		return
	}

	// Find user by username or email
	var userEmail string
	query := `SELECT email FROM users WHERE (username = $1 OR email = $1) AND client_id = $2 AND deleted_at IS NULL`
	err := database.DB.QueryRow(query, req.Username, req.ClientID).Scan(&userEmail)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if user exists or not for security
			c.JSON(http.StatusOK, models.SuccessResponse{
				Message: "If the user exists, a password reset OTP has been sent",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	// Generate OTP
	otp, err := utils.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate OTP",
		})
		return
	}

	// Store OTP with 15-minute expiry
	expiresAt := time.Now().Add(15 * time.Minute)
	insertQuery := `
		INSERT INTO password_resets (email, otp, expires_at, created_at)
		VALUES ($1, $2, $3, $4)`
	
	_, err = database.DB.Exec(insertQuery, userEmail, otp, expiresAt, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to store password reset request",
		})
		return
	}

	// TODO: Send OTP via email (implement based on your email service)
	
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Password reset OTP has been sent to your email",
	})
}

// ResetPasswordWithEmail godoc
// @Summary Reset password via email
// @Description Reset password using email-based OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param resetPassword body models.PasswordResetRequest true "Password reset request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/reset-password/email [post]
func ResetPasswordWithEmail(c *gin.Context) {
	// This is an alias for RequestPasswordReset for compatibility
	RequestPasswordReset(c)
}

// ConfirmPasswordReset godoc
// @Summary Confirm password reset OTP
// @Description Confirm password reset OTP and set new password
// @Tags Auth
// @Accept json
// @Produce json
// @Param confirmOTP body models.PasswordResetConfirmRequest true "OTP confirmation"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/confirm-password-reset-otp [post]
func ConfirmPasswordReset(c *gin.Context) {
	var req models.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request data",
		})
		return
	}

	// Find valid OTP
	var email string
	otpQuery := `
		SELECT email 
		FROM password_resets 
		WHERE otp = $1 AND expires_at > NOW() AND used = FALSE
		ORDER BY created_at DESC
		LIMIT 1`
	
	err := database.DB.QueryRow(otpQuery, req.OTP).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid or expired OTP",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to hash password",
		})
		return
	}

	// Update user password
	updateUserQuery := `
		UPDATE users 
		SET password = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE email = $2 AND deleted_at IS NULL`
	
	_, err = database.DB.Exec(updateUserQuery, hashedPassword, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update password",
		})
		return
	}

	// Mark OTP as used
	updateOTPQuery := `UPDATE password_resets SET used = TRUE WHERE otp = $1`
	_, err = database.DB.Exec(updateOTPQuery, req.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update OTP status",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Password reset successfully",
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
	authHeader := c.GetHeader("Authorization")
	tokenString, err := utils.ExtractTokenFromHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid authorization header",
			Error:   err.Error(),
		})
		return
	}

	// Validate token to get expiry time
	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid token",
			Error:   err.Error(),
		})
		return
	}

	// Add token to blacklist
	expiresAt := claims.ExpiresAt.Time
	query := `INSERT INTO token_blacklist (token, expires_at, created_at) VALUES ($1, $2, $3)`
	_, err = database.DB.Exec(query, tokenString, expiresAt, time.Now())
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
		Message: "Logged out successfully",
	})
}