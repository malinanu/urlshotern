package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminAuditHandlers handles admin audit log operations
type AdminAuditHandlers struct {
	auditService *services.AdminAuditService
}

// NewAdminAuditHandlers creates a new admin audit handlers instance
func NewAdminAuditHandlers(auditService *services.AdminAuditService) *AdminAuditHandlers {
	return &AdminAuditHandlers{
		auditService: auditService,
	}
}

// AdminListAuditLogs lists audit logs with filtering and pagination
// GET /api/v1/admin/audit/logs
func (h *AdminAuditHandlers) AdminListAuditLogs(c *gin.Context) {
	// Parse pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Parse filters
	filter := models.AuditLogFilter{
		Username:     c.Query("username"),
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		ResourceID:   c.Query("resource_id"),
		Status:       c.Query("status"),
		Severity:     c.Query("severity"),
		IPAddress:    c.Query("ip_address"),
		Limit:        limit,
		Offset:       offset,
	}

	// Parse optional user_id
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			filter.UserID = &userID
		}
	}

	// Parse date range
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = t
		}
	}

	response, err := h.auditService.ListAuditLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AdminGetAuditLog retrieves a specific audit log by ID
// GET /api/v1/admin/audit/logs/:id
func (h *AdminAuditHandlers) AdminGetAuditLog(c *gin.Context) {
	logID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_log_id",
			Message: "Invalid audit log ID",
		})
		return
	}

	log, err := h.auditService.GetAuditLog(c.Request.Context(), logID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "log_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"log": log,
	})
}

// AdminGetUserAuditTrail retrieves audit trail for a specific user
// GET /api/v1/admin/audit/users/:userId
func (h *AdminAuditHandlers) AdminGetUserAuditTrail(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response, err := h.auditService.GetUserAuditTrail(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve user audit trail",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AdminGetResourceAuditTrail retrieves audit trail for a specific resource
// GET /api/v1/admin/audit/resources/:resourceType/:resourceId
func (h *AdminAuditHandlers) AdminGetResourceAuditTrail(c *gin.Context) {
	resourceType := c.Param("resourceType")
	resourceID := c.Param("resourceId")

	if resourceType == "" || resourceID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_parameters",
			Message: "Resource type and ID are required",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	response, err := h.auditService.GetResourceAuditTrail(c.Request.Context(), resourceType, resourceID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve resource audit trail",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AdminGetCriticalActions retrieves recent critical audit actions
// GET /api/v1/admin/audit/critical
func (h *AdminAuditHandlers) AdminGetCriticalActions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, err := h.auditService.GetRecentCriticalActions(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve critical actions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
	})
}

// AdminGetAuditStatistics retrieves audit log statistics
// GET /api/v1/admin/audit/statistics
func (h *AdminAuditHandlers) AdminGetAuditStatistics(c *gin.Context) {
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

	stats, err := h.auditService.GetAuditStatistics(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve audit statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
		"period": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}

// AdminExportAuditLogs exports audit logs (placeholder for CSV/JSON export)
// GET /api/v1/admin/audit/export
func (h *AdminAuditHandlers) AdminExportAuditLogs(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Parse filters (same as list)
	filter := models.AuditLogFilter{
		Username:     c.Query("username"),
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		Status:       c.Query("status"),
		Severity:     c.Query("severity"),
		Limit:        1000, // Max export limit
		Offset:       0,
	}

	// Get audit logs
	response, err := h.auditService.ListAuditLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "export_failed",
			Message: err.Error(),
		})
		return
	}

	// Log the export action
	userID := adminID.(int64)
	_ = h.auditService.LogAction(c.Request.Context(), models.CreateAuditLogRequest{
		UserID:       &userID,
		Action:       "admin.audit.export",
		ResourceType: "audit_logs",
		Status:       models.AuditStatusSuccess,
		Severity:     models.AuditSeverityInfo,
	})

	// TODO: Convert to CSV format
	// For now, return JSON
	c.JSON(http.StatusOK, response)
}

// AdminCleanupAuditLogs deletes old audit logs
// POST /api/v1/admin/audit/cleanup
func (h *AdminAuditHandlers) AdminCleanupAuditLogs(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req struct {
		RetentionDays int `json:"retention_days" binding:"required,min=30"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	deletedCount, err := h.auditService.CleanupOldLogs(c.Request.Context(), req.RetentionDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "cleanup_failed",
			Message: err.Error(),
		})
		return
	}

	// Log the cleanup action
	userID := adminID.(int64)
	_ = h.auditService.LogAction(c.Request.Context(), models.CreateAuditLogRequest{
		UserID:       &userID,
		Action:       "admin.audit.cleanup",
		ResourceType: "audit_logs",
		Details: map[string]interface{}{
			"retention_days": req.RetentionDays,
			"deleted_count":  deletedCount,
		},
		Status:   models.AuditStatusSuccess,
		Severity: models.AuditSeverityWarning,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":       "Audit logs cleaned up successfully",
		"deleted_count": deletedCount,
		"retention_days": req.RetentionDays,
	})
}
