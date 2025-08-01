package handlers

import (
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

// @Summary Get all users
// @Description Get all users
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param pageNumber query int false "Page Number"
// @Param pageSize query int false "Page Size"
// @Param from query string false "From Date"
// @Param to query string false "To Date"
// @Param search query string false "Search"
// @Param role query string false "Role" Enums(user, admin, super_admin, )
// @Param status query string false "Status" Enums(active, inactive, deleted)
// @Success 200 {object} models.PaginatedResponse
// @Failure 500 {object} models.ErrorResponse
// @Router / [get]
func GetAllUsers(c *gin.Context) {
	pageNumber := 1
	pageSize := 10

	if p, err := strconv.Atoi(c.DefaultQuery("pageNumber", "1")); err == nil && p > 0 {
		pageNumber = p
	}
	if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "10")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	filter := models.UserFilter{PageNumber: pageNumber, PageSize: pageSize}

	// Role filter
	if roleStr := c.Query("role"); roleStr != "" {
		role := models.UserRole(roleStr)
		filter.Role = &role
	}

	// Status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status := models.UserStatus(statusStr)
		filter.Status = &status
	}

	// Search filter
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	// Date filters
	from := c.Query("from")
	to := c.Query("to")

	baseQuery := `SELECT id, client_id, email, first_name, last_name, phone, username, role, status, slug, created_at, updated_at FROM users WHERE deleted_at IS NULL`
	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"

	var conditions []string
	var args []interface{}
	argIndex := 1

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

	if filter.Search != nil {
		searchPattern := "%" + *filter.Search + "%"
		conditions = append(conditions, fmt.Sprintf("(email ILIKE $%d OR username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	if from != "" {
		conditions = append(conditions, fmt.Sprintf("DATE(created_at) >= $%d", argIndex))
		args = append(args, from)
		argIndex++
	}

	if to != "" {
		conditions = append(conditions, fmt.Sprintf("DATE(created_at) <= $%d", argIndex))
		args = append(args, to)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	var total int64
	database.DB.QueryRow(countQuery, args...).Scan(&total)

	offset := (pageNumber - 1) * pageSize
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	rows, _ := database.DB.Query(baseQuery, args...)
	defer rows.Close()

	var users []models.UserListResponse
	for rows.Next() {
		var user models.User
		rows.Scan(&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Username, &user.Role, &user.Status, &user.Slug, &user.CreatedAt, &user.UpdatedAt)

		userResponse := models.UserListResponse{
			ClientID:  user.ClientID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Phone:     user.Phone,
			Username:  user.Username,
			Role:      user.Role,
			Status:    user.Status,
			Slug:      user.Slug,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		users = append(users, userResponse)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	// Calculate nextPage and previousPage
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

// @Summary Get User by ID
// @Description Get a single user by their ID
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.UserResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/{id} [get]
func GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	err := database.DB.QueryRow(`SELECT id, client_id, email, first_name, last_name, national_id, passport_number, phone, profile_picture, username FROM users WHERE id = $1 AND deleted_at IS NULL`, userID).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName, &user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture, &user.Username)

	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	userResponse := models.UserResponse{
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       user.Username,
	}

	c.JSON(http.StatusOK, userResponse)
}

// @Summary Get User by Slug
// @Description Get a single user by their slug
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param slug path string true "User Slug"
// @Success 200 {object} models.UserResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/slug/{slug} [get]
func GetUserBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var user models.User
	err := database.DB.QueryRow(`SELECT id, client_id, email, first_name, last_name, phone, username, role, status, slug, created_at, updated_at FROM users WHERE slug = $1 AND deleted_at IS NULL`, slug).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Username, &user.Role, &user.Status, &user.Slug, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	userResponse := models.UserResponse{
		ClientID:       user.ClientID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Username:       user.Username,
	}

	c.JSON(http.StatusOK, userResponse)
}

// @Summary Register a new user
// @Description Register a new user
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User"
// @Success 200 {object} models.RegisterResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/register [post]
func RegisterNewUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request data"})
		return
	}

	hashedPassword, _ := utils.HashPassword(req.Password)

	username := req.Username
	if username == "" {
		username = utils.GenerateUsername(req.FirstName, req.LastName)
	}

	var createdUser models.User

	database.DB.QueryRow(`INSERT INTO users (client_id, email, first_name, last_name, national_id, passport_number, password, phone, profile_picture, username) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at, updated_at`,
		req.ClientID, req.Email, req.FirstName, req.LastName, req.NationalID, req.PassportNumber, hashedPassword, req.Phone, req.ProfilePicture, username).Scan(&createdUser.ID, &createdUser.CreatedAt, &createdUser.UpdatedAt)

	userResponse := models.RegisterResponse{
		ClientID:       req.ClientID,
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		NationalID:     req.NationalID,
		PassportNumber: req.PassportNumber,
		Password:       req.Password,
		Phone:          req.Phone,
		ProfilePicture: req.ProfilePicture,
		Username:       username,
	}

	c.JSON(http.StatusOK, userResponse)
}

