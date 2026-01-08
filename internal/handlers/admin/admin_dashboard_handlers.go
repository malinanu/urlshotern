package admin

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminDashboardHandlers handles admin dashboard data
type AdminDashboardHandlers struct {
	analyticsService  *services.AdminAnalyticsService
	monitoringService *services.MonitoringService
	auditService      *services.AdminAuditService
}

// NewAdminDashboardHandlers creates a new admin dashboard handlers instance
func NewAdminDashboardHandlers(
	analyticsService *services.AdminAnalyticsService,
	monitoringService *services.MonitoringService,
	auditService *services.AdminAuditService,
) *AdminDashboardHandlers {
	return &AdminDashboardHandlers{
		analyticsService:  analyticsService,
		monitoringService: monitoringService,
		auditService:      auditService,
	}
}

// GetDashboardOverview returns admin dashboard summary
// GET /api/v1/admin
func (h *AdminDashboardHandlers) GetDashboardOverview(c *gin.Context) {
	ctx := c.Request.Context()

	// Get system health
	health, err := h.monitoringService.GetSystemHealth(ctx)
	if err != nil {
		log.Printf("ERROR: GetSystemHealth failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve system health: " + err.Error(),
		})
		return
	}

	// Get dashboard statistics
	dashboard, err := h.analyticsService.GetDashboard(ctx)
	if err != nil {
		log.Printf("ERROR: GetDashboard failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve dashboard statistics: " + err.Error(),
		})
		return
	}

	// Get system metrics
	metrics, err := h.monitoringService.GetSystemMetrics(ctx)
	if err != nil {
		log.Printf("WARNING: GetSystemMetrics failed (non-critical): %v", err)
		metrics = nil // Just skip if unavailable
	}

	// Combine into response
	response := gin.H{
		"health": gin.H{
			"status":   health.Status,
			"database": health.Database.Status,
			"redis":    health.Redis.Status,
			"uptime":   health.Uptime,
		},
		"users": gin.H{
			"total":     dashboard.TotalUsers,
			"active":    dashboard.ActiveUsers,
			"new_today": dashboard.NewUsersToday,
		},
		"urls": gin.H{
			"total":         dashboard.TotalURLs,
			"total_clicks":  dashboard.TotalClicks,
			"clicks_today":  dashboard.TodayClicks,
			"top_performer": "", // Can be added if needed
		},
		"system": metrics,
	}

	c.JSON(http.StatusOK, response)
}

// AdminDashboardStats returns detailed dashboard statistics
// GET /api/v1/admin/stats
func (h *AdminDashboardHandlers) AdminDashboardStats(c *gin.Context) {
	ctx := context.Background()

	stats := gin.H{
		"timestamp": time.Now().Unix(),
	}

	// Dashboard stats
	dashboard, err := h.analyticsService.GetDashboard(ctx)
	if err == nil {
		stats["users"] = gin.H{
			"total":     dashboard.TotalUsers,
			"active":    dashboard.ActiveUsers,
			"new_today": dashboard.NewUsersToday,
		}
		stats["urls"] = gin.H{
			"total":        dashboard.TotalURLs,
			"total_clicks": dashboard.TotalClicks,
			"clicks_today": dashboard.TodayClicks,
		}
	}

	// System metrics
	metrics, err := h.monitoringService.GetSystemMetrics(ctx)
	if err == nil {
		stats["system"] = metrics
	}

	c.JSON(http.StatusOK, stats)
}
