package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"urlshortener/internal/models"
	"urlshortener/internal/services"
)

type AnalyticsHandlers struct {
	analyticsService *services.UserAnalyticsService
}

func NewAnalyticsHandlers(analyticsService *services.UserAnalyticsService) *AnalyticsHandlers {
	return &AnalyticsHandlers{
		analyticsService: analyticsService,
	}
}

// GetUserAnalyticsSummary returns comprehensive analytics summary for a user
func (h *AnalyticsHandlers) GetUserAnalyticsSummary(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Parse date range from query parameters
	dateRange := h.parseDateRange(c)

	summary, err := h.analyticsService.GetUserAnalyticsSummary(userID.(int64), dateRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "analytics_error",
			Message: "Failed to get analytics summary",
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetUserEngagementMetrics returns detailed engagement metrics
func (h *AnalyticsHandlers) GetUserEngagementMetrics(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Parse date range from query parameters
	dateRange := h.parseDateRange(c)

	metrics, err := h.analyticsService.GetUserEngagementMetrics(userID.(int64), dateRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "analytics_error",
			Message: "Failed to get engagement metrics",
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetUserActivityLog returns paginated user activity log
func (h *AnalyticsHandlers) GetUserActivityLog(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Parse request parameters
	req := &models.UserActivityLogRequest{}

	// Activity type filter
	if activityType := c.Query("activity_type"); activityType != "" {
		req.ActivityType = activityType
	}

	// URL ID filter
	if urlIdStr := c.Query("url_id"); urlIdStr != "" {
		if urlId, err := strconv.ParseInt(urlIdStr, 10, 64); err == nil {
			req.URLId = &urlId
		}
	}

	// Date range filters
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = endDate
		}
	}

	// Pagination parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	if req.Limit == 0 {
		req.Limit = 50 // Default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.ParseInt(offsetStr, 10, 64); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	// Sorting parameters
	if sortBy := c.Query("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	response, err := h.analyticsService.GetUserActivityLog(userID.(int64), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "analytics_error",
			Message: "Failed to get activity log",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDashboardAnalytics returns analytics data formatted for dashboard display
func (h *AnalyticsHandlers) GetDashboardAnalytics(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Get period from query parameter (default to "month")
	period := c.DefaultQuery("period", "month")
	validPeriods := map[string]bool{
		"today": true,
		"week":  true,
		"month": true,
		"year":  true,
	}

	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_period",
			Message: "Period must be one of: today, week, month, year",
		})
		return
	}

	analytics, err := h.analyticsService.GetDashboardAnalytics(userID.(int64), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "analytics_error",
			Message: "Failed to get dashboard analytics",
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// LogUserActivity allows manual logging of user activities (for client-side events)
func (h *AnalyticsHandlers) LogUserActivity(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req struct {
		ActivityType string                 `json:"activity_type" binding:"required"`
		Description  string                 `json:"description"`
		URLId        *int64                 `json:"url_id,omitempty"`
		Metadata     map[string]interface{} `json:"metadata,omitempty"`
		DurationMs   *int64                 `json:"duration_ms,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get session ID from context or generate a new one
	sessionID, _ := c.Get("sessionID")
	sessionIDStr, ok := sessionID.(string)
	if !ok || sessionIDStr == "" {
		sessionIDStr = "web-session-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	activity := &models.UserActivityLog{
		UserID:       userID.(int64),
		SessionID:    sessionIDStr,
		ActivityType: req.ActivityType,
		Description:  req.Description,
		URLId:        req.URLId,
		Metadata:     req.Metadata,
		IPAddress:    &c.ClientIP(),
		UserAgent:    &c.Request.UserAgent(),
		DurationMs:   req.DurationMs,
		CreatedAt:    time.Now(),
	}

	// Parse user agent for additional info
	h.parseUserAgent(activity, c.Request.UserAgent())

	err := h.analyticsService.LogUserActivity(activity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "logging_error",
			Message: "Failed to log activity",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Activity logged successfully",
	})
}

// StartSession manually starts a user session (for SPA applications)
func (h *AnalyticsHandlers) StartSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req struct {
		ScreenResolution *string `json:"screen_resolution,omitempty"`
		Timezone         *string `json:"timezone,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Ignore bind errors for optional data
	}

	sessionID := "web-session-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	session := &models.UserSession{
		ID:               sessionID,
		UserID:           userID.(int64),
		IPAddress:        c.ClientIP(),
		UserAgent:        c.Request.UserAgent(),
		ScreenResolution: req.ScreenResolution,
		Timezone:         req.Timezone,
		StartedAt:        time.Now(),
	}

	// Parse user agent for additional info
	h.parseUserAgentForSession(session, c.Request.UserAgent())

	err := h.analyticsService.StartUserSession(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "session_error",
			Message: "Failed to start session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"message":    "Session started successfully",
	})
}

// EndSession manually ends a user session
func (h *AnalyticsHandlers) EndSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	err := h.analyticsService.EndUserSession(req.SessionID, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "session_error",
			Message: "Failed to end session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session ended successfully",
	})
}

