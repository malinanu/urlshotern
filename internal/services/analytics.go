package services

import (
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type AnalyticsService struct {
	db    *storage.PostgresStorage
	redis *storage.RedisStorage
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *storage.PostgresStorage) *AnalyticsService {
	return &AnalyticsService{
		db: db,
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
	var dateFormat string
	var intervalClause string
	
	switch period {
	case "hour":
		dateFormat = "YYYY-MM-DD HH24:00:00"
		intervalClause = "24 hours"
	case "day":
		dateFormat = "YYYY-MM-DD"
		intervalClause = "30 days"
	case "week":
		dateFormat = "IYYY-IW"
		intervalClause = "12 weeks"
	case "month":
		dateFormat = "YYYY-MM"
		intervalClause = "12 months"
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