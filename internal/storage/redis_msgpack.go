package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v5"
)

// RedisMsgpackStorage is an optimized Redis storage using MessagePack for serialization
type RedisMsgpackStorage struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisMsgpackStorage creates a new Redis storage with MessagePack serialization
func NewRedisMsgpackStorage(config *configs.Config) (*RedisMsgpackStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password:     config.RedisPassword,
		DB:           config.RedisDB,
		PoolSize:     50, // Optimized connection pool
		MinIdleConns: 10, // Maintain minimum idle connections
		MaxRetries:   3,  // Retry failed operations
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	ctx := context.Background()

	// Test the connection with timeout
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctxPing).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisMsgpackStorage{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (r *RedisMsgpackStorage) Close() error {
	return r.client.Close()
}

// SetURLMapping caches a URL mapping using MessagePack serialization
func (r *RedisMsgpackStorage) SetURLMapping(shortCode string, mapping *models.URLMapping, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(r.ctx, 3*time.Second)
	defer cancel()

	// Serialize using MessagePack (3-5x faster than JSON, 30-50% smaller)
	data, err := msgpack.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal URL mapping: %w", err)
	}

	key := fmt.Sprintf("url:%s", shortCode)
	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set URL mapping in cache: %w", err)
	}

	return nil
}

// GetURLMapping retrieves a URL mapping using MessagePack deserialization
func (r *RedisMsgpackStorage) GetURLMapping(shortCode string) (*models.URLMapping, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("url:%s", shortCode)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheKeyNotFound
		}
		return nil, fmt.Errorf("failed to get URL mapping from cache: %w", err)
	}

	var mapping models.URLMapping
	if err := msgpack.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("failed to unmarshal URL mapping: %w", err)
	}

	// Check if URL has expired
	if mapping.ExpiresAt != nil && mapping.ExpiresAt.Before(time.Now()) {
		// Remove expired entry from cache
		r.client.Del(ctx, key)
		return nil, ErrURLExpired
	}

	return &mapping, nil
}

// DeleteURLMapping removes a URL mapping from Redis cache
func (r *RedisMsgpackStorage) DeleteURLMapping(shortCode string) error {
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("url:%s", shortCode)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete URL mapping from cache: %w", err)
	}
	return nil
}

// IncrementClickCount increments the click count using Redis atomic operation
func (r *RedisMsgpackStorage) IncrementClickCount(shortCode string) (int64, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("clicks:%s", shortCode)

	// Use pipeline for better performance
	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 24*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment click count in cache: %w", err)
	}

	return incrCmd.Val(), nil
}

// GetClickCount gets the cached click count
func (r *RedisMsgpackStorage) GetClickCount(shortCode string) (int64, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("clicks:%s", shortCode)
	count, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrCacheKeyNotFound
		}
		return 0, fmt.Errorf("failed to get click count from cache: %w", err)
	}
	return count, nil
}

// SetAnalytics caches analytics data using MessagePack
func (r *RedisMsgpackStorage) SetAnalytics(shortCode string, analytics *models.AnalyticsResponse, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(r.ctx, 3*time.Second)
	defer cancel()

	data, err := msgpack.Marshal(analytics)
	if err != nil {
		return fmt.Errorf("failed to marshal analytics: %w", err)
	}

	key := fmt.Sprintf("analytics:%s", shortCode)
	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set analytics in cache: %w", err)
	}

	return nil
}

// GetAnalytics retrieves cached analytics data using MessagePack
func (r *RedisMsgpackStorage) GetAnalytics(shortCode string) (*models.AnalyticsResponse, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("analytics:%s", shortCode)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheKeyNotFound
		}
		return nil, fmt.Errorf("failed to get analytics from cache: %w", err)
	}

	var analytics models.AnalyticsResponse
	if err := msgpack.Unmarshal(data, &analytics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analytics: %w", err)
	}

	return &analytics, nil
}

// BatchGet retrieves multiple keys in a single operation (pipeline)
func (r *RedisMsgpackStorage) BatchGet(keys []string) (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 3*time.Second)
	defer cancel()

	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute batch get: %w", err)
	}

	result := make(map[string][]byte)
	for i, cmd := range cmds {
		val, err := cmd.Bytes()
		if err == nil {
			result[keys[i]] = val
		}
	}

	return result, nil
}

// BatchSet sets multiple keys in a single operation (pipeline)
func (r *RedisMsgpackStorage) BatchSet(items map[string][]byte, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	pipe := r.client.Pipeline()

	for key, value := range items {
		pipe.Set(ctx, key, value, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute batch set: %w", err)
	}

	return nil
}

// IsHealthy checks if Redis is healthy
func (r *RedisMsgpackStorage) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	return r.client.Ping(ctx).Err() == nil
}

// Get retrieves a value from Redis by key
func (r *RedisMsgpackStorage) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrCacheKeyNotFound
		}
		return "", fmt.Errorf("failed to get key from cache: %w", err)
	}
	return val, nil
}

// Set stores a key-value pair in Redis with TTL
func (r *RedisMsgpackStorage) Set(key string, value interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key in cache: %w", err)
	}
	return nil
}
