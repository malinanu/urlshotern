package models

import "time"

// ============================================================================
// USER MANAGEMENT DTOs
// ============================================================================

// UserFilter represents filtering options for listing users
type UserFilter struct {
	Email       string
	Status      string // active, banned, suspended
	AccountType string
	Verified    *bool
	Search      string
	SortBy      string // created_at, email, last_login
	SortOrder   string // asc, desc
}

// AdminUpdateUserRequest represents a request to update user information
type AdminUpdateUserRequest struct {
	Email         *string `json:"email,omitempty"`
	Name          *string `json:"name,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
	PhoneVerified *bool   `json:"phone_verified,omitempty"`
	Status        *string `json:"status,omitempty"`
	AccountType   *string `json:"account_type,omitempty"`
}

// BanUserRequest represents a request to ban a user
type BanUserRequest struct {
	Reason   string `json:"reason" binding:"required"`
	Duration *int   `json:"duration_days,omitempty"` // nil = permanent
}

// UpdateUserLimitsRequest represents a request to update user rate limits
type UpdateUserLimitsRequest struct {
	MaxURLs        *int `json:"max_urls,omitempty"`
	MaxClicksMonth *int `json:"max_clicks_month,omitempty"`
	MaxAPIKeys     *int `json:"max_api_keys,omitempty"`
}

// UpdateUserPlanRequest represents a request to change user subscription plan
type UpdateUserPlanRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
	Reason string `json:"reason"`
}

// NotifyUserRequest represents a request to send notification to user
type NotifyUserRequest struct {
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
	Type    string `json:"type"` // email, sms, in-app
}

// UserActivityFilter represents filtering options for user activity
type UserActivityFilter struct {
	StartDate time.Time
	EndDate   time.Time
	Action    string // login, url_create, url_delete, etc.
	Limit     int
}

// UserWithStats extends User with usage statistics
type UserWithStats struct {
	ID            int64      `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	AccountType   string     `json:"account_type"`
	EmailVerified bool       `json:"email_verified"`
	PhoneVerified bool       `json:"phone_verified"`
	IsActive      bool       `json:"is_active"`
	IsAdmin       bool       `json:"is_admin"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	URLCount      int64      `json:"url_count"`
	TotalClicks   int64      `json:"total_clicks"`
}

// UserListResponse represents a paginated list of users
type UserListResponse struct {
	Users      []User `json:"users"`
	TotalCount int64  `json:"total_count"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}

// UserListWithStatsResponse represents a paginated list of users with usage statistics
type UserListWithStatsResponse struct {
	Users      []UserWithStats `json:"users"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// UserDetailedStats represents comprehensive statistics for a single user
type UserDetailedStats struct {
	User           User                 `json:"user"`
	Stats          UserStats            `json:"stats"`
	TopURLs        []URLWithClicks      `json:"top_urls"`
	RecentActivity []ActivityLog        `json:"recent_activity"`
	URLsOverTime   []TimeSeriesDataPoint `json:"urls_over_time"`
	ClicksOverTime []TimeSeriesDataPoint `json:"clicks_over_time"`
}

// UserStats represents aggregated statistics for a user
type UserStats struct {
	TotalURLs       int64   `json:"total_urls"`
	TotalClicks     int64   `json:"total_clicks"`
	AvgClicksPerURL float64 `json:"avg_clicks_per_url"`
	ActiveURLs      int64   `json:"active_urls"`
	ExpiredURLs     int64   `json:"expired_urls"`
	TotalRevenue    float64 `json:"total_revenue,omitempty"`
}

// URLWithClicks represents a URL with its click count
type URLWithClicks struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	ClickCount  int64     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ActivityLog represents a user activity entry
type ActivityLog struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type,omitempty"`
	ResourceID   *int64    `json:"resource_id,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ============================================================================
// URL MANAGEMENT DTOs
// ============================================================================

// URLFilter represents filtering options for listing URLs
type URLFilter struct {
	UserID      *int64
	Status      string // active, expired, flagged, suspended
	Flagged     *bool
	Search      string // search in original_url or short_code
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	SortBy      string // created_at, click_count, expires_at
	SortOrder   string // asc, desc
}

// UpdateURLStatusRequest represents a request to update URL status
type UpdateURLStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active suspended deleted"`
	Reason string `json:"reason"`
}

