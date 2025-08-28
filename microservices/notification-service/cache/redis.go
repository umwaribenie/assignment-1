package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"notification-service/config"
	"time"

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

	log.Println("Notification Service connected to Redis successfully")
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

// Cache key generators for notifications
func NotificationCacheKey(notificationID string) string {
	return fmt.Sprintf("notification:%s", notificationID)
}

func UserNotificationsCacheKey(userID string) string {
	return fmt.Sprintf("user:notifications:%s", userID)
}

func TemplateCacheKey(templateID string) string {
	return fmt.Sprintf("template:%s", templateID)
}

// Rate limiting cache keys
func EmailRateLimitKey(email string) string {
	return fmt.Sprintf("rate_limit:email:%s", email)
}

func SMSRateLimitKey(phone string) string {
	return fmt.Sprintf("rate_limit:sms:%s", phone)
}

// Rate limiting functions
func CheckEmailRateLimit(email string, maxEmails int, window time.Duration) bool {
	key := EmailRateLimitKey(email)
	current, err := RedisClient.Incr(ctx, key).Result()
	if err != nil {
		return false
	}

	if current == 1 {
		RedisClient.Expire(ctx, key, window)
	}

	return current <= int64(maxEmails)
}

func CheckSMSRateLimit(phone string, maxSMS int, window time.Duration) bool {
	key := SMSRateLimitKey(phone)
	current, err := RedisClient.Incr(ctx, key).Result()
	if err != nil {
		return false
	}

	if current == 1 {
		RedisClient.Expire(ctx, key, window)
	}

	return current <= int64(maxSMS)
}