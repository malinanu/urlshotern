package models

import (
	"time"
)

// URLMapping represents a URL mapping in the database
type URLMapping struct {
	ID          int64     `json:"id" db:"id"`
	ShortCode   string    `json:"short_code" db:"short_code"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	ClickCount  int64     `json:"click_count" db:"click_count"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedByIP string    `json:"created_by_ip,omitempty" db:"created_by_ip"`
	UserID      *int64    `json:"user_id,omitempty" db:"user_id"`
}

// ClickEvent represents a click tracking event
type ClickEvent struct {
	ID          int64     `json:"id" db:"id"`
	ShortCode   string    `json:"short_code" db:"short_code"`
	ClickedAt   time.Time `json:"clicked_at" db:"clicked_at"`
	IPAddress   string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   string    `json:"user_agent,omitempty" db:"user_agent"`
	Referrer    string    `json:"referrer,omitempty" db:"referrer"`
	CountryCode string    `json:"country_code,omitempty" db:"country_code"`
}

// ShortenRequest represents the request payload for shortening URLs
type ShortenRequest struct {
	URL       string     `json:"url" binding:"required,url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CustomCode string    `json:"custom_code,omitempty"`
}

// ShortenResponse represents the response for URL shortening
type ShortenResponse struct {
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// AnalyticsResponse represents analytics data for a short URL
type AnalyticsResponse struct {
	ShortCode    string    `json:"short_code"`
	OriginalURL  string    `json:"original_url"`
	TotalClicks  int64     `json:"total_clicks"`
	CreatedAt    time.Time `json:"created_at"`
	LastClickAt  *time.Time `json:"last_click_at,omitempty"`
	DailyClicks  []DailyClick `json:"daily_clicks,omitempty"`
	CountryStats []CountryStat `json:"country_stats,omitempty"`
}

// DailyClick represents click statistics per day
type DailyClick struct {
	Date   string `json:"date"`
	Clicks int64  `json:"clicks"`
}

// CountryStat represents click statistics per country
type CountryStat struct {
	CountryCode string `json:"country_code"`
	Clicks      int64  `json:"clicks"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// DashboardStats represents overall dashboard statistics
type DashboardStats struct {
	TotalURLs   int64    `json:"total_urls"`
	TotalClicks int64    `json:"total_clicks"`
	TopURLs     []TopURL `json:"top_urls"`
}

// TopURL represents a top performing URL
type TopURL struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	Clicks      int64  `json:"clicks"`
}

// UpdateURLRequest represents a request to update URL properties
type UpdateURLRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsPublic    *bool      `json:"is_public,omitempty"`
}

// UserURLResponse represents a URL in the user's URL list
type UserURLResponse struct {
	ID          int64      `json:"id"`
	ShortCode   string     `json:"short_code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	ClickCount  int64      `json:"click_count"`
	IsActive    bool       `json:"is_active"`
	IsPublic    bool       `json:"is_public"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// UserDashboardStats represents user-specific dashboard statistics
type UserDashboardStats struct {
	TotalURLs    int64 `json:"total_urls"`
	TotalClicks  int64 `json:"total_clicks"`
	ActiveURLs   int64 `json:"active_urls"`
	TodayClicks  int64 `json:"today_clicks"`
	MonthClicks  int64 `json:"month_clicks"`
}

// ClickTrend represents click trends over time
type ClickTrend struct {
	Period string `json:"period"`
	Clicks int64  `json:"clicks"`
}

// ReferrerStat represents referrer statistics
type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Clicks   int64  `json:"clicks"`
}

// UserAgentStats represents user agent statistics
type UserAgentStats struct {
	Browsers []BrowserStat `json:"browsers"`
	Devices  []DeviceStat  `json:"devices"`
	OS       []OSStat      `json:"operating_systems"`
}

// BrowserStat represents browser statistics
type BrowserStat struct {
	Browser string `json:"browser"`
	Clicks  int64  `json:"clicks"`
}

// DeviceStat represents device statistics
type DeviceStat struct {
	Device string `json:"device"`
	Clicks int64  `json:"clicks"`
}

// OSStat represents operating system statistics
type OSStat struct {
	OS     string `json:"operating_system"`
	Clicks int64  `json:"clicks"`
}

// Enhanced Geographic Analytics Models

// GeographicAnalytics represents detailed geographic analytics
type GeographicAnalytics struct {
	ShortCode      string           `json:"short_code"`
	TotalClicks    int64           `json:"total_clicks"`
	Countries      []CountryDetail  `json:"countries"`
	Regions        []RegionDetail   `json:"regions"`
	Cities         []CityDetail     `json:"cities"`
	MapData        []MapPoint       `json:"map_data"`
}

