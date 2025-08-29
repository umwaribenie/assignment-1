package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"notification-service/internal/models"
	"notification-service/internal/repository"

	"gopkg.in/gomail.v2"
)

// NotificationService interface defines the business logic methods
type NotificationService interface {
	CreateNotification(ctx context.Context, req models.CreateNotificationRequest) (*models.NotificationResponse, error)
	GetNotificationByID(ctx context.Context, id string) (*models.NotificationResponse, error)
	GetAllNotifications(ctx context.Context, filter models.NotificationFilter) (*models.PaginatedResponse, error)
	GetNotificationsByUserID(ctx context.Context, userID string, filter models.NotificationFilter) (*models.PaginatedResponse, error)
	ProcessNotification(ctx context.Context, notification *models.Notification) error
	RetryFailedNotifications(ctx context.Context) error
}

// notificationService implements NotificationService
type notificationService struct {
	notificationRepo repository.NotificationRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(notificationRepo repository.NotificationRepository) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
	}
}

// CreateNotification creates a new notification
func (s *notificationService) CreateNotification(ctx context.Context, req models.CreateNotificationRequest) (*models.NotificationResponse, error) {
	// Create notification
	notification := &models.Notification{
		UserID:    req.UserID,
		Type:      req.Type,
		Subject:   req.Subject,
		Message:   req.Message,
		Recipient: req.Recipient,
		Template:  req.Template,
		Variables: models.JSONB(req.Variables),
		Status:    models.NotificationStatusPending,
		RetryCount: 0,
		MaxRetries: 3,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Process the notification asynchronously
	go func() {
		if err := s.ProcessNotification(context.Background(), notification); err != nil {
			log.Printf("Failed to process notification %s: %v", notification.ID, err)
		}
	}()

	return s.toNotificationResponse(notification), nil
}

// GetNotificationByID retrieves a notification by ID
func (s *notificationService) GetNotificationByID(ctx context.Context, id string) (*models.NotificationResponse, error) {
	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return s.toNotificationResponse(notification), nil
}

// GetAllNotifications retrieves all notifications with pagination and filtering
func (s *notificationService) GetAllNotifications(ctx context.Context, filter models.NotificationFilter) (*models.PaginatedResponse, error) {
	notifications, total, err := s.notificationRepo.GetAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// Convert to response format
	notificationResponses := make([]models.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		notificationResponses[i] = *s.toNotificationResponse(&notification)
	}

	// Calculate pagination
	lastPage := int(math.Ceil(float64(total) / float64(filter.PageSize)))
	if lastPage == 0 && total > 0 {
		lastPage = 1
	}

	var nextPage, previousPage *int
	if filter.PageNumber < lastPage {
		next := filter.PageNumber + 1
		nextPage = &next
	}
	if filter.PageNumber > 1 {
		prev := filter.PageNumber - 1
		previousPage = &prev
	}

	return &models.PaginatedResponse{
		CurrentPage:  filter.PageNumber,
		LastPage:     lastPage,
		List:         notificationResponses,
		NextPage:     nextPage,
		PreviousPage: previousPage,
		Status:       "success",
		Total:        total,
	}, nil
}

// GetNotificationsByUserID retrieves notifications for a specific user
func (s *notificationService) GetNotificationsByUserID(ctx context.Context, userID string, filter models.NotificationFilter) (*models.PaginatedResponse, error) {
	notifications, total, err := s.notificationRepo.GetByUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notifications: %w", err)
	}

	// Convert to response format
	notificationResponses := make([]models.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		notificationResponses[i] = *s.toNotificationResponse(&notification)
	}

	// Calculate pagination
	lastPage := int(math.Ceil(float64(total) / float64(filter.PageSize)))
	if lastPage == 0 && total > 0 {
		lastPage = 1
	}

	var nextPage, previousPage *int
	if filter.PageNumber < lastPage {
		next := filter.PageNumber + 1
		nextPage = &next
	}
	if filter.PageNumber > 1 {
		prev := filter.PageNumber - 1
		previousPage = &prev
	}

	return &models.PaginatedResponse{
		CurrentPage:  filter.PageNumber,
		LastPage:     lastPage,
		List:         notificationResponses,
		NextPage:     nextPage,
		PreviousPage: previousPage,
		Status:       "success",
		Total:        total,
	}, nil
}

