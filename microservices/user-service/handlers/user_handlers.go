package handlers

import (
	"fmt"
	"math"
	"net/http"
	"shared"
	"strconv"
	"strings"
	"time"
	"user-service/cache"
	"user-service/database"
	"user-service/kafka"
	"user-service/models"
	"user-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAllUsers retrieves paginated list of users with caching
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

	if roleStr := c.Query("role"); roleStr != "" {
		role := shared.UserRole(roleStr)
		if utils.ValidateRole(role) {
			filter.Role = &role
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := shared.UserStatus(statusStr)
		if utils.ValidateStatus(status) {
			filter.Status = &status
		}
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	// Try to get from cache first
	cacheKey := cache.UsersListCacheKey(pageNumber, pageSize, fmt.Sprintf("%v", filter))
	var cachedResponse models.PaginatedResponse
	if err := cache.GetCache(cacheKey, &cachedResponse); err == nil {
		c.JSON(http.StatusOK, cachedResponse)
		return
	}

	// Build query
	baseQuery := `SELECT id, client_id, email, first_name, last_name, phone, username, role, status, slug, created_at, updated_at FROM users WHERE deleted_at IS NULL`
	countQuery := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	
	var conditions []string
	var args []interface{}
	argIndex := 1

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
		conditions = append(conditions, fmt.Sprintf("(email ILIKE $%d OR username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int64
	err := database.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Failed to count users"})
		return
	}

	// Add pagination
	offset := (pageNumber - 1) * pageSize
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d OFFSET %d", pageSize, offset)

	rows, err := database.DB.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Failed to fetch users"})
		return
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Username, &user.Role, &user.Status, &user.Slug, &user.CreatedAt, &user.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Failed to scan user"})
			return
		}
		
		userResponse := models.UserResponse{
			ID: user.ID, ClientID: user.ClientID, Email: user.Email, FirstName: user.FirstName, LastName: user.LastName, Phone: user.Phone, Username: user.Username, Role: user.Role, Status: user.Status, Slug: user.Slug, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
		}
		users = append(users, userResponse)
	}

	// Calculate pagination info
	lastPage := int(math.Ceil(float64(total) / float64(pageSize)))
	var nextPage, previousPage *int
	
	if pageNumber < lastPage {
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
		LastPage:     lastPage,
		NextPage:     nextPage,
		PreviousPage: previousPage,
		Total:        total,
		Status:       "success",
	}

	// Cache the response for 5 minutes
	cache.SetCache(cacheKey, response, 5*time.Minute)

	c.JSON(http.StatusOK, response)
}

// GetUserByID retrieves a user by ID with caching
func GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	// Try cache first
	cacheKey := cache.UserCacheKey(userID)
	var cachedUser models.UserResponse
	if err := cache.GetCache(cacheKey, &cachedUser); err == nil {
		c.JSON(http.StatusOK, shared.APIResponse{Success: true, Data: cachedUser})
		return
	}

	var user models.User
	err := database.DB.QueryRow(`SELECT id, client_id, email, first_name, last_name, phone, username, role, status, slug, created_at, updated_at FROM users WHERE id = $1 AND deleted_at IS NULL`, userID).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName, &user.Phone, &user.Username, &user.Role, &user.Status, &user.Slug, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Error: "User not found"})
		return
	}

	userResponse := models.UserResponse{ID: user.ID, ClientID: user.ClientID, Email: user.Email, FirstName: user.FirstName, LastName: user.LastName, Phone: user.Phone, Username: user.Username, Role: user.Role, Status: user.Status, Slug: user.Slug, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}

	// Cache for 10 minutes
	cache.SetCache(cacheKey, userResponse, 10*time.Minute)

	c.JSON(http.StatusOK, shared.APIResponse{Success: true, Data: userResponse})
}

// CreateUser creates a new user and publishes event
func CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Error: "Invalid request data"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Failed to process password"})
		return
	}

	// Generate username if not provided
	username := req.Username
	if username == "" {
		username = strings.ToLower(req.FirstName + req.LastName)
	}

	// Generate slug
	slug := utils.GenerateSlug(req.FirstName, req.LastName)

	// Set defaults
	if req.Role == "" {
		req.Role = shared.RoleUser
	}
	if req.Status == "" {
		req.Status = shared.StatusActive
	}

	userID := uuid.New()
	now := time.Now()

	var createdUser models.User
	err = database.DB.QueryRow(`INSERT INTO users (id, client_id, email, first_name, last_name, national_id, passport_number, password, phone, username, role, status, slug, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id, created_at, updated_at`,
		userID, req.ClientID, req.Email, req.FirstName, req.LastName, req.NationalID, req.PassportNumber, hashedPassword, req.Phone, username, req.Role, req.Status, slug, now, now).Scan(&createdUser.ID, &createdUser.CreatedAt, &createdUser.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, shared.ErrorResponse{Error: "User with this email or username already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Failed to create user"})
		}
		return
	}

	// Set additional fields for response
	createdUser.ClientID = req.ClientID
	createdUser.Email = req.Email
	createdUser.FirstName = req.FirstName
	createdUser.LastName = req.LastName
	createdUser.Phone = req.Phone
	createdUser.Username = username
	createdUser.Role = req.Role
	createdUser.Status = req.Status
	createdUser.Slug = slug

	// Publish user registered event to Kafka
	sharedUser := shared.User{
		ID:        createdUser.ID,
		ClientID:  createdUser.ClientID,
		Email:     createdUser.Email,
		FirstName: createdUser.FirstName,
		LastName:  createdUser.LastName,
		Phone:     createdUser.Phone,
		Username:  createdUser.Username,
		Role:      createdUser.Role,
		Status:    createdUser.Status,
		Slug:      createdUser.Slug,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
	}

	if err := kafka.PublishUserRegisteredEvent(sharedUser); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish user registered event: %v\n", err)
	}

	userResponse := models.UserResponse{
		ID: createdUser.ID, ClientID: createdUser.ClientID, Email: createdUser.Email, FirstName: createdUser.FirstName, LastName: createdUser.LastName, Phone: createdUser.Phone, Username: createdUser.Username, Role: createdUser.Role, Status: createdUser.Status, Slug: createdUser.Slug, CreatedAt: createdUser.CreatedAt, UpdatedAt: createdUser.UpdatedAt,
	}

	c.JSON(http.StatusCreated, shared.APIResponse{Success: true, Message: "User created successfully", Data: userResponse})
}