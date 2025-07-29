package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"generalusermanagement/database"
	"generalusermanagement/models"
	"generalusermanagement/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// @Summary Login a user
// @Description Login a user
// @Tags auth
// @Accept json
// @Produce json
// @Param loginData body models.LoginData true "Login Data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /login [post]
func Login(c *gin.Context) {
	var loginData models.LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Query user by username or email
	var user models.User
	query := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
			   password, phone, profile_picture, username, role, status, slug, notes,
			   institution_id, subscription_status, has_active_subscription, is_active,
			   otp_required, referral_code, created_by, created_at, updated_at, deleted_at
		FROM users 
		WHERE (username = $1 OR email = $1) AND deleted_at IS NULL
	`

	err := database.DB.QueryRow(query, loginData.Username).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Password, &user.Phone,
		&user.ProfilePicture, &user.Username, &user.Role, &user.Status, &user.Slug,
		&user.Notes, &user.InstitutionID, &user.SubscriptionStatus,
		&user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
		&user.ReferralCode, &user.CreatedBy, &user.CreatedAt, &user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid credentials",
			})
			return
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	// Check password
	if err := utils.CheckPassword(loginData.Password, user.Password); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Invalid credentials",
		})
		return
	}

	// Check if user is active
	if user.Status != models.StatusActive || !user.IsActive {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Account is not active",
		})
		return
	}

	// Check if OTP is required
	if user.OTPRequired {
		// Generate and store OTP
		otp := utils.GenerateOTP()
		expiresAt := time.Now().Add(5 * time.Minute)

		// Store OTP in database
		_, err = database.DB.Exec(`
			INSERT INTO login_otps (user_id, otp, expires_at) 
			VALUES ($1, $2, $3)
		`, user.ID, otp, expiresAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to generate OTP",
			})
			return
		}

		// In a real application, you would send the OTP via SMS/Email
		// For demo purposes, we'll return a success message
		c.JSON(http.StatusOK, models.SuccessResponse{
			Message: "OTP sent successfully. Please verify to complete login.",
		})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.String(), string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate token",
		})
		return
	}

	// Create user response
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
		Slug:                  user.Slug,
		Notes:                 user.Notes,
		InstitutionID:         user.InstitutionID,
		SubscriptionStatus:    user.SubscriptionStatus,
		HasActiveSubscription: user.HasActiveSubscription,
		IsActive:             user.IsActive,
		OTPRequired:          user.OTPRequired,
		ReferralCode:         user.ReferralCode,
		CreatedBy:            user.CreatedBy,
		CreatedAt:            user.CreatedAt,
		UpdatedAt:            user.UpdatedAt,
	}

	c.JSON(http.StatusOK, models.LoginSuccessResponse{
		AccessToken: token,
		User:        userResponse,
	})
}

// @Summary Verify login OTP
// @Description Verifies user login OTP and returns JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param otpData body models.ConfirmLoginOTPRequest true "OTP verification"
// @Success 200 {object} models.LoginSuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /verify-login-otp [post]
func VerifyLoginOTP(c *gin.Context) {
	var otpRequest models.ConfirmLoginOTPRequest
	if err := c.ShouldBindJSON(&otpRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Find valid OTP
	var userID uuid.UUID
	var expiresAt time.Time
	query := `
		SELECT user_id, expires_at 
		FROM login_otps 
		WHERE otp = $1 AND used = false AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := database.DB.QueryRow(query, otpRequest.OTP).Scan(&userID, &expiresAt)
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
	_, err = database.DB.Exec(`
		UPDATE login_otps 
		SET used = true 
		WHERE otp = $1 AND user_id = $2
	`, otpRequest.OTP, userID)

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
			   password, phone, profile_picture, username, role, status, slug, notes,
			   institution_id, subscription_status, has_active_subscription, is_active,
			   otp_required, referral_code, created_by, created_at, updated_at, deleted_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL
	`

	err = database.DB.QueryRow(userQuery, userID).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Password, &user.Phone,
		&user.ProfilePicture, &user.Username, &user.Role, &user.Status, &user.Slug,
		&user.Notes, &user.InstitutionID, &user.SubscriptionStatus,
		&user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
		&user.ReferralCode, &user.CreatedBy, &user.CreatedAt, &user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to get user details",
		})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.String(), string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate token",
		})
		return
	}

	// Create user response
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
		Slug:                  user.Slug,
		Notes:                 user.Notes,
		InstitutionID:         user.InstitutionID,
		SubscriptionStatus:    user.SubscriptionStatus,
		HasActiveSubscription: user.HasActiveSubscription,
		IsActive:             user.IsActive,
		OTPRequired:          user.OTPRequired,
		ReferralCode:         user.ReferralCode,
		CreatedBy:            user.CreatedBy,
		CreatedAt:            user.CreatedAt,
		UpdatedAt:            user.UpdatedAt,
	}

	c.JSON(http.StatusOK, models.LoginSuccessResponse{
		AccessToken: token,
		User:        userResponse,
	})
}

// @Summary Request password reset
// @Description Request password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param passwordReset body models.PasswordResetRequest true "Password Reset Request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /password-reset [post]
func RequestPasswordReset(c *gin.Context) {
	var request models.PasswordResetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Check if user exists
	var userEmail string
	query := `
		SELECT email FROM users 
		WHERE (username = $1 OR email = $1) AND client_id = $2 AND deleted_at IS NULL
	`
	err := database.DB.QueryRow(query, request.Username, request.ClientID).Scan(&userEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if user exists or not
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
	otp := utils.GenerateOTP()
	expiresAt := time.Now().Add(15 * time.Minute)

	// Store OTP
	_, err = database.DB.Exec(`
		INSERT INTO password_resets (email, otp, expires_at) 
		VALUES ($1, $2, $3)
	`, userEmail, otp, expiresAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to generate password reset OTP",
		})
		return
	}

	// In a real application, send OTP via email/SMS
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Password reset OTP sent successfully",
	})
}

// @Summary Reset password via email
// @Description Reset password via email
// @Tags auth
// @Accept json
// @Produce json
// @Param resetPassword body models.PasswordResetRequest true "Reset Password Request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /reset-password/email [post]
func ResetPasswordEmail(c *gin.Context) {
	// This is the same as RequestPasswordReset for now
	RequestPasswordReset(c)
}

// @Summary Confirm password reset OTP
// @Description Confirm password reset OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param confirmOTP body models.ConfirmOTPRequest true "Confirm OTP Request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /confirm-password-reset-otp [post]
func ConfirmPasswordReset(c *gin.Context) {
	var request models.ConfirmOTPRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Find valid OTP
	var email string
	var expiresAt time.Time
	query := `
		SELECT email, expires_at 
		FROM password_resets 
		WHERE otp = $1 AND used = false AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := database.DB.QueryRow(query, request.OTP).Scan(&email, &expiresAt)
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
	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to hash password",
		})
		return
	}

	// Update user password
	_, err = database.DB.Exec(`
		UPDATE users 
		SET password = $1, updated_at = NOW() 
		WHERE email = $2
	`, hashedPassword, email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update password",
		})
		return
	}

	// Mark OTP as used
	_, err = database.DB.Exec(`
		UPDATE password_resets 
		SET used = true 
		WHERE otp = $1 AND email = $2
	`, request.OTP, email)

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

