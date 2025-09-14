package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// StaticPage represents a CMS page
type StaticPage struct {
	ID                int64      `json:"id" db:"id"`
	Slug              string     `json:"slug" db:"slug"`
	Title             string     `json:"title" db:"title"`
	Content           string     `json:"content" db:"content"`
	ContentBlocks     *string    `json:"content_blocks,omitempty" db:"content_blocks"` // JSON for structured content
	MetaDescription   *string    `json:"meta_description,omitempty" db:"meta_description"`
	MetaKeywords      *string    `json:"meta_keywords,omitempty" db:"meta_keywords"`
	MetaImage         *string    `json:"meta_image,omitempty" db:"meta_image"`
	MetaTitle         *string    `json:"meta_title,omitempty" db:"meta_title"`
	CanonicalURL      *string    `json:"canonical_url,omitempty" db:"canonical_url"`
	SchemaMarkup      *string    `json:"schema_markup,omitempty" db:"schema_markup"`
	IsPublished       bool       `json:"is_published" db:"is_published"`
	SortOrder         int        `json:"sort_order" db:"sort_order"`
	Template          string     `json:"template" db:"template"`
	AuthorID          *int64     `json:"author_id,omitempty" db:"author_id"`
	AuthorName        string     `json:"author_name,omitempty"` // Joined field
	PublishedAt       *time.Time `json:"published_at,omitempty" db:"published_at"`
	ScheduledAt       *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	FeaturedImageID   *int64     `json:"featured_image_id,omitempty" db:"featured_image_id"`
	FeaturedImageURL  string     `json:"featured_image_url,omitempty"` // Joined field
	Category          *string    `json:"category,omitempty" db:"category"`
	Tags              *string    `json:"tags,omitempty" db:"tags"` // JSON array
	CustomFields      *string    `json:"custom_fields,omitempty" db:"custom_fields"` // JSON for additional fields
	ViewCount         int64      `json:"view_count" db:"view_count"`
	IsSticky          bool       `json:"is_sticky" db:"is_sticky"`
	AllowComments     bool       `json:"allow_comments" db:"allow_comments"`
	Language          string     `json:"language" db:"language"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateStaticPageRequest represents the request to create a new page
type CreateStaticPageRequest struct {
	Slug              string     `json:"slug" validate:"required,min=1,max=100"`
	Title             string     `json:"title" validate:"required,min=1,max=200"`
	Content           string     `json:"content" validate:"required"`
	ContentBlocks     *string    `json:"content_blocks,omitempty"`
	MetaDescription   *string    `json:"meta_description,omitempty"`
	MetaKeywords      *string    `json:"meta_keywords,omitempty"`
	MetaImage         *string    `json:"meta_image,omitempty"`
	MetaTitle         *string    `json:"meta_title,omitempty"`
	CanonicalURL      *string    `json:"canonical_url,omitempty"`
	SchemaMarkup      *string    `json:"schema_markup,omitempty"`
	IsPublished       bool       `json:"is_published"`
	SortOrder         int        `json:"sort_order"`
	Template          string     `json:"template"`
	ScheduledAt       *time.Time `json:"scheduled_at,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	FeaturedImageID   *int64     `json:"featured_image_id,omitempty"`
	Category          *string    `json:"category,omitempty"`
	Tags              *string    `json:"tags,omitempty"`
	CustomFields      *string    `json:"custom_fields,omitempty"`
	IsSticky          bool       `json:"is_sticky"`
	AllowComments     bool       `json:"allow_comments"`
	Language          string     `json:"language"`
}

