package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(config *configs.Config) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	ctx := context.Background()

	// Test the connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
}

// SetURLMapping caches a URL mapping in Redis with TTL
func (r *RedisStorage) SetURLMapping(shortCode string, mapping *models.URLMapping, ttl time.Duration) error {
	data, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal URL mapping: %w", err)
	}

	key := fmt.Sprintf("url:%s", shortCode)
	err = r.client.Set(r.ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set URL mapping in cache: %w", err)
	}

	return nil
}

// GetURLMapping retrieves a URL mapping from Redis cache
func (r *RedisStorage) GetURLMapping(shortCode string) (*models.URLMapping, error) {
	key := fmt.Sprintf("url:%s", shortCode)
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheKeyNotFound
		}
		return nil, fmt.Errorf("failed to get URL mapping from cache: %w", err)
	}

	var mapping models.URLMapping
	if err := json.Unmarshal([]byte(data), &mapping); err != nil {
		return nil, fmt.Errorf("failed to unmarshal URL mapping: %w", err)
	}

	// Check if URL has expired
	if mapping.ExpiresAt != nil && mapping.ExpiresAt.Before(time.Now()) {
		// Remove expired entry from cache
		r.client.Del(r.ctx, key)
		return nil, ErrURLExpired
	}

	return &mapping, nil
}

// DeleteURLMapping removes a URL mapping from Redis cache
func (r *RedisStorage) DeleteURLMapping(shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete URL mapping from cache: %w", err)
	}
	return nil
}

// IncrementClickCount increments the click count in cache
func (r *RedisStorage) IncrementClickCount(shortCode string) (int64, error) {
	key := fmt.Sprintf("clicks:%s", shortCode)
	count, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment click count in cache: %w", err)
	}

	// Set expiry for click counter (24 hours)
	r.client.Expire(r.ctx, key, 24*time.Hour)

	return count, nil
}

// GetClickCount gets the cached click count
func (r *RedisStorage) GetClickCount(shortCode string) (int64, error) {
	key := fmt.Sprintf("clicks:%s", shortCode)
	count, err := r.client.Get(r.ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrCacheKeyNotFound
		}
		return 0, fmt.Errorf("failed to get click count from cache: %w", err)
	}
	return count, nil
}

// SetRecentClicks caches recent click data for analytics
func (r *RedisStorage) SetRecentClicks(shortCode string, clicks []models.DailyClick, ttl time.Duration) error {
	data, err := json.Marshal(clicks)
	if err != nil {
		return fmt.Errorf("failed to marshal recent clicks: %w", err)
	}

	key := fmt.Sprintf("recent_clicks:%s", shortCode)
	err = r.client.Set(r.ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set recent clicks in cache: %w", err)
	}

	return nil
}

// GetRecentClicks retrieves cached recent click data
func (r *RedisStorage) GetRecentClicks(shortCode string) ([]models.DailyClick, error) {
	key := fmt.Sprintf("recent_clicks:%s", shortCode)
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheKeyNotFound
		}
		return nil, fmt.Errorf("failed to get recent clicks from cache: %w", err)
	}

	var clicks []models.DailyClick
	if err := json.Unmarshal([]byte(data), &clicks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recent clicks: %w", err)
	}

	return clicks, nil
}

// SetAnalytics caches analytics data
func (r *RedisStorage) SetAnalytics(shortCode string, analytics *models.AnalyticsResponse, ttl time.Duration) error {
	data, err := json.Marshal(analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	key := fmt.Sprintf("analytics:%s", shortCode)
	err = r.client.Set(r.ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set analytics in cache: %w", err)
	}

	return nil
}

// GetAnalytics retrieves cached analytics data
func (r *RedisStorage) GetAnalytics(shortCode string) (*models.AnalyticsResponse, error) {
	key := fmt.Sprintf("analytics:%s", shortCode)
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheKeyNotFound
		}
		return nil, fmt.Errorf("failed to get analytics from cache: %w", err)
	}

	var analytics models.AnalyticsResponse
	if err := json.Unmarshal([]byte(data), &analytics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analytics: %w", err)
	}

	return &analytics, nil
}

// IsHealthy checks if Redis is healthy
func (r *RedisStorage) IsHealthy() bool {
	return r.client.Ping(r.ctx).Err() == nil
}

// FlushDB clears all keys in the current database (use for testing)
func (r *RedisStorage) FlushDB() error {
	return r.client.FlushDB(r.ctx).Err()
}

// Get retrieves a value from Redis by key
func (r *RedisStorage) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrCacheKeyNotFound
		}
		return "", fmt.Errorf("failed to get key from cache: %w", err)
	}
	return val, nil
}

// Set stores a key-value pair in Redis with TTL
func (r *RedisStorage) Set(key string, value interface{}, ttl time.Duration) error {
	err := r.client.Set(r.ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key in cache: %w", err)
	}
	return nil
}

// Custom error for cache key not found
var ErrCacheKeyNotFound = &CacheError{Message: "cache key not found"}

type CacheError struct {
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}