// TransferURLRequest represents a request to transfer URL ownership
type TransferURLRequest struct {
	NewOwnerID int64  `json:"new_owner_id" binding:"required"`
	Reason     string `json:"reason"`
}

// ExtendExpirationRequest represents a request to extend URL expiration
type ExtendExpirationRequest struct {
	ExpiresAt time.Time `json:"expires_at" binding:"required"`
}

// FlagURLRequest represents a request to flag a URL
type FlagURLRequest struct {
	Reason   string `json:"reason" binding:"required"`
	Category string `json:"category"` // spam, malware, phishing, copyright, other
}

// URLListResponse represents a paginated list of URLs
type URLListResponse struct {
	URLs       []URLMapping `json:"urls"`
	TotalCount int64        `json:"total_count"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

// ============================================================================
// DOMAIN MANAGEMENT DTOs
// ============================================================================

// BlockDomainRequest represents a request to block a domain
type BlockDomainRequest struct {
	Domain string `json:"domain" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

// BlockedDomain represents a blocked domain entry
type BlockedDomain struct {
	ID        int64     `json:"id" db:"id"`
	Domain    string    `json:"domain" db:"domain"`
	Reason    string    `json:"reason" db:"reason"`
	BlockedBy int64     `json:"blocked_by" db:"blocked_by"`
	BlockedAt time.Time `json:"blocked_at" db:"blocked_at"`
}

// BlockedDomainListResponse represents a paginated list of blocked domains
type BlockedDomainListResponse struct {
	Domains    []BlockedDomain `json:"domains"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// ============================================================================
// ANALYTICS DTOs
// ============================================================================

// AnalyticsDateRange represents a date range for analytics queries
type AnalyticsDateRange struct {
	StartDate time.Time `json:"start_date" form:"start_date"`
	EndDate   time.Time `json:"end_date" form:"end_date"`
}

// DashboardResponse represents the admin dashboard overview
type DashboardResponse struct {
	TotalUsers          int64   `json:"total_users"`
	ActiveUsers         int64   `json:"active_users"`
	NewUsersToday       int64   `json:"new_users_today"`
	NewUsersThisWeek    int64   `json:"new_users_this_week"`
	NewUsersThisMonth   int64   `json:"new_users_this_month"`
	TotalURLs           int64   `json:"total_urls"`
	ActiveURLs          int64   `json:"active_urls"`
	NewURLsToday        int64   `json:"new_urls_today"`
	TotalClicks         int64   `json:"total_clicks"`
	TodayClicks         int64   `json:"today_clicks"`
	ThisWeekClicks      int64   `json:"this_week_clicks"`
	ThisMonthClicks     int64   `json:"this_month_clicks"`
	TotalRevenue        float64 `json:"total_revenue"`
	MonthlyRevenue      float64 `json:"monthly_revenue"`
	ActiveSubscriptions int64   `json:"active_subscriptions"`
	TrialSubscriptions  int64   `json:"trial_subscriptions"`
	CancelledToday      int64   `json:"cancelled_today"`
}

// UserAnalyticsResponse represents user growth analytics
type UserAnalyticsResponse struct {
	TotalUsers       int64                    `json:"total_users"`
	GrowthRate       float64                  `json:"growth_rate"`
	RegistrationData []TimeSeriesDataPoint    `json:"registration_data"`
	ActiveUserData   []TimeSeriesDataPoint    `json:"active_user_data"`
	UsersByPlan      []UsersByPlanDataPoint   `json:"users_by_plan"`
	RetentionRate    float64                  `json:"retention_rate"`
}

// TimeSeriesDataPoint represents a data point in time series
type TimeSeriesDataPoint struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
	Value float64   `json:"value,omitempty"`
}

// UsersByPlanDataPoint represents user distribution by plan
type UsersByPlanDataPoint struct {
	Plan  string `json:"plan"`
	Count int64  `json:"count"`
}

// URLAnalyticsResponse represents URL creation and usage analytics
type URLAnalyticsResponse struct {
	TotalURLs        int64                 `json:"total_urls"`
	ActiveURLs       int64                 `json:"active_urls"`
	ExpiredURLs      int64                 `json:"expired_urls"`
	FlaggedURLs      int64                 `json:"flagged_urls"`
	CreationData     []TimeSeriesDataPoint `json:"creation_data"`
	AvgClicksPerURL  float64               `json:"avg_clicks_per_url"`
	TopDomains       []DomainStat          `json:"top_domains"`
}

// DomainStat represents statistics for a domain
type DomainStat struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}

// ExportReportRequest represents a request to export analytics report
type ExportReportRequest struct {
	ReportType string                 `json:"report_type" binding:"required"` // users, urls, clicks, revenue
	Format     string                 `json:"format" binding:"required,oneof=csv json xlsx"`
	StartDate  time.Time              `json:"start_date"`
	EndDate    time.Time              `json:"end_date"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// ============================================================================
// BILLING DTOs
// ============================================================================

// SubscriptionFilter represents filtering options for subscriptions
type SubscriptionFilter struct {
	UserID    *int64
	PlanType  string // free, premium, enterprise
	Status    string // active, cancelled, expired, trial
	Search    string
	SortBy    string // created_at, status, plan_type
	SortOrder string // asc, desc
}

// AdminUpdateSubscriptionRequest represents a request to update subscription
type AdminUpdateSubscriptionRequest struct {
	PlanType    *string    `json:"plan_type,omitempty"`
	Status      *string    `json:"status,omitempty"`
	TrialEnd    *time.Time `json:"trial_end,omitempty"`
	CancelAtEnd *bool      `json:"cancel_at_end,omitempty"`
}

// RefundSubscriptionRequest represents a request to refund a subscription
type RefundSubscriptionRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Reason string  `json:"reason" binding:"required"`
}

// CreatePlanRequest represents a request to create a subscription plan
type CreatePlanRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	Price       float64    `json:"price" binding:"required,gte=0"`
	Currency    string     `json:"currency" binding:"required"`
	Interval    string     `json:"interval" binding:"required,oneof=month year"`
	Features    []string   `json:"features"`
	Limits      PlanLimits `json:"limits"`
}

