package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"notification-service/internal/models"
	"notification-service/internal/repository"
)

// NotificationService represents the notification service
type NotificationService struct {
	notificationRepo *repository.NotificationRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(notificationRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, req models.CreateNotificationRequest) error {
	// Set default max retries if not provided
	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}

	notification := &models.Notification{
		UserID:    req.UserID,
		Type:      req.Type,
		Subject:   req.Subject,
		Message:   req.Message,
		Template:  req.Template,
		Variables: req.Variables,
		MaxRetries: req.MaxRetries,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	log.Printf("Created notification: %s for user: %s", notification.ID, notification.UserID)

	// In a real implementation, you would trigger the notification processing
	// This could be done via a background worker or message queue
	go s.processNotification(context.Background(), notification.ID)

	return nil
}

// GetNotificationByID retrieves a notification by ID
func (s *NotificationService) GetNotificationByID(ctx context.Context, id string) (*models.NotificationResponse, error) {
	notificationID, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid notification ID: %w", err)
	}

	notification, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		return nil, err
	}

	return s.toNotificationResponse(notification), nil
}

// GetAllNotifications retrieves all notifications with filtering and pagination
func (s *NotificationService) GetAllNotifications(ctx context.Context, filter models.NotificationFilter) (*models.PaginatedResponse, error) {
	response, err := s.notificationRepo.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Convert notifications to responses
	if notifications, ok := response.Data.([]models.Notification); ok {
		var notificationResponses []models.NotificationResponse
		for _, notification := range notifications {
			notificationResponses = append(notificationResponses, *s.toNotificationResponse(&notification))
		}
		response.Data = notificationResponses
	}

	return response, nil
}

// processNotification processes a notification (sends it)
func (s *NotificationService) processNotification(ctx context.Context, notificationID string) {
	// Parse UUID
	id, err := parseUUID(notificationID)
	if err != nil {
		log.Printf("Invalid notification ID: %v", err)
		return
	}

	// Get notification
	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Failed to get notification: %v", err)
		return
	}

	// Check if already processed
	if notification.Status != models.NotificationStatusPending {
		return
	}

	// Send notification based on type
	var err2 error
	switch notification.Type {
	case models.NotificationTypeEmail:
		err2 = s.sendEmailNotification(ctx, notification)
	case models.NotificationTypeSMS:
		err2 = s.sendSMSNotification(ctx, notification)
	case models.NotificationTypePush:
		err2 = s.sendPushNotification(ctx, notification)
	default:
		err2 = fmt.Errorf("unknown notification type: %s", notification.Type)
	}

	// Update notification status
	updates := make(map[string]interface{})
	if err2 != nil {
		// Increment retry count
		notification.RetryCount++
		updates["retry_count"] = notification.RetryCount

		// Check if max retries reached
		if notification.RetryCount >= notification.MaxRetries {
			updates["status"] = models.NotificationStatusFailed
			updates["error"] = err2.Error()
			log.Printf("Notification failed after %d retries: %v", notification.RetryCount, err2)
		} else {
			// Schedule retry
			log.Printf("Notification failed, will retry (%d/%d): %v", notification.RetryCount, notification.MaxRetries, err2)
			// In a real implementation, you would schedule a retry
		}
	} else {
		// Success
		now := time.Now()
		updates["status"] = models.NotificationStatusSent
		updates["sent_at"] = &now
		log.Printf("Notification sent successfully: %s", notification.ID)
	}

	if err := s.notificationRepo.Update(ctx, notification.ID, updates); err != nil {
		log.Printf("Failed to update notification status: %v", err)
	}
}

// sendEmailNotification sends an email notification
func (s *NotificationService) sendEmailNotification(ctx context.Context, notification *models.Notification) error {
	// In a real implementation, you would integrate with an email service
	// like SendGrid, AWS SES, or SMTP server
	
	log.Printf("Sending email notification to user %s: %s", notification.UserID, notification.Subject)
	
	// Simulate email sending
	time.Sleep(100 * time.Millisecond)
	
	// Simulate occasional failure for testing
	if time.Now().UnixNano()%10 == 0 {
		return fmt.Errorf("simulated email sending failure")
	}
	
	return nil
}

// sendSMSNotification sends an SMS notification
func (s *NotificationService) sendSMSNotification(ctx context.Context, notification *models.Notification) error {
	// In a real implementation, you would integrate with an SMS service
	// like Twilio, AWS SNS, or other SMS providers
	
	log.Printf("Sending SMS notification to user %s: %s", notification.UserID, notification.Message)
	
	// Simulate SMS sending
	time.Sleep(50 * time.Millisecond)
	
	// Simulate occasional failure for testing
	if time.Now().UnixNano()%15 == 0 {
		return fmt.Errorf("simulated SMS sending failure")
	}
	
	return nil
}

// sendPushNotification sends a push notification
func (s *NotificationService) sendPushNotification(ctx context.Context, notification *models.Notification) error {
	// In a real implementation, you would integrate with push notification services
	// like Firebase Cloud Messaging, Apple Push Notification Service, etc.
	
	log.Printf("Sending push notification to user %s: %s", notification.UserID, notification.Subject)
	
	// Simulate push notification sending
	time.Sleep(30 * time.Millisecond)
	
	// Simulate occasional failure for testing
	if time.Now().UnixNano()%20 == 0 {
		return fmt.Errorf("simulated push notification failure")
	}
	
	return nil
}

// toNotificationResponse converts a Notification to NotificationResponse
func (s *NotificationService) toNotificationResponse(notification *models.Notification) *models.NotificationResponse {
	return &models.NotificationResponse{
		ID:          notification.ID,
		UserID:      notification.UserID,
		Type:        notification.Type,
		Subject:     notification.Subject,
		Message:     notification.Message,
		Template:    notification.Template,
		Variables:   notification.Variables,
		Status:      notification.Status,
		RetryCount:  notification.RetryCount,
		MaxRetries:  notification.MaxRetries,
		SentAt:      notification.SentAt,
		DeliveredAt: notification.DeliveredAt,
		Error:       notification.Error,
		CreatedAt:   notification.CreatedAt,
		UpdatedAt:   notification.UpdatedAt,
	}
}

// Helper function to parse UUID
func parseUUID(id string) (string, error) {
	// In a real implementation, you would use github.com/google/uuid
	// For now, we'll just return the string as is
	return id, nil
}