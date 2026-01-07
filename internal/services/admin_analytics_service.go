package services

import (
	"context"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// AdminAnalyticsService handles admin analytics operations
type AdminAnalyticsService struct {
	dbManager *storage.DatabaseManager
}

// NewAdminAnalyticsService creates a new admin analytics service
func NewAdminAnalyticsService(dbManager *storage.DatabaseManager) *AdminAnalyticsService {
	return &AdminAnalyticsService{
		dbManager: dbManager,
	}
}

// GetDashboard retrieves dashboard overview statistics
func (s *AdminAnalyticsService) GetDashboard(ctx context.Context) (*models.DashboardResponse, error) {
	dash := &models.DashboardResponse{}

	// Total users
	err := s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&dash.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	// Active users (logged in last 30 days)
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE last_login_at > NOW() - INTERVAL '30 days'").Scan(&dash.ActiveUsers)
	if err != nil {
		dash.ActiveUsers = 0
	}

	// New users today
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE").Scan(&dash.NewUsersToday)
	if err != nil {
		dash.NewUsersToday = 0
	}

	// New users this week
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= DATE_TRUNC('week', CURRENT_DATE)").Scan(&dash.NewUsersThisWeek)
	if err != nil {
		dash.NewUsersThisWeek = 0
	}

	// New users this month
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= DATE_TRUNC('month', CURRENT_DATE)").Scan(&dash.NewUsersThisMonth)
	if err != nil {
		dash.NewUsersThisMonth = 0
	}

	// Total URLs
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM url_mappings").Scan(&dash.TotalURLs)
	if err != nil {
		dash.TotalURLs = 0
	}

	// Active URLs
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM url_mappings WHERE is_active = TRUE AND (expires_at IS NULL OR expires_at > NOW())").Scan(&dash.ActiveURLs)
	if err != nil {
		dash.ActiveURLs = 0
	}

	// New URLs today
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM url_mappings WHERE created_at >= CURRENT_DATE").Scan(&dash.NewURLsToday)
	if err != nil {
		dash.NewURLsToday = 0
	}

	// Total clicks
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COALESCE(SUM(click_count), 0) FROM url_mappings").Scan(&dash.TotalClicks)
	if err != nil {
		dash.TotalClicks = 0
	}

	// Today clicks
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM click_events WHERE clicked_at >= CURRENT_DATE").Scan(&dash.TodayClicks)
	if err != nil {
		dash.TodayClicks = 0
	}

	// This week clicks
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM click_events WHERE clicked_at >= DATE_TRUNC('week', CURRENT_DATE)").Scan(&dash.ThisWeekClicks)
	if err != nil {
		dash.ThisWeekClicks = 0
	}

	// This month clicks
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM click_events WHERE clicked_at >= DATE_TRUNC('month', CURRENT_DATE)").Scan(&dash.ThisMonthClicks)
	if err != nil {
		dash.ThisMonthClicks = 0
	}

	// Revenue metrics (placeholder - depends on billing implementation)
	dash.TotalRevenue = 0
	dash.MonthlyRevenue = 0
	dash.ActiveSubscriptions = 0
	dash.TrialSubscriptions = 0
	dash.CancelledToday = 0

	return dash, nil
}

// GetUserAnalytics retrieves user growth analytics
func (s *AdminAnalyticsService) GetUserAnalytics(ctx context.Context, startDate, endDate time.Time) (*models.UserAnalyticsResponse, error) {
	analytics := &models.UserAnalyticsResponse{}

	// Total users
	err := s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&analytics.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Registration data (daily counts)
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM users
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	rows, err := s.dbManager.ReadOnlyQueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query registration data: %w", err)
	}
	defer rows.Close()

	analytics.RegistrationData = []models.TimeSeriesDataPoint{}
	for rows.Next() {
		var point models.TimeSeriesDataPoint
		var count int64
		err := rows.Scan(&point.Date, &count)
		if err != nil {
			continue
		}
		point.Count = count
		analytics.RegistrationData = append(analytics.RegistrationData, point)
	}

	// Users by plan
	planQuery := `
		SELECT subscription_tier, COUNT(*) as count
		FROM users
		WHERE deleted_at IS NULL
		GROUP BY subscription_tier
	`

	planRows, err := s.dbManager.ReadOnlyQueryContext(ctx, planQuery)
	if err == nil {
		defer planRows.Close()
		analytics.UsersByPlan = []models.UsersByPlanDataPoint{}
		for planRows.Next() {
			var point models.UsersByPlanDataPoint
			err := planRows.Scan(&point.Plan, &point.Count)
			if err != nil {
				continue
			}
			analytics.UsersByPlan = append(analytics.UsersByPlan, point)
		}
	}

	analytics.GrowthRate = 0    // Calculate based on data
	analytics.RetentionRate = 0 // Calculate based on login data

	return analytics, nil
}

