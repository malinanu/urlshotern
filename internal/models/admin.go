package models

import "time"

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalUsers       int64  `json:"total_users"`
	ActiveUsers      int64  `json:"active_users"`
	TotalURLs        int64  `json:"total_urls"`
	TotalClicks      int64  `json:"total_clicks"`
	StorageUsed      string `json:"storage_used"`       // e.g., "1.2 GB"
	DailySignups     int64  `json:"daily_signups"`
	DailyURLsCreated int64  `json:"daily_urls_created"`
	SystemStatus     string `json:"system_status"` // "healthy", "warning", "error"
}

// AdminUser represents user data for admin panel
type AdminUser struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	IsActive        bool       `json:"is_active"`
	IsEmailVerified bool       `json:"is_email_verified"`
	AccountType     string     `json:"account_type"`
	CreatedAt       time.Time  `json:"created_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	URLsCount       int64      `json:"urls_count"`
	TotalClicks     int64      `json:"total_clicks"`
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Users  []*AdminUser `json:"users"`
	Total  int64        `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

// RecentActivity represents an activity log entry
type RecentActivity struct {
	ID           int64      `json:"id"`
	UserID       *int64     `json:"user_id,omitempty"`
	UserName     string     `json:"user_name"`
	ActivityType string     `json:"activity_type"`
	Description  string     `json:"description"`
	CreatedAt    time.Time  `json:"created_at"`
	IPAddress    *string    `json:"ip_address,omitempty"`
}

// CMS Models

// CMSPage represents a CMS page for admin panel
type CMSPage struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	IsPublished bool       `json:"is_published"`
	AuthorID    *int64     `json:"author_id,omitempty"`
	ViewCount   int64      `json:"view_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CMSBlock represents a content block for admin panel
type CMSBlock struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Identifier string    `json:"identifier"`
	Type       string    `json:"type"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CMSTemplate represents a page template for admin panel
type CMSTemplate struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