// UpdateStaticPageRequest represents the request to update a page
type UpdateStaticPageRequest struct {
	Title           *string    `json:"title,omitempty"`
	Content         *string    `json:"content,omitempty"`
	ContentBlocks   *string    `json:"content_blocks,omitempty"`
	MetaDescription *string    `json:"meta_description,omitempty"`
	MetaKeywords    *string    `json:"meta_keywords,omitempty"`
	MetaImage       *string    `json:"meta_image,omitempty"`
	MetaTitle       *string    `json:"meta_title,omitempty"`
	CanonicalURL    *string    `json:"canonical_url,omitempty"`
	SchemaMarkup    *string    `json:"schema_markup,omitempty"`
	IsPublished     *bool      `json:"is_published,omitempty"`
	SortOrder       *int       `json:"sort_order,omitempty"`
	Template        *string    `json:"template,omitempty"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	FeaturedImageID *int64     `json:"featured_image_id,omitempty"`
	Category        *string    `json:"category,omitempty"`
	Tags            *string    `json:"tags,omitempty"`
	CustomFields    *string    `json:"custom_fields,omitempty"`
	IsSticky        *bool      `json:"is_sticky,omitempty"`
	AllowComments   *bool      `json:"allow_comments,omitempty"`
	Language        *string    `json:"language,omitempty"`
}

// PageRevision represents a page revision for version control
type PageRevision struct {
	ID              int64     `json:"id" db:"id"`
	PageID          int64     `json:"page_id" db:"page_id"`
	Title           string    `json:"title" db:"title"`
	Content         string    `json:"content" db:"content"`
	MetaDescription *string   `json:"meta_description,omitempty" db:"meta_description"`
	MetaKeywords    *string   `json:"meta_keywords,omitempty" db:"meta_keywords"`
	RevisionNumber  int       `json:"revision_number" db:"revision_number"`
	AuthorID        *int64    `json:"author_id,omitempty" db:"author_id"`
	AuthorName      string    `json:"author_name,omitempty"` // Joined field
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// PageVisit represents a page visit for analytics
type PageVisit struct {
	ID          int64     `json:"id" db:"id"`
	PageID      int64     `json:"page_id" db:"page_id"`
	IPAddress   *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   *string   `json:"user_agent,omitempty" db:"user_agent"`
	Referrer    *string   `json:"referrer,omitempty" db:"referrer"`
	CountryCode *string   `json:"country_code,omitempty" db:"country_code"`
	VisitedAt   time.Time `json:"visited_at" db:"visited_at"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID          int64         `json:"id" db:"id"`
	UserID      int64         `json:"user_id" db:"user_id"`
	KeyHash     string        `json:"-" db:"key_hash"` // Never expose in JSON
	KeyPrefix   string        `json:"key_prefix" db:"key_prefix"`
	Name        string        `json:"name" db:"name"`
	Description *string       `json:"description,omitempty" db:"description"`
	Permissions APIKeyPermissions `json:"permissions" db:"permissions"`
	RateLimit   int           `json:"rate_limit" db:"rate_limit"`
	LastUsedAt  *time.Time    `json:"last_used_at,omitempty" db:"last_used_at"`
	LastUsedIP  *string       `json:"last_used_ip,omitempty" db:"last_used_ip"`
	ExpiresAt   *time.Time    `json:"expires_at,omitempty" db:"expires_at"`
	IsActive    bool          `json:"is_active" db:"is_active"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// APIKeyPermissions represents the permissions for an API key
type APIKeyPermissions []string

// Value implements the driver.Valuer interface
func (p APIKeyPermissions) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface
func (p *APIKeyPermissions) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return nil
	}
}

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=100"`
	Description *string           `json:"description,omitempty"`
	Permissions APIKeyPermissions `json:"permissions"`
	RateLimit   int               `json:"rate_limit"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	APIKey    *APIKey `json:"api_key"`
	PlainKey  string  `json:"plain_key"` // Only returned once
	KeyPrefix string  `json:"key_prefix"`
}

// APIKeyUsage represents API key usage statistics
type APIKeyUsage struct {
	ID             int64     `json:"id" db:"id"`
	APIKeyID       int64     `json:"api_key_id" db:"api_key_id"`
	Endpoint       string    `json:"endpoint" db:"endpoint"`
	Method         string    `json:"method" db:"method"`
	StatusCode     *int      `json:"status_code,omitempty" db:"status_code"`
	ResponseTimeMs *int      `json:"response_time_ms,omitempty" db:"response_time_ms"`
	IPAddress      *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      *string   `json:"user_agent,omitempty" db:"user_agent"`
	RequestSize    *int      `json:"request_size,omitempty" db:"request_size"`
	ResponseSize   *int      `json:"response_size,omitempty" db:"response_size"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// APIKeyStats represents API key usage statistics
type APIKeyStats struct {
	APIKeyID       int64 `json:"api_key_id"`
	TotalRequests  int64 `json:"total_requests"`
	TodayRequests  int64 `json:"today_requests"`
	WeekRequests   int64 `json:"week_requests"`
	MonthRequests  int64 `json:"month_requests"`
	SuccessRate    float64 `json:"success_rate"`
	AvgResponseTime float64 `json:"avg_response_time"`
}

// PageAnalytics represents analytics for a static page
type PageAnalytics struct {
	PageID        int64 `json:"page_id"`
	TotalVisits   int64 `json:"total_visits"`
	TodayVisits   int64 `json:"today_visits"`
	WeekVisits    int64 `json:"week_visits"`
	MonthVisits   int64 `json:"month_visits"`
	UniqueVisitors int64 `json:"unique_visitors"`
	BounceRate     float64 `json:"bounce_rate"`
}

// ListPagesQuery represents query parameters for listing pages
type ListPagesQuery struct {
	Published   *bool   `json:"published,omitempty"`
	Template    *string `json:"template,omitempty"`
	AuthorID    *int64  `json:"author_id,omitempty"`
	Search      *string `json:"search,omitempty"`
	SortBy      string  `json:"sort_by"`
	SortOrder   string  `json:"sort_order"`
	Limit       int     `json:"limit"`
	Offset      int     `json:"offset"`
}

// ListAPIKeysQuery represents query parameters for listing API keys
type ListAPIKeysQuery struct {
	Active    *bool   `json:"active,omitempty"`
	Expired   *bool   `json:"expired,omitempty"`
	Search    *string `json:"search,omitempty"`
	SortBy    string  `json:"sort_by"`
	SortOrder string  `json:"sort_order"`
	Limit     int     `json:"limit"`
	Offset    int     `json:"offset"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
}

// MediaFile represents an uploaded media file
type MediaFile struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Filename    string    `json:"filename" db:"filename"`
	OriginalName string   `json:"original_name" db:"original_name"`
	FilePath    string    `json:"file_path" db:"file_path"`
	FileURL     string    `json:"file_url" db:"file_url"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	MimeType    string    `json:"mime_type" db:"mime_type"`
	FileType    string    `json:"file_type" db:"file_type"` // image, document, video, etc.
	Width       *int      `json:"width,omitempty" db:"width"`
	Height      *int      `json:"height,omitempty" db:"height"`
	Alt         *string   `json:"alt,omitempty" db:"alt"`
	Caption     *string   `json:"caption,omitempty" db:"caption"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// GlobalSetting represents site-wide configurable settings
type GlobalSetting struct {
	ID          int64     `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	Value       string    `json:"value" db:"value"`
	Type        string    `json:"type" db:"type"` // text, html, json, boolean, number
	Category    string    `json:"category" db:"category"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Description *string   `json:"description,omitempty" db:"description"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ContentBlock represents reusable content components
type ContentBlock struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Identifier  string    `json:"identifier" db:"identifier"`
	Content     string    `json:"content" db:"content"`
	Type        string    `json:"type" db:"type"` // hero, testimonial, cta, footer, etc.
	IsActive    bool      `json:"is_active" db:"is_active"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// PageTemplate represents available page templates
type PageTemplate struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description *string   `json:"description,omitempty" db:"description"`
	Content     string    `json:"content" db:"content"` // Template HTML with placeholders
	Fields      *string   `json:"fields,omitempty" db:"fields"` // JSON defining custom fields
	PreviewURL  *string   `json:"preview_url,omitempty" db:"preview_url"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// NavigationMenu represents dynamic site navigation
type NavigationMenu struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Location  string    `json:"location" db:"location"` // header, footer, sidebar, etc.
	Items     string    `json:"items" db:"items"` // JSON array of menu items
	IsActive  bool      `json:"is_active" db:"is_active"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// MediaUploadRequest represents media upload request
type MediaUploadRequest struct {
	Alt       *string `json:"alt,omitempty"`
	Caption   *string `json:"caption,omitempty"`
	IsPublic  bool    `json:"is_public"`
}

// GlobalSettingRequest represents global setting update request  
type GlobalSettingRequest struct {
	Key         string  `json:"key" validate:"required"`
	Value       string  `json:"value"`
	Type        string  `json:"type" validate:"required"`
	Category    string  `json:"category"`
	DisplayName string  `json:"display_name"`
	Description *string `json:"description,omitempty"`
	IsPublic    bool    `json:"is_public"`
	SortOrder   int     `json:"sort_order"`
}

// ContentBlockRequest represents content block request
type ContentBlockRequest struct {
	Name       string `json:"name" validate:"required"`
	Identifier string `json:"identifier" validate:"required"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	IsActive   bool   `json:"is_active"`
	SortOrder  int    `json:"sort_order"`
}

// Default permissions for API keys
var (
	DefaultAPIKeyPermissions = APIKeyPermissions{
		"url.create",
		"url.read",
		"url.update", 
		"url.delete",
		"analytics.read",
	}

	PremiumAPIKeyPermissions = APIKeyPermissions{
		"url.create",
		"url.read", 
		"url.update",
		"url.delete",
		"analytics.read",
		"domain.read",
		"domain.create",
		"domain.update",
		"domain.delete",
	}

	AdminAPIKeyPermissions = APIKeyPermissions{
		"*", // All permissions
	}
)