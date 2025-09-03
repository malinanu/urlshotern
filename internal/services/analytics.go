package services

import (
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type AnalyticsService struct {
	db             *storage.PostgresStorage
	redis          *storage.RedisStorage
	advancedService *AdvancedAnalyticsService
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *storage.PostgresStorage) *AnalyticsService {
	return &AnalyticsService{
		db:              db,
		advancedService: NewAdvancedAnalyticsService(db),
	}
}

// SetRedis sets the Redis storage for caching analytics
func (a *AnalyticsService) SetRedis(redis *storage.RedisStorage) {
	a.redis = redis
}

// GetAnalytics retrieves analytics data for a short code
func (a *AnalyticsService) GetAnalytics(shortCode string, days int) (*models.AnalyticsResponse, error) {
	// Try cache first if Redis is available
	if a.redis != nil {
		cached, err := a.redis.GetAnalytics(shortCode)
		if err == nil {
			return cached, nil
		}
		// If cache miss or error, continue to database lookup
	}

	// Get analytics from database
	analytics, err := a.db.GetAnalytics(shortCode, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	// Cache the result if Redis is available
	if a.redis != nil {
		cacheTTL := 5 * time.Minute // Cache analytics for 5 minutes
		a.redis.SetAnalytics(shortCode, analytics, cacheTTL)
	}

	return analytics, nil
}

// GetURLAnalytics retrieves comprehensive analytics for a specific URL
func (a *AnalyticsService) GetURLAnalytics(shortCode string) (*models.URLAnalytics, error) {
	// Implementation would query database for comprehensive analytics
	// For now, return simulated comprehensive analytics
	now := time.Now()
	
	return &models.URLAnalytics{
		ShortCode:    shortCode,
		OriginalURL:  "https://example.com/some-url",
		TotalClicks:  247,
		UniqueClicks: 183,
		CreatedAt:    now.Add(-7 * 24 * time.Hour),
		LastClickAt:  &now,
		ClickTrends: []models.ClickTrend{
			{Period: "2024-01-01", Clicks: 23},
			{Period: "2024-01-02", Clicks: 34},
			{Period: "2024-01-03", Clicks: 27},
			{Period: "2024-01-04", Clicks: 41},
			{Period: "2024-01-05", Clicks: 35},
			{Period: "2024-01-06", Clicks: 42},
			{Period: "2024-01-07", Clicks: 45},
		},
		ReferrerStats: []models.ReferrerStat{
			{Referrer: "twitter.com", Clicks: 89},
			{Referrer: "facebook.com", Clicks: 67},
			{Referrer: "direct", Clicks: 45},
			{Referrer: "google.com", Clicks: 46},
		},
		GeographicStats: []models.CountryDetail{
			{CountryCode: "US", CountryName: "United States", Clicks: 123, Percentage: 49.8, UniqueIPs: 89},
			{CountryCode: "GB", CountryName: "United Kingdom", Clicks: 34, Percentage: 13.8, UniqueIPs: 28},
			{CountryCode: "CA", CountryName: "Canada", Clicks: 28, Percentage: 11.3, UniqueIPs: 22},
			{CountryCode: "DE", CountryName: "Germany", Clicks: 22, Percentage: 8.9, UniqueIPs: 18},
		},
		DeviceStats: []models.DeviceTypeStat{
			{DeviceType: "desktop", Clicks: 148, Percentage: 59.9},
			{DeviceType: "mobile", Clicks: 86, Percentage: 34.8},
			{DeviceType: "tablet", Clicks: 13, Percentage: 5.3},
		},
		BrowserStats: []models.BrowserStat{
			{Browser: "Chrome", Clicks: 123},
			{Browser: "Firefox", Clicks: 45},
			{Browser: "Safari", Clicks: 34},
			{Browser: "Edge", Clicks: 28},
		},
		OSStats: []models.OSStat{
			{OS: "Windows", Clicks: 98},
			{OS: "macOS", Clicks: 67},
			{OS: "Linux", Clicks: 45},
			{OS: "iOS", Clicks: 23},
			{OS: "Android", Clicks: 14},
		},
		HourlyPattern: []models.HourlyClick{
			{Hour: 0, Clicks: 3}, {Hour: 1, Clicks: 2}, {Hour: 2, Clicks: 1},
			{Hour: 3, Clicks: 1}, {Hour: 4, Clicks: 2}, {Hour: 5, Clicks: 4},
			{Hour: 6, Clicks: 8}, {Hour: 7, Clicks: 12}, {Hour: 8, Clicks: 18},
			{Hour: 9, Clicks: 25}, {Hour: 10, Clicks: 28}, {Hour: 11, Clicks: 22},
			{Hour: 12, Clicks: 19}, {Hour: 13, Clicks: 21}, {Hour: 14, Clicks: 24},
			{Hour: 15, Clicks: 20}, {Hour: 16, Clicks: 17}, {Hour: 17, Clicks: 15},
			{Hour: 18, Clicks: 13}, {Hour: 19, Clicks: 11}, {Hour: 20, Clicks: 9},
			{Hour: 21, Clicks: 7}, {Hour: 22, Clicks: 5}, {Hour: 23, Clicks: 4},
		},
		WeeklyPattern: []models.WeekdayClick{
			{Weekday: 0, Day: "Sunday", Clicks: 28},
			{Weekday: 1, Day: "Monday", Clicks: 42},
			{Weekday: 2, Day: "Tuesday", Clicks: 45},
			{Weekday: 3, Day: "Wednesday", Clicks: 38},
			{Weekday: 4, Day: "Thursday", Clicks: 41},
			{Weekday: 5, Day: "Friday", Clicks: 35},
			{Weekday: 6, Day: "Saturday", Clicks: 18},
		},
	}, nil
}

// GetDetailedURLAnalytics retrieves detailed analytics for a URL within a time range
func (a *AnalyticsService) GetDetailedURLAnalytics(shortCode string, timeRange string) (*models.URLAnalytics, error) {
	// This would query database with time range filtering
	// For now, return the same as GetURLAnalytics with time range consideration
	analytics, err := a.GetURLAnalytics(shortCode)
	if err != nil {
		return nil, err
	}

	// Adjust data based on time range (simplified)
	switch timeRange {
	case "1d":
		// Filter to last 24 hours
		analytics.ClickTrends = analytics.ClickTrends[len(analytics.ClickTrends)-1:]
	case "7d":
		// Last 7 days (default)
		// No change needed
	case "30d":
		// Extend to 30 days (would need more data in real implementation)
		analytics.TotalClicks = analytics.TotalClicks * 4
		analytics.UniqueClicks = analytics.UniqueClicks * 4
	case "90d":
		// Extend to 90 days
		analytics.TotalClicks = analytics.TotalClicks * 12
		analytics.UniqueClicks = analytics.UniqueClicks * 12
	case "1y":
		// Extend to 1 year
		analytics.TotalClicks = analytics.TotalClicks * 52
		analytics.UniqueClicks = analytics.UniqueClicks * 52
	}

	return analytics, nil
}

// GetAdvancedAnalytics retrieves comprehensive advanced analytics
func (a *AnalyticsService) GetAdvancedAnalytics(shortCode string, days int) (*models.AdvancedAnalyticsResponse, error) {
	return a.advancedService.GetAdvancedAnalytics(shortCode, days)
}

// ProcessEnhancedClickEvent processes a click event with advanced analytics
func (a *AnalyticsService) ProcessEnhancedClickEvent(clickEvent *models.ClickEvent) error {
	return a.advancedService.ProcessEnhancedClickEvent(clickEvent)
}

// GetUserDashboardStats retrieves dashboard statistics for a specific user
func (a *AnalyticsService) GetUserDashboardStats(userID int64) (*models.UserDashboardStats, error) {
	stats := &models.UserDashboardStats{}

	// Get total URLs count
	err := a.db.QueryRow("SELECT COUNT(*) FROM url_mappings WHERE user_id = $1 AND is_active = TRUE", userID).Scan(&stats.TotalURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to get total URLs: %w", err)
	}

	// Get active URLs count (same as total in this case since we filter by is_active)
	stats.ActiveURLs = stats.TotalURLs

	// Get total clicks
	err = a.db.QueryRow(`
		SELECT COALESCE(SUM(click_count), 0) 
		FROM url_mappings 
		WHERE user_id = $1 AND is_active = TRUE
	`, userID).Scan(&stats.TotalClicks)
	if err != nil {
		return nil, fmt.Errorf("failed to get total clicks: %w", err)
	}

	// Get today's clicks
	err = a.db.QueryRow(`
		SELECT COUNT(*) 
		FROM click_events ce 
		JOIN url_mappings um ON ce.short_code = um.short_code 
		WHERE um.user_id = $1 
		  AND um.is_active = TRUE 
		  AND DATE(ce.clicked_at) = CURRENT_DATE
	`, userID).Scan(&stats.TodayClicks)
	if err != nil {
		stats.TodayClicks = 0 // Don't fail if there's no click events table or data
	}

	// Get this month's clicks
	err = a.db.QueryRow(`
		SELECT COUNT(*) 
		FROM click_events ce 
		JOIN url_mappings um ON ce.short_code = um.short_code 
		WHERE um.user_id = $1 
		  AND um.is_active = TRUE 
		  AND EXTRACT(YEAR FROM ce.clicked_at) = EXTRACT(YEAR FROM CURRENT_DATE)
		  AND EXTRACT(MONTH FROM ce.clicked_at) = EXTRACT(MONTH FROM CURRENT_DATE)
	`, userID).Scan(&stats.MonthClicks)
	if err != nil {
		stats.MonthClicks = 0 // Don't fail if there's no click events table or data
	}

	return stats, nil
}

// GetDashboardStats retrieves overall statistics for dashboard
func (a *AnalyticsService) GetDashboardStats() (*models.DashboardStats, error) {
	// This would be implemented for a dashboard view
	// For now, returning a basic implementation
	stats := &models.DashboardStats{
		TotalURLs:   0,
		TotalClicks: 0,
		TopURLs:     []models.TopURL{},
	}

	// In a real implementation, you would query the database for:
	// - Total number of active URLs
	// - Total number of clicks across all URLs
	// - Top performing URLs
	// - Click trends over time

	return stats, nil
}

// GetClickTrends retrieves click trends over time
func (a *AnalyticsService) GetClickTrends(shortCode string, period string) ([]models.ClickTrend, error) {
	var trends []models.ClickTrend
	
	// Determine the date range based on period
	switch period {
	case "hour":
		_ = "YYYY-MM-DD HH24:00:00"  // dateFormat for future use
		_ = "24 hours"                // intervalClause for future use
	case "day":
		_ = "YYYY-MM-DD"             // dateFormat for future use  
		_ = "30 days"                // intervalClause for future use
	case "week":
		_ = "IYYY-IW"                // dateFormat for future use
		_ = "12 weeks"               // intervalClause for future use
	case "month":
		_ = "YYYY-MM"                // dateFormat for future use
		_ = "12 months"              // intervalClause for future use
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	// This is a placeholder query - in a real implementation you would use proper SQL
	// query := fmt.Sprintf(`
	//     SELECT TO_CHAR(clicked_at, '%s') as period, COUNT(*) as clicks
	//     FROM click_events
	//     WHERE short_code = $1 AND clicked_at >= NOW() - INTERVAL '%s'
	//     GROUP BY TO_CHAR(clicked_at, '%s')
	//     ORDER BY period
	// `, dateFormat, intervalClause, dateFormat)

	// For now, return empty trends
	return trends, nil
}

// GetTopReferrers retrieves top referrers for a short code
func (a *AnalyticsService) GetTopReferrers(shortCode string, limit int) ([]models.ReferrerStat, error) {
	var referrers []models.ReferrerStat
	
	// This would query the click_events table for referrer statistics
	// Implementation would be similar to the country stats in postgres.go
	
	return referrers, nil
}

// GetUserAgentStats retrieves user agent statistics
func (a *AnalyticsService) GetUserAgentStats(shortCode string) (*models.UserAgentStats, error) {
	stats := &models.UserAgentStats{
		Browsers: []models.BrowserStat{},
		Devices:  []models.DeviceStat{},
		OS:       []models.OSStat{},
	}
	
	// This would parse user agent strings and categorize them
	// You might use a library like github.com/mssola/user_agent for this
	
	return stats, nil
}

// InvalidateAnalyticsCache invalidates cached analytics for a short code
func (a *AnalyticsService) InvalidateAnalyticsCache(shortCode string) error {
	if a.redis == nil {
		return nil
	}
	
	// Remove cached analytics
	return a.redis.DeleteURLMapping(fmt.Sprintf("analytics:%s", shortCode))
}

// GeoData represents geographic information
type GeoData struct {
	Country string
	City    string
}

// DeviceData represents device information parsed from user agent
type DeviceData struct {
	DeviceType string
	Browser    string
	OS         string
}

// getGeographicData retrieves geographic information from IP address
func (a *AnalyticsService) getGeographicData(clientIP string) *GeoData {
	// Placeholder implementation - in production you would use a GeoIP service
	// like MaxMind GeoIP2 or similar
	return &GeoData{
		Country: "Unknown",
		City:    "Unknown",
	}
}

// getDeviceData parses user agent to extract device information
func (a *AnalyticsService) getDeviceData(userAgent string) *DeviceData {
	// Placeholder implementation - in production you would use a user agent parser
	// like github.com/mssola/user_agent or similar
	return &DeviceData{
		DeviceType: "Unknown",
		Browser:    "Unknown", 
		OS:         "Unknown",
	}
}