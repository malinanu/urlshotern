package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminAnalyticsHandlers handles admin analytics operations
type AdminAnalyticsHandlers struct {
	analyticsService *services.AdminAnalyticsService
}

// NewAdminAnalyticsHandlers creates a new admin analytics handlers instance
func NewAdminAnalyticsHandlers(analyticsService *services.AdminAnalyticsService) *AdminAnalyticsHandlers {
	return &AdminAnalyticsHandlers{
		analyticsService: analyticsService,
	}
}

// AdminGetDashboard retrieves dashboard overview statistics
// GET /api/v1/admin/analytics/dashboard
func (h *AdminAnalyticsHandlers) AdminGetDashboard(c *gin.Context) {
	dashboard, err := h.analyticsService.GetDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve dashboard data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dashboard": dashboard,
	})
}

// AdminGetUserAnalytics retrieves user growth analytics
// GET /api/v1/admin/analytics/users
func (h *AdminAnalyticsHandlers) AdminGetUserAnalytics(c *gin.Context) {
	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}

	analytics, err := h.analyticsService.GetUserAnalytics(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve user analytics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analytics": analytics,
		"period": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

// AdminGetURLAnalytics retrieves URL creation and usage analytics
// GET /api/v1/admin/analytics/urls
func (h *AdminAnalyticsHandlers) AdminGetURLAnalytics(c *gin.Context) {
	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}

	analytics, err := h.analyticsService.GetURLAnalytics(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve URL analytics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analytics": analytics,
		"period": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

// AdminGetClickAnalytics retrieves click statistics
// GET /api/v1/admin/analytics/clicks
func (h *AdminAnalyticsHandlers) AdminGetClickAnalytics(c *gin.Context) {
	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}

	clickData, err := h.analyticsService.GetClickAnalytics(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve click analytics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clicks": clickData,
		"period": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

// AdminGetTopURLs retrieves top performing URLs
// GET /api/v1/admin/analytics/top-urls
func (h *AdminAnalyticsHandlers) AdminGetTopURLs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}

	topURLs, err := h.analyticsService.GetTopURLs(c.Request.Context(), limit, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve top URLs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"urls":  topURLs,
		"limit": limit,
		"period": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

// AdminExportReport generates and exports an analytics report
// POST /api/v1/admin/analytics/export
func (h *AdminAnalyticsHandlers) AdminExportReport(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req models.ExportReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	data, err := h.analyticsService.ExportReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "export_failed",
			Message: err.Error(),
		})
		return
	}

	// Determine content type based on format
	var contentType string
	var filename string
	switch req.Format {
	case "csv":
		contentType = "text/csv"
		filename = "analytics_report.csv"
	case "json":
		contentType = "application/json"
		filename = "analytics_report.json"
	case "xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "analytics_report.xlsx"
	default:
		contentType = "text/csv"
		filename = "analytics_report.csv"
	}

	_ = adminID // TODO: Log to audit trail

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

// AdminGetGeographicAnalytics retrieves geographic analytics
// GET /api/v1/admin/analytics/geography
func (h *AdminAnalyticsHandlers) AdminGetGeographicAnalytics(c *gin.Context) {
	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}

	// TODO: Query from click_analytics table with IP geolocation data
	// For now, return placeholder data showing geographic distribution
	geography := gin.H{
		"countries": []gin.H{
			{
				"country_code": "US",
				"country_name": "United States",
				"clicks":       1250,
				"percentage":   35.5,
			},
			{
				"country_code": "GB",
				"country_name": "United Kingdom",
				"clicks":       520,
				"percentage":   14.8,
			},
			{
				"country_code": "CA",
				"country_name": "Canada",
				"clicks":       380,
				"percentage":   10.8,
			},
			{
				"country_code": "DE",
				"country_name": "Germany",
				"clicks":       290,
				"percentage":   8.2,
			},
			{
				"country_code": "FR",
				"country_name": "France",
				"clicks":       245,
				"percentage":   7.0,
			},
		},
		"regions": []gin.H{
			{"region": "North America", "clicks": 1630, "percentage": 46.3},
			{"region": "Europe", "clicks": 1125, "percentage": 32.0},
			{"region": "Asia", "clicks": 520, "percentage": 14.8},
			{"region": "Other", "clicks": 245, "percentage": 6.9},
		},
		"total_clicks": 3520,
		"period": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
	}

	c.JSON(http.StatusOK, geography)
}

// AdminGetDeviceAnalytics retrieves device/browser analytics
// GET /api/v1/admin/analytics/devices
func (h *AdminAnalyticsHandlers) AdminGetDeviceAnalytics(c *gin.Context) {
	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}

	// TODO: Parse User-Agent from click_analytics table
	// For now, return placeholder data showing device/browser breakdown
	devices := gin.H{
		"devices": []gin.H{
			{
				"device_type": "Desktop",
				"clicks":      1980,
				"percentage":  56.3,
			},
			{
				"device_type": "Mobile",
				"clicks":      1230,
				"percentage":  35.0,
			},
			{
				"device_type": "Tablet",
				"clicks":      310,
				"percentage":  8.7,
			},
		},
		"browsers": []gin.H{
			{
				"browser": "Chrome",
				"clicks":  1560,
				"percentage": 44.3,
			},
			{
				"browser": "Safari",
				"clicks":  920,
				"percentage": 26.1,
			},
			{
				"browser": "Firefox",
				"clicks":  530,
				"percentage": 15.1,
			},
			{
				"browser": "Edge",
				"clicks":  380,
				"percentage": 10.8,
			},
			{
				"browser": "Other",
				"clicks":  130,
				"percentage": 3.7,
			},
		},
		"operating_systems": []gin.H{
			{
				"os": "Windows",
				"clicks": 1450,
				"percentage": 41.2,
			},
			{
				"os": "macOS",
				"clicks": 870,
				"percentage": 24.7,
			},
			{
				"os": "iOS",
				"clicks": 640,
				"percentage": 18.2,
			},
			{
				"os": "Android",
				"clicks": 450,
				"percentage": 12.8,
			},
			{
				"os": "Linux",
				"clicks": 110,
				"percentage": 3.1,
			},
		},
		"total_clicks": 3520,
		"period": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
	}

	c.JSON(http.StatusOK, devices)
}

// AdminGetTrendAnalytics retrieves trend analytics
// GET /api/v1/admin/analytics/trends
func (h *AdminAnalyticsHandlers) AdminGetTrendAnalytics(c *gin.Context) {
	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = t
		}
	}

	// Get metric type (default to clicks)
	metric := c.DefaultQuery("metric", "clicks")

	// TODO: Query time-series data from analytics tables
	// For now, return placeholder data showing trends over time

	// Generate sample daily data for the period
	days := int(endDate.Sub(startDate).Hours() / 24)
	dailyData := []gin.H{}

	// Simulate trend data
	for i := 0; i <= days; i++ {
		date := startDate.AddDate(0, 0, i)
		value := 100 + (i * 5) + (i % 7 * 10) // Simulate growth with weekly patterns

		dailyData = append(dailyData, gin.H{
			"date":  date.Format("2006-01-02"),
			"value": value,
		})
	}

	trends := gin.H{
		"metric": metric,
		"data":   dailyData,
		"summary": gin.H{
			"total":          3520,
			"average_daily":  117,
			"growth_rate":    12.5,  // Percentage
			"peak_value":     185,
			"peak_date":      startDate.AddDate(0, 0, 15).Format("2006-01-02"),
		},
		"period": gin.H{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
			"days":       days,
		},
	}

	c.JSON(http.StatusOK, trends)
}
