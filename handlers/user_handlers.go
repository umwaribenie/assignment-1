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
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param role query string false "Filter by role" Enums(user, admin)
// @Param status query string false "Filter by status" Enums(active, inactive, suspended)
// @Param search query string false "Search in email, username, first_name, last_name"
// @Success 200 {object} models.APIResponse{data=models.PaginatedResponse}
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Router /users [get]
func GetAllUsers(c *gin.Context) {
	// Parse query parameters
	page := 1
	pageSize := 10
	
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	
	if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	filter := models.UserFilter{
		Page:     page,
		PageSize: pageSize,
	}

	// Parse filters
	if role := c.Query("role"); role != "" {
		userRole := models.UserRole(role)
		filter.Role = &userRole
	}

	if status := c.Query("status"); status != "" {
		userStatus := models.UserStatus(status)
		filter.Status = &userStatus
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	// Build query
	baseQuery := `SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
	              phone, profile_picture, username, role, status, created_at, updated_at 
	              FROM users WHERE deleted_at IS NULL`
	
	countQuery := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	
	var args []interface{}
	var conditions []string
	argIndex := 1

	// Add filters
	if filter.Role != nil {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, *filter.Role)
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Search != nil {
		searchCondition := fmt.Sprintf("(email ILIKE $%d OR username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", 
			argIndex, argIndex, argIndex, argIndex)
		conditions = append(conditions, searchCondition)
		args = append(args, "%"+*filter.Search+"%")
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
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	// Add pagination
	offset := (page - 1) * pageSize
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Execute query
	rows, err := database.DB.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
			&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
			&user.Username, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Error scanning user data",
				Error:   err.Error(),
			})
			return
		}

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
		users = append(users, userResponse)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	paginatedResponse := models.PaginatedResponse{
		Data:       users,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    paginatedResponse,
	})
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
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /users/{id} [get]
func GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
		return
	}

	var user models.User
	query := `SELECT id, client_id, email, first_name, last_name, national_id, passport_number, 
	          phone, profile_picture, username, role, status, created_at, updated_at 
	          FROM users WHERE id = $1 AND deleted_at IS NULL`
	
	row := database.DB.QueryRow(query, userUUID)
	err = row.Scan(&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
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
		Message: "User retrieved successfully",
		Data:    userResponse,
	})
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param user body models.UpdateUserRequest true "Updated user data"
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Router /users/{id} [put]
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
		return
	}

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
	err = database.DB.QueryRow(checkQuery, userUUID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Check for username conflicts (if username is being updated)
	if req.Username != nil {
		var usernameExists bool
		usernameCheckQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2 AND deleted_at IS NULL)`
		err = database.DB.QueryRow(usernameCheckQuery, *req.Username, userUUID).Scan(&usernameExists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Database error",
				Error:   err.Error(),
			})
		return
		}

		if usernameExists {
			c.JSON(http.StatusConflict, models.APIResponse{
				Success: false,
				Message: "Username already exists",
			})
			return
		}
	}

	// Build update query dynamically
	var setParts []string
	var args []interface{}
	argIndex := 1

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
		setParts = append(setParts, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, *req.Role)
		argIndex++
	}

	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
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
	args = append(args, userUUID)

	updateQuery := fmt.Sprintf(`UPDATE users SET %s WHERE id = $%d 
	                           RETURNING id, client_id, email, first_name, last_name, national_id, passport_number, 
	                           phone, profile_picture, username, role, status, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	var user models.User
	err = database.DB.QueryRow(updateQuery, args...).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture,
		&user.Username, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
		return
	}

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
		Message: "User updated successfully",
		Data:    userResponse,
	})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Soft delete a user (sets deleted_at timestamp)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
		return
	}

	// Check if user exists and is not already deleted
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`
	err = database.DB.QueryRow(checkQuery, userUUID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Soft delete user
	deleteQuery := `UPDATE users SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err = database.DB.Exec(deleteQuery, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// UpdatePassword godoc
// @Summary Update user password
// @Description Update the password of the authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body models.PasswordUpdateRequest true "Password update data"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /users/password [put]
func UpdatePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")
	
	var req models.PasswordUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	// Get current password
	var currentPasswordHash string
	query := `SELECT password FROM users WHERE id = $1 AND deleted_at IS NULL`
	err = database.DB.QueryRow(query, userUUID).Scan(&currentPasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, models.APIResponse{
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

	// Verify current password
	if !utils.CheckPasswordHash(req.CurrentPassword, currentPasswordHash) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Current password is incorrect",
		})
		return
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to hash new password",
			Error:   err.Error(),
		})
		return
	}

	// Update password
	updateQuery := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = database.DB.Exec(updateQuery, newPasswordHash, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update password",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password updated successfully",
	})
}