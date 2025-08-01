package handlers

import (
	"generalusermanagement/database"
	"generalusermanagement/models"
	"generalusermanagement/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary Login a user
// @Description Login a user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param loginData body models.UserLoginRequest true "Login Data"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/login [post]
func Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request data"})
		return
	}

	var user models.User
	query := `SELECT id, client_id, email, first_name, last_name, national_id, passport_number, password, phone, profile_picture, username, role, status, slug, created_at, updated_at FROM users WHERE (username = $1 OR email = $1) AND deleted_at IS NULL`

	err := database.DB.QueryRow(query, req.Username).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName, &user.NationalID, &user.PassportNumber, &user.Password, &user.Phone, &user.ProfilePicture, &user.Username, &user.Role, &user.Status, &user.Slug, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid username or password"})
		return
	}

	if user.Status != models.StatusActive {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Account is not active"})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid username or password"})
		return
	}

	if req.ClientID != user.ClientID {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid client ID"})
		return
	}

	// Generate JWT token
	token, expiresAt, err := utils.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	userResponse := models.UserResponse{
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Phone:          user.Phone,
		Username:       user.Username,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		ProfilePicture: user.ProfilePicture,
	}

	response := models.LoginResponse{
		AccessToken: token,
		User:        userResponse,
		ExpiresAt:   expiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Check Authentication
// @Description Verify if user is authenticated
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/check [get]
func CheckAuth(c *gin.Context) {
	c.JSON(http.StatusOK, models.SuccessResponse{Message: "User is authenticated"})
}

// @Summary Request Password Reset
// @Description Request password reset OTP
// @Tags Authentication
// @Accept json
// @Produce json
// @Param passwordResetRequest body models.PasswordResetRequest true "Password reset request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/password-reset [post]
func RequestPasswordReset(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request data"})
		return
	}

	var userEmail string
	err := database.DB.QueryRow(`SELECT email FROM users WHERE (username = $1 OR email = $1) AND client_id = $2 AND deleted_at IS NULL`, req.Username, req.ClientID).Scan(&userEmail)
	if err != nil {
		c.JSON(http.StatusOK, models.SuccessResponse{Message: "If the user exists, a password reset OTP has been sent"})
		return
	}

	otp, _ := utils.GenerateOTP()
	expiresAt := time.Now().Add(15 * time.Minute)
	database.DB.Exec(`INSERT INTO password_resets (email, otp, expires_at, created_at) VALUES ($1, $2, $3, $4)`, userEmail, otp, expiresAt, time.Now())

	// Send OTP email
	go utils.SendOTPEmail(userEmail, otp, "password_reset")

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Password reset OTP has been sent to your email"})
}

// @Summary Confirm Password Reset
// @Description Confirm password reset with OTP
// @Tags Authentication
// @Accept json
// @Produce json
// @Param passwordResetConfirmRequest body models.PasswordResetConfirmRequest true "Password reset confirmation"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/confirm-password-reset-otp [post]
func ConfirmPasswordReset(c *gin.Context) {
	var req models.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request data"})
		return
	}

	var email string
	err := database.DB.QueryRow(`SELECT email FROM password_resets WHERE otp = $1 AND expires_at > NOW() AND used = FALSE ORDER BY created_at DESC LIMIT 1`, req.OTP).Scan(&email)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid or expired OTP"})
		return
	}

	hashedPassword, _ := utils.HashPassword(req.Password)
	database.DB.Exec(`UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE email = $2 AND deleted_at IS NULL`, hashedPassword, email)
	database.DB.Exec(`UPDATE password_resets SET used = TRUE WHERE otp = $1`, req.OTP)

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Password reset successfully"})
}

// @Summary User Logout
// @Description Logout user and blacklist token
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /users/logout [post]
func Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	tokenString, err := utils.ExtractTokenFromHeader(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Invalid authorization header"})
		return
	}

	claims, err := utils.ValidateJWT(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Invalid token"})
		return
	}

	expiresAt := claims.ExpiresAt.Time
	database.DB.Exec(`INSERT INTO token_blacklist (token, expires_at, created_at) VALUES ($1, $2, $3)`, tokenString, expiresAt, time.Now())

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Logged out successfully"})
}

// @Summary Reset password via email
// @Description Reset password via email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param resetPassword body models.ResetPasswordEmailRequest true "Reset Password Request"
// @Success 200 {object} models.ResetPasswordEmailResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /reset-password/email [post]
func ResetPasswordWithEmail(c *gin.Context) {
	var req models.ResetPasswordEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request data"})
		return
	}

	var userEmail string
	err := database.DB.QueryRow(`SELECT email FROM users WHERE (username = $1 OR email = $1) AND client_id = $2 AND deleted_at IS NULL`, req.Username, req.ClientID).Scan(&userEmail)
	if err != nil {
		c.JSON(http.StatusOK, models.ResetPasswordEmailResponse{Message: "If the user exists, a password reset email has been sent"})
		return
	}

	otp, _ := utils.GenerateOTP()
	expiresAt := time.Now().Add(15 * time.Minute)
	database.DB.Exec(`INSERT INTO password_resets (email, otp, expires_at, created_at) VALUES ($1, $2, $3, $4)`, userEmail, otp, expiresAt, time.Now())

	go utils.SendOTPEmail(userEmail, otp, "password_reset")

	c.JSON(http.StatusOK, models.ResetPasswordEmailResponse{Message: "Password reset email has been sent"})
}

// @Summary Update password
// @Description Update password
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param updatePassword body models.AuthUpdatePasswordRequest true "Update Password Request"
// @Success 200 {object} models.AuthUpdatePasswordResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /update-password [post]
func AuthUpdatePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req models.AuthUpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request data"})
		return
	}

	var currentPasswordHash string
	database.DB.QueryRow(`SELECT password FROM users WHERE id = $1`, userID).Scan(&currentPasswordHash)

	if !utils.CheckPasswordHash(req.OldPassword, currentPasswordHash) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Current password is incorrect"})
		return
	}

	newPasswordHash, _ := utils.HashPassword(req.NewPassword)
	database.DB.Exec(`UPDATE users SET password = $1 WHERE id = $2`, newPasswordHash, userID)

	c.JSON(http.StatusOK, models.AuthUpdatePasswordResponse{Message: "Password updated successfully"})
}