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
	// TODO: Implement geographic analytics
	c.JSON(http.StatusOK, gin.H{
		"countries": []interface{}{},
	})
}

// AdminGetDeviceAnalytics retrieves device/browser analytics
// GET /api/v1/admin/analytics/devices
func (h *AdminAnalyticsHandlers) AdminGetDeviceAnalytics(c *gin.Context) {
	// TODO: Implement device analytics
	c.JSON(http.StatusOK, gin.H{
		"devices": []interface{}{},
	})
}

// AdminGetTrendAnalytics retrieves trend analytics
// GET /api/v1/admin/analytics/trends
func (h *AdminAnalyticsHandlers) AdminGetTrendAnalytics(c *gin.Context) {
	// TODO: Implement trend analytics
	c.JSON(http.StatusOK, gin.H{
		"trends": []interface{}{},
	})
}
