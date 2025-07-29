package handlers

import (
	"database/sql"
	"fmt"
	"generalusermanagement/database"
	"generalusermanagement/models"
	"generalusermanagement/utils"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAllUsers godoc
// @Summary Get all users
// @Description Get paginated list of users with optional filtering
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param pageNumber query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param role query string false "Filter by role" Enums(user, admin, super_admin, trainer, instructor, frontdesk, finance, seler, member)
// @Param status query string false "Filter by status" Enums(active, inactive, suspended, deleted)
// @Param subscriptionStatus query string false "Filter by subscription status" Enums(active, inactive, expired, onhold, paused, canceled)
// @Param institutionId query string false "Filter by institution ID"
// @Param search query string false "Search in email, username, first_name, last_name"
// @Param from query string false "Filter from date (YYYY-MM-DD)"
// @Param to query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} models.APIResponse{data=models.PaginatedResponse}
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Router /users [get]
func GetAllUsers(c *gin.Context) {
	// Parse query parameters
	pageNumber := 1
	pageSize := 10
	
	if p, err := strconv.Atoi(c.DefaultQuery("pageNumber", "1")); err == nil && p > 0 {
		pageNumber = p
	}
	
	if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "10")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	filter := models.UserFilter{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}

	// Parse role filter
	if roleStr := c.Query("role"); roleStr != "" {
		role := models.UserRole(roleStr)
		if utils.ValidateRole(role) {
			filter.Role = &role
		}
	}

	// Parse status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := models.UserStatus(statusStr)
		filter.Status = &status
	}

	// Parse subscription status filter
	if subStatusStr := c.Query("subscriptionStatus"); subStatusStr != "" {
		subStatus := models.SubscriptionStatus(subStatusStr)
		if utils.ValidateSubscriptionStatus(subStatus) {
			filter.SubscriptionStatus = &subStatus
		}
	}

	// Parse institution ID filter
	if institutionID := c.Query("institutionId"); institutionID != "" {
		filter.InstitutionID = &institutionID
	}

	// Parse search filter
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	// Parse date filters
	if from := c.Query("from"); from != "" {
		filter.From = &from
	}
	if to := c.Query("to"); to != "" {
		filter.To = &to
	}

	// Build base query
	baseQuery := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
		       phone, profile_picture, username, role, status, subscription_status, 
		       institution_id, commission_percentage, seler_type, specialization, notes, 
		       slug, referral_code, has_active_subscription, is_active, otp_required, 
		       created_by, created_at, updated_at, deleted_at 
		FROM users WHERE deleted_at IS NULL`

	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add filters
	if filter.Role != nil {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, string(*filter.Role))
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*filter.Status))
		argIndex++
	}

	if filter.SubscriptionStatus != nil {
		conditions = append(conditions, fmt.Sprintf("subscription_status = $%d", argIndex))
		args = append(args, string(*filter.SubscriptionStatus))
		argIndex++
	}

	if filter.InstitutionID != nil {
		conditions = append(conditions, fmt.Sprintf("institution_id = $%d", argIndex))
		args = append(args, *filter.InstitutionID)
		argIndex++
	}

	if filter.Search != nil {
		searchPattern := "%" + *filter.Search + "%"
		conditions = append(conditions, fmt.Sprintf(
			"(email ILIKE $%d OR username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)",
			argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	if filter.From != nil {
		conditions = append(conditions, fmt.Sprintf("DATE(created_at) >= $%d", argIndex))
		args = append(args, *filter.From)
		argIndex++
	}

	if filter.To != nil {
		conditions = append(conditions, fmt.Sprintf("DATE(created_at) <= $%d", argIndex))
		args = append(args, *filter.To)
		argIndex++
	}

	// Add conditions to queries
	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int64
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to get user count",
			Error:   err.Error(),
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
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to fetch users",
			Error:   err.Error(),
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
			&user.Username, &user.Role, &user.Status, &user.SubscriptionStatus,
			&user.InstitutionID, &user.CommissionPercentage, &user.SelerType,
			&user.Specialization, &user.Notes, &user.Slug, &user.ReferralCode,
			&user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
			&user.CreatedBy, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to scan user data",
				Error:   err.Error(),
			})
			return
		}

		// Convert to UserResponse
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

		users = append(users, userResponse)
	}

	// Calculate pagination info
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	var nextPage, previousPage *int
	if pageNumber < totalPages {
		next := pageNumber + 1
		nextPage = &next
	}
	if pageNumber > 1 {
		prev := pageNumber - 1
		previousPage = &prev
	}

	response := models.PaginatedResponse{
		List:         users,
		CurrentPage:  pageNumber,
		LastPage:     totalPages,
		NextPage:     nextPage,
		PreviousPage: previousPage,
		Total:        total,
		Status:       "success",
	}

	c.JSON(http.StatusOK, response)
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get a specific user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{id} [get]
func GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	query := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
		       phone, profile_picture, username, role, status, subscription_status, 
		       institution_id, commission_percentage, seler_type, specialization, notes, 
		       slug, referral_code, has_active_subscription, is_active, otp_required, 
		       created_by, created_at, updated_at, deleted_at 
		FROM users WHERE id = $1 AND deleted_at IS NULL`

	var user models.User
	err := database.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.SubscriptionStatus,
		&user.InstitutionID, &user.CommissionPercentage, &user.SelerType,
		&user.Specialization, &user.Notes, &user.Slug, &user.ReferralCode,
		&user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
		&user.CreatedBy, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to fetch user",
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

	c.JSON(http.StatusOK, userResponse)
}

