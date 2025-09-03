package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"urlshortener/internal/models"
)

type UserAnalyticsService struct {
	db *sqlx.DB
}

func NewUserAnalyticsService(db *sqlx.DB) *UserAnalyticsService {
	return &UserAnalyticsService{
		db: db,
	}
}

// LogUserActivity logs a user activity event
func (s *UserAnalyticsService) LogUserActivity(activity *models.UserActivityLog) error {
	query := `
		INSERT INTO user_activity_logs (
			user_id, session_id, activity_type, description,
			url_id, metadata, ip_address, user_agent,
			device_type, browser, platform, location,
			duration_ms, screen_resolution, timezone,
			is_mobile, is_bot, created_at
		) VALUES (
			:user_id, :session_id, :activity_type, :description,
			:url_id, :metadata, :ip_address, :user_agent,
			:device_type, :browser, :platform, :location,
			:duration_ms, :screen_resolution, :timezone,
			:is_mobile, :is_bot, :created_at
		)`

	_, err := s.db.NamedExec(query, activity)
	if err != nil {
		return fmt.Errorf("failed to log user activity: %w", err)
	}

	// Update user's last activity timestamp
	_, err = s.db.Exec("UPDATE users SET last_login_at = NOW() WHERE id = $1", activity.UserID)
	if err != nil {
		// Log warning but don't fail the activity logging
		fmt.Printf("Warning: failed to update user last activity: %v\n", err)
	}

	return nil
}

// StartUserSession creates a new user session
func (s *UserAnalyticsService) StartUserSession(session *models.UserSession) error {
	query := `
		INSERT INTO user_sessions (
			id, user_id, ip_address, user_agent,
			device_type, browser, platform, location,
			is_mobile, screen_resolution, timezone,
			started_at
		) VALUES (
			:id, :user_id, :ip_address, :user_agent,
			:device_type, :browser, :platform, :location,
			:is_mobile, :screen_resolution, :timezone,
			:started_at
		)`

	_, err := s.db.NamedExec(query, session)
	if err != nil {
		return fmt.Errorf("failed to start user session: %w", err)
	}

	return nil
}

// EndUserSession ends a user session
func (s *UserAnalyticsService) EndUserSession(sessionID string, endedAt time.Time) error {
	query := `
		UPDATE user_sessions 
		SET ended_at = $1, duration_minutes = EXTRACT(EPOCH FROM ($1 - started_at))/60
		WHERE id = $2 AND ended_at IS NULL`

	_, err := s.db.Exec(query, endedAt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to end user session: %w", err)
	}

	return nil
}

// GetUserAnalyticsSummary returns comprehensive analytics for a user
func (s *UserAnalyticsService) GetUserAnalyticsSummary(userID int64, dateRange models.DateRange) (*models.UserAnalyticsSummary, error) {
	summary := &models.UserAnalyticsSummary{
		UserID: userID,
	}

	// Get user info
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	summary.AccountCreatedAt = user.CreatedAt
	summary.LastActiveAt = user.LastLoginAt

	whereClause, args := s.buildDateRangeClause(dateRange, userID)

	// Get session stats
	sessionQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_sessions,
			AVG(duration_minutes) as avg_duration,
			SUM(duration_minutes) as total_duration,
			COUNT(DISTINCT DATE(started_at)) as active_days
		FROM user_sessions 
		WHERE user_id = $1 %s`, whereClause)

	var sessionStats struct {
		TotalSessions   int64           `db:"total_sessions"`
		AvgDuration     sql.NullFloat64 `db:"avg_duration"`
		TotalDuration   sql.NullFloat64 `db:"total_duration"`
		ActiveDays      int64           `db:"active_days"`
	}

	err = s.db.Get(&sessionStats, sessionQuery, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}

	summary.TotalSessions = sessionStats.TotalSessions
	summary.TotalActiveDays = sessionStats.ActiveDays
	if sessionStats.AvgDuration.Valid {
		summary.AvgSessionDuration = sessionStats.AvgDuration.Float64
	}
	if sessionStats.TotalDuration.Valid {
		summary.TotalTimeSpent = sessionStats.TotalDuration.Float64
	}

	// Get URL stats
	urlQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_urls,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '30 days') as urls_last_30_days,
			SUM(click_count) as total_clicks,
			SUM(click_count) FILTER (WHERE created_at >= NOW() - INTERVAL '30 days') as clicks_last_30_days
		FROM urls 
		WHERE user_id = $1 %s`, strings.ReplaceAll(whereClause, "started_at", "created_at"))

	var urlStats struct {
		TotalURLs       int64 `db:"total_urls"`
		URLsLast30Days  int64 `db:"urls_last_30_days"`
		TotalClicks     int64 `db:"total_clicks"`
		ClicksLast30Days int64 `db:"clicks_last_30_days"`
	}

	err = s.db.Get(&urlStats, urlQuery, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get URL stats: %w", err)
	}

	summary.TotalURLsCreated = urlStats.TotalURLs
	summary.URLsCreatedLast30Days = urlStats.URLsLast30Days
	summary.TotalClicks = urlStats.TotalClicks
	summary.ClicksLast30Days = urlStats.ClicksLast30Days

	// Calculate engagement score
	summary.EngagementScore = s.calculateEngagementScore(summary)

	// Get top devices and browsers
	summary.TopDevices = s.getTopDevices(userID, dateRange)
	summary.TopBrowsers = s.getTopBrowsers(userID, dateRange)
	summary.TopLocations = s.getTopLocations(userID, dateRange)

	return summary, nil
}

