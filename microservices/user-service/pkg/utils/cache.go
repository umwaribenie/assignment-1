package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheClient represents a Redis cache client
type CacheClient struct {
	client *redis.Client
}

// NewCacheClient creates a new Redis cache client
func NewCacheClient(addr, password string, db int) *CacheClient {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &CacheClient{
		client: client,
	}
}

// Set stores a key-value pair in Redis with optional expiration
func (c *CacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, jsonValue, expiration).Err()
}

// Get retrieves a value from Redis by key
func (c *CacheClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key from Redis
func (c *CacheClient) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (c *CacheClient) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists checks if a key exists in Redis
func (c *CacheClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return result > 0, nil
}

// SetNX sets a key-value pair only if the key doesn't exist
func (c *CacheClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.SetNX(ctx, key, jsonValue, expiration).Result()
}

// GenerateUserCacheKey generates a cache key for a specific user
func GenerateUserCacheKey(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

// GenerateUserListCacheKey generates a cache key for user list queries
func GenerateUserListCacheKey(filter string) string {
	return fmt.Sprintf("users:list:%s", filter)
}

// GenerateUserByEmailCacheKey generates a cache key for user by email
func GenerateUserByEmailCacheKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

// GenerateUserByUsernameCacheKey generates a cache key for user by username
func GenerateUserByUsernameCacheKey(username string) string {
	return fmt.Sprintf("user:username:%s", username)
}

// GenerateUserBySlugCacheKey generates a cache key for user by slug
func GenerateUserBySlugCacheKey(slug string) string {
	return fmt.Sprintf("user:slug:%s", slug)
}