package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type DashboardHandlers struct {
	shortenerService  *services.ShortenerService
	analyticsService  *services.AnalyticsService
	userService       *services.UserService
	rbacService       *services.RBACService
}

// NewDashboardHandlers creates new dashboard handlers
func NewDashboardHandlers(
	shortenerService *services.ShortenerService,
	analyticsService *services.AnalyticsService,
	userService *services.UserService,
	rbacService *services.RBACService,
) *DashboardHandlers {
	return &DashboardHandlers{
		shortenerService: shortenerService,
		analyticsService: analyticsService,
		userService:      userService,
		rbacService:      rbacService,
	}
}

// Dashboard Overview

// GetDashboardOverview provides overview statistics for user dashboard
// GET /api/v1/dashboard/overview
func (h *DashboardHandlers) GetDashboardOverview(c *gin.Context) {
	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get user's URL statistics
	stats, err := h.getUserURLStats(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "stats_failed",
			Message: "Failed to retrieve user statistics",
		})
		return
	}

	// Get recent activity
	recentURLs, err := h.getRecentURLs(userID.(int64), 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "recent_urls_failed",
			Message: "Failed to retrieve recent URLs",
		})
		return
	}

	// Get analytics summary
	analyticsSummary, err := h.getAnalyticsSummary(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "analytics_failed",
			Message: "Failed to retrieve analytics summary",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
		"recent_urls": recentURLs,
		"analytics": analyticsSummary,
		"account_info": h.getAccountInfo(c),
	})
}

// URL Management

// GetUserURLs retrieves all URLs created by the user
// GET /api/v1/dashboard/urls
func (h *DashboardHandlers) GetUserURLs(c *gin.Context) {
	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sortBy := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get user's URLs with pagination and filtering
	urls, total, err := h.getUserURLsWithPagination(userID.(int64), page, limit, sortBy, order, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_urls_failed",
			Message: "Failed to retrieve user URLs",
		})
		return
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"urls": urls,
		"pagination": gin.H{
			"current_page":  page,
			"total_pages":   totalPages,
			"total_items":   total,
			"items_per_page": limit,
			"has_next":      hasNext,
			"has_previous":  hasPrev,
		},
		"filters": gin.H{
			"search":  search,
			"sort_by": sortBy,
			"order":   order,
		},
	})
}

// GetUserURL retrieves a specific URL owned by the user
// GET /api/v1/dashboard/urls/:short_code
func (h *DashboardHandlers) GetUserURL(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL and verify ownership
	url, err := h.shortenerService.GetURLByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "url_not_found",
			Message: "URL not found",
		})
		return
	}

	// Check ownership or permissions
	if !h.userCanAccessURL(userID.(int64), url) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "access_denied",
			Message: "You don't have access to this URL",
		})
		return
	}

	// Get URL analytics if requested
	includeAnalytics := c.Query("analytics") == "true"
	var analytics *models.URLAnalytics
	
	if includeAnalytics {
		analytics, err = h.analyticsService.GetURLAnalytics(shortCode)
		if err != nil {
			// Don't fail if analytics retrieval fails, just log and continue
			analytics = nil
		}
	}

	response := gin.H{
		"url": url,
	}
	
	if analytics != nil {
		response["analytics"] = analytics
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUserURL updates a URL owned by the user
// PUT /api/v1/dashboard/urls/:short_code
func (h *DashboardHandlers) UpdateUserURL(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	var req models.UpdateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL and verify ownership
	url, err := h.shortenerService.GetURLByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "url_not_found",
			Message: "URL not found",
		})
		return
	}

	// Check ownership or edit permissions
	if !h.userCanEditURL(userID.(int64), url) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "edit_denied",
			Message: "You don't have permission to edit this URL",
		})
		return
	}

	// Update URL
	updatedURL, err := h.shortenerService.UpdateURL(shortCode, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: "Failed to update URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URL updated successfully",
		"url":     updatedURL,
	})
}

// DeleteUserURL deletes a URL owned by the user
// DELETE /api/v1/dashboard/urls/:short_code
func (h *DashboardHandlers) DeleteUserURL(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL and verify ownership
	url, err := h.shortenerService.GetURLByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "url_not_found",
			Message: "URL not found",
		})
		return
	}

	// Check ownership or delete permissions
	if !h.userCanDeleteURL(userID.(int64), url) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "delete_denied",
			Message: "You don't have permission to delete this URL",
		})
		return
	}

	// Delete URL
	err = h.shortenerService.DeleteURL(shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_failed",
			Message: "Failed to delete URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URL deleted successfully",
	})
}