// GetURLAnalytics retrieves URL creation and usage analytics
func (s *AdminAnalyticsService) GetURLAnalytics(ctx context.Context, startDate, endDate time.Time) (*models.URLAnalyticsResponse, error) {
	analytics := &models.URLAnalyticsResponse{}

	// Total URLs
	err := s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM url_mappings").Scan(&analytics.TotalURLs)
	if err != nil {
		analytics.TotalURLs = 0
	}

	// Active URLs
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM url_mappings WHERE is_active = TRUE").Scan(&analytics.ActiveURLs)
	if err != nil {
		analytics.ActiveURLs = 0
	}

	// Expired URLs
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM url_mappings WHERE expires_at IS NOT NULL AND expires_at <= NOW()").Scan(&analytics.ExpiredURLs)
	if err != nil {
		analytics.ExpiredURLs = 0
	}

	// Flagged URLs
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COUNT(*) FROM admin.flagged_urls WHERE resolved_at IS NULL").Scan(&analytics.FlaggedURLs)
	if err != nil {
		analytics.FlaggedURLs = 0
	}

	// Average clicks per URL
	err = s.dbManager.ReadOnlyQueryRowContext(ctx, "SELECT COALESCE(AVG(click_count), 0) FROM url_mappings").Scan(&analytics.AvgClicksPerURL)
	if err != nil {
		analytics.AvgClicksPerURL = 0
	}

	// Creation data (daily counts)
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM url_mappings
		WHERE created_at BETWEEN $1 AND $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	rows, err := s.dbManager.ReadOnlyQueryContext(ctx, query, startDate, endDate)
	if err == nil {
		defer rows.Close()
		analytics.CreationData = []models.TimeSeriesDataPoint{}
		for rows.Next() {
			var point models.TimeSeriesDataPoint
			var count int64
			err := rows.Scan(&point.Date, &count)
			if err != nil {
				continue
			}
			point.Count = count
			analytics.CreationData = append(analytics.CreationData, point)
		}
	}

	return analytics, nil
}

// GetClickAnalytics retrieves click statistics
func (s *AdminAnalyticsService) GetClickAnalytics(ctx context.Context, startDate, endDate time.Time) ([]models.TimeSeriesDataPoint, error) {
	query := `
		SELECT DATE(clicked_at) as date, COUNT(*) as count
		FROM click_events
		WHERE clicked_at BETWEEN $1 AND $2
		GROUP BY DATE(clicked_at)
		ORDER BY date
	`

	rows, err := s.dbManager.ReadOnlyQueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query click analytics: %w", err)
	}
	defer rows.Close()

	data := []models.TimeSeriesDataPoint{}
	for rows.Next() {
		var point models.TimeSeriesDataPoint
		var count int64
		err := rows.Scan(&point.Date, &count)
		if err != nil {
			continue
		}
		point.Count = count
		data = append(data, point)
	}

	return data, nil
}

// GetTopURLs retrieves top performing URLs
func (s *AdminAnalyticsService) GetTopURLs(ctx context.Context, limit int, startDate, endDate time.Time) ([]models.URLMapping, error) {
	query := `
		SELECT id, short_code, original_url, user_id, created_at, expires_at, click_count, is_active
		FROM url_mappings
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY click_count DESC
		LIMIT $3
	`

	rows, err := s.dbManager.ReadOnlyQueryContext(ctx, query, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top URLs: %w", err)
	}
	defer rows.Close()

	urls := []models.URLMapping{}
	for rows.Next() {
		var url models.URLMapping
		err := rows.Scan(
			&url.ID, &url.ShortCode, &url.OriginalURL, &url.UserID,
			&url.CreatedAt, &url.ExpiresAt, &url.ClickCount, &url.IsActive,
		)
		if err != nil {
			continue
		}
		urls = append(urls, url)
	}

	return urls, nil
}

// ExportReport generates an export report (placeholder)
func (s *AdminAnalyticsService) ExportReport(ctx context.Context, req models.ExportReportRequest) ([]byte, error) {
	// This would generate CSV/JSON/XLSX based on report type
	// Placeholder implementation
	return []byte{}, nil
}