// @Summary Update password
// @Description Update password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param updatePassword body models.UpdatePasswordRequest true "Update Password Request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /update-password [post]
func UpdatePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	var request models.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Get current password
	var currentPassword string
	err := database.DB.QueryRow(`
		SELECT password FROM users WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&currentPassword)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to get user information",
		})
		return
	}

	// Verify old password
	if err := utils.CheckPassword(request.OldPassword, currentPassword); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Current password is incorrect",
		})
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to hash password",
		})
		return
	}

	// Update password
	_, err = database.DB.Exec(`
		UPDATE users 
		SET password = $1, updated_at = NOW() 
		WHERE id = $2
	`, hashedPassword, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Password updated successfully",
	})
}

// @Summary Check if the user is authenticated
// @Description Check if the user is authenticated
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Router /check [get]
func CheckAuth(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Authenticated",
	})
}

// @Summary User logout
// @Description User logout (requires authentication)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /logout [post]
func Logout(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "No token provided",
		})
		return
	}

	// Remove "Bearer " prefix
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Parse token to get expiration time
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(utils.GetJWTSecret()), nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid token",
		})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid token claims",
		})
		return
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid token expiration",
		})
		return
	}

	expirationTime := time.Unix(int64(exp), 0)

	// Add token to blacklist
	_, err = database.DB.Exec(`
		INSERT INTO token_blacklist (token, expires_at) 
		VALUES ($1, $2)
	`, tokenString, expirationTime)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to logout",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Logged out successfully",
	})
}