// GetUserEngagementMetrics returns detailed engagement metrics
func (s *UserAnalyticsService) GetUserEngagementMetrics(userID int64, dateRange models.DateRange) (*models.UserEngagementMetrics, error) {
	metrics := &models.UserEngagementMetrics{
		UserID: userID,
	}

	whereClause, args := s.buildDateRangeClause(dateRange, userID)

	// Get activity counts by type
	activityQuery := fmt.Sprintf(`
		SELECT 
			activity_type,
			COUNT(*) as count
		FROM user_activity_logs 
		WHERE user_id = $1 %s
		GROUP BY activity_type`, whereClause)

	rows, err := s.db.Query(activityQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity counts: %w", err)
	}
	defer rows.Close()

	activityCounts := make(map[string]int64)
	for rows.Next() {
		var activityType string
		var count int64
		if err := rows.Scan(&activityType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan activity count: %w", err)
		}
		activityCounts[activityType] = count
	}

	// Map activity types to metrics
	metrics.LoginCount = activityCounts["user_login"]
	metrics.URLCreatedCount = activityCounts["url_created"] + activityCounts["url_shortened"]
	metrics.URLViewedCount = activityCounts["url_clicked"] + activityCounts["url_accessed"]
	metrics.URLSharedCount = activityCounts["url_shared"]
	metrics.CommentCount = activityCounts["comment_created"]
	metrics.BookmarkCount = activityCounts["bookmark_created"]
	metrics.CollectionCount = activityCounts["collection_created"]

	// Get feature usage
	featureQuery := fmt.Sprintf(`
		SELECT DISTINCT activity_type 
		FROM user_activity_logs 
		WHERE user_id = $1 %s`, whereClause)

	var features []string
	err = s.db.Select(&features, featureQuery, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get features used: %w", err)
	}
	metrics.FeaturesUsed = features

	// Calculate scores
	metrics.ActivityScore = s.calculateActivityScore(activityCounts)
	metrics.ConsistencyScore = s.calculateConsistencyScore(userID, dateRange)
	metrics.DiversityScore = s.calculateDiversityScore(features)

	return metrics, nil
}

