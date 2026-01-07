package admin

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminMonitoringHandlers handles admin monitoring operations
type AdminMonitoringHandlers struct {
	monitoringService *services.MonitoringService
}

// NewAdminMonitoringHandlers creates a new admin monitoring handlers instance
func NewAdminMonitoringHandlers(monitoringService *services.MonitoringService) *AdminMonitoringHandlers {
	return &AdminMonitoringHandlers{
		monitoringService: monitoringService,
	}
}

// AdminGetSystemHealth retrieves overall system health status
// GET /api/v1/admin/monitoring/health
func (h *AdminMonitoringHandlers) AdminGetSystemHealth(c *gin.Context) {
	health, err := h.monitoringService.GetSystemHealth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve system health",
		})
		return
	}

	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == "degraded" {
		statusCode = http.StatusOK // Still OK, but degraded
	}

	c.JSON(statusCode, gin.H{
		"health": health,
	})
}

// AdminGetDatabaseHealth retrieves database health status
// GET /api/v1/admin/monitoring/health/database
func (h *AdminMonitoringHandlers) AdminGetDatabaseHealth(c *gin.Context) {
	health, err := h.monitoringService.GetDatabaseHealth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "database_unhealthy",
			Message: err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"database": health,
	})
}

// AdminGetRedisHealth retrieves Redis health status
// GET /api/v1/admin/monitoring/health/redis
func (h *AdminMonitoringHandlers) AdminGetRedisHealth(c *gin.Context) {
	health, err := h.monitoringService.GetRedisHealth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "redis_unhealthy",
			Message: err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == "unavailable" {
		statusCode = http.StatusOK // Redis is optional
	}

	c.JSON(statusCode, gin.H{
		"redis": health,
	})
}

// AdminGetSystemMetrics retrieves system resource metrics
// GET /api/v1/admin/monitoring/metrics
func (h *AdminMonitoringHandlers) AdminGetSystemMetrics(c *gin.Context) {
	metrics, err := h.monitoringService.GetSystemMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve system metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// AdminListAlerts retrieves system alerts with filtering
// GET /api/v1/admin/monitoring/alerts
func (h *AdminMonitoringHandlers) AdminListAlerts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	filter := models.AlertFilter{
		Severity:  c.Query("severity"),
		Status:    c.Query("status"),
		Service:   c.Query("service"),
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	alerts, total, err := h.monitoringService.ListAlerts(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve alerts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts":      alerts,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// AdminGetAlert retrieves a specific alert by ID
// GET /api/v1/admin/monitoring/alerts/:id
func (h *AdminMonitoringHandlers) AdminGetAlert(c *gin.Context) {
	alertID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_alert_id",
			Message: "Invalid alert ID",
		})
		return
	}

	alert, err := h.monitoringService.GetAlert(c.Request.Context(), alertID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "alert_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alert": alert,
	})
}

// AdminAcknowledgeAlert marks an alert as acknowledged
// POST /api/v1/admin/monitoring/alerts/:id/acknowledge
func (h *AdminMonitoringHandlers) AdminAcknowledgeAlert(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	alertID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_alert_id",
			Message: "Invalid alert ID",
		})
		return
	}

	var req struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Note = ""
	}

	if err := h.monitoringService.AcknowledgeAlert(c.Request.Context(), alertID, adminID.(int64), req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "acknowledgment_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert acknowledged successfully",
	})
}

// AdminResolveAlert marks an alert as resolved
// POST /api/v1/admin/monitoring/alerts/:id/resolve
func (h *AdminMonitoringHandlers) AdminResolveAlert(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	alertID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_alert_id",
			Message: "Invalid alert ID",
		})
		return
	}

	var req struct {
		Resolution string `json:"resolution" binding:"required"`
		Note       string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.monitoringService.ResolveAlert(c.Request.Context(), alertID, adminID.(int64), req.Resolution, req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "resolution_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert resolved successfully",
	})
}

// AdminGetErrorLogs retrieves recent error logs
// GET /api/v1/admin/monitoring/errors
func (h *AdminMonitoringHandlers) AdminGetErrorLogs(c *gin.Context) {
	level := c.DefaultQuery("level", "error")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit < 1 || limit > 1000 {
		limit = 100
	}

	logs, err := h.monitoringService.GetErrorLogs(c.Request.Context(), level, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve error logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"level": level,
		"limit": limit,
	})
}

// AdminGetServicesHealth retrieves health status of all services
// GET /api/v1/admin/monitoring/health/services
func (h *AdminMonitoringHandlers) AdminGetServicesHealth(c *gin.Context) {
	// TODO: Implement services health check
	c.JSON(http.StatusOK, gin.H{
		"services": gin.H{
			"database": "healthy",
			"redis":    "healthy",
			"api":      "healthy",
		},
	})
}

// AdminGetPerformanceMetrics retrieves performance metrics
// GET /api/v1/admin/monitoring/metrics/performance
func (h *AdminMonitoringHandlers) AdminGetPerformanceMetrics(c *gin.Context) {
	// TODO: Implement performance metrics
	c.JSON(http.StatusOK, gin.H{
		"cpu_usage":    0,
		"memory_usage": 0,
		"response_time": 0,
	})
}
