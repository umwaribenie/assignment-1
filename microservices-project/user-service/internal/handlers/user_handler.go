package handlers

import (
	"net/http"
	"strconv"
	"user-service/internal/models"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register handles user registration
func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Registration failed",
			Message: err.Error(),
		})
		return
	}

	// Convert to response
	userResponse := models.UserResponse{
		ID:               user.ID.String(),
		Email:            user.Email,
		Username:         user.Username,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		PhoneNumber:      user.PhoneNumber,
		Role:             user.Role,
		Status:           user.Status,
		EmailVerified:    user.EmailVerified,
		PhoneVerified:    user.PhoneVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    userResponse,
	})
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Get client IP
	clientIP := c.ClientIP()

	loginResponse, err := h.userService.Login(c.Request.Context(), req, clientIP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "Login failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Login successful",
		Data:    loginResponse,
	})
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "User not found",
			Message: err.Error(),
		})
		return
	}

	// Convert to response
	userResponse := models.UserResponse{
		ID:               user.ID.String(),
		Email:            user.Email,
		Username:         user.Username,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		PhoneNumber:      user.PhoneNumber,
		ProfilePicture:   user.ProfilePicture,
		Role:             user.Role,
		Status:           user.Status,
		EmailVerified:    user.EmailVerified,
		PhoneVerified:    user.PhoneVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    userResponse,
	})
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Update failed",
			Message: err.Error(),
		})
		return
	}

	// Convert to response
	userResponse := models.UserResponse{
		ID:               user.ID.String(),
		Email:            user.Email,
		Username:         user.Username,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		PhoneNumber:      user.PhoneNumber,
		ProfilePicture:   user.ProfilePicture,
		Role:             user.Role,
		Status:           user.Status,
		EmailVerified:    user.EmailVerified,
		PhoneVerified:    user.PhoneVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    userResponse,
	})
}

// ChangePassword handles password change
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.Param("id")
	
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Password change failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// ForgotPassword handles forgot password request
func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if err := h.userService.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to process request",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if err := h.userService.ResetPassword(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Password reset failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Delete failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// GetAllUsers retrieves all users with pagination
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

	// Apply filters
	if roleStr := c.Query("role"); roleStr != "" {
		role := models.UserRole(roleStr)
		filter.Role = &role
	}
	if statusStr := c.Query("status"); statusStr != "" {
		status := models.UserStatus(statusStr)
		filter.Status = &status
	}
	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	response, err := h.userService.GetAllUsers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to retrieve users",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    response,
	})
}