// GetUserActivityLog returns user activity log with pagination
func (s *UserAnalyticsService) GetUserActivityLog(userID int64, req *models.UserActivityLogRequest) (*models.UserActivityLogResponse, error) {
	baseQuery := "FROM user_activity_logs WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	// Add filters
	if req.ActivityType != "" {
		baseQuery += fmt.Sprintf(" AND activity_type = $%d", argIndex)
		args = append(args, req.ActivityType)
		argIndex++
	}

	if req.URLId != nil {
		baseQuery += fmt.Sprintf(" AND url_id = $%d", argIndex)
		args = append(args, *req.URLId)
		argIndex++
	}

	if !req.StartDate.IsZero() {
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, req.StartDate)
		argIndex++
	}

	if !req.EndDate.IsZero() {
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, req.EndDate)
		argIndex++
	}

	// Get total count
	var totalCount int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := s.db.Get(&totalCount, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity count: %w", err)
	}

	// Get activities with pagination
	orderBy := "created_at DESC"
	if req.SortBy != "" {
		orderBy = req.SortBy
		if req.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT * %s 
		ORDER BY %s 
		LIMIT $%d OFFSET $%d`, 
		baseQuery, orderBy, argIndex, argIndex+1)
	
	args = append(args, limit, offset)

	var activities []models.UserActivityLog
	err = s.db.Select(&activities, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	return &models.UserActivityLogResponse{
		Activities:  activities,
		TotalCount:  totalCount,
		Limit:       limit,
		Offset:      offset,
		HasMore:     offset+int64(len(activities)) < totalCount,
	}, nil
}

// GetDashboardAnalytics returns analytics data for the dashboard
func (s *UserAnalyticsService) GetDashboardAnalytics(userID int64, period string) (*models.UserDashboardAnalytics, error) {
	analytics := &models.UserDashboardAnalytics{
		UserID: userID,
		Period: period,
	}

	// Get date range for the period
	var startDate time.Time
	switch period {
	case "today":
		startDate = time.Now().Truncate(24 * time.Hour)
	case "week":
		startDate = time.Now().AddDate(0, 0, -7)
	case "month":
		startDate = time.Now().AddDate(0, -1, 0)
	case "year":
		startDate = time.Now().AddDate(-1, 0, 0)
	default:
		startDate = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	dateRange := models.DateRange{
		StartDate: startDate,
		EndDate:   time.Now(),
	}

	// Get basic metrics
	summary, err := s.GetUserAnalyticsSummary(userID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics summary: %w", err)
	}

	analytics.TotalURLs = summary.TotalURLsCreated
	analytics.TotalClicks = summary.TotalClicks
	analytics.TotalSessions = summary.TotalSessions
	analytics.AvgSessionDuration = summary.AvgSessionDuration

	// Get chart data
	analytics.UrlCreationChart = s.getURLCreationChart(userID, dateRange, period)
	analytics.ClicksChart = s.getClicksChart(userID, dateRange, period)
	analytics.ActivityChart = s.getActivityChart(userID, dateRange, period)
	analytics.DeviceBreakdown = summary.TopDevices
	analytics.BrowserBreakdown = summary.TopBrowsers

	// Get top performing URLs
	analytics.TopURLs = s.getTopURLs(userID, dateRange)

	// Get recent activities
	activityReq := &models.UserActivityLogRequest{
		StartDate: startDate,
		EndDate:   time.Now(),
		Limit:     10,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	
	activityLog, err := s.GetUserActivityLog(userID, activityReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activities: %w", err)
	}
	analytics.RecentActivities = activityLog.Activities

	return analytics, nil
}

// Helper functions

func (s *UserAnalyticsService) buildDateRangeClause(dateRange models.DateRange, userID int64) (string, []interface{}) {
	var whereClause string
	args := []interface{}{userID}

	if !dateRange.StartDate.IsZero() && !dateRange.EndDate.IsZero() {
		whereClause = " AND started_at BETWEEN $2 AND $3"
		args = append(args, dateRange.StartDate, dateRange.EndDate)
	} else if !dateRange.StartDate.IsZero() {
		whereClause = " AND started_at >= $2"
		args = append(args, dateRange.StartDate)
	} else if !dateRange.EndDate.IsZero() {
		whereClause = " AND started_at <= $2"
		args = append(args, dateRange.EndDate)
	}

	return whereClause, args
}

func (s *UserAnalyticsService) calculateEngagementScore(summary *models.UserAnalyticsSummary) float64 {
	// Engagement score algorithm based on multiple factors
	score := 0.0

	// Sessions factor (0-30 points)
	if summary.TotalSessions > 0 {
		sessionScore := float64(summary.TotalSessions) * 2
		if sessionScore > 30 {
			sessionScore = 30
		}
		score += sessionScore
	}

	// URL creation factor (0-25 points)
	if summary.TotalURLsCreated > 0 {
		urlScore := float64(summary.TotalURLsCreated) * 1.5
		if urlScore > 25 {
			urlScore = 25
		}
		score += urlScore
	}

	// Click factor (0-20 points)
	if summary.TotalClicks > 0 {
		clickScore := float64(summary.TotalClicks) * 0.5
		if clickScore > 20 {
			clickScore = 20
		}
		score += clickScore
	}

	// Session duration factor (0-15 points)
	if summary.AvgSessionDuration > 0 {
		durationScore := summary.AvgSessionDuration * 0.3
		if durationScore > 15 {
			durationScore = 15
		}
		score += durationScore
	}

	// Active days factor (0-10 points)
	if summary.TotalActiveDays > 0 {
		activeScore := float64(summary.TotalActiveDays) * 0.5
		if activeScore > 10 {
			activeScore = 10
		}
		score += activeScore
	}

	// Normalize to 0-100 scale
	return score
}

func (s *UserAnalyticsService) calculateActivityScore(activityCounts map[string]int64) float64 {
	total := int64(0)
	for _, count := range activityCounts {
		total += count
	}
	
	if total == 0 {
		return 0
	}

	// Higher activity = higher score (max 100)
	score := float64(total) * 2
	if score > 100 {
		score = 100
	}
	
	return score
}

func (s *UserAnalyticsService) calculateConsistencyScore(userID int64, dateRange models.DateRange) float64 {
	// Calculate how consistently the user is active
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as activities
		FROM user_activity_logs 
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY DATE(created_at)
		ORDER BY date`

	rows, err := s.db.Query(query, userID, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		return 0
	}
	defer rows.Close()

	var dailyActivities []int64
	for rows.Next() {
		var date time.Time
		var count int64
		if err := rows.Scan(&date, &count); err != nil {
			continue
		}
		dailyActivities = append(dailyActivities, count)
	}

	if len(dailyActivities) == 0 {
		return 0
	}

	// Calculate consistency based on variance
	// Lower variance = higher consistency
	mean := float64(0)
	for _, count := range dailyActivities {
		mean += float64(count)
	}
	mean /= float64(len(dailyActivities))

	variance := float64(0)
	for _, count := range dailyActivities {
		variance += (float64(count) - mean) * (float64(count) - mean)
	}
	variance /= float64(len(dailyActivities))

	// Convert to score (0-100, higher is better)
	if variance == 0 {
		return 100 // Perfect consistency
	}
	
	score := 100 - (variance * 10) // Adjust multiplier as needed
	if score < 0 {
		score = 0
	}
	
	return score
}

