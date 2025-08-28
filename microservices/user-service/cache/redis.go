package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"user-service/config"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var ctx = context.Background()

// InitRedis initializes Redis connection
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.RedisURL,
		Password: config.AppConfig.RedisPassword,
		DB:       config.AppConfig.RedisDB,
	})

	// Test connection
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println("Connected to Redis successfully")
}

// CloseRedis closes Redis connection
func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
	}
}

// SetCache sets a value in cache with expiration
func SetCache(key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return RedisClient.Set(ctx, key, jsonValue, expiration).Err()
}

// GetCache gets a value from cache
func GetCache(key string, dest interface{}) error {
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// DeleteCache deletes a key from cache
func DeleteCache(key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// ExistsCache checks if a key exists in cache
func ExistsCache(key string) bool {
	val, err := RedisClient.Exists(ctx, key).Result()
	return err == nil && val > 0
}

// Cache key generators
func UserCacheKey(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

func UserSlugCacheKey(slug string) string {
	return fmt.Sprintf("user:slug:%s", slug)
}

func UserEmailCacheKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

func UsersListCacheKey(page, size int, filters string) string {
	return fmt.Sprintf("users:list:page:%d:size:%d:filters:%s", page, size, filters)
}