// Bulk Operations

// BulkDeleteURLs deletes multiple URLs owned by the user
// POST /api/v1/dashboard/urls/bulk-delete
func (h *DashboardHandlers) BulkDeleteURLs(c *gin.Context) {
	var req struct {
		ShortCodes []string `json:"short_codes" validate:"required,min=1,max=50"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	successCount := 0
	failedCodes := []string{}

	for _, shortCode := range req.ShortCodes {
		// Get URL and verify ownership
		url, err := h.shortenerService.GetURLByShortCode(shortCode)
		if err != nil {
			failedCodes = append(failedCodes, shortCode)
			continue
		}

		// Check ownership or delete permissions
		if !h.userCanDeleteURL(userID.(int64), url) {
			failedCodes = append(failedCodes, shortCode)
			continue
		}

		// Delete URL
		err = h.shortenerService.DeleteURL(shortCode)
		if err != nil {
			failedCodes = append(failedCodes, shortCode)
			continue
		}

		successCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Bulk delete completed",
		"deleted":      successCount,
		"failed":       len(failedCodes),
		"failed_codes": failedCodes,
	})
}

// Analytics and Reports

// GetURLAnalytics retrieves detailed analytics for a specific URL
// GET /api/v1/dashboard/urls/:short_code/analytics
func (h *DashboardHandlers) GetURLAnalytics(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get URL and verify ownership/access
	url, err := h.shortenerService.GetURLByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "url_not_found",
			Message: "URL not found",
		})
		return
	}

	if !h.userCanAccessURL(userID.(int64), url) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "access_denied",
			Message: "You don't have access to this URL's analytics",
		})
		return
	}

	// Parse time range parameters
	timeRange := c.DefaultQuery("range", "7d") // 1d, 7d, 30d, 90d, 1y
	
	// Get detailed analytics
	analytics, err := h.analyticsService.GetDetailedURLAnalytics(shortCode, timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "analytics_failed",
			Message: "Failed to retrieve URL analytics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":       url,
		"analytics": analytics,
		"time_range": timeRange,
	})
}

// ExportUserData exports user's URL data
// GET /api/v1/dashboard/export
func (h *DashboardHandlers) ExportUserData(c *gin.Context) {
	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	format := c.DefaultQuery("format", "json") // json, csv
	includeAnalytics := c.Query("analytics") == "true"

	// Get all user URLs
	urls, _, err := h.getUserURLsWithPagination(userID.(int64), 1, 10000, "created_at", "desc", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "export_failed",
			Message: "Failed to export user data",
		})
		return
	}

	// Include analytics if requested
	if includeAnalytics {
		for _, url := range urls {
			analytics, err := h.analyticsService.GetURLAnalytics(url.ShortCode)
			if err == nil {
				// Add analytics data to URL (would need to extend URL model)
				_ = analytics
			}
		}
	}

	// Set appropriate headers for download
	filename := "urls-export-" + time.Now().Format("2006-01-02") + "." + format
	c.Header("Content-Disposition", "attachment; filename="+filename)

	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		// Convert to CSV format (simplified)
		c.String(http.StatusOK, "Short Code,Original URL,Title,Clicks,Created At,Updated At\n")
		for _, url := range urls {
			c.String(http.StatusOK, "%s,%s,%s,%d,%s,%s\n",
				url.ShortCode, url.OriginalURL, url.Title, url.ClickCount,
				url.CreatedAt.Format("2006-01-02 15:04:05"),
				url.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	} else {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, gin.H{
			"exported_at": time.Now(),
			"user_id":     userID,
			"total_urls":  len(urls),
			"urls":        urls,
		})
	}
}

// Helper Methods

// getUserURLStats gets URL statistics for a user
func (h *DashboardHandlers) getUserURLStats(userID int64) (map[string]interface{}, error) {
	// Implementation would query database for user statistics
	// For now, return simulated statistics
	return map[string]interface{}{
		"total_urls":          25,
		"total_clicks":        1847,
		"urls_created_today":  3,
		"clicks_today":        127,
		"urls_created_this_week": 8,
		"clicks_this_week":    456,
		"active_urls":         23,
		"expired_urls":        2,
		"custom_domains":      1,
	}, nil
}

// getRecentURLs gets recently created URLs for a user
func (h *DashboardHandlers) getRecentURLs(userID int64, limit int) ([]*models.URL, error) {
	// Implementation would query database
	// For now, return simulated recent URLs
	now := time.Now()
	return []*models.URL{
		{
			ID:          1,
			OriginalURL: "https://example.com/very-long-url-path",
			ShortCode:   "abc123",
			Title:       "Example Website",
			CreatedBy:   &userID,
			ClickCount:  45,
			CreatedAt:   now.Add(-2 * time.Hour),
			UpdatedAt:   now.Add(-1 * time.Hour),
		},
		{
			ID:          2,
			OriginalURL: "https://docs.example.com/documentation",
			ShortCode:   "doc456",
			Title:       "Documentation",
			CreatedBy:   &userID,
			ClickCount:  23,
			CreatedAt:   now.Add(-1 * 24 * time.Hour),
			UpdatedAt:   now.Add(-1 * 24 * time.Hour),
		},
	}, nil
}

// getAnalyticsSummary gets analytics summary for a user
func (h *DashboardHandlers) getAnalyticsSummary(userID int64) (map[string]interface{}, error) {
	// Implementation would aggregate analytics data
	// For now, return simulated analytics summary
	return map[string]interface{}{
		"top_performing_url": map[string]interface{}{
			"short_code": "abc123",
			"title":      "Example Website",
			"clicks":     45,
		},
		"click_trends": []map[string]interface{}{
			{"date": "2024-01-01", "clicks": 23},
			{"date": "2024-01-02", "clicks": 34},
			{"date": "2024-01-03", "clicks": 27},
		},
		"top_referrers": []map[string]interface{}{
			{"domain": "twitter.com", "clicks": 123},
			{"domain": "facebook.com", "clicks": 89},
			{"domain": "direct", "clicks": 67},
		},
		"device_breakdown": map[string]int{
			"desktop": 60,
			"mobile":  35,
			"tablet":  5,
		},
	}, nil
}

// getAccountInfo gets account information for dashboard
func (h *DashboardHandlers) getAccountInfo(c *gin.Context) map[string]interface{} {
	user, exists := c.Get("user")
	if !exists {
		return map[string]interface{}{}
	}

	u := user.(*models.User)
	return map[string]interface{}{
		"account_type":    u.AccountType,
		"email_verified":  u.EmailVerified,
		"phone_verified":  u.PhoneVerified,
		"member_since":    u.CreatedAt,
		"last_login":      u.LastLoginAt,
	}
}

// getUserURLsWithPagination gets user URLs with pagination and filtering
func (h *DashboardHandlers) getUserURLsWithPagination(userID int64, page, limit int, sortBy, order, search string) ([]*models.URL, int, error) {
	// Implementation would query database with proper pagination
	// For now, return simulated paginated results
	urls := []*models.URL{
		{
			ID:          1,
			OriginalURL: "https://example.com/search-result-1",
			ShortCode:   "search1",
			Title:       "Search Result 1",
			CreatedBy:   &userID,
			ClickCount:  15,
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          2,
			OriginalURL: "https://example.com/search-result-2",
			ShortCode:   "search2",
			Title:       "Search Result 2",
			CreatedBy:   &userID,
			ClickCount:  8,
			CreatedAt:   time.Now().Add(-3 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
	}

	// Filter by search if provided
	if search != "" {
		// Implementation would filter results
		_ = search
	}

	// Apply sorting
	// Implementation would sort results based on sortBy and order
	_ = sortBy
	_ = order

	// Calculate offset for pagination
	offset := (page - 1) * limit
	if offset >= len(urls) {
		return []*models.URL{}, len(urls), nil
	}

	end := offset + limit
	if end > len(urls) {
		end = len(urls)
	}

	return urls[offset:end], len(urls), nil
}

// Permission check helpers

func (h *DashboardHandlers) userCanAccessURL(userID int64, url *models.URL) bool {
	// User can access their own URLs
	if url.CreatedBy != nil && *url.CreatedBy == userID {
		return true
	}

	// Check if user has system permissions
	hasPermission, _ := h.rbacService.UserHasPermission(userID, models.PermissionURLsRead)
	return hasPermission
}

func (h *DashboardHandlers) userCanEditURL(userID int64, url *models.URL) bool {
	// User can edit their own URLs
	if url.CreatedBy != nil && *url.CreatedBy == userID {
		return true
	}

	// Check if user has system permissions
	hasPermission, _ := h.rbacService.UserHasPermission(userID, models.PermissionURLsUpdate)
	return hasPermission
}

func (h *DashboardHandlers) userCanDeleteURL(userID int64, url *models.URL) bool {
	// User can delete their own URLs
	if url.CreatedBy != nil && *url.CreatedBy == userID {
		return true
	}

	// Check if user has system permissions
	hasPermission, _ := h.rbacService.UserHasPermission(userID, models.PermissionURLsDelete)
	return hasPermission
}