// GetUserBySlug godoc
// @Summary Find a user by slug
// @Description Get a specific user by their slug
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "User Slug"
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 404 {object} models.ErrorResponse
// @Router /users/slug/{slug} [get]
func GetUserBySlug(c *gin.Context) {
	slug := c.Param("slug")

	query := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
		       phone, profile_picture, username, role, status, subscription_status, 
		       institution_id, commission_percentage, seler_type, specialization, notes, 
		       slug, referral_code, has_active_subscription, is_active, otp_required, 
		       created_by, created_at, updated_at, deleted_at 
		FROM users WHERE slug = $1 AND deleted_at IS NULL`

	var user models.User
	err := database.DB.QueryRow(query, slug).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.SubscriptionStatus,
		&user.InstitutionID, &user.CommissionPercentage, &user.SelerType,
		&user.Specialization, &user.Notes, &user.Slug, &user.ReferralCode,
		&user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
		&user.CreatedBy, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to fetch user",
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

	c.JSON(http.StatusOK, userResponse)
}

// RegisterUser godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User registration data"
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /users/register [post]
func RegisterUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
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

	// Generate unique identifiers
	userID := uuid.New()
	
	// Generate username if not provided
	username := req.Username
	if username == "" {
		username = utils.GenerateUsername(req.FirstName, req.LastName)
	}

	// Generate slug
	slug := utils.GenerateUniqueSlug(username, userID.String())

	// Generate referral code
	referralCode, err := utils.GenerateReferralCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate referral code",
			Error:   err.Error(),
		})
		return
	}

	// Set default values
	role := req.Role
	if role == "" {
		role = models.RoleUser
	}

	status := req.Status
	if status == "" {
		status = models.StatusActive
	}

	// Insert user into database
	query := `
		INSERT INTO users (id, client_id, email, first_name, last_name, national_id, passport_number, 
		                  password, phone, username, role, status, slug, referral_code, 
		                  has_active_subscription, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	var createdUser models.User
	err = database.DB.QueryRow(query,
		userID, req.ClientID, req.Email, req.FirstName, req.LastName,
		req.NationalID, req.PassportNumber, hashedPassword, req.Phone,
		username, role, status, slug, referralCode, false, true, now, now,
	).Scan(&createdUser.ID, &createdUser.CreatedAt, &createdUser.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "User with this email, username, or client ID already exists",
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	// Prepare response
	userResponse := models.UserResponse{
		ID:                    userID,
		ClientID:              req.ClientID,
		Email:                 req.Email,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		NationalID:            req.NationalID,
		PassportNumber:        req.PassportNumber,
		Phone:                 req.Phone,
		Username:              username,
		Role:                  role,
		Status:                status,
		Slug:                  slug,
		ReferralCode:          &referralCode,
		HasActiveSubscription: false,
		IsActive:              true,
		CreatedAt:             createdUser.CreatedAt,
		UpdatedAt:             createdUser.UpdatedAt,
	}

	c.JSON(http.StatusOK, userResponse)
}

// CreateUserByAdmin godoc
// @Summary Register a new user by admin
// @Description Create a new user account with admin privileges
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.CreateUserByAdminRequest true "User creation data"
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /users/registerusersbyadmin [post]
func CreateUserByAdmin(c *gin.Context) {
	var req models.CreateUserByAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Get current user from context (set by middleware)
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	// Validate role
	if !utils.ValidateRole(req.Role) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid role specified",
		})
		return
	}

	// Validate commission percentage for sellers
	if req.Role == models.RoleSeler {
		if req.CommissionPercentage == nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Commission percentage is required for sellers",
			})
			return
		}
		if req.SelerType == nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Seller type is required for sellers",
			})
			return
		}
		if !utils.ValidateSelerType(*req.SelerType) {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Invalid seller type",
			})
			return
		}
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

	// Generate unique identifiers
	userID := uuid.New()
	clientID := utils.GenerateClientID()
	
	// Generate username if not provided
	username := ""
	if req.Username != nil {
		username = *req.Username
	} else {
		username = utils.GenerateUsername(req.FirstName, req.LastName)
	}

	// Generate slug
	slug := utils.GenerateUniqueSlug(username, userID.String())

	// Generate referral code
	referralCode, err := utils.GenerateReferralCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate referral code",
			Error:   err.Error(),
		})
		return
	}

	// Insert user into database
	query := `
		INSERT INTO users (id, client_id, email, first_name, last_name, national_id, passport_number, 
		                  password, phone, username, role, status, institution_id, commission_percentage,
		                  seler_type, specialization, profile_picture, slug, referral_code, 
		                  has_active_subscription, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	var createdUser models.User
	err = database.DB.QueryRow(query,
		userID, clientID, req.Email, req.FirstName, req.LastName,
		req.NationalID, req.PassportNumber, hashedPassword, req.Phone,
		username, req.Role, models.StatusActive, req.InstitutionID, req.CommissionPercentage,
		req.SelerType, req.Specialization, req.ProfilePicture, slug, referralCode,
		false, true, currentUserID, now, now,
	).Scan(&createdUser.ID, &createdUser.CreatedAt, &createdUser.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "User with this email or username already exists",
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	// Prepare response
	userResponse := models.CreateUserByAdminRequest{
		Email:                req.Email,
		FirstName:            req.FirstName,
		LastName:             req.LastName,
		NationalID:           req.NationalID,
		PassportNumber:       req.PassportNumber,
		Password:             "", // Don't return password
		Phone:                req.Phone,
		Username:             &username,
		Role:                 req.Role,
		InstitutionID:        req.InstitutionID,
		CommissionPercentage: req.CommissionPercentage,
		SelerType:            req.SelerType,
		Specialization:       req.Specialization,
		ProfilePicture:       req.ProfilePicture,
	}

	c.JSON(http.StatusOK, userResponse)
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update user information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param user body models.UpdateUserRequest true "Updated user data"
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /users/{id} [patch]
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Check if user exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`
	err := database.DB.QueryRow(checkQuery, userID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "User not found",
		})
		return
	}

	// Build update query dynamically
	var setParts []string
	var args []interface{}
	argIndex := 1

	if req.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *req.Email)
		argIndex++
	}

	if req.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, *req.FirstName)
		argIndex++
	}

	if req.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, *req.LastName)
		argIndex++
	}

	if req.NationalID != nil {
		setParts = append(setParts, fmt.Sprintf("national_id = $%d", argIndex))
		args = append(args, *req.NationalID)
		argIndex++
	}

	if req.PassportNumber != nil {
		setParts = append(setParts, fmt.Sprintf("passport_number = $%d", argIndex))
		args = append(args, *req.PassportNumber)
		argIndex++
	}

	if req.Phone != nil {
		setParts = append(setParts, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, *req.Phone)
		argIndex++
	}

	if req.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *req.Username)
		argIndex++
	}

	if req.Role != nil {
		if !utils.ValidateRole(*req.Role) {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Invalid role specified",
			})
			return
		}
		setParts = append(setParts, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, string(*req.Role))
		argIndex++
	}

	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*req.Status))
		argIndex++
	}

	if req.Notes != nil {
		setParts = append(setParts, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, *req.Notes)
		argIndex++
	}

	if req.Specialization != nil {
		setParts = append(setParts, fmt.Sprintf("specialization = $%d", argIndex))
		args = append(args, *req.Specialization)
		argIndex++
	}

	if req.ProfilePicture != nil {
		setParts = append(setParts, fmt.Sprintf("profile_picture = $%d", argIndex))
		args = append(args, *req.ProfilePicture)
		argIndex++
	}

	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to hash password",
				Error:   err.Error(),
			})
			return
		}
		setParts = append(setParts, fmt.Sprintf("password = $%d", argIndex))
		args = append(args, hashedPassword)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "No fields to update",
		})
		return
	}

	// Add updated_at
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	// Add user ID for WHERE clause
	args = append(args, userID)

	updateQuery := fmt.Sprintf(`
		UPDATE users SET %s WHERE id = $%d 
		RETURNING id, client_id, email, first_name, last_name, national_id, passport_number, 
		          phone, profile_picture, username, role, status, subscription_status, 
		          institution_id, commission_percentage, seler_type, specialization, notes, 
		          slug, referral_code, has_active_subscription, is_active, otp_required, 
		          created_by, created_at, updated_at, deleted_at`,
		strings.Join(setParts, ", "), argIndex)

	var user models.User
	err = database.DB.QueryRow(updateQuery, args...).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.SubscriptionStatus,
		&user.InstitutionID, &user.CommissionPercentage, &user.SelerType,
		&user.Specialization, &user.Notes, &user.Slug, &user.ReferralCode,
		&user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
		&user.CreatedBy, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
		return
	}

	userResponse := models.UpdateUserRequest{
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		NationalID:     req.NationalID,
		PassportNumber: req.PassportNumber,
		Phone:          req.Phone,
		Username:       req.Username,
		Role:           req.Role,
		Status:         req.Status,
		Notes:          req.Notes,
		Specialization: req.Specialization,
		ProfilePicture: req.ProfilePicture,
		Password:       nil, // Don't return password
	}

	c.JSON(http.StatusOK, userResponse)
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Soft delete a user (sets deleted_at timestamp)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Check if user exists and is not already deleted
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`
	err := database.DB.QueryRow(checkQuery, userID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to check user existence",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "User not found",
		})
		return
	}

	// Soft delete user
	deleteQuery := `UPDATE users SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = database.DB.Exec(deleteQuery, userID)
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

