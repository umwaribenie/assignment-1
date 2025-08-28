package handlers

import (
	"net/http"
	"strconv"

	"user-service/internal/models"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler represents the user handler
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles user creation
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User creation request"
// @Success 201 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "user with email "+req.Email+" already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "User created successfully",
		Data:    user,
	})
}

// GetUserByID handles getting a user by ID
// @Summary Get user by ID
// @Description Get user information by user ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Success 200 {object} models.SuccessResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// GetUserByEmail handles getting a user by email
// @Summary Get user by email
// @Description Get user information by email address
// @Tags Users
// @Produce json
// @Param email query string true "User email"
// @Success 200 {object} models.SuccessResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/email [get]
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Email parameter is required"})
		return
	}

	user, err := h.userService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// UpdateUser handles user updates
// @Summary Update user
// @Description Update user information
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param user body models.UpdateUserRequest true "User update request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "User updated successfully",
		Data:    user,
	})
}

// DeleteUser handles user deletion
// @Summary Delete user
// @Description Soft delete a user
// @Tags Users
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "User deleted successfully",
	})
}

// GetAllUsers handles getting all users with filtering and pagination
// @Summary Get all users
// @Description Get paginated list of users with optional filtering
// @Tags Users
// @Produce json
// @Param pageNumber query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Param search query string false "Search in email, username, first_name, last_name"
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
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

	if roleStr := c.Query("role"); roleStr != "" {
		filter.Role = &roleStr
	}

	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = &statusStr
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	response, err := h.userService.GetAllUsers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get users: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}