package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"microservices/user-service/internal/models"
	"microservices/user-service/pkg/utils"

	"gorm.io/gorm"
)

// UserRepository interface defines the methods for user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetBySlug(ctx context.Context, slug string) (*models.User, error)
	Update(ctx context.Context, id string, user *models.User) error
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context, filter models.UserFilter) ([]models.User, int64, error)
	UpdatePassword(ctx context.Context, id string, password string) error
}

// userRepository implements UserRepository
type userRepository struct {
	db    *gorm.DB
	cache *utils.CacheClient
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, cache *utils.CacheClient) UserRepository {
	return &userRepository{
		db:    db,
		cache: cache,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Invalidate cache after creating user
	r.invalidateUserCache(ctx, user.ID)
	r.invalidateUserListCache(ctx)

	return nil
}

// GetByID retrieves a user by ID with caching
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	// Try to get from cache first
	cacheKey := utils.GenerateUserCacheKey(id)
	var user models.User
	
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	// If not in cache, get from database
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Store in cache for 30 minutes
	r.cache.Set(ctx, cacheKey, user, 30*time.Minute)

	return &user, nil
}

// GetByEmail retrieves a user by email with caching
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	// Try to get from cache first
	cacheKey := utils.GenerateUserByEmailCacheKey(email)
	var user models.User
	
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	// If not in cache, get from database
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Store in cache for 30 minutes
	r.cache.Set(ctx, cacheKey, user, 30*time.Minute)

	return &user, nil
}

// GetByUsername retrieves a user by username with caching
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	// Try to get from cache first
	cacheKey := utils.GenerateUserByUsernameCacheKey(username)
	var user models.User
	
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	// If not in cache, get from database
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	// Store in cache for 30 minutes
	r.cache.Set(ctx, cacheKey, user, 30*time.Minute)

	return &user, nil
}

// GetBySlug retrieves a user by slug with caching
func (r *userRepository) GetBySlug(ctx context.Context, slug string) (*models.User, error) {
	// Try to get from cache first
	cacheKey := utils.GenerateUserBySlugCacheKey(slug)
	var user models.User
	
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	// If not in cache, get from database
	if err := r.db.First(&user, "slug = ?", slug).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by slug: %w", err)
	}

	// Store in cache for 30 minutes
	r.cache.Set(ctx, cacheKey, user, 30*time.Minute)

	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, id string, user *models.User) error {
	if err := r.db.Model(&models.User{}).Where("id = ?", id).Updates(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Invalidate cache after updating user
	r.invalidateUserCache(ctx, id)
	r.invalidateUserListCache(ctx)

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Invalidate cache after deleting user
	r.invalidateUserCache(ctx, id)
	r.invalidateUserListCache(ctx)

	return nil
}

// GetAll retrieves all users with pagination and filtering
func (r *userRepository) GetAll(ctx context.Context, filter models.UserFilter) ([]models.User, int64, error) {
	// Generate cache key based on filter
	filterJSON, _ := json.Marshal(filter)
	cacheKey := utils.GenerateUserListCacheKey(string(filterJSON))
	
	// Try to get from cache first
	var cachedResult struct {
		Users []models.User `json:"users"`
		Total int64         `json:"total"`
	}
	
	if err := r.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
		return cachedResult.Users, cachedResult.Total, nil
	}

	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters
	if filter.From != "" && filter.To != "" {
		query = query.Where("created_at BETWEEN ? AND ?", filter.From, filter.To)
	}
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR username LIKE ?", 
			search, search, search, search)
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Apply pagination
	if filter.PageSize == 0 {
		filter.PageSize = 10
	}
	if filter.PageNumber == 0 {
		filter.PageNumber = 1
	}
	offset := (filter.PageNumber - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	// Get users
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	// Store in cache for 15 minutes
	cachedResult.Users = users
	cachedResult.Total = total
	r.cache.Set(ctx, cacheKey, cachedResult, 15*time.Minute)

	return users, total, nil
}

// UpdatePassword updates a user's password
func (r *userRepository) UpdatePassword(ctx context.Context, id string, password string) error {
	if err := r.db.Model(&models.User{}).Where("id = ?", id).Update("password", password).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate cache after updating password
	r.invalidateUserCache(ctx, id)

	return nil
}

// invalidateUserCache invalidates user-specific cache entries
func (r *userRepository) invalidateUserCache(ctx context.Context, userID string) {
	r.cache.Delete(ctx, utils.GenerateUserCacheKey(userID))
}

// invalidateUserListCache invalidates user list cache entries
func (r *userRepository) invalidateUserListCache(ctx context.Context) {
	r.cache.DeletePattern(ctx, "users:list:*")
}