func (s *UserAnalyticsService) calculateDiversityScore(features []string) float64 {
	// Score based on number of different features used
	uniqueFeatures := len(features)
	
	// Assuming there are ~15 different activity types
	maxFeatures := 15.0
	score := (float64(uniqueFeatures) / maxFeatures) * 100
	
	if score > 100 {
		score = 100
	}
	
	return score
}

func (s *UserAnalyticsService) getTopDevices(userID int64, dateRange models.DateRange) []models.DeviceBreakdown {
	whereClause, args := s.buildDateRangeClause(dateRange, userID)
	query := fmt.Sprintf(`
		SELECT device_type, COUNT(*) as count
		FROM user_sessions 
		WHERE user_id = $1 %s AND device_type IS NOT NULL
		GROUP BY device_type 
		ORDER BY count DESC 
		LIMIT 5`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []models.DeviceBreakdown{}
	}
	defer rows.Close()

	var devices []models.DeviceBreakdown
	var total int64
	for rows.Next() {
		var device models.DeviceBreakdown
		if err := rows.Scan(&device.DeviceType, &device.Count); err != nil {
			continue
		}
		devices = append(devices, device)
		total += device.Count
	}

	// Calculate percentages
	for i := range devices {
		if total > 0 {
			devices[i].Percentage = float64(devices[i].Count) / float64(total) * 100
		}
	}

	return devices
}

func (s *UserAnalyticsService) getTopBrowsers(userID int64, dateRange models.DateRange) []models.BrowserBreakdown {
	whereClause, args := s.buildDateRangeClause(dateRange, userID)
	query := fmt.Sprintf(`
		SELECT browser, COUNT(*) as count
		FROM user_sessions 
		WHERE user_id = $1 %s AND browser IS NOT NULL
		GROUP BY browser 
		ORDER BY count DESC 
		LIMIT 5`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []models.BrowserBreakdown{}
	}
	defer rows.Close()

	var browsers []models.BrowserBreakdown
	var total int64
	for rows.Next() {
		var browser models.BrowserBreakdown
		if err := rows.Scan(&browser.Browser, &browser.Count); err != nil {
			continue
		}
		browsers = append(browsers, browser)
		total += browser.Count
	}

	// Calculate percentages
	for i := range browsers {
		if total > 0 {
			browsers[i].Percentage = float64(browsers[i].Count) / float64(total) * 100
		}
	}

	return browsers
}

func (s *UserAnalyticsService) getTopLocations(userID int64, dateRange models.DateRange) []models.LocationBreakdown {
	whereClause, args := s.buildDateRangeClause(dateRange, userID)
	query := fmt.Sprintf(`
		SELECT location, COUNT(*) as count
		FROM user_sessions 
		WHERE user_id = $1 %s AND location IS NOT NULL
		GROUP BY location 
		ORDER BY count DESC 
		LIMIT 10`, whereClause)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []models.LocationBreakdown{}
	}
	defer rows.Close()

	var locations []models.LocationBreakdown
	var total int64
	for rows.Next() {
		var location models.LocationBreakdown
		if err := rows.Scan(&location.Location, &location.Count); err != nil {
			continue
		}
		locations = append(locations, location)
		total += location.Count
	}

	// Calculate percentages
	for i := range locations {
		if total > 0 {
			locations[i].Percentage = float64(locations[i].Count) / float64(total) * 100
		}
	}

	return locations
}

