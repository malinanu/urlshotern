package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/google/uuid"
)

// Constants for API key statistics time intervals
const (
	statsQueryTimeout = 5 * time.Second // Max time for stats query execution
	dayInterval       = 1                // Current day
	weekInterval      = 7                // Last 7 days
	monthInterval     = 30               // Last 30 days
)

type APIKeyService struct {
	db *storage.PostgresStorage
}

func NewAPIKeyService(db *storage.PostgresStorage) *APIKeyService {
	return &APIKeyService{
		db: db,
	}
}

// CreateAPIKey creates a new API key for a user
func (s *APIKeyService) CreateAPIKey(userID int64, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	// Generate API key
	plainKey, keyHash, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create prefix for display
	keyPrefix := plainKey[:8] + "..." + plainKey[len(plainKey)-4:]

	// Set default values
	if req.RateLimit <= 0 {
		req.RateLimit = 1000 // Default 1000 requests per hour
	}
	if req.Permissions == nil || len(req.Permissions) == 0 {
		req.Permissions = models.DefaultAPIKeyPermissions
	}

	now := time.Now()
	newID := uuid.New()
	query := `
		INSERT INTO api_keys (id, user_id, key_hash, key_prefix, name, description, permissions,
			rate_limit, expires_at, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	var apiKey models.APIKey
	err = s.db.QueryRow(query,
		newID, userID, keyHash, keyPrefix, req.Name, req.Description, req.Permissions,
		req.RateLimit, req.ExpiresAt, true, now, now,
	).Scan(&apiKey.ID, &apiKey.CreatedAt, &apiKey.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Populate the API key with request data
	apiKey.UserID = userID
	apiKey.KeyHash = keyHash
	apiKey.KeyPrefix = keyPrefix
	apiKey.Name = req.Name
	apiKey.Description = req.Description
	apiKey.Permissions = req.Permissions
	apiKey.RateLimit = req.RateLimit
	apiKey.ExpiresAt = req.ExpiresAt
	apiKey.IsActive = true

	return &models.CreateAPIKeyResponse{
		APIKey:    &apiKey,
		PlainKey:  plainKey,
		KeyPrefix: keyPrefix,
	}, nil
}

// GetAPIKey retrieves an API key by ID
func (s *APIKeyService) GetAPIKey(id uuid.UUID, userID int64) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, key_prefix, name, description, permissions,
			rate_limit, last_used_at, last_used_ip, expires_at, is_active,
			created_at, updated_at
		FROM api_keys
		WHERE id = $1 AND user_id = $2`

	var apiKey models.APIKey
	err := s.db.QueryRow(query, id, userID).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.KeyPrefix,
		&apiKey.Name, &apiKey.Description, &apiKey.Permissions,
		&apiKey.RateLimit, &apiKey.LastUsedAt, &apiKey.LastUsedIP,
		&apiKey.ExpiresAt, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &apiKey, nil
}

// ValidateAPIKey validates an API key and returns the associated user ID
func (s *APIKeyService) ValidateAPIKey(key string) (*models.APIKey, error) {
	// Hash the provided key
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	query := `
		SELECT id, user_id, key_hash, key_prefix, name, description, permissions,
			rate_limit, last_used_at, last_used_ip, expires_at, is_active,
			created_at, updated_at
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true`

	var apiKey models.APIKey
	err := s.db.QueryRow(query, keyHash).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.KeyPrefix,
		&apiKey.Name, &apiKey.Description, &apiKey.Permissions,
		&apiKey.RateLimit, &apiKey.LastUsedAt, &apiKey.LastUsedIP,
		&apiKey.ExpiresAt, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}

	// Check if key is expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Use constant time comparison for security
	if subtle.ConstantTimeCompare([]byte(keyHash), []byte(apiKey.KeyHash)) != 1 {
		return nil, fmt.Errorf("invalid API key")
	}

	return &apiKey, nil
}