// ProcessNotification processes a notification by sending it via the appropriate channel
func (s *notificationService) ProcessNotification(ctx context.Context, notification *models.Notification) error {
	var err error

	switch notification.Type {
	case models.NotificationTypeEmail:
		err = s.sendEmailNotification(notification)
	case models.NotificationTypeSMS:
		err = s.sendSMSNotification(notification)
	case models.NotificationTypePush:
		err = s.sendPushNotification(notification)
	default:
		err = fmt.Errorf("unsupported notification type: %s", notification.Type)
	}

	// Update notification status
	if err != nil {
		notification.Status = models.NotificationStatusFailed
		notification.RetryCount++
		notification.FailedAt = &time.Time{}
		notification.ErrorMsg = err.Error()
	} else {
		notification.Status = models.NotificationStatusSent
		notification.SentAt = &time.Time{}
	}

	// Update the notification in the database
	if updateErr := s.notificationRepo.Update(ctx, notification.ID, notification); updateErr != nil {
		log.Printf("Failed to update notification status: %v", updateErr)
	}

	return err
}

// RetryFailedNotifications retries failed notifications
func (s *notificationService) RetryFailedNotifications(ctx context.Context) error {
	notifications, err := s.notificationRepo.GetPendingNotifications(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	for _, notification := range notifications {
		if err := s.ProcessNotification(ctx, &notification); err != nil {
			log.Printf("Failed to retry notification %s: %v", notification.ID, err)
		}
	}

	return nil
}

// sendEmailNotification sends an email notification
func (s *notificationService) sendEmailNotification(notification *models.Notification) error {
	// Get SMTP configuration from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("FROM_EMAIL")

	if smtpHost == "" || smtpPortStr == "" || smtpUsername == "" || smtpPassword == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %w", err)
	}

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", notification.Recipient)
	m.SetHeader("Subject", notification.Subject)
	m.SetBody("text/html", notification.Message)

	// Create dialer
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully to %s", notification.Recipient)
	return nil
}

// sendSMSNotification sends an SMS notification
func (s *notificationService) sendSMSNotification(notification *models.Notification) error {
	// For now, we'll just log the SMS
	// In a real implementation, you would integrate with an SMS provider like Twilio
	log.Printf("SMS would be sent to %s: %s", notification.Recipient, notification.Message)
	
	// Simulate SMS sending
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// sendPushNotification sends a push notification
func (s *notificationService) sendPushNotification(notification *models.Notification) error {
	// For now, we'll just log the push notification
	// In a real implementation, you would integrate with FCM, APNS, etc.
	log.Printf("Push notification would be sent to %s: %s", notification.Recipient, notification.Message)
	
	// Simulate push notification sending
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// toNotificationResponse converts a Notification model to NotificationResponse
func (s *notificationService) toNotificationResponse(notification *models.Notification) *models.NotificationResponse {
	return &models.NotificationResponse{
		ID:          notification.ID,
		UserID:      notification.UserID,
		Type:        notification.Type,
		Subject:     notification.Subject,
		Message:     notification.Message,
		Recipient:   notification.Recipient,
		Status:      notification.Status,
		Template:    notification.Template,
		Variables:   map[string]interface{}(notification.Variables),
		RetryCount:  notification.RetryCount,
		MaxRetries:  notification.MaxRetries,
		SentAt:      notification.SentAt,
		FailedAt:    notification.FailedAt,
		ErrorMsg:    notification.ErrorMsg,
		CreatedAt:   notification.CreatedAt,
		UpdatedAt:   notification.UpdatedAt,
	}
}