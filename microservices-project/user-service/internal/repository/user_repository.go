package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"user-service/internal/models"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context, filter models.UserFilter) ([]models.User, int64, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
}

type userRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewUserRepository(db *gorm.DB, redis *redis.Client) UserRepository {
	return &userRepository{
		db:    db,
		redis: redis,
	}
}

// Cache keys
func userCacheKey(id uuid.UUID) string {
	return fmt.Sprintf("user:id:%s", id.String())
}

func userEmailCacheKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

func userUsernameCacheKey(username string) string {
	return fmt.Sprintf("user:username:%s", username)
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return err
	}

	// Cache the new user
	r.cacheUser(ctx, user)
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	cacheKey := userCacheKey(id)

	// Try to get from cache
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	// If not in cache, get from database
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Cache the result
	r.cacheUser(ctx, &user)
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	cacheKey := userEmailCacheKey(email)

	// Try to get from cache
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	// If not in cache, get from database
	var user models.User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}

	// Cache the result
	r.cacheUser(ctx, &user)
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	cacheKey := userUsernameCacheKey(username)

	// Try to get from cache
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var user models.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	// If not in cache, get from database
	var user models.User
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}

	// Cache the result
	r.cacheUser(ctx, &user)
	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, user)
	
	// Re-cache the updated user
	r.cacheUser(ctx, user)
	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Get user first to invalidate all cache entries
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.db.Delete(&models.User{}, id).Error; err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateUserCache(ctx, user)
	return nil
}

// GetAll retrieves all users with pagination and filters
func (r *userRepository) GetAll(ctx context.Context, filter models.UserFilter) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters
	if filter.Role != nil {
		query = query.Where("role = ?", *filter.Role)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("email ILIKE ? OR username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Count total records
	query.Count(&total)

	// Apply pagination
	offset := (filter.PageNumber - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	// Fetch users
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateLastLogin updates user's last login information
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_login_at":     now,
		"last_login_ip":     ip,
		"login_count":       gorm.Expr("login_count + ?", 1),
		"failed_login_count": 0,
	}

	if err := r.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return err
	}

	// Invalidate cache for this user
	r.redis.Del(ctx, userCacheKey(userID))
	return nil
}

// UpdatePassword updates user's password
func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"password":            hashedPassword,
		"password_changed_at": now,
		"password_reset_token": "",
		"password_reset_expiry": nil,
	}

	if err := r.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return err
	}

	// Invalidate cache for this user
	r.redis.Del(ctx, userCacheKey(userID))
	return nil
}

// Helper methods for caching
func (r *userRepository) cacheUser(ctx context.Context, user *models.User) {
	userData, err := json.Marshal(user)
	if err != nil {
		return
	}

	ttl := 15 * time.Minute

	// Cache by ID
	r.redis.Set(ctx, userCacheKey(user.ID), userData, ttl)
	
	// Cache by email
	r.redis.Set(ctx, userEmailCacheKey(user.Email), userData, ttl)
	
	// Cache by username
	r.redis.Set(ctx, userUsernameCacheKey(user.Username), userData, ttl)
}

func (r *userRepository) invalidateUserCache(ctx context.Context, user *models.User) {
	// Delete all cache entries for this user
	r.redis.Del(ctx, 
		userCacheKey(user.ID),
		userEmailCacheKey(user.Email),
		userUsernameCacheKey(user.Username),
	)
}