// UpdateAPIKeyUsage updates the last used timestamp and IP for an API key
func (s *APIKeyService) UpdateAPIKeyUsage(keyID uuid.UUID, ipAddress string) error {
	query := `
		UPDATE api_keys 
		SET last_used_at = $1, last_used_ip = $2, updated_at = $3
		WHERE id = $4`

	now := time.Now()
	_, err := s.db.Exec(query, now, ipAddress, now, keyID)
	if err != nil {
		return fmt.Errorf("failed to update API key usage: %w", err)
	}

	return nil
}

// RecordAPIKeyUsage records API key usage for analytics
func (s *APIKeyService) RecordAPIKeyUsage(keyID uuid.UUID, endpoint, method string, statusCode, responseTimeMs int, ipAddress, userAgent string) error {
	query := `
		INSERT INTO api_key_usage (api_key_id, endpoint, method, status_code, response_time_ms,
			ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := s.db.Exec(query,
		keyID, endpoint, method, statusCode, responseTimeMs,
		ipAddress, userAgent, time.Now(),
	)

	return err
}

// ListAPIKeys lists API keys for a user
func (s *APIKeyService) ListAPIKeys(userID int64, query *models.ListAPIKeysQuery) (*models.PaginatedResponse, error) {
	baseQuery := `
		SELECT id, user_id, key_hash, key_prefix, name, description, permissions,
			rate_limit, last_used_at, last_used_ip, expires_at, is_active,
			created_at, updated_at
		FROM api_keys
		WHERE user_id = $1`

	countQuery := `SELECT COUNT(*) FROM api_keys WHERE user_id = $1`

	// Build WHERE conditions
	var conditions []string
	var args []interface{}
	args = append(args, userID)
	argIndex := 2

	if query.Active != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *query.Active)
		argIndex++
	}
	if query.Expired != nil {
		if *query.Expired {
			conditions = append(conditions, fmt.Sprintf("expires_at IS NOT NULL AND expires_at < $%d", argIndex))
		} else {
			conditions = append(conditions, fmt.Sprintf("expires_at IS NULL OR expires_at > $%d", argIndex))
		}
		args = append(args, time.Now())
		argIndex++
	}
	if query.Search != nil && *query.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*query.Search+"%")
		argIndex++
	}

	// Add WHERE clause if needed
	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count API keys: %w", err)
	}

	// Add ORDER BY
	orderBy := "created_at DESC"
	if query.SortBy != "" {
		direction := "ASC"
		if query.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", query.SortBy, direction)
	}
	baseQuery += " ORDER BY " + orderBy

	// Add LIMIT and OFFSET
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, query.Limit, query.Offset)

	// Execute query
	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []models.APIKey
	for rows.Next() {
		var apiKey models.APIKey

		err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.KeyPrefix,
			&apiKey.Name, &apiKey.Description, &apiKey.Permissions,
			&apiKey.RateLimit, &apiKey.LastUsedAt, &apiKey.LastUsedIP,
			&apiKey.ExpiresAt, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return &models.PaginatedResponse{
		Data:    apiKeys,
		Total:   total,
		Limit:   query.Limit,
		Offset:  query.Offset,
		HasMore: query.Offset+query.Limit < int(total),
	}, nil
}

// RevokeAPIKey deactivates an API key
func (s *APIKeyService) RevokeAPIKey(id uuid.UUID, userID int64) error {
	query := `UPDATE api_keys SET is_active = false, updated_at = $1 WHERE id = $2 AND user_id = $3`
	result, err := s.db.Exec(query, time.Now(), id, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found or not owned by user")
	}

	return nil
}

// DeleteAPIKey permanently deletes an API key
func (s *APIKeyService) DeleteAPIKey(id uuid.UUID, userID int64) error {
	query := `DELETE FROM api_keys WHERE id = $1 AND user_id = $2`
	result, err := s.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found or not owned by user")
	}

	return nil
}

// GetAPIKeyStats gets usage statistics for an API key with production-ready features:
// - Context timeout to prevent long-running queries
// - Comprehensive error logging for debugging
// - Input validation
// - NULL-safe SQL with COALESCE
func (s *APIKeyService) GetAPIKeyStats(keyID uuid.UUID, userID int64) (*models.APIKeyStats, error) {
	// Input validation
	if keyID == uuid.Nil {
		log.Printf("[ERROR] GetAPIKeyStats: invalid API key ID (nil UUID) for user %d", userID)
		return nil, fmt.Errorf("invalid API key ID")
	}
	if userID <= 0 {
		log.Printf("[ERROR] GetAPIKeyStats: invalid user ID %d for API key %s", userID, keyID)
		return nil, fmt.Errorf("invalid user ID")
	}

	// Verify the key belongs to the user (authorization check)
	apiKey, err := s.GetAPIKey(keyID, userID)
	if err != nil {
		log.Printf("[WARN] GetAPIKeyStats: API key %s not found or not owned by user %d: %v", keyID, userID, err)
		return nil, fmt.Errorf("API key not found or access denied")
	}

	// Create context with timeout to prevent slow queries from blocking
	ctx, cancel := context.WithTimeout(context.Background(), statsQueryTimeout)
	defer cancel()

	// Note: COALESCE ensures 0.0 is returned instead of NULL for keys with no usage
	// This prevents Scan errors when an API key has never been used
	// Uses parameterized query to prevent SQL injection
	query := `
		SELECT
			COUNT(*) as total_requests,
			COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END) as today_requests,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '` + fmt.Sprintf("%d days", weekInterval) + `' THEN 1 END) as week_requests,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '` + fmt.Sprintf("%d days", monthInterval) + `' THEN 1 END) as month_requests,
			COALESCE(
				ROUND(
					COUNT(CASE WHEN status_code >= 200 AND status_code < 400 THEN 1 END) * 100.0 /
					NULLIF(COUNT(*), 0), 2
				), 0.0
			) as success_rate,
			COALESCE(ROUND(AVG(response_time_ms), 2), 0.0) as avg_response_time
		FROM api_key_usage
		WHERE api_key_id = $1`

	var stats models.APIKeyStats
	err = s.db.QueryRowContext(ctx, query, keyID).Scan(
		&stats.TotalRequests, &stats.TodayRequests, &stats.WeekRequests,
		&stats.MonthRequests, &stats.SuccessRate, &stats.AvgResponseTime,
	)
	if err != nil {
		// Check if context timeout occurred
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("[ERROR] GetAPIKeyStats: query timeout after %v for API key %s (user %d)",
				statsQueryTimeout, keyID, userID)
			return nil, fmt.Errorf("stats query timeout: please try again later")
		}

		// Log detailed error for debugging
		log.Printf("[ERROR] GetAPIKeyStats: database error for API key %s (user %d, name: %s): %v",
			keyID, userID, apiKey.Name, err)
		return nil, fmt.Errorf("failed to retrieve API key statistics: %w", err)
	}

	stats.APIKeyID = keyID

	// Log successful stats retrieval for audit/debugging
	log.Printf("[INFO] GetAPIKeyStats: retrieved stats for API key %s (user %d, name: %s): %d total requests",
		keyID, userID, apiKey.Name, stats.TotalRequests)

	return &stats, nil
}

// CheckAPIKeyPermission checks if an API key has a specific permission
func (s *APIKeyService) CheckAPIKeyPermission(apiKey *models.APIKey, permission string) bool {
	// Admin keys with "*" permission have access to everything
	for _, perm := range apiKey.Permissions {
		if perm == "*" || perm == permission {
			return true
		}
	}
	return false
}

// generateAPIKey generates a new API key and its hash
func (s *APIKeyService) generateAPIKey() (string, string, error) {
	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// Create key with prefix
	plainKey := "us_" + hex.EncodeToString(bytes)
	
	// Create hash for storage
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])

	return plainKey, keyHash, nil
}

// CleanupExpiredAPIKeys removes expired API keys
func (s *APIKeyService) CleanupExpiredAPIKeys() error {
	query := `DELETE FROM api_keys WHERE expires_at IS NOT NULL AND expires_at < $1`
	_, err := s.db.Exec(query, time.Now())
	return err
}

// ========== Admin Methods ==========

// AdminListAPIKeys lists all API keys across users with optional user filter
func (s *APIKeyService) AdminListAPIKeys(userID *int64, page, limit int) ([]models.APIKey, int64, error) {
	offset := (page - 1) * limit

	// Build base query
	baseQuery := `
		SELECT id, user_id, key_hash, key_prefix, name, description, permissions,
			rate_limit, last_used_at, last_used_ip, expires_at, is_active,
			created_at, updated_at
		FROM api_keys`
	countQuery := `SELECT COUNT(*) FROM api_keys`

	// Build WHERE clause
	var args []interface{}
	whereClause := ""
	if userID != nil {
		whereClause = " WHERE user_id = $1"
		args = append(args, *userID)
	}

	baseQuery += whereClause + " ORDER BY created_at DESC"
	countQuery += whereClause

	// Get total count
	var total int64
	if len(args) > 0 {
		err := s.db.QueryRow(countQuery, args...).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count API keys: %w", err)
		}
	} else {
		err := s.db.QueryRow(countQuery).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count API keys: %w", err)
		}
	}

	// Add pagination
	if len(args) > 0 {
		baseQuery += " LIMIT $2 OFFSET $3"
		args = append(args, limit, offset)
	} else {
		baseQuery += " LIMIT $1 OFFSET $2"
		args = append(args, limit, offset)
	}

	// Execute query
	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []models.APIKey
	for rows.Next() {
		var apiKey models.APIKey
		err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.KeyPrefix,
			&apiKey.Name, &apiKey.Description, &apiKey.Permissions,
			&apiKey.RateLimit, &apiKey.LastUsedAt, &apiKey.LastUsedIP,
			&apiKey.ExpiresAt, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan API key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, total, nil
}

// AdminGetAPIKey retrieves an API key by ID without user restriction
func (s *APIKeyService) AdminGetAPIKey(id uuid.UUID) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, key_prefix, name, description, permissions,
			rate_limit, last_used_at, last_used_ip, expires_at, is_active,
			created_at, updated_at
		FROM api_keys
		WHERE id = $1`

	var apiKey models.APIKey
	err := s.db.QueryRow(query, id).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.KeyHash, &apiKey.KeyPrefix,
		&apiKey.Name, &apiKey.Description, &apiKey.Permissions,
		&apiKey.RateLimit, &apiKey.LastUsedAt, &apiKey.LastUsedIP,
		&apiKey.ExpiresAt, &apiKey.IsActive, &apiKey.CreatedAt, &apiKey.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &apiKey, nil
}

