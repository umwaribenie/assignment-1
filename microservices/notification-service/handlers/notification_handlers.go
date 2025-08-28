package handlers

import (
	"net/http"
	"notification-service/cache"
	"notification-service/models"
	"notification-service/services"
	"shared"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	emailService *services.EmailService
	smsService   *services.SMSService
}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{
		emailService: services.NewEmailService(),
		smsService:   services.NewSMSService(),
	}
}

// SendEmail sends an email
func (nh *NotificationHandler) SendEmail(c *gin.Context) {
	var req models.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Error: "Invalid request data"})
		return
	}

	// Check rate limiting
	if !cache.CheckEmailRateLimit(req.To, 10, time.Hour) {
		c.JSON(http.StatusTooManyRequests, shared.ErrorResponse{Error: "Too many emails sent to this address"})
		return
	}

	// Send email
	err := nh.emailService.SendEmail(req.To, req.Subject, req.Content, req.IsHTML)
	if err != nil {
		c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Failed to send email"})
		return
	}

	// Create notification record (in a real implementation, you'd save this to a database)
	notification := models.Notification{
		ID:        uuid.New(),
		Type:      models.NotificationTypeEmail,
		Recipient: req.To,
		Subject:   req.Subject,
		Content:   req.Content,
		Status:    models.NotificationStatusSent,
		UserID:    req.UserID,
		Metadata:  req.Metadata,
		SentAt:    &[]time.Time{time.Now()}[0],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Cache the notification
	cache.SetCache(cache.NotificationCacheKey(notification.ID.String()), notification, 24*time.Hour)

	c.JSON(http.StatusOK, shared.APIResponse{
		Success: true,
		Message: "Email sent successfully",
		Data:    notification,
	})
}

// SendSMS sends an SMS
func (nh *NotificationHandler) SendSMS(c *gin.Context) {
	var req models.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Error: "Invalid request data"})
		return
	}

	// Check rate limiting
	if !cache.CheckSMSRateLimit(req.To, 5, time.Hour) {
		c.JSON(http.StatusTooManyRequests, shared.ErrorResponse{Error: "Too many SMS sent to this number"})
		return
	}

	// Send SMS
	err := nh.smsService.SendSMS(req.To, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Failed to send SMS"})
		return
	}

	// Create notification record
	notification := models.Notification{
		ID:        uuid.New(),
		Type:      models.NotificationTypeSMS,
		Recipient: req.To,
		Content:   req.Content,
		Status:    models.NotificationStatusSent,
		UserID:    req.UserID,
		Metadata:  req.Metadata,
		SentAt:    &[]time.Time{time.Now()}[0],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Cache the notification
	cache.SetCache(cache.NotificationCacheKey(notification.ID.String()), notification, 24*time.Hour)

	c.JSON(http.StatusOK, shared.APIResponse{
		Success: true,
		Message: "SMS sent successfully",
		Data:    notification,
	})
}

// SendTemplateEmail sends an email using a template
func (nh *NotificationHandler) SendTemplateEmail(c *gin.Context) {
	var req models.SendTemplateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, shared.ErrorResponse{Error: "Invalid request data"})
		return
	}

	// Check rate limiting
	if !cache.CheckEmailRateLimit(req.To, 10, time.Hour) {
		c.JSON(http.StatusTooManyRequests, shared.ErrorResponse{Error: "Too many emails sent to this address"})
		return
	}

	// Send template email
	err := nh.emailService.SendTemplateEmail(req.To, req.TemplateID, req.Variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, shared.ErrorResponse{Error: "Failed to send email: " + err.Error()})
		return
	}

	// Create notification record
	notification := models.Notification{
		ID:         uuid.New(),
		Type:       models.NotificationTypeEmail,
		Recipient:  req.To,
		Status:     models.NotificationStatusSent,
		UserID:     req.UserID,
		TemplateID: req.TemplateID,
		Metadata:   req.Metadata,
		SentAt:     &[]time.Time{time.Now()}[0],
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Cache the notification
	cache.SetCache(cache.NotificationCacheKey(notification.ID.String()), notification, 24*time.Hour)

	c.JSON(http.StatusOK, shared.APIResponse{
		Success: true,
		Message: "Template email sent successfully",
		Data:    notification,
	})
}

// GetEmailTemplates returns available email templates
func (nh *NotificationHandler) GetEmailTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, shared.APIResponse{
		Success: true,
		Data:    models.EmailTemplates,
	})
}

// GetSMSTemplates returns available SMS templates
func (nh *NotificationHandler) GetSMSTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, shared.APIResponse{
		Success: true,
		Data:    models.SMSTemplates,
	})
}

// GetNotification gets a notification by ID
func (nh *NotificationHandler) GetNotification(c *gin.Context) {
	notificationID := c.Param("id")

	// Try cache first
	var notification models.Notification
	if err := cache.GetCache(cache.NotificationCacheKey(notificationID), &notification); err != nil {
		c.JSON(http.StatusNotFound, shared.ErrorResponse{Error: "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, shared.APIResponse{
		Success: true,
		Data:    notification,
	})
}