package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"user-service/internal/models"
	"user-service/pkg/utils"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UserRepository represents the user repository with caching
type UserRepository struct {
	db    *sql.DB
	cache *utils.CacheClient
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB, cache *utils.CacheClient) *UserRepository {
	return &UserRepository{
		db:    db,
		cache: cache,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, client_id, email, first_name, last_name, national_id, passport_number,
			password, phone, profile_picture, username, role, status, slug,
			referral_code, has_active_subscription, is_active, otp_required,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		)
	`

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.ClientID, user.Email, user.FirstName, user.LastName,
		user.NationalID, user.PassportNumber, user.Password, user.Phone,
		user.ProfilePicture, user.Username, user.Role, user.Status, user.Slug,
		user.ReferralCode, user.HasActiveSubscription, user.IsActive, user.OTPRequired,
		user.CreatedBy, user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, user.ID.String())
	r.invalidateUserListCache(ctx)

	return nil
}

// GetByID retrieves a user by ID with caching
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// Try to get from cache first
	cacheKey := utils.GenerateUserCacheKey(id.String())
	var user models.User
	
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	// If not in cache, get from database
	query := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number,
			   password, phone, profile_picture, username, role, status, slug,
			   referral_code, has_active_subscription, is_active, otp_required,
			   created_by, created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`

	user = models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Password, &user.Phone,
		&user.ProfilePicture, &user.Username, &user.Role, &user.Status, &user.Slug,
		&user.ReferralCode, &user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
		&user.CreatedBy, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Store in cache for 30 minutes
	r.cache.Set(ctx, cacheKey, user, 30*time.Minute)

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, client_id, email, first_name, last_name, national_id, passport_number,
			   password, phone, profile_picture, username, role, status, slug,
			   referral_code, has_active_subscription, is_active, otp_required,
			   created_by, created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`

	user := models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
		&user.NationalID, &user.PassportNumber, &user.Password, &user.Phone,
		&user.ProfilePicture, &user.Username, &user.Role, &user.Status, &user.Slug,
		&user.ReferralCode, &user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
		&user.CreatedBy, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now()

	var setClauses []string
	var args []interface{}
	argIndex := 1

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE users 
		SET %s 
		WHERE id = $%d AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, id.String())
	r.invalidateUserListCache(ctx)

	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users 
		SET deleted_at = $1, updated_at = $2 
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, id.String())
	r.invalidateUserListCache(ctx)

	return nil
}

// GetAll retrieves all users with filtering and pagination
func (r *UserRepository) GetAll(ctx context.Context, filter models.UserFilter) (*models.PaginatedResponse, error) {
	// Try to get from cache first
	cacheKey := utils.GenerateUserListCacheKey(fmt.Sprintf("%d-%d-%s-%s-%s", 
		filter.PageNumber, filter.PageSize, 
		stringValue(filter.Role), stringValue(filter.Status), stringValue(filter.Search)))
	
	var response models.PaginatedResponse
	if err := r.cache.Get(ctx, cacheKey, &response); err == nil {
		return &response, nil
	}

	baseQuery := `SELECT id, client_id, email, first_name, last_name, national_id, passport_number,
						  password, phone, profile_picture, username, role, status, slug,
						  referral_code, has_active_subscription, is_active, otp_required,
						  created_by, created_at, updated_at FROM users WHERE deleted_at IS NULL`
	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"

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
		searchPattern := "%" + *filter.Search + "%"
		conditions = append(conditions, fmt.Sprintf("(email ILIKE $%d OR username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", 
			argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Calculate pagination
	offset := (filter.PageNumber - 1) * filter.PageSize
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PageSize, offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.ClientID, &user.Email, &user.FirstName, &user.LastName,
			&user.NationalID, &user.PassportNumber, &user.Password, &user.Phone,
			&user.ProfilePicture, &user.Username, &user.Role, &user.Status, &user.Slug,
			&user.ReferralCode, &user.HasActiveSubscription, &user.IsActive, &user.OTPRequired,
			&user.CreatedBy, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))

	response = models.PaginatedResponse{
		Data:       users,
		Total:      total,
		PageNumber: filter.PageNumber,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}

	// Store in cache for 5 minutes
	r.cache.Set(ctx, cacheKey, response, 5*time.Minute)

	return &response, nil
}

// invalidateUserCache invalidates user-specific cache
func (r *UserRepository) invalidateUserCache(ctx context.Context, userID string) {
	cacheKey := utils.GenerateUserCacheKey(userID)
	r.cache.Delete(ctx, cacheKey)
}

// invalidateUserListCache invalidates user list cache
func (r *UserRepository) invalidateUserListCache(ctx context.Context) {
	pattern := utils.UserListCacheKey + ":*"
	r.cache.DeletePattern(ctx, pattern)
}

// Helper function to safely get string value from pointer
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}