// AdminRevokeAPIKey revokes an API key (admin version - no user check)
func (s *APIKeyService) AdminRevokeAPIKey(id uuid.UUID, reason string) error {
	query := `UPDATE api_keys SET is_active = false, updated_at = $1 WHERE id = $2`
	result, err := s.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	log.Printf("[INFO] Admin revoked API key %s. Reason: %s", id, reason)
	return nil
}

// AdminDeleteAPIKey permanently deletes an API key (admin version - no user check)
func (s *APIKeyService) AdminDeleteAPIKey(id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	log.Printf("[INFO] Admin deleted API key %s", id)
	return nil
}

// AdminGetAPIKeyUsage retrieves usage statistics for an API key (admin version)
func (s *APIKeyService) AdminGetAPIKeyUsage(keyID uuid.UUID, days int) (*models.APIKeyStats, error) {
	if keyID == uuid.Nil {
		return nil, fmt.Errorf("invalid API key ID")
	}

	// Verify the key exists
	_, err := s.AdminGetAPIKey(keyID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), statsQueryTimeout)
	defer cancel()

	// Query with custom day range
	query := `
		SELECT
			COUNT(*) as total_requests,
			COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END) as today_requests,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '` + fmt.Sprintf("%d days", days) + `' THEN 1 END) as custom_requests,
			COALESCE(
				ROUND(
					COUNT(CASE WHEN status_code >= 200 AND status_code < 400 THEN 1 END) * 100.0 /
					NULLIF(COUNT(*), 0), 2
				), 0.0
			) as success_rate,
			COALESCE(ROUND(AVG(response_time_ms), 2), 0.0) as avg_response_time
		FROM api_key_usage
		WHERE api_key_id = $1`

	var stats models.APIKeyStats
	var customRequests int64
	err = s.db.QueryRowContext(ctx, query, keyID).Scan(
		&stats.TotalRequests, &stats.TodayRequests, &customRequests,
		&stats.SuccessRate, &stats.AvgResponseTime,
	)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("stats query timeout")
		}
		return nil, fmt.Errorf("failed to retrieve API key statistics: %w", err)
	}

	stats.APIKeyID = keyID
	stats.WeekRequests = customRequests // Use custom period

	return &stats, nil
}

