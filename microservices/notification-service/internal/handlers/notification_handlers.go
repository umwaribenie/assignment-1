package handlers

import (
	"net/http"
	"strconv"

	"notification-service/internal/models"
	"notification-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NotificationHandler represents the notification handler
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// CreateNotification handles notification creation
// @Summary Create a new notification
// @Description Create a new notification with the provided information
// @Tags Notifications
// @Accept json
// @Produce json
// @Param notification body models.CreateNotificationRequest true "Notification creation request"
// @Success 201 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications [post]
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	if err := h.notificationService.CreateNotification(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create notification: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Notification created successfully",
	})
}

// GetNotificationByID handles getting a notification by ID
// @Summary Get notification by ID
// @Description Get notification information by notification ID
// @Tags Notifications
// @Produce json
// @Param id path string true "Notification ID" format(uuid)
// @Success 200 {object} models.SuccessResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications/{id} [get]
func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	idStr := c.Param("id")
	if _, err := uuid.Parse(idStr); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid notification ID format"})
		return
	}

	notification, err := h.notificationService.GetNotificationByID(c.Request.Context(), idStr)
	if err != nil {
		if err.Error() == "notification not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get notification: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Notification retrieved successfully",
		Data:    notification,
	})
}

// GetAllNotifications handles getting all notifications with filtering and pagination
// @Summary Get all notifications
// @Description Get paginated list of notifications with optional filtering
// @Tags Notifications
// @Produce json
// @Param pageNumber query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param user_id query string false "Filter by user ID" format(uuid)
// @Param type query string false "Filter by notification type"
// @Param status query string false "Filter by notification status"
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications [get]
func (h *NotificationHandler) GetAllNotifications(c *gin.Context) {
	pageNumber := 1
	pageSize := 10

	if p, err := strconv.Atoi(c.DefaultQuery("pageNumber", "1")); err == nil && p > 0 {
		pageNumber = p
	}
	if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "10")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	filter := models.NotificationFilter{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID format"})
			return
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		notificationType := models.NotificationType(typeStr)
		filter.Type = &notificationType
	}

	if statusStr := c.Query("status"); statusStr != "" {
		notificationStatus := models.NotificationStatus(statusStr)
		filter.Status = &notificationStatus
	}

	response, err := h.notificationService.GetAllNotifications(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get notifications: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetNotificationsByUserID handles getting notifications for a specific user
// @Summary Get notifications by user ID
// @Description Get all notifications for a specific user
// @Tags Notifications
// @Produce json
// @Param user_id path string true "User ID" format(uuid)
// @Param pageNumber query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications/user/{user_id} [get]
func (h *NotificationHandler) GetNotificationsByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	pageNumber := 1
	pageSize := 10

	if p, err := strconv.Atoi(c.DefaultQuery("pageNumber", "1")); err == nil && p > 0 {
		pageNumber = p
	}
	if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "10")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	filter := models.NotificationFilter{
		PageNumber: pageNumber,
		PageSize:   pageSize,
		UserID:     &userID,
	}

	response, err := h.notificationService.GetAllNotifications(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get notifications: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}