// PlanLimits represents resource limits for a plan
type PlanLimits struct {
	MaxURLs        int `json:"max_urls"`
	MaxClicksMonth int `json:"max_clicks_month"`
	MaxAPIKeys     int `json:"max_api_keys"`
	MaxTeamMembers int `json:"max_team_members"`
}

// UpdatePlanRequest represents a request to update a subscription plan
type UpdatePlanRequest struct {
	Name        *string     `json:"name,omitempty"`
	Description *string     `json:"description,omitempty"`
	Price       *float64    `json:"price,omitempty"`
	Features    *[]string   `json:"features,omitempty"`
	Limits      *PlanLimits `json:"limits,omitempty"`
	Active      *bool       `json:"active,omitempty"`
}

// SubscriptionListResponse represents a paginated list of subscriptions
type SubscriptionListResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
	TotalCount    int64          `json:"total_count"`
	Page          int            `json:"page"`
	PageSize      int            `json:"page_size"`
	TotalPages    int            `json:"total_pages"`
}

// Subscription represents a user subscription (placeholder - define properly based on billing system)
type Subscription struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	PlanType  string    `json:"plan_type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ============================================================================
// SETTINGS DTOs
// ============================================================================

// BulkUpdateSettingsRequest represents a request to update multiple settings
type BulkUpdateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// ResetSettingsRequest represents a request to reset settings to defaults
type ResetSettingsRequest struct {
	Category string `json:"category"` // empty = all categories
	Confirm  bool   `json:"confirm" binding:"required"`
}

// SettingsCategory represents a group of related settings
type SettingsCategory struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Settings    []GlobalSetting `json:"settings"`
}

// ============================================================================
// MONITORING DTOs
// ============================================================================

// AlertFilter represents filtering options for system alerts
type AlertFilter struct {
	Severity  string // info, warning, critical
	Status    string // active, acknowledged, resolved
	Service   string
	StartDate time.Time
	EndDate   time.Time
	SortBy    string // created_at, severity
	SortOrder string // asc, desc
}

// SystemHealthResponse represents overall system health status
type SystemHealthResponse struct {
	Status    string               `json:"status"` // healthy, degraded, unhealthy
	Timestamp time.Time            `json:"timestamp"`
	Uptime    int64                `json:"uptime_seconds"`
	Version   string               `json:"version"`
	Database  DatabaseHealthStatus `json:"database"`
	Redis     RedisHealthStatus    `json:"redis"`
	Services  map[string]ServiceStatus `json:"services"`
}

// DatabaseHealthStatus represents database health
type DatabaseHealthStatus struct {
	Status          string `json:"status"`
	ConnectionsUsed int    `json:"connections_used"`
	ConnectionsMax  int    `json:"connections_max"`
	ResponseTimeMs  int64  `json:"response_time_ms"`
}

// RedisHealthStatus represents Redis cache health
type RedisHealthStatus struct {
	Status         string `json:"status"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	MemoryUsedMB   int64  `json:"memory_used_mb"`
	KeyCount       int64  `json:"key_count"`
}

