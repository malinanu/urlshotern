package models

import (
	"time"
)

// UserActivityLog represents a user activity log entry
type UserActivityLog struct {
	ID            int64                  `json:"id" db:"id"`
	UserID        int64                  `json:"user_id" db:"user_id"`
	SessionID     string                 `json:"session_id" db:"session_id"`
	ActivityType  string                 `json:"activity_type" db:"activity_type"`
	ResourceType  string                 `json:"resource_type" db:"resource_type"` // url, user, team, domain, etc.
	ResourceID    *string                `json:"resource_id,omitempty" db:"resource_id"`
	Action        string                 `json:"action" db:"action"` // create, read, update, delete, share, etc.
	Description   string                 `json:"description" db:"description"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"` // Additional context data
	IPAddress     *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string                `json:"user_agent,omitempty" db:"user_agent"`
	Referrer      *string                `json:"referrer,omitempty" db:"referrer"`
	Location      *string                `json:"location,omitempty" db:"location"` // Country/city
	DeviceType    *string                `json:"device_type,omitempty" db:"device_type"`
	Browser       *string                `json:"browser,omitempty" db:"browser"`
	OS            *string                `json:"os,omitempty" db:"os"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// UserActivityInfo represents detailed user activity information
type UserActivityInfo struct {
	*UserActivityLog
	User         *PublicUser `json:"user,omitempty"`
	ResourceName *string     `json:"resource_name,omitempty"`
}

// UserSession represents a user session for analytics
type UserSession struct {
	ID              string                 `json:"id" db:"id"`
	UserID          *int64                 `json:"user_id,omitempty" db:"user_id"`
	StartTime       time.Time              `json:"start_time" db:"start_time"`
	EndTime         *time.Time             `json:"end_time,omitempty" db:"end_time"`
	Duration        *int64                 `json:"duration_seconds,omitempty" db:"duration_seconds"`
	ActivityCount   int                    `json:"activity_count" db:"activity_count"`
	URLsCreated     int                    `json:"urls_created" db:"urls_created"`
	URLsClicked     int                    `json:"urls_clicked" db:"urls_clicked"`
	IPAddress       *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent       *string                `json:"user_agent,omitempty" db:"user_agent"`
	DeviceType      *string                `json:"device_type,omitempty" db:"device_type"`
	Browser         *string                `json:"browser,omitempty" db:"browser"`
	OS              *string                `json:"os,omitempty" db:"os"`
	Location        *string                `json:"location,omitempty" db:"location"`
	Referrer        *string                `json:"referrer,omitempty" db:"referrer"`
	ExitPage        *string                `json:"exit_page,omitempty" db:"exit_page"`
	SessionData     map[string]interface{} `json:"session_data,omitempty" db:"session_data"`
	IsActive        bool                   `json:"is_active" db:"is_active"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// UserAnalyticsSummary represents comprehensive user analytics
type UserAnalyticsSummary struct {
	UserID              int64                    `json:"user_id"`
	AccountCreatedAt    time.Time                `json:"account_created_at"`
	LastActiveAt        *time.Time               `json:"last_active_at"`
	TotalSessions       int64                    `json:"total_sessions"`
	TotalActiveDays     int64                    `json:"total_active_days"`
	AvgSessionDuration  float64                  `json:"avg_session_duration_minutes"`
	TotalURLsCreated    int64                    `json:"total_urls_created"`
	TotalClicks         int64                    `json:"total_clicks"`
	TotalShares         int64                    `json:"total_shares"`
	TotalComments       int64                    `json:"total_comments"`
	TotalBookmarks      int64                    `json:"total_bookmarks"`
	TotalCollections    int64                    `json:"total_collections"`
	TopDevices          []DeviceUsageStat        `json:"top_devices"`
	TopBrowsers         []BrowserUsageStat       `json:"top_browsers"`
	TopLocations        []LocationUsageStat      `json:"top_locations"`
	ActivityTrends      []ActivityTrend          `json:"activity_trends"`
	RecentActivity      []*UserActivityInfo      `json:"recent_activity"`
	UsagePatterns       UserUsagePatterns        `json:"usage_patterns"`
	EngagementMetrics   UserEngagementMetrics    `json:"engagement_metrics"`
}

// DeviceUsageStat represents device usage statistics
type DeviceUsageStat struct {
	DeviceType   string  `json:"device_type"`
	Sessions     int64   `json:"sessions"`
	Duration     int64   `json:"total_duration_minutes"`
	Percentage   float64 `json:"percentage"`
}

// BrowserUsageStat represents browser usage statistics
type BrowserUsageStat struct {
	Browser      string  `json:"browser"`
	Sessions     int64   `json:"sessions"`
	Duration     int64   `json:"total_duration_minutes"`
	Percentage   float64 `json:"percentage"`
}

// LocationUsageStat represents location usage statistics
type LocationUsageStat struct {
	Location     string  `json:"location"`
	Sessions     int64   `json:"sessions"`
	Duration     int64   `json:"total_duration_minutes"`
	Percentage   float64 `json:"percentage"`
}

// ActivityTrend represents activity trends over time
type ActivityTrend struct {
	Date            string `json:"date"`
	Sessions        int64  `json:"sessions"`
	URLsCreated     int64  `json:"urls_created"`
	TotalClicks     int64  `json:"total_clicks"`
	ActiveDuration  int64  `json:"active_duration_minutes"`
}

// UserUsagePatterns represents user behavior patterns
type UserUsagePatterns struct {
	PeakHours       []HourUsage     `json:"peak_hours"`
	PeakDays        []DayUsage      `json:"peak_days"`
	AvgSessionTime  float64         `json:"avg_session_time_minutes"`
	AvgURLsPerDay   float64         `json:"avg_urls_per_day"`
	MostActiveHour  int             `json:"most_active_hour"`
	MostActiveDay   string          `json:"most_active_day"`
	PreferredDevice string          `json:"preferred_device"`
	PreferredBrowser string         `json:"preferred_browser"`
}

// HourUsage represents usage by hour
type HourUsage struct {
	Hour            int     `json:"hour"`
	Sessions        int64   `json:"sessions"`
	ActivityCount   int64   `json:"activity_count"`
	AvgDuration     float64 `json:"avg_duration_minutes"`
}

// DayUsage represents usage by day of week
type DayUsage struct {
	Day             string  `json:"day"`
	Weekday         int     `json:"weekday"` // 0-6
	Sessions        int64   `json:"sessions"`
	ActivityCount   int64   `json:"activity_count"`
	AvgDuration     float64 `json:"avg_duration_minutes"`
}

// UserEngagementMetrics represents user engagement metrics
type UserEngagementMetrics struct {
	RetentionRate         float64 `json:"retention_rate_percentage"`
	EngagementScore       float64 `json:"engagement_score"` // 0-100
	DaysActive            int64   `json:"days_active"`
	DaysInactive          int64   `json:"days_inactive"`
	LongestStreak         int64   `json:"longest_active_streak_days"`
	CurrentStreak         int64   `json:"current_active_streak_days"`
	AvgActionsPerSession  float64 `json:"avg_actions_per_session"`
	FeatureAdoptionRate   float64 `json:"feature_adoption_rate_percentage"`
	CollaborationRate     float64 `json:"collaboration_rate_percentage"`
	URLSuccessRate        float64 `json:"url_success_rate_percentage"`
}

// UserGoal represents user-defined goals and tracking
type UserGoal struct {
	ID              int64      `json:"id" db:"id"`
	UserID          int64      `json:"user_id" db:"user_id"`
	GoalType        string     `json:"goal_type" db:"goal_type"` // url_creation, clicks, shares, etc.
	TargetValue     int64      `json:"target_value" db:"target_value"`
	CurrentValue    int64      `json:"current_value" db:"current_value"`
	TargetDate      *time.Time `json:"target_date,omitempty" db:"target_date"`
	Description     *string    `json:"description,omitempty" db:"description"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	IsAchieved      bool       `json:"is_achieved" db:"is_achieved"`
	AchievedAt      *time.Time `json:"achieved_at,omitempty" db:"achieved_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// UserAchievement represents user achievements and milestones
type UserAchievement struct {
	ID              int64                  `json:"id" db:"id"`
	UserID          int64                  `json:"user_id" db:"user_id"`
	AchievementType string                 `json:"achievement_type" db:"achievement_type"`
	AchievementName string                 `json:"achievement_name" db:"achievement_name"`
	Description     string                 `json:"description" db:"description"`
	Icon            string                 `json:"icon" db:"icon"`
	Value           int64                  `json:"value" db:"value"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	EarnedAt        time.Time              `json:"earned_at" db:"earned_at"`
}

// UserInsight represents personalized insights for users
type UserInsight struct {
	ID           int64                  `json:"id" db:"id"`
	UserID       int64                  `json:"user_id" db:"user_id"`
	InsightType  string                 `json:"insight_type" db:"insight_type"`
	Title        string                 `json:"title" db:"title"`
	Description  string                 `json:"description" db:"description"`
	ActionText   *string                `json:"action_text,omitempty" db:"action_text"`
	ActionURL    *string                `json:"action_url,omitempty" db:"action_url"`
	Priority     int                    `json:"priority" db:"priority"` // 1-5
	Category     string                 `json:"category" db:"category"` // performance, engagement, feature, etc.
	Data         map[string]interface{} `json:"data,omitempty" db:"data"`
	IsRead       bool                   `json:"is_read" db:"is_read"`
	IsDismissed  bool                   `json:"is_dismissed" db:"is_dismissed"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// User Analytics Dashboard Models

// UserDashboardData represents data for user dashboard
type UserDashboardData struct {
	Summary      UserAnalyticsSummary  `json:"summary"`
	RecentURLs   []*URL                `json:"recent_urls"`
	TopURLs      []*URL                `json:"top_urls"`
	Goals        []*UserGoal           `json:"goals"`
	Achievements []*UserAchievement    `json:"achievements"`
	Insights     []*UserInsight        `json:"insights"`
	QuickStats   UserQuickStats        `json:"quick_stats"`
}

// UserQuickStats represents quick statistics for dashboard
type UserQuickStats struct {
	URLsThisWeek      int64   `json:"urls_this_week"`
	ClicksThisWeek    int64   `json:"clicks_this_week"`
	SharesThisWeek    int64   `json:"shares_this_week"`
	GrowthRate        float64 `json:"growth_rate_percentage"`
	EngagementRate    float64 `json:"engagement_rate_percentage"`
	SuccessRate       float64 `json:"success_rate_percentage"`
}

// Request/Response Models

// GetUserActivityRequest represents request parameters for user activity
type GetUserActivityRequest struct {
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	ActivityTypes []string   `json:"activity_types,omitempty"`
	ResourceTypes []string   `json:"resource_types,omitempty"`
	Limit         int        `json:"limit,omitempty"`
	Offset        int        `json:"offset,omitempty"`
}

// CreateUserGoalRequest represents a request to create a user goal
type CreateUserGoalRequest struct {
	GoalType     string     `json:"goal_type" validate:"required"`
	TargetValue  int64      `json:"target_value" validate:"required,min=1"`
	TargetDate   *time.Time `json:"target_date,omitempty"`
	Description  *string    `json:"description,omitempty"`
}

// UpdateUserGoalRequest represents a request to update a user goal
type UpdateUserGoalRequest struct {
	TargetValue  *int64     `json:"target_value,omitempty" validate:"omitempty,min=1"`
	TargetDate   *time.Time `json:"target_date,omitempty"`
	Description  *string    `json:"description,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

// AnalyticsExportRequest represents a request to export analytics data
type AnalyticsExportRequest struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Format    string     `json:"format" validate:"required,oneof=json csv pdf"`
	Sections  []string   `json:"sections,omitempty"` // activity, sessions, urls, etc.
}

// Constants for user analytics

const (
	// Activity types
	ActivityTypeLogin          = "login"
	ActivityTypeLogout         = "logout"
	ActivityTypeURLCreate      = "url_create"
	ActivityTypeURLUpdate      = "url_update"
	ActivityTypeURLDelete      = "url_delete"
	ActivityTypeURLView        = "url_view"
	ActivityTypeURLShare       = "url_share"
	ActivityTypeURLClick       = "url_click"
	ActivityTypeComment        = "comment"
	ActivityTypeBookmark       = "bookmark"
	ActivityTypeCollection     = "collection"
	ActivityTypeTeamJoin       = "team_join"
	ActivityTypeTeamLeave      = "team_leave"
	ActivityTypeDomainAdd      = "domain_add"
	ActivityTypeDomainVerify   = "domain_verify"
	ActivityTypeProfileUpdate  = "profile_update"
	ActivityTypeSettingsUpdate = "settings_update"
	ActivityTypePasswordChange = "password_change"
	ActivityTypeEmailVerify    = "email_verify"
	ActivityTypePhoneVerify    = "phone_verify"

	// Resource types
	ResourceTypeURL        = "url"
	ResourceTypeUser       = "user"
	ResourceTypeTeam       = "team"
	ResourceTypeDomain     = "domain"
	ResourceTypeComment    = "comment"
	ResourceTypeBookmark   = "bookmark"
	ResourceTypeCollection = "collection"
	ResourceTypeShare      = "share"
	ResourceTypeGoal       = "goal"
	ResourceTypeSession    = "session"

	// Actions
	ActionCreate   = "create"
	ActionRead     = "read"
	ActionUpdate   = "update"
	ActionDelete   = "delete"
	ActionShare    = "share"
	ActionClick    = "click"
	ActionView     = "view"
	ActionJoin     = "join"
	ActionLeave    = "leave"
	ActionVerify   = "verify"
	ActionComplete = "complete"

	// Goal types
	GoalTypeURLCreation  = "url_creation"
	GoalTypeClicks       = "clicks"
	GoalTypeShares       = "shares"
	GoalTypeEngagement   = "engagement"
	GoalTypeRetention    = "retention"
	GoalTypeCollaboration = "collaboration"

	// Achievement types
	AchievementTypeURLs         = "urls"
	AchievementTypeClicks       = "clicks"
	AchievementTypeShares       = "shares"
	AchievementTypeEngagement   = "engagement"
	AchievementTypeCollaboration = "collaboration"
	AchievementTypeMilestone    = "milestone"
	AchievementTypeStreak       = "streak"
	AchievementTypeFeature      = "feature"

	// Insight types
	InsightTypePerformance = "performance"
	InsightTypeEngagement  = "engagement"
	InsightTypeFeature     = "feature"
	InsightTypeOptimization = "optimization"
	InsightTypeGoal        = "goal"
	InsightTypeTrend       = "trend"
	InsightTypeAlert       = "alert"

	// Insight categories
	InsightCategoryPerformance = "performance"
	InsightCategoryEngagement  = "engagement"
	InsightCategoryFeature     = "feature"
	InsightCategoryGrowth      = "growth"
	InsightCategoryOptimization = "optimization"

	// Device types
	DeviceTypeDesktop = "desktop"
	DeviceTypeMobile  = "mobile"
	DeviceTypeTablet  = "tablet"
	DeviceTypeOther   = "other"

	// Session status
	SessionStatusActive   = "active"
	SessionStatusEnded    = "ended"
	SessionStatusTimedOut = "timed_out"
)

// Analytics helper functions

// CalculateEngagementScore calculates user engagement score based on various metrics
func CalculateEngagementScore(metrics UserEngagementMetrics) float64 {
	// Weighted scoring algorithm
	score := 0.0
	
	// Active days weight: 30%
	activeDaysScore := float64(metrics.DaysActive) / (float64(metrics.DaysActive + metrics.DaysInactive)) * 30
	
	// Actions per session weight: 25%
	actionsScore := (metrics.AvgActionsPerSession / 10.0) * 25
	if actionsScore > 25 {
		actionsScore = 25
	}
	
	// Feature adoption weight: 20%
	featureScore := metrics.FeatureAdoptionRate / 100 * 20
	
	// Collaboration weight: 15%
	collabScore := metrics.CollaborationRate / 100 * 15
	
	// Success rate weight: 10%
	successScore := metrics.URLSuccessRate / 100 * 10
	
	score = activeDaysScore + actionsScore + featureScore + collabScore + successScore
	
	if score > 100 {
		score = 100
	}
	
	return score
}

// GetEngagementLevel returns engagement level based on score
func GetEngagementLevel(score float64) string {
	switch {
	case score >= 80:
		return "Very High"
	case score >= 60:
		return "High"
	case score >= 40:
		return "Medium"
	case score >= 20:
		return "Low"
	default:
		return "Very Low"
	}
}