// UpdatePassword godoc
// @Summary Update password
// @Description Update user password with current password verification
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param updatePassword body models.PasswordUpdateRequest true "Password update data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/update-password [post]
func UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	var req models.PasswordUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request data",
		})
		return
	}

	// Get current password hash
	var currentPasswordHash string
	query := `SELECT password FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := database.DB.QueryRow(query, userID).Scan(&currentPasswordHash)
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

	// Verify current password
	if !utils.CheckPasswordHash(req.OldPassword, currentPasswordHash) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Current password is incorrect",
		})
		return
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to hash new password",
		})
		return
	}

	// Update password
	updateQuery := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = database.DB.Exec(updateQuery, newPasswordHash, userID)
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

// UpdatePasswordByAdmin godoc
// @Summary Update password by admin
// @Description Update user password by admin without current password verification
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param updatePassword body models.PasswordUpdateRequest true "Password update data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/update-password/admin [post]
func UpdatePasswordByAdmin(c *gin.Context) {
	var req models.PasswordUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request data",
		})
		return
	}

	// For admin password update, we use the old_password field as the user ID
	targetUserID := req.OldPassword

	// Check if target user exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`
	err := database.DB.QueryRow(checkQuery, targetUserID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "User not found",
		})
		return
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to hash new password",
		})
		return
	}

	// Update password
	updateQuery := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = database.DB.Exec(updateQuery, newPasswordHash, targetUserID)
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