// AdminResetAPIKeyQuota resets rate limit quota for an API key
// In this implementation, we simply update last_used_at to reset the rate limit window
func (s *APIKeyService) AdminResetAPIKeyQuota(keyID uuid.UUID) error {
	// Verify key exists
	_, err := s.AdminGetAPIKey(keyID)
	if err != nil {
		return err
	}

	// Reset last_used_at to allow fresh quota
	// Note: Actual rate limiting is typically done in Redis/cache
	query := `UPDATE api_keys SET last_used_at = NULL, updated_at = $1 WHERE id = $2`
	_, err = s.db.Exec(query, time.Now(), keyID)
	if err != nil {
		return fmt.Errorf("failed to reset API key quota: %w", err)
	}

	log.Printf("[INFO] Admin reset quota for API key %s", keyID)
	return nil
}

// AdminCreateSystemAPIKey creates a system-level API key (no user association)
func (s *APIKeyService) AdminCreateSystemAPIKey(name, description string, permissions []string, rateLimit int) (*models.CreateAPIKeyResponse, error) {
	// Generate API key
	plainKey, keyHash, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create prefix for display
	keyPrefix := plainKey[:8] + "..." + plainKey[len(plainKey)-4:]

	// Set defaults
	if rateLimit <= 0 {
		rateLimit = 10000 // Higher limit for system keys
	}
	if len(permissions) == 0 {
		permissions = []string{"*"} // Full access for system keys
	}

	now := time.Now()
	newID := uuid.New()
	query := `
		INSERT INTO api_keys (id, user_id, key_hash, key_prefix, name, description, permissions,
			rate_limit, expires_at, is_active, created_at, updated_at)
		VALUES ($1, NULL, $2, $3, $4, $5, $6, $7, NULL, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	var apiKey models.APIKey
	err = s.db.QueryRow(query,
		newID, keyHash, keyPrefix, name, description, permissions,
		rateLimit, true, now, now,
	).Scan(&apiKey.ID, &apiKey.CreatedAt, &apiKey.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create system API key: %w", err)
	}

	// Populate the API key
	apiKey.KeyHash = keyHash
	apiKey.KeyPrefix = keyPrefix
	apiKey.Name = name
	if description != "" {
		apiKey.Description = &description
	}
	apiKey.Permissions = permissions
	apiKey.RateLimit = rateLimit
	apiKey.IsActive = true

	log.Printf("[INFO] Admin created system API key: %s (%s)", name, newID)

	return &models.CreateAPIKeyResponse{
		APIKey:    &apiKey,
		PlainKey:  plainKey,
		KeyPrefix: keyPrefix,
	}, nil
}