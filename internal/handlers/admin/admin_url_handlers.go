package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminURLHandlers handles admin URL management operations
type AdminURLHandlers struct {
	adminURLService *services.AdminURLService
	auditService    *services.AdminAuditService
}

// NewAdminURLHandlers creates a new admin URL handlers instance
func NewAdminURLHandlers(adminURLService *services.AdminURLService, auditService *services.AdminAuditService) *AdminURLHandlers {
	return &AdminURLHandlers{
		adminURLService: adminURLService,
		auditService:    auditService,
	}
}

// AdminListURLs lists all URLs with filtering and pagination
// GET /api/v1/admin/urls
func (h *AdminURLHandlers) AdminListURLs(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	filter := models.URLFilter{
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	// Parse optional user_id filter
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			filter.UserID = &userID
		}
	}

	// Parse date range
	if createdFrom := c.Query("created_from"); createdFrom != "" {
		if t, err := time.Parse(time.RFC3339, createdFrom); err == nil {
			filter.CreatedFrom = &t
		}
	}
	if createdTo := c.Query("created_to"); createdTo != "" {
		if t, err := time.Parse(time.RFC3339, createdTo); err == nil {
			filter.CreatedTo = &t
		}
	}

	urls, total, err := h.adminURLService.ListURLs(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve URLs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"urls":        urls,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// AdminGetURL retrieves a specific URL by short code
// GET /api/v1/admin/urls/:shortCode
func (h *AdminURLHandlers) AdminGetURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	url, err := h.adminURLService.GetURLByShortCode(c.Request.Context(), shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "url_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

// AdminUpdateURLStatus updates the status of a URL (activate/suspend)
// PUT /api/v1/admin/urls/:shortCode/status
func (h *AdminURLHandlers) AdminUpdateURLStatus(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	var req models.UpdateURLStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.adminURLService.UpdateURLStatus(c.Request.Context(), shortCode, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.url.update_status", "url", nil, gin.H{
		"short_code": shortCode,
		"status":     req.Status,
		"reason":     req.Reason,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "URL status updated successfully",
	})
}

// AdminDeleteURL deletes a URL permanently
// DELETE /api/v1/admin/urls/:shortCode
func (h *AdminURLHandlers) AdminDeleteURL(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	if err := h.adminURLService.DeleteURL(c.Request.Context(), shortCode); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.url.delete", "url", nil, gin.H{
		"short_code": shortCode,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "URL deleted successfully",
	})
}

// AdminTransferURL transfers URL ownership to a different user
// PUT /api/v1/admin/urls/:shortCode/transfer
func (h *AdminURLHandlers) AdminTransferURL(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	var req models.TransferURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.adminURLService.TransferURLOwnership(c.Request.Context(), shortCode, req.NewOwnerID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "transfer_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.url.transfer", "url", nil, gin.H{
		"short_code":   shortCode,
		"new_owner_id": req.NewOwnerID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "URL ownership transferred successfully",
	})
}

// AdminExtendExpiration extends the expiration date of a URL
// PUT /api/v1/admin/urls/:shortCode/extend
func (h *AdminURLHandlers) AdminExtendExpiration(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	var req models.ExtendExpirationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.adminURLService.ExtendExpiration(c.Request.Context(), shortCode, req.ExpiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "extension_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.url.extend_expiration", "url", nil, gin.H{
		"short_code": shortCode,
		"new_expiry": req.ExpiresAt,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "URL expiration extended successfully",
	})
}

// AdminGetFlaggedURLs retrieves all flagged URLs
// GET /api/v1/admin/urls/flagged
func (h *AdminURLHandlers) AdminGetFlaggedURLs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	urls, total, err := h.adminURLService.GetFlaggedURLs(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve flagged URLs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"urls":        urls,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// AdminFlagURL flags a URL as problematic
// POST /api/v1/admin/urls/:shortCode/flag
func (h *AdminURLHandlers) AdminFlagURL(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	var req models.FlagURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.adminURLService.FlagURL(c.Request.Context(), shortCode, req, adminID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "flag_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.url.flag", "url", nil, gin.H{
		"short_code": shortCode,
		"reason":     req.Reason,
		"category":   req.Category,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "URL flagged successfully",
	})
}

// AdminUnflagURL removes flag from a URL
// POST /api/v1/admin/urls/:shortCode/unflag
func (h *AdminURLHandlers) AdminUnflagURL(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	if err := h.adminURLService.UnflagURL(c.Request.Context(), shortCode, adminID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "unflag_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.url.unflag", "url", nil, gin.H{
		"short_code": shortCode,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "URL unflagged successfully",
	})
}
