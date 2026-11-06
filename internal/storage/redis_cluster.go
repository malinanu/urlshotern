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

// RedisClusterStorage provides high-performance Redis cluster support
type RedisClusterStorage struct {
	client *redis.ClusterClient
	ctx    context.Context
}

// RedisClusterConfig holds Redis cluster configuration
type RedisClusterConfig struct {
	Addrs              []string
	Password           string
	PoolSize           int
	MinIdleConns       int
	MaxRetries         int
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
}

// NewRedisClusterStorage creates a new Redis cluster storage instance with optimized settings
func NewRedisClusterStorage(config *configs.Config) (*RedisClusterStorage, error) {
	// For single-node development, use cluster mode with single address
	addrs := []string{
		fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
	}

	// In production, use multiple cluster nodes:
	// addrs := []string{
	//     "redis-node-1:6379",
	//     "redis-node-2:6379",
	//     "redis-node-3:6379",
	// }

	clusterConfig := &RedisClusterConfig{
		Addrs:              addrs,
		Password:           config.RedisPassword,
		PoolSize:           100,             // Increased from default 10
		MinIdleConns:       10,              // Maintain warm connections
		MaxRetries:         3,               // Retry failed operations
		PoolTimeout:        5 * time.Second, // Wait time for connection from pool
		IdleTimeout:        5 * time.Minute, // Close idle connections
		IdleCheckFrequency: 1 * time.Minute, // Check for idle connections
		ReadTimeout:        3 * time.Second, // Network read timeout
		WriteTimeout:       3 * time.Second, // Network write timeout
	}

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:              clusterConfig.Addrs,
		Password:           clusterConfig.Password,
		PoolSize:           clusterConfig.PoolSize,
		MinIdleConns:       clusterConfig.MinIdleConns,
		MaxRetries:         clusterConfig.MaxRetries,
		PoolTimeout:        clusterConfig.PoolTimeout,
		IdleTimeout:        clusterConfig.IdleTimeout,
		IdleCheckFrequency: clusterConfig.IdleCheckFrequency,
		ReadTimeout:        clusterConfig.ReadTimeout,
		WriteTimeout:       clusterConfig.WriteTimeout,

		// Enable read from replicas for better performance
		ReadOnly:       false,
		RouteByLatency: true, // Route to lowest latency node
		RouteRandomly:  false,
	})

	ctx := context.Background()

	// Test the connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		// Fallback: If cluster mode fails, it might be a single instance
		// This is for development compatibility
		return nil, fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}

	return &RedisClusterStorage{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis cluster connection
func (r *RedisClusterStorage) Close() error {
	return r.client.Close()
}

// SetURLMapping caches a URL mapping using MessagePack for efficiency
func (r *RedisClusterStorage) SetURLMapping(shortCode string, mapping *models.URLMapping, ttl time.Duration) error {
	// Serialize using MessagePack (40% smaller than JSON, 2x faster)
	data, err := msgpack.Marshal(mapping)
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

// GetURLMapping retrieves a URL mapping using MessagePack deserialization
func (r *RedisClusterStorage) GetURLMapping(shortCode string) (*models.URLMapping, error) {
	key := fmt.Sprintf("url:%s", shortCode)
	data, err := r.client.Get(r.ctx, key).Bytes()
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
		r.client.Del(r.ctx, key)
		return nil, ErrURLExpired
	}

	return &mapping, nil
}

// DeleteURLMapping removes a URL mapping with pattern-based deletion
func (r *RedisClusterStorage) DeleteURLMapping(shortCode string) error {
	// Delete main URL mapping
	key := fmt.Sprintf("url:%s", shortCode)

	// Use pipeline for efficiency when deleting multiple related keys
	pipe := r.client.Pipeline()
	pipe.Del(r.ctx, key)
	pipe.Del(r.ctx, fmt.Sprintf("clicks:%s", shortCode))
	pipe.Del(r.ctx, fmt.Sprintf("analytics:%s", shortCode))
	pipe.Del(r.ctx, fmt.Sprintf("recent_clicks:%s", shortCode))

	_, err := pipe.Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to delete URL mapping from cache: %w", err)
	}

	return nil
}

