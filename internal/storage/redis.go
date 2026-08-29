package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client redis.UniversalClient
	ctx    context.Context
}

// NewRedisStorage creates a new Redis storage instance.
// Set REDIS_CLUSTER_ADDRS to a comma-separated list of nodes to enable cluster mode.
func NewRedisStorage(config *configs.Config) (*RedisStorage, error) {
	ctx := context.Background()

	var rdb redis.UniversalClient
	if config.RedisClusterAddrs != "" {
		// Cluster mode: split comma-separated addresses
		addrs := splitAddrs(config.RedisClusterAddrs)
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: config.RedisPassword,
		})
		log.Printf("🔗 Redis cluster mode: %v", addrs)
	} else {
		// Single-node mode (default)
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})
	}

	// Test the connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{
		client: rdb,
		ctx:    ctx,
	}, nil
}

func splitAddrs(s string) []string {
	var out []string
	for _, a := range strings.Split(s, ",") {
		a = strings.TrimSpace(a)
		if a != "" {
			out = append(out, a)
		}
	}
	return out
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

// Del deletes one or more keys from Redis
func (r *RedisStorage) Del(keys ...string) error {
	err := r.client.Del(r.ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete keys from cache: %w", err)
	}
	return nil
}

// BlacklistToken adds a token to the blacklist with TTL matching token expiry
func (r *RedisStorage) BlacklistToken(token string, expiresAt time.Time) error {
	key := fmt.Sprintf("blacklist:%s", token)
	ttl := time.Until(expiresAt)

	// Don't blacklist if already expired
	if ttl <= 0 {
		return nil
	}

	err := r.client.Set(r.ctx, key, "blacklisted", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

// IncrBotRateCheck increments and returns the click count for an IP+shortCode pair
// within a 60-second window. Used for bot rate detection.
func (r *RedisStorage) IncrBotRateCheck(ip, shortCode string) (int64, error) {
	key := fmt.Sprintf("bot_rate:%s:%s", ip, shortCode)
	count, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		r.client.Expire(r.ctx, key, 60*time.Second)
	}
	return count, nil
}

// CheckAndSetUniqueVisitor returns true if this is the first visit for the
// given ipHash+shortCode combination today. Sets the key on first visit.
func (r *RedisStorage) CheckAndSetUniqueVisitor(ipHash, shortCode string) (bool, error) {
	key := fmt.Sprintf("unique:%s:%s", shortCode, ipHash)
	set, err := r.client.SetNX(r.ctx, key, 1, 24*time.Hour).Result()
	if err != nil {
		return false, err
	}
	return set, nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *RedisStorage) IsTokenBlacklisted(token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	exists, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	return exists > 0, nil
}

// SetForceLogout marks that all tokens for userID issued before now are invalid.
func (r *RedisStorage) SetForceLogout(userID int64) error {
	key := fmt.Sprintf("force_logout:%d", userID)
	return r.client.Set(r.ctx, key, time.Now().Unix(), 24*time.Hour).Err()
}

// GetForceLogout returns the force-logout timestamp for userID, or zero time if not set.
func (r *RedisStorage) GetForceLogout(userID int64) (time.Time, error) {
	key := fmt.Sprintf("force_logout:%d", userID)
	val, err := r.client.Get(r.ctx, key).Int64()
	if err == redis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(val, 0), nil
}

// LPush pushes a value onto a list
func (r *RedisStorage) LPush(key string, value interface{}) error {
	return r.client.LPush(r.ctx, key, value).Err()
}

// RPop pops a value from a list
func (r *RedisStorage) RPop(key string) (string, error) {
	return r.client.RPop(r.ctx, key).Result()
}