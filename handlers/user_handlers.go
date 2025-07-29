package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"generalusermanagement/database"
	"generalusermanagement/models"
	"generalusermanagement/utils"

	"github.com/gin-gonic/gin"
)

// @Summary Register a new user
// @Description Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.UserDto true "User"
// @Success 200 {object} models.UserDto
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /register [post]
func Register(c *gin.Context) {
	var userDto models.UserDto
	if err := c.ShouldBindJSON(&userDto); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Check if user already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)`
	username := userDto.Username
	if username == nil {
		// Generate username from email if not provided
		emailParts := strings.Split(userDto.Email, "@")
		generatedUsername := emailParts[0]
		username = &generatedUsername
	}

	err := database.DB.QueryRow(checkQuery, userDto.Email, *username).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "User with this email or username already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(userDto.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to hash password",
		})
		return
	}

	// Generate slug and other fields
	slug := utils.GenerateUniqueSlug(userDto.FirstName + " " + userDto.LastName)
	referralCode := utils.GenerateReferralCode()

	// Insert user
	var user models.User
	insertQuery := `
		INSERT INTO users (client_id, email, first_name, last_name, national_id, passport_number, 
		                   password, phone, profile_picture, username, role, status, slug, 
		                   has_active_subscription, is_active, otp_required, referral_code) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) 
		RETURNING id, client_id, email, first_name, last_name, national_id, passport_number, 
		          phone, profile_picture, username, role, status, slug, notes, institution_id,
		          subscription_status, has_active_subscription, is_active, otp_required, 
		          referral_code, created_by, created_at, updated_at`

	err = database.DB.QueryRow(insertQuery, 
		userDto.ClientID, userDto.Email, userDto.FirstName, userDto.LastName,
		userDto.NationalID, userDto.PassportNumber, hashedPassword, userDto.Phone,
		userDto.ProfilePicture, *username, models.RoleUser, models.StatusActive, 
		slug, false, true, false, referralCode).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.Slug, &user.Notes,
		&user.InstitutionID, &user.SubscriptionStatus, &user.HasActiveSubscription,
		&user.IsActive, &user.OTPRequired, &user.ReferralCode, &user.CreatedBy,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create user",
		})
		return
	}

	// Return user data (matching UserDto structure)
	responseDto := models.UserDto{
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       &user.Username,
	}

	c.JSON(http.StatusOK, responseDto)
}

// @Summary Register a new user by admin
// @Description Register a new user by admin
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.CreateUserByAdminDto true "User"
// @Success 200 {object} models.CreateUserByAdminDto
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /registerusersbyadmin [post]
func RegisterUserByAdmin(c *gin.Context) {
	// Check if user is admin
	userRole := c.GetString("user_role")
	if userRole != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied. Admin role required",
		})
		return
	}

	var createDto models.CreateUserByAdminDto
	if err := c.ShouldBindJSON(&createDto); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Check if user already exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := database.DB.QueryRow(checkQuery, createDto.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "User with this email already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(createDto.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to hash password",
		})
		return
	}

	// Generate required fields
	clientID := utils.GenerateClientID()
	slug := utils.GenerateUniqueSlug(createDto.FirstName + " " + createDto.LastName)
	referralCode := utils.GenerateReferralCode()
	username := createDto.Username
	if username == nil {
		// Generate username from email if not provided
		emailParts := strings.Split(createDto.Email, "@")
		generatedUsername := emailParts[0]
		username = &generatedUsername
	}

	// Get current user ID for created_by field
	createdBy := c.GetString("user_id")

	// Insert user
	var user models.User
	insertQuery := `
		INSERT INTO users (client_id, email, first_name, last_name, national_id, passport_number, 
		                   password, phone, profile_picture, username, role, status, slug, 
		                   institution_id, has_active_subscription, is_active, otp_required, 
		                   referral_code, created_by) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19) 
		RETURNING id, client_id, email, first_name, last_name, national_id, passport_number, 
		          phone, profile_picture, username, role, status, slug, notes, institution_id,
		          subscription_status, has_active_subscription, is_active, otp_required, 
		          referral_code, created_by, created_at, updated_at`

	err = database.DB.QueryRow(insertQuery,
		clientID, createDto.Email, createDto.FirstName, createDto.LastName,
		createDto.NationalID, createDto.PassportNumber, hashedPassword, createDto.Phone,
		createDto.ProfilePicture, *username, createDto.Role, models.StatusActive,
		slug, createDto.InstitutionID, false, true, false, referralCode, createdBy).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.Slug, &user.Notes,
		&user.InstitutionID, &user.SubscriptionStatus, &user.HasActiveSubscription,
		&user.IsActive, &user.OTPRequired, &user.ReferralCode, &user.CreatedBy,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create user",
		})
		return
	}

	// Return created user data
	responseDto := models.CreateUserByAdminDto{
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Role:           user.Role,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       &user.Username,
		InstitutionID:  user.InstitutionID,
	}

	c.JSON(http.StatusOK, responseDto)
}

// @Summary Get all users
// @Description Get all users
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param pageNumber query int false "Page Number"
// @Param pageSize query int false "Page Size"
// @Param from query string false "From Date"
// @Param to query string false "To Date"
// @Param search query string false "Search"
// @Param role query string false "Role" Enums(user, admin)
// @Param status query string false "Status" Enums(active, inactive, deleted)
// @Param subscriptionStatus query string false "Subscription Status" Enums(active, inactive, expired, onhold, paused, canceled)
// @Param institutionId query string false "Institution ID"
// @Success 200 {object} models.PaginationResponse
// @Failure 500 {object} models.ErrorResponse
// @Router / [get]
func GetAllUsers(c *gin.Context) {
	// Parse query parameters
	pageNumber, _ := strconv.Atoi(c.DefaultQuery("pageNumber", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")
	role := c.Query("role")
	status := c.Query("status")
	subscriptionStatus := c.Query("subscriptionStatus")
	institutionId := c.Query("institutionId")
	from := c.Query("from")
	to := c.Query("to")

	// Build query
	baseQuery := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
		       phone, profile_picture, username, role, status, slug, notes, institution_id,
		       subscription_status, has_active_subscription, is_active, otp_required, 
		       referral_code, created_by, created_at, updated_at
		FROM users 
		WHERE deleted_at IS NULL
	`

	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add filters
	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d OR username ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if role != "" {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, role)
		argIndex++
	}

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if subscriptionStatus != "" {
		conditions = append(conditions, fmt.Sprintf("subscription_status = $%d", argIndex))
		args = append(args, subscriptionStatus)
		argIndex++
	}

	if institutionId != "" {
		conditions = append(conditions, fmt.Sprintf("institution_id = $%d", argIndex))
		args = append(args, institutionId)
		argIndex++
	}

	if from != "" {
		fromDate, err := utils.ParseDateFilter(from)
		if err == nil && fromDate != nil {
			conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
			args = append(args, *fromDate)
			argIndex++
		}
	}

	if to != "" {
		toDate, err := utils.ParseDateFilter(to)
		if err == nil && toDate != nil {
			conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
			args = append(args, toDate.Add(24*time.Hour)) // Include the entire day
			argIndex++
		}
	}

	// Add conditions to queries
	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to get user count",
		})
		return
	}

	// Add pagination
	offset := (pageNumber - 1) * pageSize
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Execute query
	rows, err := database.DB.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to get users",
		})
		return
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
			&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
			&user.Username, &user.Role, &user.Status, &user.Slug, &user.Notes,
			&user.InstitutionID, &user.SubscriptionStatus, &user.HasActiveSubscription,
			&user.IsActive, &user.OTPRequired, &user.ReferralCode, &user.CreatedBy,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to scan user data",
			})
			return
		}

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
		users = append(users, userResponse)
	}

	// Calculate pagination
	totalPages, nextPage, prevPage := utils.CalculatePagination(pageNumber, pageSize, total)

	response := models.PaginationResponse{
		List:         users,
		Total:        total,
		CurrentPage:  pageNumber,
		LastPage:     totalPages,
		NextPage:     nextPage,
		PreviousPage: prevPage,
		Status:       "success",
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get a user by ID
// @Description Get a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.UserDto
// @Failure 404 {object} models.ErrorResponse
// @Router /{id} [get]
func GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	
	var user models.User
	query := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
		       phone, profile_picture, username, role, status, slug, notes, institution_id,
		       subscription_status, has_active_subscription, is_active, otp_required, 
		       referral_code, created_by, created_at, updated_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := database.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.Slug, &user.Notes,
		&user.InstitutionID, &user.SubscriptionStatus, &user.HasActiveSubscription,
		&user.IsActive, &user.OTPRequired, &user.ReferralCode, &user.CreatedBy,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	// Return as UserDto format
	userDto := models.UserDto{
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       &user.Username,
	}

	c.JSON(http.StatusOK, userDto)
}