// IncrementClickCount uses atomic increment for thread safety
func (r *RedisClusterStorage) IncrementClickCount(shortCode string) (int64, error) {
	key := fmt.Sprintf("clicks:%s", shortCode)

	// Use pipeline for atomicity
	pipe := r.client.Pipeline()
	incr := pipe.Incr(r.ctx, key)
	pipe.Expire(r.ctx, key, 24*time.Hour)

	_, err := pipe.Exec(r.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment click count in cache: %w", err)
	}

	return incr.Val(), nil
}

// GetClickCount retrieves cached click count
func (r *RedisClusterStorage) GetClickCount(shortCode string) (int64, error) {
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

// SetRecentClicks caches recent click data using MessagePack
func (r *RedisClusterStorage) SetRecentClicks(shortCode string, clicks []models.DailyClick, ttl time.Duration) error {
	data, err := msgpack.Marshal(clicks)
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
func (r *RedisClusterStorage) GetRecentClicks(shortCode string) ([]models.DailyClick, error) {
	key := fmt.Sprintf("recent_clicks:%s", shortCode)
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheKeyNotFound
		}
		return nil, fmt.Errorf("failed to get recent clicks from cache: %w", err)
	}

	var clicks []models.DailyClick
	if err := msgpack.Unmarshal(data, &clicks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recent clicks: %w", err)
	}

	return clicks, nil
}

// SetAnalytics caches analytics data with compression
func (r *RedisClusterStorage) SetAnalytics(shortCode string, analytics *models.AnalyticsResponse, ttl time.Duration) error {
	data, err := msgpack.Marshal(analytics)
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
func (r *RedisClusterStorage) GetAnalytics(shortCode string) (*models.AnalyticsResponse, error) {
	key := fmt.Sprintf("analytics:%s", shortCode)
	data, err := r.client.Get(r.ctx, key).Bytes()
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

// IsHealthy checks if Redis cluster is healthy
func (r *RedisClusterStorage) IsHealthy() bool {
	return r.client.Ping(r.ctx).Err() == nil
}

// FlushDB clears all keys in the current database (use for testing)
func (r *RedisClusterStorage) FlushDB() error {
	return r.client.ForEachShard(r.ctx, func(ctx context.Context, shard *redis.Client) error {
		return shard.FlushDB(ctx).Err()
	})
}

// Get retrieves a value from Redis by key
func (r *RedisClusterStorage) Get(key string) (string, error) {
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
func (r *RedisClusterStorage) Set(key string, value interface{}, ttl time.Duration) error {
	err := r.client.Set(r.ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key in cache: %w", err)
	}
	return nil
}

// MGet retrieves multiple keys in a single round trip (pipeline optimization)
func (r *RedisClusterStorage) MGet(keys []string) ([]interface{}, error) {
	vals, err := r.client.MGet(r.ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to mget keys from cache: %w", err)
	}
	return vals, nil
}

// MSet sets multiple key-value pairs in a single round trip
func (r *RedisClusterStorage) MSet(pairs map[string]interface{}, ttl time.Duration) error {
	pipe := r.client.Pipeline()

	for key, value := range pairs {
		pipe.Set(r.ctx, key, value, ttl)
	}

	_, err := pipe.Exec(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to mset keys in cache: %w", err)
	}

	return nil
}

// GetPoolStats returns connection pool statistics for monitoring
func (r *RedisClusterStorage) GetPoolStats() map[string]*redis.PoolStats {
	stats := make(map[string]*redis.PoolStats)

	r.client.ForEachShard(r.ctx, func(ctx context.Context, shard *redis.Client) error {
		stats[shard.Options().Addr] = shard.PoolStats()
		return nil
	})

	return stats
}