// CountryDetail represents detailed country analytics
type CountryDetail struct {
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	Clicks      int64   `json:"clicks"`
	Percentage  float64 `json:"percentage"`
	UniqueIPs   int64   `json:"unique_ips"`
	LastClick   *time.Time `json:"last_click,omitempty"`
}

// RegionDetail represents regional analytics within a country
type RegionDetail struct {
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	Clicks      int64   `json:"clicks"`
	Percentage  float64 `json:"percentage"`
}

// CityDetail represents city-level analytics
type CityDetail struct {
	CountryCode string   `json:"country_code"`
	Region      string   `json:"region,omitempty"`
	City        string   `json:"city"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	Clicks      int64    `json:"clicks"`
	Percentage  float64  `json:"percentage"`
}

// MapPoint represents a point on the map for visualization
type MapPoint struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Clicks      int64   `json:"clicks"`
	Location    string  `json:"location"` // City, Country format
	CountryCode string  `json:"country_code"`
}

// Enhanced Time-Based Analytics Models

// TimeAnalytics represents comprehensive time-based analytics
type TimeAnalytics struct {
	ShortCode       string           `json:"short_code"`
	HourlyPattern   []HourlyClick    `json:"hourly_pattern"`
	WeeklyPattern   []WeekdayClick   `json:"weekly_pattern"`
	MonthlyTrend    []MonthlyClick   `json:"monthly_trend"`
	HeatmapData     []HeatmapPoint   `json:"heatmap_data"`
	PeakTimes       PeakTimeInfo     `json:"peak_times"`
}

// HourlyClick represents clicks by hour of day
type HourlyClick struct {
	Hour   int   `json:"hour"` // 0-23
	Clicks int64 `json:"clicks"`
}

// WeekdayClick represents clicks by day of week
type WeekdayClick struct {
	Weekday int   `json:"weekday"` // 0-6 (Sunday-Saturday)
	Day     string `json:"day"`     // Sunday, Monday, etc.
	Clicks  int64 `json:"clicks"`
}

// MonthlyClick represents clicks by month
type MonthlyClick struct {
	Month  time.Time `json:"month"`
	Clicks int64     `json:"clicks"`
}

// HeatmapPoint represents a point in the time heatmap
type HeatmapPoint struct {
	Date   string `json:"date"`   // YYYY-MM-DD format
	Hour   int    `json:"hour"`   // 0-23
	Clicks int64  `json:"clicks"`
}

// PeakTimeInfo represents peak usage information
type PeakTimeInfo struct {
	PeakHour    int    `json:"peak_hour"`
	PeakWeekday int    `json:"peak_weekday"`
	PeakDay     string `json:"peak_day"`
	MaxClicks   int64  `json:"max_clicks"`
}

// Enhanced Device Analytics Models

// DeviceAnalytics represents comprehensive device analytics
type DeviceAnalytics struct {
	ShortCode     string               `json:"short_code"`
	DeviceTypes   []DeviceTypeStat     `json:"device_types"`
	DeviceBrands  []DeviceBrandStat    `json:"device_brands"`
	DeviceModels  []DeviceModelStat    `json:"device_models"`
	OSStats       []OSDetailStat       `json:"operating_systems"`
	BrowserStats  []BrowserDetailStat  `json:"browsers"`
	ScreenSizes   []ScreenSizeStat     `json:"screen_sizes"`
}

// DeviceTypeStat represents device type statistics
type DeviceTypeStat struct {
	DeviceType string  `json:"device_type"` // mobile, desktop, tablet
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

// DeviceBrandStat represents device brand statistics
type DeviceBrandStat struct {
	Brand      string  `json:"brand"` // Apple, Samsung, etc.
	DeviceType string  `json:"device_type"`
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

// DeviceModelStat represents device model statistics
type DeviceModelStat struct {
	Brand      string  `json:"brand"`
	Model      string  `json:"model"`
	DeviceType string  `json:"device_type"`
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

// OSDetailStat represents detailed OS statistics
type OSDetailStat struct {
	OSName     string  `json:"os_name"`
	OSVersion  string  `json:"os_version"`
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

// BrowserDetailStat represents detailed browser statistics
type BrowserDetailStat struct {
	BrowserName    string  `json:"browser_name"`
	BrowserVersion string  `json:"browser_version"`
	Clicks         int64   `json:"clicks"`
	Percentage     float64 `json:"percentage"`
}

// ScreenSizeStat represents screen resolution statistics
type ScreenSizeStat struct {
	Resolution string  `json:"resolution"` // 1920x1080, etc.
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

// Advanced Analytics Response Models

// AdvancedAnalyticsResponse represents comprehensive analytics
type AdvancedAnalyticsResponse struct {
	ShortCode   string              `json:"short_code"`
	OriginalURL string              `json:"original_url"`
	TotalClicks int64               `json:"total_clicks"`
	CreatedAt   time.Time           `json:"created_at"`
	Geographic  GeographicAnalytics `json:"geographic"`
	TimeData    TimeAnalytics       `json:"time_analytics"`
	DeviceData  DeviceAnalytics     `json:"device_analytics"`
	Referrers   []ReferrerStat      `json:"referrers"`
	LastUpdated time.Time           `json:"last_updated"`
}

// Conversion Tracking Models

// ConversionGoal represents a conversion goal configuration
type ConversionGoal struct {
	ID                int64     `json:"id" db:"id"`
	UserID            int64     `json:"user_id" db:"user_id"`
	GoalName          string    `json:"goal_name" db:"goal_name"`
	GoalType          string    `json:"goal_type" db:"goal_type"` // url_visit, custom_event, form_submit, purchase
	TargetURL         *string   `json:"target_url,omitempty" db:"target_url"`
	CustomEventName   *string   `json:"custom_event_name,omitempty" db:"custom_event_name"`
	GoalValue         float64   `json:"goal_value" db:"goal_value"` // monetary value
	AttributionWindow int       `json:"attribution_window" db:"attribution_window"` // days
	IsActive          bool      `json:"is_active" db:"is_active"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// CreateConversionGoalRequest represents a request to create a conversion goal
type CreateConversionGoalRequest struct {
	GoalName          string  `json:"goal_name" binding:"required,min=1,max=100"`
	GoalType          string  `json:"goal_type" binding:"required,oneof=url_visit custom_event form_submit purchase"`
	TargetURL         *string `json:"target_url,omitempty"`
	CustomEventName   *string `json:"custom_event_name,omitempty"`
	GoalValue         float64 `json:"goal_value"`
	AttributionWindow int     `json:"attribution_window" binding:"min=1,max=365"`
}

// UpdateConversionGoalRequest represents a request to update a conversion goal
type UpdateConversionGoalRequest struct {
	GoalName          *string  `json:"goal_name,omitempty"`
	GoalType          *string  `json:"goal_type,omitempty"`
	TargetURL         *string  `json:"target_url,omitempty"`
	CustomEventName   *string  `json:"custom_event_name,omitempty"`
	GoalValue         *float64 `json:"goal_value,omitempty"`
	AttributionWindow *int     `json:"attribution_window,omitempty"`
	IsActive          *bool    `json:"is_active,omitempty"`
}

// Conversion represents a conversion event
type Conversion struct {
	ID                int64     `json:"id" db:"id"`
	ShortCode         string    `json:"short_code" db:"short_code"`
	GoalID            int64     `json:"goal_id" db:"goal_id"`
	ConversionID      string    `json:"conversion_id" db:"conversion_id"`
	ConversionType    string    `json:"conversion_type" db:"conversion_type"`
	ConversionValue   float64   `json:"conversion_value" db:"conversion_value"`
	UserIP            string    `json:"user_ip,omitempty" db:"user_ip"`
	UserAgent         string    `json:"user_agent,omitempty" db:"user_agent"`
	Referrer          string    `json:"referrer,omitempty" db:"referrer"`
	SessionID         string    `json:"session_id,omitempty" db:"session_id"`
	ClickID           *int64    `json:"click_id,omitempty" db:"click_id"`
	ConversionTime    time.Time `json:"conversion_time" db:"conversion_time"`
	AttributionModel  string    `json:"attribution_model" db:"attribution_model"`
	TimeToConversion  int       `json:"time_to_conversion" db:"time_to_conversion"` // minutes
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// ConversionTrackingRequest represents a request to track a conversion
type ConversionTrackingRequest struct {
	GoalID          int64   `json:"goal_id" binding:"required"`
	ConversionID    string  `json:"conversion_id" binding:"required"`
	ConversionValue float64 `json:"conversion_value"`
	SessionID       string  `json:"session_id"`
	CustomData      map[string]interface{} `json:"custom_data,omitempty"`
}

// ConversionStats represents conversion statistics
type ConversionStats struct {
	GoalID            int64   `json:"goal_id"`
	GoalName          string  `json:"goal_name"`
	TotalConversions  int64   `json:"total_conversions"`
	ConversionRate    float64 `json:"conversion_rate"` // percentage
	TotalValue        float64 `json:"total_value"`
	AvgValue          float64 `json:"avg_value"`
	AvgTimeToConvert  float64 `json:"avg_time_to_convert"` // minutes
}

// A/B Testing Models

// ABTest represents an A/B test configuration
type ABTest struct {
	ID               int64      `json:"id" db:"id"`
	UserID           int64      `json:"user_id" db:"user_id"`
	TestName         string     `json:"test_name" db:"test_name"`
	TestType         string     `json:"test_type" db:"test_type"`
	Description      *string    `json:"description,omitempty" db:"description"`
	ShortCodeA       string     `json:"short_code_a" db:"short_code_a"` // Control
	ShortCodeB       string     `json:"short_code_b" db:"short_code_b"` // Variant
	TrafficSplit     string     `json:"traffic_split" db:"traffic_split"` // JSON configuration
	StartDate        *time.Time `json:"start_date,omitempty" db:"start_date"`
	EndDate          *time.Time `json:"end_date,omitempty" db:"end_date"`
	SampleSize       int        `json:"sample_size" db:"sample_size"`
	Confidence       float64    `json:"confidence" db:"confidence"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	Status           string     `json:"status" db:"status"`
	Winner           *string    `json:"winner,omitempty" db:"winner"`
	ConfidenceLevel  *float64   `json:"confidence_level,omitempty" db:"confidence_level"`
	MinSampleSize    int        `json:"min_sample_size" db:"min_sample_size"`
	ConversionGoalID *int64     `json:"conversion_goal_id,omitempty" db:"conversion_goal_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateABTestRequest represents a request to create an A/B test
type CreateABTestRequest struct {
	TestName         string                        `json:"test_name" binding:"required,min=1,max=100"`
	TestType         string                        `json:"test_type" binding:"required"`
	Description      *string                       `json:"description,omitempty"`
	ShortCodeA       string                        `json:"short_code_a" binding:"required"`
	ShortCodeB       string                        `json:"short_code_b" binding:"required"`
	TrafficSplit     string                        `json:"traffic_split"`
	StartDate        *time.Time                    `json:"start_date,omitempty"`
	EndDate          *time.Time                    `json:"end_date,omitempty"`
	SampleSize       int                           `json:"sample_size" binding:"min=10,max=100000"`
	Confidence       float64                       `json:"confidence"`
	MinSampleSize    int                           `json:"min_sample_size" binding:"min=10,max=100000"`
	ConversionGoalID *int64                        `json:"conversion_goal_id,omitempty"`
	Variants         []CreateABTestVariantRequest  `json:"variants" binding:"required,min=2"`
}

// UpdateABTestRequest represents a request to update an A/B test
type UpdateABTestRequest struct {
	TestName         *string    `json:"test_name,omitempty"`
	Description      *string    `json:"description,omitempty"`
	TrafficSplit     *int       `json:"traffic_split,omitempty"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	Status           *string    `json:"status,omitempty"`
	MinSampleSize    *int       `json:"min_sample_size,omitempty"`
	ConversionGoalID *int64     `json:"conversion_goal_id,omitempty"`
}

// ABTestResult represents A/B test results for a specific variant and date
type ABTestResult struct {
	ID              int64     `json:"id" db:"id"`
	TestID          int64     `json:"test_id" db:"test_id"`
	Variant         string    `json:"variant" db:"variant"` // 'A' or 'B'
	Date            time.Time `json:"date" db:"date"`
	Clicks          int       `json:"clicks" db:"clicks"`
	Conversions     int       `json:"conversions" db:"conversions"`
	ConversionRate  float64   `json:"conversion_rate" db:"conversion_rate"`
	Revenue         float64   `json:"revenue" db:"revenue"`
	UniqueVisitors  int       `json:"unique_visitors" db:"unique_visitors"`
	BounceRate      float64   `json:"bounce_rate" db:"bounce_rate"`
	AvgTimeOnSite   int       `json:"avg_time_on_site" db:"avg_time_on_site"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ABTestSummary represents a summary of A/B test performance
type ABTestSummary struct {
	Test            ABTest                 `json:"test"`
	VariantA        ABTestVariantSummary   `json:"variant_a"`
	VariantB        ABTestVariantSummary   `json:"variant_b"`
	StatisticalData ABTestStatisticalData  `json:"statistical_data"`
	IsSignificant   bool                   `json:"is_significant"`
	Recommendation  string                 `json:"recommendation"`
}

// ABTestVariantSummary represents performance summary for a test variant
type ABTestVariantSummary struct {
	Variant           string  `json:"variant"`
	ShortCode         string  `json:"short_code"`
	TotalClicks       int64   `json:"total_clicks"`
	TotalConversions  int64   `json:"total_conversions"`
	ConversionRate    float64 `json:"conversion_rate"`
	TotalRevenue      float64 `json:"total_revenue"`
	AvgOrderValue     float64 `json:"avg_order_value"`
	UniqueVisitors    int64   `json:"unique_visitors"`
	BounceRate        float64 `json:"bounce_rate"`
	AvgTimeOnSite     float64 `json:"avg_time_on_site"`
}

// ABTestStatisticalData represents statistical analysis of A/B test
type ABTestStatisticalData struct {
	SampleSizeA      int64   `json:"sample_size_a"`
	SampleSizeB      int64   `json:"sample_size_b"`
	ConfidenceLevel  float64 `json:"confidence_level"`
	PValue           float64 `json:"p_value"`
	ZScore           float64 `json:"z_score"`
	MarginOfError    float64 `json:"margin_of_error"`
	MinimumDetectable float64 `json:"minimum_detectable_effect"`
	TestPower        float64 `json:"test_power"`
}

// Attribution Models

// AttributionTouchpoint represents a marketing touchpoint in the customer journey
type AttributionTouchpoint struct {
	ID              int64      `json:"id" db:"id"`
	SessionID       string     `json:"session_id" db:"session_id"`
	ShortCode       string     `json:"short_code" db:"short_code"`
	UserIP          string     `json:"user_ip,omitempty" db:"user_ip"`
	UserAgent       string     `json:"user_agent,omitempty" db:"user_agent"`
	Referrer        string     `json:"referrer,omitempty" db:"referrer"`
	CampaignSource  string     `json:"campaign_source,omitempty" db:"campaign_source"`
	CampaignMedium  string     `json:"campaign_medium,omitempty" db:"campaign_medium"`
	CampaignName    string     `json:"campaign_name,omitempty" db:"campaign_name"`
	TouchpointOrder int        `json:"touchpoint_order" db:"touchpoint_order"`
	TouchpointTime  time.Time  `json:"touchpoint_time" db:"touchpoint_time"`
	ConversionID    *string    `json:"conversion_id,omitempty" db:"conversion_id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// ConversionJourney represents the complete customer journey to conversion
type ConversionJourney struct {
	ConversionID string                  `json:"conversion_id"`
	SessionID    string                  `json:"session_id"`
	Touchpoints  []AttributionTouchpoint `json:"touchpoints"`
	TotalTouches int                     `json:"total_touches"`
	JourneyTime  int                     `json:"journey_time_minutes"`
	Conversion   Conversion              `json:"conversion"`
}

// Real-time Analytics Models

// RealtimeUpdate represents a real-time analytics update sent via WebSocket
type RealtimeUpdate struct {
	Type      string      `json:"type"`       // "click", "analytics_update", "conversion"
	ShortCode string      `json:"short_code"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// WebSocketMessage represents incoming WebSocket messages from clients
type WebSocketMessage struct {
	Type      string `json:"type"`       // "subscribe", "unsubscribe", "ping"
	ShortCode string `json:"short_code,omitempty"`
}

// RealtimeClickData represents real-time click event data
type RealtimeClickData struct {
	ShortCode string    `json:"short_code"`
	ClientIP  string    `json:"client_ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Referrer  string    `json:"referrer,omitempty"`
	Country   string    `json:"country,omitempty"`
	City      string    `json:"city,omitempty"`
	Device    string    `json:"device,omitempty"`
	Browser   string    `json:"browser,omitempty"`
	OS        string    `json:"os,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// A/B Testing Models (Additional)

// ABTestVariant represents an A/B test variant configuration
type ABTestVariant struct {
	ID                int64     `json:"id" db:"id"`
	TestID            int64     `json:"test_id" db:"test_id"`
	VariantName       string    `json:"variant_name" db:"variant_name"`
	ShortCode         string    `json:"short_code" db:"short_code"`
	TrafficAllocation int       `json:"traffic_allocation" db:"traffic_allocation"` // percentage
	IsControl         bool      `json:"is_control" db:"is_control"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// ABTestResults represents comprehensive A/B test results
type ABTestResults struct {
	TestID           int64            `json:"test_id"`
	TestName         string           `json:"test_name"`
	Status           string           `json:"status"`
	StartDate        *time.Time       `json:"start_date"`
	EndDate          *time.Time       `json:"end_date"`
	TotalSessions    int              `json:"total_sessions"`
	TotalConversions int              `json:"total_conversions"`
	OverallCR        float64          `json:"overall_conversion_rate"`
	VariantResults   []*VariantResult `json:"variant_results"`
	IsSignificant    bool             `json:"is_significant"`
	PValue           float64          `json:"p_value"`
	ConfidenceLevel  float64          `json:"confidence_level"`
	Winner           *string          `json:"winner,omitempty"`
	Recommendation   string           `json:"recommendation"`
}

// VariantResult represents performance metrics for an A/B test variant
type VariantResult struct {
	VariantID      int64   `json:"variant_id"`
	VariantName    string  `json:"variant_name"`
	IsControl      bool    `json:"is_control"`
	Sessions       int     `json:"sessions"`
	Conversions    int     `json:"conversions"`
	ConversionRate float64 `json:"conversion_rate"`
	Revenue        float64 `json:"revenue"`
	AOV            float64 `json:"average_order_value"`
}

// SignificanceResult represents statistical significance analysis
type SignificanceResult struct {
	IsSignificant              bool      `json:"is_significant"`
	PValue                     float64   `json:"p_value"`
	ZScore                     float64   `json:"z_score"`
	ControlCR                  float64   `json:"control_conversion_rate"`
	VariantCR                  float64   `json:"variant_conversion_rate"`
	Improvement                float64   `json:"improvement_percentage"`
	EffectSize                 float64   `json:"effect_size"`                    // Cohen's h
	ConfidenceInterval         [2]float64 `json:"confidence_interval"`           // [lower, upper] bounds
	MinDetectableEffect        float64   `json:"minimum_detectable_effect"`
	SampleSizeRecommendation   int64     `json:"sample_size_recommendation"`
}

// SequentialTestResult represents results from sequential testing for early stopping
type SequentialTestResult struct {
	CanStop      bool    `json:"can_stop"`
	Decision     string  `json:"decision"`      // "continue", "test_wins", "control_wins"
	Confidence   float64 `json:"confidence"`
	LogLR        float64 `json:"log_likelihood_ratio"`
	UpperBound   float64 `json:"upper_bound"`   // Upper decision boundary
	LowerBound   float64 `json:"lower_bound"`   // Lower decision boundary
	Reason       string  `json:"reason"`        // Human-readable explanation
}

// CreateABTestVariantRequest represents a variant in an A/B test creation request
type CreateABTestVariantRequest struct {
	VariantName       string `json:"variant_name" binding:"required"`
	ShortCode         string `json:"short_code" binding:"required"`
	TrafficAllocation int    `json:"traffic_allocation" binding:"required,min=0,max=100"`
	IsControl         bool   `json:"is_control"`
}

// URL represents a URL in the system (used by dashboard)
type URL struct {
	ID          int64      `json:"id" db:"id"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	Title       *string    `json:"title,omitempty" db:"title"`
	Description *string    `json:"description,omitempty" db:"description"`
	CreatedBy   *int64     `json:"created_by,omitempty" db:"created_by"`
	ClickCount  int64      `json:"click_count" db:"click_count"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	IsPublic    bool       `json:"is_public" db:"is_public"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// URLAnalytics represents comprehensive analytics for a URL
type URLAnalytics struct {
	ShortCode        string             `json:"short_code"`
	OriginalURL      string             `json:"original_url"`
	TotalClicks      int64              `json:"total_clicks"`
	UniqueClicks     int64              `json:"unique_clicks"`
	CreatedAt        time.Time          `json:"created_at"`
	LastClickAt      *time.Time         `json:"last_click_at,omitempty"`
	ClickTrends      []ClickTrend       `json:"click_trends"`
	ReferrerStats    []ReferrerStat     `json:"referrer_stats"`
	GeographicStats  []CountryDetail    `json:"geographic_stats"`
	DeviceStats      []DeviceTypeStat   `json:"device_stats"`
	BrowserStats     []BrowserStat      `json:"browser_stats"`
	OSStats          []OSStat           `json:"os_stats"`
	HourlyPattern    []HourlyClick      `json:"hourly_pattern"`
	WeeklyPattern    []WeekdayClick     `json:"weekly_pattern"`
}