package handlers

import (
	"net/http"
	"strconv"
	"time"

	"notification-service/internal/models"
	"notification-service/internal/service"

	"github.com/gin-gonic/gin"
)

// NotificationHandler handles HTTP requests for notification operations
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// CreateNotification handles notification creation
// @Summary Create a new notification
// @Description Creates a new notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body models.CreateNotificationRequest true "Notification data"
// @Success 201 {object} models.NotificationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications [post]
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	notification, err := h.notificationService.CreateNotification(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// GetNotificationByID handles getting a notification by ID
// @Summary Get notification by ID
// @Description Retrieves a notification by its ID
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} models.NotificationResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications/{id} [get]
func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "notification ID is required"})
		return
	}

	notification, err := h.notificationService.GetNotificationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "notification not found"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// GetAllNotifications handles getting all notifications with pagination and filtering
// @Summary Get all notifications
// @Description Retrieves all notifications with pagination and filtering
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param type query string false "Filter by type" Enums(email, sms, push)
// @Param status query string false "Filter by status" Enums(pending, sent, failed, cancelled)
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications [get]
func (h *NotificationHandler) GetAllNotifications(c *gin.Context) {
	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	notificationType := c.Query("type")
	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 10
	}

	filter := models.NotificationFilter{
		PageNumber: page,
		PageSize:   size,
		Type:       models.NotificationType(notificationType),
		Status:     models.NotificationStatus(status),
		From:       from,
		To:         to,
	}

	notifications, err := h.notificationService.GetAllNotifications(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// GetNotificationsByUserID handles getting notifications for a specific user
// @Summary Get notifications by user ID
// @Description Retrieves notifications for a specific user
// @Tags notifications
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param type query string false "Filter by type" Enums(email, sms, push)
// @Param status query string false "Filter by status" Enums(pending, sent, failed, cancelled)
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications/user/{userId} [get]
func (h *NotificationHandler) GetNotificationsByUserID(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "user ID is required"})
		return
	}

	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")
	notificationType := c.Query("type")
	status := c.Query("status")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 10
	}

	filter := models.NotificationFilter{
		PageNumber: page,
		PageSize:   size,
		Type:       models.NotificationType(notificationType),
		Status:     models.NotificationStatus(status),
	}

	notifications, err := h.notificationService.GetNotificationsByUserID(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// RetryFailedNotifications handles retrying failed notifications
// @Summary Retry failed notifications
// @Description Retries all failed notifications
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /notifications/retry [post]
func (h *NotificationHandler) RetryFailedNotifications(c *gin.Context) {
	err := h.notificationService.RetryFailedNotifications(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Failed notifications retry initiated"})
}

// HealthCheck handles health check requests
// @Summary Health check
// @Description Returns the health status of the notification service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *NotificationHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "notification-service",
		"timestamp": time.Now().UTC(),
	})
}