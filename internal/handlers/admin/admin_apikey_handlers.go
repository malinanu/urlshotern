package admin

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminAPIKeyHandlers handles admin API key operations
type AdminAPIKeyHandlers struct {
	apiKeyService *services.APIKeyService
	auditService  *services.AdminAuditService
}

// NewAdminAPIKeyHandlers creates a new admin API key handlers instance
func NewAdminAPIKeyHandlers(apiKeyService *services.APIKeyService, auditService *services.AdminAuditService) *AdminAPIKeyHandlers {
	return &AdminAPIKeyHandlers{
		apiKeyService: apiKeyService,
		auditService:  auditService,
	}
}

// AdminListAPIKeys lists all API keys across all users
// GET /api/v1/admin/apikeys
func (h *AdminAPIKeyHandlers) AdminListAPIKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Optional user filter
	var userID *int64
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if uid, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = &uid
		}
	}

	apiKeys, total, err := h.apiKeyService.AdminListAPIKeys(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list API keys",
		})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{
		"api_keys":    apiKeys,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// AdminGetAPIKey retrieves a specific API key
// GET /api/v1/admin/apikeys/:id
func (h *AdminAPIKeyHandlers) AdminGetAPIKey(c *gin.Context) {
	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_key_id",
			Message: "Invalid API key ID format",
		})
		return
	}

	apiKey, err := h.apiKeyService.AdminGetAPIKey(keyID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "api_key_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": apiKey,
	})
}

// AdminRevokeAPIKey revokes an API key
// POST /api/v1/admin/apikeys/:id/revoke
func (h *AdminAPIKeyHandlers) AdminRevokeAPIKey(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_key_id",
			Message: "Invalid API key ID format",
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = ""
	}

	if err := h.apiKeyService.AdminRevokeAPIKey(keyID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "revoke_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "apikey.revoke", "api_key", nil, map[string]interface{}{
		"key_id": keyID.String(),
		"reason": req.Reason,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "API key revoked successfully",
	})
}

// AdminDeleteAPIKey permanently deletes an API key
// DELETE /api/v1/admin/apikeys/:id
func (h *AdminAPIKeyHandlers) AdminDeleteAPIKey(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_key_id",
			Message: "Invalid API key ID format",
		})
		return
	}

	if err := h.apiKeyService.AdminDeleteAPIKey(keyID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "apikey.delete", "api_key", nil, map[string]interface{}{
		"key_id": keyID.String(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "API key deleted successfully",
	})
}

// AdminGetAPIKeyUsage retrieves usage statistics for an API key
// GET /api/v1/admin/apikeys/:id/usage
func (h *AdminAPIKeyHandlers) AdminGetAPIKeyUsage(c *gin.Context) {
	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_key_id",
			Message: "Invalid API key ID format",
		})
		return
	}

	// Parse time range
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	usage, err := h.apiKeyService.AdminGetAPIKeyUsage(keyID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve API key usage",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"usage": usage,
	})
}

// AdminResetAPIKeyQuota resets the rate limit quota for an API key
// POST /api/v1/admin/apikeys/:id/reset-quota
func (h *AdminAPIKeyHandlers) AdminResetAPIKeyQuota(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_key_id",
			Message: "Invalid API key ID format",
		})
		return
	}

	if err := h.apiKeyService.AdminResetAPIKeyQuota(keyID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "apikey.reset_quota", "api_key", nil, map[string]interface{}{
		"key_id": keyID.String(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "API key quota reset successfully",
	})
}

// AdminCreateSystemAPIKey creates a system-level API key
// POST /api/v1/admin/api-keys/system
func (h *AdminAPIKeyHandlers) AdminCreateSystemAPIKey(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
		RateLimit   int      `json:"rate_limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	response, err := h.apiKeyService.AdminCreateSystemAPIKey(req.Name, req.Description, req.Permissions, req.RateLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "apikey.create_system", "api_key", nil, map[string]interface{}{
		"key_id":      response.APIKey.ID.String(),
		"name":        req.Name,
		"permissions": req.Permissions,
		"rate_limit":  req.RateLimit,
	})

	c.JSON(http.StatusCreated, response)
}
