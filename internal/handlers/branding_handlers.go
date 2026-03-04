package handlers

import (
	"net/http"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// BrandingHandler handles brand settings API endpoints (enterprise only).
type BrandingHandler struct {
	svc *services.BrandingService
}

// NewBrandingHandler creates a new BrandingHandler.
func NewBrandingHandler(svc *services.BrandingService) *BrandingHandler {
	return &BrandingHandler{svc: svc}
}

// GetBranding returns the current user's brand settings.
// GET /api/v1/branding
func (h *BrandingHandler) GetBranding(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	b, err := h.svc.GetSettings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, b)
}

// UpdateBranding saves brand settings for the current user.
// PUT /api/v1/branding
func (h *BrandingHandler) UpdateBranding(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req models.UpdateBrandSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	b, err := h.svc.UpdateSettings(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, b)
}