func (s *UserAnalyticsService) getURLCreationChart(userID int64, dateRange models.DateRange, period string) []models.ChartDataPoint {
	var groupBy string
	switch period {
	case "today":
		groupBy = "DATE_TRUNC('hour', created_at)"
	case "week", "month":
		groupBy = "DATE_TRUNC('day', created_at)"
	case "year":
		groupBy = "DATE_TRUNC('month', created_at)"
	default:
		groupBy = "DATE_TRUNC('day', created_at)"
	}

	query := fmt.Sprintf(`
		SELECT %s as period, COUNT(*) as value
		FROM urls 
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY %s
		ORDER BY period`, groupBy, groupBy)

	rows, err := s.db.Query(query, userID, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		return []models.ChartDataPoint{}
	}
	defer rows.Close()

	var data []models.ChartDataPoint
	for rows.Next() {
		var point models.ChartDataPoint
		var period time.Time
		if err := rows.Scan(&period, &point.Value); err != nil {
			continue
		}
		point.Label = s.formatChartLabel(period, period)
		data = append(data, point)
	}

	return data
}

func (s *UserAnalyticsService) getClicksChart(userID int64, dateRange models.DateRange, period string) []models.ChartDataPoint {
	var groupBy string
	switch period {
	case "today":
		groupBy = "DATE_TRUNC('hour', created_at)"
	case "week", "month":
		groupBy = "DATE_TRUNC('day', created_at)"
	case "year":
		groupBy = "DATE_TRUNC('month', created_at)"
	default:
		groupBy = "DATE_TRUNC('day', created_at)"
	}

	query := fmt.Sprintf(`
		SELECT %s as period, COALESCE(SUM(click_count), 0) as value
		FROM urls 
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY %s
		ORDER BY period`, groupBy, groupBy)

	rows, err := s.db.Query(query, userID, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		return []models.ChartDataPoint{}
	}
	defer rows.Close()

	var data []models.ChartDataPoint
	for rows.Next() {
		var point models.ChartDataPoint
		var period time.Time
		if err := rows.Scan(&period, &point.Value); err != nil {
			continue
		}
		point.Label = s.formatChartLabel(period, period)
		data = append(data, point)
	}

	return data
}

func (s *UserAnalyticsService) getActivityChart(userID int64, dateRange models.DateRange, period string) []models.ChartDataPoint {
	var groupBy string
	switch period {
	case "today":
		groupBy = "DATE_TRUNC('hour', created_at)"
	case "week", "month":
		groupBy = "DATE_TRUNC('day', created_at)"
	case "year":
		groupBy = "DATE_TRUNC('month', created_at)"
	default:
		groupBy = "DATE_TRUNC('day', created_at)"
	}

	query := fmt.Sprintf(`
		SELECT %s as period, COUNT(*) as value
		FROM user_activity_logs 
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY %s
		ORDER BY period`, groupBy, groupBy)

	rows, err := s.db.Query(query, userID, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		return []models.ChartDataPoint{}
	}
	defer rows.Close()

	var data []models.ChartDataPoint
	for rows.Next() {
		var point models.ChartDataPoint
		var period time.Time
		if err := rows.Scan(&period, &point.Value); err != nil {
			continue
		}
		point.Label = s.formatChartLabel(period, period)
		data = append(data, point)
	}

	return data
}

func (s *UserAnalyticsService) getTopURLs(userID int64, dateRange models.DateRange) []models.TopURL {
	query := `
		SELECT id, short_code, original_url, title, click_count, created_at
		FROM urls 
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY click_count DESC, created_at DESC
		LIMIT 10`

	rows, err := s.db.Query(query, userID, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		return []models.TopURL{}
	}
	defer rows.Close()

	var urls []models.TopURL
	for rows.Next() {
		var url models.TopURL
		if err := rows.Scan(&url.ID, &url.ShortCode, &url.OriginalURL, &url.Title, &url.ClickCount, &url.CreatedAt); err != nil {
			continue
		}
		urls = append(urls, url)
	}

	return urls
}

func (s *UserAnalyticsService) formatChartLabel(period time.Time, periodType string) string {
	switch periodType {
	case "today":
		return period.Format("15:04")
	case "week", "month":
		return period.Format("Jan 2")
	case "year":
		return period.Format("Jan 2006")
	default:
		return period.Format("Jan 2")
	}
}