package repository

import (
	"context"
	"fmt"
	"time"

	"notification-service/internal/models"

	"gorm.io/gorm"
)

// NotificationRepository interface defines the methods for notification data access
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id string) (*models.Notification, error)
	Update(ctx context.Context, id string, notification *models.Notification) error
	GetAll(ctx context.Context, filter models.NotificationFilter) ([]models.Notification, int64, error)
	GetPendingNotifications(ctx context.Context, limit int) ([]models.Notification, error)
	GetByUserID(ctx context.Context, userID string, filter models.NotificationFilter) ([]models.Notification, int64, error)
}

// notificationRepository implements NotificationRepository
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{
		db: db,
	}
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	if err := r.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// GetByID retrieves a notification by ID
func (r *notificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	var notification models.Notification
	if err := r.db.First(&notification, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get notification by ID: %w", err)
	}
	return &notification, nil
}

// Update updates a notification
func (r *notificationRepository) Update(ctx context.Context, id string, notification *models.Notification) error {
	if err := r.db.Model(&models.Notification{}).Where("id = ?", id).Updates(notification).Error; err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}
	return nil
}

// GetAll retrieves all notifications with pagination and filtering
func (r *notificationRepository) GetAll(ctx context.Context, filter models.NotificationFilter) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{})

	// Apply filters
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.From != "" && filter.To != "" {
		query = query.Where("created_at BETWEEN ? AND ?", filter.From, filter.To)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Apply pagination
	if filter.PageSize == 0 {
		filter.PageSize = 10
	}
	if filter.PageNumber == 0 {
		filter.PageNumber = 1
	}
	offset := (filter.PageNumber - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize).Order("created_at DESC")

	// Get notifications
	if err := query.Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}

	return notifications, total, nil
}

// GetPendingNotifications retrieves pending notifications for processing
func (r *notificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]models.Notification, error) {
	var notifications []models.Notification

	query := r.db.Where("status = ? AND retry_count < max_retries", models.NotificationStatusPending)
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("created_at ASC").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}

	return notifications, nil
}

// GetByUserID retrieves notifications for a specific user
func (r *notificationRepository) GetByUserID(ctx context.Context, userID string, filter models.NotificationFilter) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{}).Where("user_id = ?", userID)

	// Apply filters
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.From != "" && filter.To != "" {
		query = query.Where("created_at BETWEEN ? AND ?", filter.From, filter.To)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user notifications: %w", err)
	}

	// Apply pagination
	if filter.PageSize == 0 {
		filter.PageSize = 10
	}
	if filter.PageNumber == 0 {
		filter.PageNumber = 1
	}
	offset := (filter.PageNumber - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize).Order("created_at DESC")

	// Get notifications
	if err := query.Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user notifications: %w", err)
	}

	return notifications, total, nil
}