// @Summary Register a new user by admin
// @Description Register a new user by admin
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.AdminRegisterRequest true "User"
// @Success 200 {object} models.AdminRegisterResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/registerusersbyadmin [post]
func CreateUserByAdmin(c *gin.Context) {
	var req models.AdminRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request data"})
		return
	}

	hashedPassword, _ := utils.HashPassword(req.Password)
	userID := uuid.New()
	clientID := utils.GenerateClientID()

	username := req.Username
	if username == "" {
		username = utils.GenerateUsername(req.FirstName, req.LastName)
	}

	slug := utils.GenerateUniqueSlug(username, userID.String())

	now := time.Now()
	database.DB.QueryRow(`INSERT INTO users (id, client_id, email, first_name, last_name, national_id, passport_number, password, phone, profile_picture, username, role, status, slug, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING id`,
		userID, clientID, req.Email, req.FirstName, req.LastName, req.NationalID, req.PassportNumber, hashedPassword, req.Phone, req.ProfilePicture, username, models.RoleUser, models.StatusActive, slug, now, now)

	response := models.AdminRegisterResponse{
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		NationalID:     req.NationalID,
		PassportNumber: req.PassportNumber,
		Password:       req.Password,
		Phone:          req.Phone,
		ProfilePicture: req.ProfilePicture,
		Username:       username,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Update a user
// @Description Update a user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param user body models.UserUpdateRequest true "User"
// @Success 200 {object} models.UserUpdateResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [patch]
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request data"})
		return
	}

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
	if req.Password != nil {
		hashedPassword, _ := utils.HashPassword(*req.Password)
		setParts = append(setParts, fmt.Sprintf("password = $%d", argIndex))
		args = append(args, hashedPassword)
		argIndex++
	}
	if req.Phone != nil {
		setParts = append(setParts, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, *req.Phone)
		argIndex++
	}
	if req.ProfilePicture != nil {
		setParts = append(setParts, fmt.Sprintf("profile_picture = $%d", argIndex))
		args = append(args, *req.ProfilePicture)
		argIndex++
	}
	if req.Role != nil {
		setParts = append(setParts, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, string(*req.Role))
		argIndex++
	}
	if req.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, *req.Username)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "No fields to update"})
		return
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID)

	updateQuery := fmt.Sprintf(`UPDATE users SET %s WHERE id = $%d RETURNING email, first_name, last_name, national_id, passport_number, phone, profile_picture, role, username`, strings.Join(setParts, ", "), argIndex)

	var user models.User
	err := database.DB.QueryRow(updateQuery, args...).Scan(&user.Email, &user.FirstName, &user.LastName, &user.NationalID, &user.PassportNumber, &user.Phone, &user.ProfilePicture, &user.Role, &user.Username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update user"})
		return
	}

	passwordToReturn := ""
	if req.Password != nil {
		passwordToReturn = *req.Password
	}

	response := models.UserUpdateResponse{
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NationalID:     user.NationalID,
		PassportNumber: user.PassportNumber,
		Password:       passwordToReturn,
		Phone:          user.Phone,
		ProfilePicture: user.ProfilePicture,
		Role:           user.Role,
		Username:       user.Username,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Delete User
// @Description Soft delete a user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	database.DB.Exec(`UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1`, userID)
	c.JSON(http.StatusOK, models.SuccessResponse{Message: "User deleted successfully"})
}

// @Summary Update Password
// @Description Update user's own password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param passwordUpdateRequest body models.PasswordUpdateRequest true "Password update data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/update-password [post]
func UpdatePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req models.PasswordUpdateRequest
	c.ShouldBindJSON(&req)

	var currentPasswordHash string
	database.DB.QueryRow(`SELECT password FROM users WHERE id = $1`, userID).Scan(&currentPasswordHash)

	if !utils.CheckPasswordHash(req.OldPassword, currentPasswordHash) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Current password is incorrect"})
		return
	}

	newPasswordHash, _ := utils.HashPassword(req.NewPassword)
	database.DB.Exec(`UPDATE users SET password = $1 WHERE id = $2`, newPasswordHash, userID)

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Password updated successfully"})
}