package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"notification-service/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// NotificationRepository represents the notification repository
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (
			id, user_id, type, subject, message, template, variables, status,
			retry_count, max_retries, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	notification.ID = uuid.New()
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()
	notification.Status = models.NotificationStatusPending
	notification.RetryCount = 0

	// Convert variables map to JSON
	var variablesJSON []byte
	if notification.Variables != nil {
		var err error
		variablesJSON, err = json.Marshal(notification.Variables)
		if err != nil {
			return fmt.Errorf("failed to marshal variables: %w", err)
		}
	}

	_, err := r.db.ExecContext(ctx, query,
		notification.ID, notification.UserID, notification.Type, notification.Subject,
		notification.Message, notification.Template, variablesJSON, notification.Status,
		notification.RetryCount, notification.MaxRetries, notification.CreatedAt, notification.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT id, user_id, type, subject, message, template, variables, status,
			   retry_count, max_retries, sent_at, delivered_at, error, created_at, updated_at
		FROM notifications WHERE id = $1
	`

	notification := models.Notification{}
	var variablesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID, &notification.UserID, &notification.Type, &notification.Subject,
		&notification.Message, &notification.Template, &variablesJSON, &notification.Status,
		&notification.RetryCount, &notification.MaxRetries, &notification.SentAt,
		&notification.DeliveredAt, &notification.Error, &notification.CreatedAt, &notification.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Parse variables JSON
	if len(variablesJSON) > 0 {
		if err := json.Unmarshal(variablesJSON, &notification.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}
	}

	return &notification, nil
}

// Update updates a notification
func (r *NotificationRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now()

	var setClauses []string
	var args []interface{}
	argIndex := 1

	for field, value := range updates {
		if field == "variables" {
			// Handle variables map conversion
			if variables, ok := value.(map[string]string); ok {
				variablesJSON, err := json.Marshal(variables)
				if err != nil {
					return fmt.Errorf("failed to marshal variables: %w", err)
				}
				setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, variablesJSON)
				argIndex++
			}
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE notifications 
		SET %s 
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// GetAll retrieves all notifications with filtering and pagination
func (r *NotificationRepository) GetAll(ctx context.Context, filter models.NotificationFilter) (*models.PaginatedResponse, error) {
	baseQuery := `SELECT id, user_id, type, subject, message, template, variables, status,
						  retry_count, max_retries, sent_at, delivered_at, error, created_at, updated_at
				   FROM notifications`
	countQuery := "SELECT COUNT(*) FROM notifications"

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, string(*filter.Type))
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*filter.Status))
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Calculate pagination
	offset := (filter.PageNumber - 1) * filter.PageSize
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PageSize, offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		var variablesJSON []byte

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.Type, &notification.Subject,
			&notification.Message, &notification.Template, &variablesJSON, &notification.Status,
			&notification.RetryCount, &notification.MaxRetries, &notification.SentAt,
			&notification.DeliveredAt, &notification.Error, &notification.CreatedAt, &notification.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Parse variables JSON
		if len(variablesJSON) > 0 {
			if err := json.Unmarshal(variablesJSON, &notification.Variables); err != nil {
				return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
			}
		}

		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))

	response := models.PaginatedResponse{
		Data:       notifications,
		Total:      total,
		PageNumber: filter.PageNumber,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}

	return &response, nil
}

// GetPendingNotifications retrieves pending notifications for processing
func (r *NotificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]models.Notification, error) {
	query := `
		SELECT id, user_id, type, subject, message, template, variables, status,
			   retry_count, max_retries, sent_at, delivered_at, error, created_at, updated_at
		FROM notifications 
		WHERE status = $1 AND retry_count < max_retries
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, models.NotificationStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		var variablesJSON []byte

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.Type, &notification.Subject,
			&notification.Message, &notification.Template, &variablesJSON, &notification.Status,
			&notification.RetryCount, &notification.MaxRetries, &notification.SentAt,
			&notification.DeliveredAt, &notification.Error, &notification.CreatedAt, &notification.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Parse variables JSON
		if len(variablesJSON) > 0 {
			if err := json.Unmarshal(variablesJSON, &notification.Variables); err != nil {
				return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
			}
		}

		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}