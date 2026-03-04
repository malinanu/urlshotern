package handlers

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

type AdvancedAnalyticsHandler struct {
	service *services.AdvancedAnalyticsService
}

func NewAdvancedAnalyticsHandler(svc *services.AdvancedAnalyticsService) *AdvancedAnalyticsHandler {
	return &AdvancedAnalyticsHandler{service: svc}
}

func parseDays(c *gin.Context) (int, bool) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_days_parameter",
			Message: "Days must be between 1 and 365",
		})
		return 0, false
	}
	return days, true
}

func (h *AdvancedAnalyticsHandler) GetAdvancedAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	days, ok := parseDays(c)
	if !ok {
		return
	}
	data, err := h.service.GetAdvancedAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AdvancedAnalyticsHandler) GetGeographicAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	days, ok := parseDays(c)
	if !ok {
		return
	}
	data, err := h.service.GetGeographicAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AdvancedAnalyticsHandler) GetTimeAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	days, ok := parseDays(c)
	if !ok {
		return
	}
	data, err := h.service.GetTimeAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AdvancedAnalyticsHandler) GetDeviceAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	days, ok := parseDays(c)
	if !ok {
		return
	}
	data, err := h.service.GetDeviceAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AdvancedAnalyticsHandler) GetReferrerAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	days, ok := parseDays(c)
	if !ok {
		return
	}
	data, err := h.service.GetReferrerAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AdvancedAnalyticsHandler) GetEnhancedReferrerAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	days, ok := parseDays(c)
	if !ok {
		return
	}
	data, err := h.service.GetEnhancedReferrerAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AdvancedAnalyticsHandler) GetUTMCampaignAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	days, ok := parseDays(c)
	if !ok {
		return
	}
	data, err := h.service.GetUTMCampaignAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AdvancedAnalyticsHandler) GetReferrerInsights(c *gin.Context) {
	shortCode := c.Param("shortCode")
	days, ok := parseDays(c)
	if !ok {
		return
	}
	data, err := h.service.GetReferrerInsights(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetHeatmapAnalytics delegates to GetTimeAnalytics (time data includes the hour×day heatmap grid).
func (h *AdvancedAnalyticsHandler) GetHeatmapAnalytics(c *gin.Context) {
	h.GetTimeAnalytics(c)
}

// GetMapAnalytics delegates to GetGeographicAnalytics (geographic data includes lat/lon for map markers).
func (h *AdvancedAnalyticsHandler) GetMapAnalytics(c *gin.Context) {
	h.GetGeographicAnalytics(c)
}