// Helper functions

func (h *AnalyticsHandlers) parseDateRange(c *gin.Context) models.DateRange {
	var dateRange models.DateRange

	// Parse start_date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			dateRange.StartDate = startDate
		}
	}

	// Parse end_date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set end of day
			dateRange.EndDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	// If no dates provided, default to last 30 days
	if dateRange.StartDate.IsZero() && dateRange.EndDate.IsZero() {
		dateRange.EndDate = time.Now()
		dateRange.StartDate = time.Now().AddDate(0, 0, -30)
	}

	return dateRange
}

func (h *AnalyticsHandlers) parseUserAgent(activity *models.UserActivityLog, userAgent string) {
	// Simple user agent parsing - in production you might want to use a library
	if userAgent == "" {
		return
	}

	// Device type detection
	if contains(userAgent, "Mobile") || contains(userAgent, "Android") || contains(userAgent, "iPhone") {
		activity.DeviceType = "mobile"
		activity.IsMobile = true
	} else if contains(userAgent, "Tablet") || contains(userAgent, "iPad") {
		activity.DeviceType = "tablet"
		activity.IsMobile = true
	} else {
		activity.DeviceType = "desktop"
		activity.IsMobile = false
	}

	// Browser detection
	if contains(userAgent, "Chrome") {
		activity.Browser = "Chrome"
	} else if contains(userAgent, "Firefox") {
		activity.Browser = "Firefox"
	} else if contains(userAgent, "Safari") && !contains(userAgent, "Chrome") {
		activity.Browser = "Safari"
	} else if contains(userAgent, "Edge") {
		activity.Browser = "Edge"
	} else {
		activity.Browser = "Unknown"
	}

	// Platform detection
	if contains(userAgent, "Windows") {
		activity.Platform = "Windows"
	} else if contains(userAgent, "Macintosh") || contains(userAgent, "Mac OS") {
		activity.Platform = "macOS"
	} else if contains(userAgent, "Linux") {
		activity.Platform = "Linux"
	} else if contains(userAgent, "Android") {
		activity.Platform = "Android"
	} else if contains(userAgent, "iOS") || contains(userAgent, "iPhone") || contains(userAgent, "iPad") {
		activity.Platform = "iOS"
	} else {
		activity.Platform = "Unknown"
	}

	// Bot detection
	if contains(userAgent, "bot") || contains(userAgent, "crawler") || contains(userAgent, "spider") {
		activity.IsBot = true
	} else {
		activity.IsBot = false
	}
}

func (h *AnalyticsHandlers) parseUserAgentForSession(session *models.UserSession, userAgent string) {
	// Similar parsing for session
	if userAgent == "" {
		return
	}

	// Device type detection
	if contains(userAgent, "Mobile") || contains(userAgent, "Android") || contains(userAgent, "iPhone") {
		session.DeviceType = "mobile"
		session.IsMobile = true
	} else if contains(userAgent, "Tablet") || contains(userAgent, "iPad") {
		session.DeviceType = "tablet"
		session.IsMobile = true
	} else {
		session.DeviceType = "desktop"
		session.IsMobile = false
	}

	// Browser detection
	if contains(userAgent, "Chrome") {
		session.Browser = "Chrome"
	} else if contains(userAgent, "Firefox") {
		session.Browser = "Firefox"
	} else if contains(userAgent, "Safari") && !contains(userAgent, "Chrome") {
		session.Browser = "Safari"
	} else if contains(userAgent, "Edge") {
		session.Browser = "Edge"
	} else {
		session.Browser = "Unknown"
	}

	// Platform detection
	if contains(userAgent, "Windows") {
		session.Platform = "Windows"
	} else if contains(userAgent, "Macintosh") || contains(userAgent, "Mac OS") {
		session.Platform = "macOS"
	} else if contains(userAgent, "Linux") {
		session.Platform = "Linux"
	} else if contains(userAgent, "Android") {
		session.Platform = "Android"
	} else if contains(userAgent, "iOS") || contains(userAgent, "iPhone") || contains(userAgent, "iPad") {
		session.Platform = "iOS"
	} else {
		session.Platform = "Unknown"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)+1] == substr+" " || 
		s[len(s)-len(substr)-1:] == " "+substr || 
		findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}