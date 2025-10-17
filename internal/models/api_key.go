package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key in the system
type APIKey struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	UserID      int64         `json:"user_id" db:"user_id"`
	KeyHash     string        `json:"-" db:"key_hash"` // Never expose the hash
	KeyPrefix   string        `json:"key_prefix" db:"key_prefix"`
	Name        string        `json:"name" db:"name"`
	Description *string       `json:"description" db:"description"`
	Permissions StringArray   `json:"permissions" db:"permissions"`
	RateLimit   int           `json:"rate_limit" db:"rate_limit"`
	LastUsedAt  *time.Time    `json:"last_used_at" db:"last_used_at"`
	LastUsedIP  *string       `json:"last_used_ip" db:"last_used_ip"`
	ExpiresAt   *time.Time    `json:"expires_at" db:"expires_at"`
	IsActive    bool          `json:"is_active" db:"is_active"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// StringArray is a custom type for handling PostgreSQL text arrays
type StringArray []string

// Value implements the driver.Valuer interface
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), s)
	}
	return json.Unmarshal(bytes, s)
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name        string      `json:"name" validate:"required,min=1,max=100"`
	Description *string     `json:"description" validate:"omitempty,max=500"`
	Permissions StringArray `json:"permissions" validate:"required,min=1"`
	RateLimit   int         `json:"rate_limit" validate:"omitempty,min=0,max=100000"`
	ExpiresAt   *time.Time  `json:"expires_at" validate:"omitempty"`
}

// CreateAPIKeyResponse represents the response when creating a new API key
type CreateAPIKeyResponse struct {
	APIKey    *APIKey `json:"api_key"`
	PlainKey  string  `json:"plain_key"`  // Only returned once during creation
	KeyPrefix string  `json:"key_prefix"`
}

// ListAPIKeysQuery represents query parameters for listing API keys
type ListAPIKeysQuery struct {
	Limit     int     `json:"limit"`
	Offset    int     `json:"offset"`
	Active    *bool   `json:"active"`
	Expired   *bool   `json:"expired"`
	Search    *string `json:"search"`
	SortBy    string  `json:"sort_by"`
	SortOrder string  `json:"sort_order"`
}

// APIKeyStats represents usage statistics for an API key
type APIKeyStats struct {
	APIKeyID        uuid.UUID `json:"api_key_id"`
	TotalRequests   int64     `json:"total_requests"`
	TodayRequests   int64     `json:"today_requests"`
	WeekRequests    int64     `json:"week_requests"`
	MonthRequests   int64     `json:"month_requests"`
	SuccessRate     float64   `json:"success_rate"`
	AvgResponseTime float64   `json:"avg_response_time"`
}

// DefaultAPIKeyPermissions contains the default permissions for a new API key
var DefaultAPIKeyPermissions = StringArray{
	"urls:create",
	"urls:read",
	"urls:update",
	"urls:delete",
	"analytics:read",
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
	HasMore bool        `json:"has_more"`
}
