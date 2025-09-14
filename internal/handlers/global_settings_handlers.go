package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type GlobalSettingsHandlers struct {
	settingsService *services.GlobalSettingsService
}

func NewGlobalSettingsHandlers(settingsService *services.GlobalSettingsService) *GlobalSettingsHandlers {
	return &GlobalSettingsHandlers{
		settingsService: settingsService,
	}
}

// GetPublicSettings returns all public settings (no auth required)
func (h *GlobalSettingsHandlers) GetPublicSettings(c *gin.Context) {
	settings, err := h.settingsService.GetPublicSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get public settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// GetSiteInfo returns basic site information (no auth required)
func (h *GlobalSettingsHandlers) GetSiteInfo(c *gin.Context) {
	info := h.settingsService.GetSiteInfo()
	socialLinks := h.settingsService.GetSocialLinks()
	
	c.JSON(http.StatusOK, gin.H{
		"site_info":     info,
		"social_links":  socialLinks,
	})
}

// GetSettings returns all settings (admin only)
func (h *GlobalSettingsHandlers) GetSettings(c *gin.Context) {
	category := c.Query("category")
	publicOnly := c.Query("public_only") == "true"

	if category == "" {
		// Return grouped by category
		settings, err := h.settingsService.GetSettingsByCategory()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"settings": settings})
		return
	}

	// Return specific category
	settings, err := h.settingsService.GetSettings(category, publicOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// GetSetting returns a specific setting (admin only)
func (h *GlobalSettingsHandlers) GetSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	setting, err := h.settingsService.GetSetting(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"setting": setting})
}

// CreateSetting creates a new global setting (admin only)
func (h *GlobalSettingsHandlers) CreateSetting(c *gin.Context) {
	var req models.GlobalSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.Key == "" || req.DisplayName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key and display name are required"})
		return
	}

	// Set defaults
	if req.Type == "" {
		req.Type = "text"
	}
	if req.Category == "" {
		req.Category = "general"
	}

	setting, err := h.settingsService.CreateSetting(&req)
	if err != nil {
		if err.Error() == "setting with key '"+req.Key+"' already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"setting": setting})
}

// UpdateSetting updates an existing global setting (admin only)
func (h *GlobalSettingsHandlers) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	var req models.GlobalSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	setting, err := h.settingsService.UpdateSetting(key, &req)
	if err != nil {
		if err.Error() == "setting with key '"+key+"' not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"setting": setting})
}

// DeleteSetting deletes a global setting (admin only)
func (h *GlobalSettingsHandlers) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	// Prevent deletion of critical settings
	protectedKeys := []string{"site_name", "maintenance_mode"}
	for _, protectedKey := range protectedKeys {
		if key == protectedKey {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This setting cannot be deleted"})
			return
		}
	}

	err := h.settingsService.DeleteSetting(key)
	if err != nil {
		if err.Error() == "setting with key '"+key+"' not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete setting", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting deleted successfully"})
}

// BulkUpdateSettings updates multiple settings at once (admin only)
func (h *GlobalSettingsHandlers) BulkUpdateSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if len(req) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No settings provided"})
		return
	}

	if len(req) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 50 settings can be updated at once"})
		return
	}

	err := h.settingsService.BulkUpdateSettings(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Settings updated successfully",
		"updated_count": len(req),
	})
}

// CheckMaintenanceMode checks if site is in maintenance mode
func (h *GlobalSettingsHandlers) CheckMaintenanceMode(c *gin.Context) {
	isMaintenanceMode := h.settingsService.IsMaintenanceMode()
	
	c.JSON(http.StatusOK, gin.H{
		"maintenance_mode": isMaintenanceMode,
	})
}