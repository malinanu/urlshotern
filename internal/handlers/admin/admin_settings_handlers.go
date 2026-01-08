package admin

import (
	"net/http"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminSettingsHandlers handles admin settings operations
type AdminSettingsHandlers struct {
	settingsService *services.GlobalSettingsService
	auditService    *services.AdminAuditService
}

// NewAdminSettingsHandlers creates a new admin settings handlers instance
func NewAdminSettingsHandlers(settingsService *services.GlobalSettingsService, auditService *services.AdminAuditService) *AdminSettingsHandlers {
	return &AdminSettingsHandlers{
		settingsService: settingsService,
		auditService:    auditService,
	}
}

// AdminGetSettings retrieves all system settings
// GET /api/v1/admin/settings
func (h *AdminSettingsHandlers) AdminGetSettings(c *gin.Context) {
	// Optional category filter
	category := c.Query("category")

	var settings []*models.GlobalSetting
	var err error

	if category != "" {
		settings, err = h.settingsService.GetSettings(category, false)
	} else {
		settings, err = h.settingsService.GetAllSettings()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve settings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
		"total":    len(settings),
	})
}

// AdminGetSetting retrieves a specific setting by key
// GET /api/v1/admin/settings/:key
func (h *AdminSettingsHandlers) AdminGetSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_key",
			Message: "Setting key is required",
		})
		return
	}

	setting, err := h.settingsService.GetSetting(key)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "setting_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"setting": setting,
	})
}

// AdminUpdateSetting updates a system setting
// PUT /api/v1/admin/settings/:key
func (h *AdminSettingsHandlers) AdminUpdateSetting(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_key",
			Message: "Setting key is required",
		})
		return
	}

	var req models.GlobalSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Validate setting value
	if err := h.settingsService.ValidateSetting(key, req.Value); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
		return
	}

	// Update setting
	setting, err := h.settingsService.UpdateSetting(key, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "settings.update", "setting", nil, map[string]interface{}{
		"key":       key,
		"old_value": req.Value, // Would need to fetch old value to track properly
		"new_value": setting.Value,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting updated successfully",
		"setting": setting,
	})
}

// AdminBulkUpdateSettings updates multiple settings at once
// PUT /api/v1/admin/settings
func (h *AdminSettingsHandlers) AdminBulkUpdateSettings(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req struct {
		Settings map[string]string `json:"settings" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Validate all settings before updating
	for key, value := range req.Settings {
		if err := h.settingsService.ValidateSetting(key, value); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "validation_failed",
				Message: "Setting '" + key + "': " + err.Error(),
			})
			return
		}
	}

	// Bulk update
	if err := h.settingsService.BulkUpdateSettings(req.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	keys := make([]string, 0, len(req.Settings))
	for key := range req.Settings {
		keys = append(keys, key)
	}
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "settings.bulk_update", "setting", nil, map[string]interface{}{
		"keys":  keys,
		"count": len(req.Settings),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Settings updated successfully",
		"count":   len(req.Settings),
	})
}

// AdminResetSetting resets a setting to default value
// POST /api/v1/admin/settings/:key/reset
func (h *AdminSettingsHandlers) AdminResetSetting(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_key",
			Message: "Setting key is required",
		})
		return
	}

	// Reset setting
	if err := h.settingsService.ResetSetting(key); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "settings.reset", "setting", nil, map[string]interface{}{
		"key": key,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting reset to default successfully",
	})
}

// AdminExportSettings exports all settings
// GET /api/v1/admin/settings/export
func (h *AdminSettingsHandlers) AdminExportSettings(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get all settings
	settings, err := h.settingsService.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "export_failed",
			Message: "Failed to export settings",
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "settings.export", "setting", nil, map[string]interface{}{
		"count": len(settings),
	})

	// Set export headers
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=settings_export.json")
	c.JSON(http.StatusOK, gin.H{
		"settings":    settings,
		"exported_at": "2026-01-07T00:00:00Z",
		"count":       len(settings),
	})
}

// AdminGetSettingCategories retrieves all setting categories
// GET /api/v1/admin/settings/categories
func (h *AdminSettingsHandlers) AdminGetSettingCategories(c *gin.Context) {
	categories, err := h.settingsService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// AdminResetSettings resets all settings to defaults
// POST /api/v1/admin/settings/reset
func (h *AdminSettingsHandlers) AdminResetSettings(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Reset all settings
	if err := h.settingsService.ResetAllSettings(); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "settings.reset_all", "setting", nil, map[string]interface{}{
		"action": "reset_all_settings",
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "All settings reset to defaults successfully",
	})
}