// ServiceStatus represents external service health
type ServiceStatus struct {
	Name           string    `json:"name"`
	Status         string    `json:"status"` // available, unavailable, degraded
	ResponseTimeMs int64     `json:"response_time_ms"`
	LastChecked    time.Time `json:"last_checked"`
	Message        string    `json:"message,omitempty"` // Optional error/status message
}

// SystemMetricsResponse represents system resource metrics
type SystemMetricsResponse struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsedMB       int64   `json:"memory_used_mb"`
	MemoryTotalMB      int64   `json:"memory_total_mb"`
	DiskUsedMB         int64   `json:"disk_used_mb"`
	DiskTotalMB        int64   `json:"disk_total_mb"`
	GoroutineCount     int     `json:"goroutine_count"`
	RequestsPerSecond  float64 `json:"requests_per_second"`
	AvgResponseTimeMs  float64 `json:"avg_response_time_ms"`
}

// Alert represents a system alert
type Alert struct {
	ID              int64     `json:"id" db:"id"`
	Severity        string    `json:"severity" db:"severity"`
	Service         string    `json:"service" db:"service"`
	Title           string    `json:"title" db:"title"`
	Message         string    `json:"message" db:"message"`
	Status          string    `json:"status" db:"status"`
	AcknowledgedBy  *int64    `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	AcknowledgedAt  *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	ResolvedBy      *int64    `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolutionNote  string    `json:"resolution_note,omitempty" db:"resolution_note"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// AcknowledgeAlertRequest represents a request to acknowledge an alert
type AcknowledgeAlertRequest struct {
	Note string `json:"note"`
}

// ResolveAlertRequest represents a request to resolve an alert
type ResolveAlertRequest struct {
	Resolution string `json:"resolution" binding:"required"`
	Note       string `json:"note"`
}

// AlertListResponse represents a paginated list of alerts
type AlertListResponse struct {
	Alerts     []Alert `json:"alerts"`
	TotalCount int64   `json:"total_count"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

// ============================================================================
// API KEY DTOs
// ============================================================================

// ListAllAPIKeysFilter represents filtering options for API keys
type ListAllAPIKeysFilter struct {
	UserID    *int64
	Active    *bool
	Expired   *bool
	Search    string
	SortBy    string // created_at, last_used_at
	SortOrder string // asc, desc
}

// CreateSystemAPIKeyRequest represents a request to create a system API key
type CreateSystemAPIKeyRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// APIKeyListResponse represents a paginated list of API keys
type APIKeyListResponse struct {
	APIKeys    []APIKey `json:"api_keys"`
	TotalCount int64    `json:"total_count"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}