// @Summary Find a user by slug
// @Description Find a user by slug
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "User Slug"
// @Success 200 {object} models.UserDto
// @Failure 404 {object} models.ErrorResponse
// @Router /slug/{slug} [get]
func GetUserBySlug(c *gin.Context) {
	slug := c.Param("slug")
	
	var user models.User
	query := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
		       phone, profile_picture, username, role, status, slug, notes, institution_id,
		       subscription_status, has_active_subscription, is_active, otp_required, 
		       referral_code, created_by, created_at, updated_at
		FROM users 
		WHERE slug = $1 AND deleted_at IS NULL
	`

	err := database.DB.QueryRow(query, slug).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.Slug, &user.Notes,
		&user.InstitutionID, &user.SubscriptionStatus, &user.HasActiveSubscription,
		&user.IsActive, &user.OTPRequired, &user.ReferralCode, &user.CreatedBy,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	// Return as UserDto format
	userDto := models.UserDto{
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       &user.Username,
	}

	c.JSON(http.StatusOK, userDto)
}

// @Summary Update a user
// @Description Update a user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param user body models.UpdateUserDto true "User"
// @Success 200 {object} models.UpdateUserDto
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /{id} [patch]
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	
	var updateDto models.UpdateUserDto
	if err := c.ShouldBindJSON(&updateDto); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// Build dynamic update query
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updateDto.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *updateDto.Email)
		argIndex++
	}

	if updateDto.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *updateDto.FirstName)
		argIndex++
	}

	if updateDto.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *updateDto.LastName)
		argIndex++
	}

	if updateDto.NationalID != nil {
		setParts = append(setParts, fmt.Sprintf("national_id = $%d", argIndex))
		args = append(args, *updateDto.NationalID)
		argIndex++
	}

	if updateDto.PassportNumber != nil {
		setParts = append(setParts, fmt.Sprintf("passport_number = $%d", argIndex))
		args = append(args, *updateDto.PassportNumber)
		argIndex++
	}

	if updateDto.Phone != nil {
		setParts = append(setParts, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, *updateDto.Phone)
		argIndex++
	}

	if updateDto.ProfilePicture != nil {
		setParts = append(setParts, fmt.Sprintf("profile_picture = $%d", argIndex))
		args = append(args, *updateDto.ProfilePicture)
		argIndex++
	}

	if updateDto.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *updateDto.Username)
		argIndex++
	}

	if updateDto.Role != nil {
		setParts = append(setParts, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, *updateDto.Role)
		argIndex++
	}

	if updateDto.Notes != nil {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, *updateDto.Notes)
		argIndex++
	}

	if updateDto.Password != nil {
		hashedPassword, err := utils.HashPassword(*updateDto.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to hash password",
			})
			return
		}
		setParts = append(setParts, fmt.Sprintf("password = $%d", argIndex))
		args = append(args, hashedPassword)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "No fields to update",
		})
		return
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add user ID for WHERE clause
	args = append(args, userID)

	updateQuery := fmt.Sprintf(`
		UPDATE users 
		SET %s 
		WHERE id = $%d AND deleted_at IS NULL
	`, strings.Join(setParts, ", "), argIndex)

	_, err := database.DB.Exec(updateQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, updateDto)
}

// @Summary Delete a user
// @Description Delete a user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /{id} [delete]
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Soft delete the user
	_, err := database.DB.Exec(`
		UPDATE users 
		SET deleted_at = NOW(), updated_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "User deleted successfully",
	})
}

// @Summary Update password by admin
// @Description Update password by admin
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param updatePassword body models.UpdatePasswordRequest true "Update Password Request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /update-password/admin [post]
func UpdatePasswordByAdmin(c *gin.Context) {
	// Check if user is admin
	userRole := c.GetString("user_role")
	if userRole != string(models.RoleAdmin) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error: "Access denied. Admin role required",
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

	// Hash new password
	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to hash password",
		})
		return
	}

	// For admin, we'll need a user ID parameter - let's assume it's in the request body or query
	// Since the swagger doesn't specify clearly, I'll add it as a query parameter
	targetUserID := c.Query("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Target user ID is required",
		})
		return
	}

	// Update password
	_, err = database.DB.Exec(`
		UPDATE users 
		SET password = $1, updated_at = NOW() 
		WHERE id = $2 AND deleted_at IS NULL
	`, hashedPassword, targetUserID)

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