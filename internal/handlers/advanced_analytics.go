package handlers

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdvancedAnalyticsHandler handles advanced analytics endpoints
type AdvancedAnalyticsHandler struct {
	analyticsService *services.AdvancedAnalyticsService
}

// NewAdvancedAnalyticsHandler creates a new advanced analytics handler
func NewAdvancedAnalyticsHandler(analyticsService *services.AdvancedAnalyticsService) *AdvancedAnalyticsHandler {
	return &AdvancedAnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetAdvancedAnalytics retrieves comprehensive analytics for a short code
// @Summary Get advanced analytics for a URL
// @Description Retrieves detailed analytics including geographic, time-based, device, and referrer data
// @Tags Analytics
// @Accept json
// @Produce json
// @Param shortCode path string true "Short code"
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {object} models.AdvancedAnalyticsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/analytics/advanced/{shortCode} [get]
func (h *AdvancedAnalyticsHandler) GetAdvancedAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get days parameter with default
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter (must be 1-365)"})
		return
	}

	analytics, err := h.analyticsService.GetAdvancedAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analytics not found"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetGeographicAnalytics retrieves geographic analytics for a short code
// @Summary Get geographic analytics for a URL
// @Description Retrieves detailed geographic analytics including countries, regions, cities, and map data
// @Tags Analytics
// @Accept json
// @Produce json
// @Param shortCode path string true "Short code"
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {object} models.GeographicAnalytics
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/analytics/geographic/{shortCode} [get]
func (h *AdvancedAnalyticsHandler) GetGeographicAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get days parameter with default
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter (must be 1-365)"})
		return
	}

	analytics, err := h.analyticsService.GetGeographicAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Geographic analytics not found"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetTimeAnalytics retrieves time-based analytics for a short code
// @Summary Get time-based analytics for a URL
// @Description Retrieves detailed time-based analytics including hourly patterns, heatmaps, and peak times
// @Tags Analytics
// @Accept json
// @Produce json
// @Param shortCode path string true "Short code"
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {object} models.TimeAnalytics
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/analytics/time/{shortCode} [get]
func (h *AdvancedAnalyticsHandler) GetTimeAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get days parameter with default
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter (must be 1-365)"})
		return
	}

	analytics, err := h.analyticsService.GetTimeAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Time analytics not found"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetDeviceAnalytics retrieves device analytics for a short code
// @Summary Get device analytics for a URL
// @Description Retrieves detailed device analytics including device types, browsers, and operating systems
// @Tags Analytics
// @Accept json
// @Produce json
// @Param shortCode path string true "Short code"
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {object} models.DeviceAnalytics
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/analytics/device/{shortCode} [get]
func (h *AdvancedAnalyticsHandler) GetDeviceAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get days parameter with default
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter (must be 1-365)"})
		return
	}

	analytics, err := h.analyticsService.GetDeviceAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device analytics not found"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetReferrerAnalytics retrieves referrer analytics for a short code
// @Summary Get referrer analytics for a URL
// @Description Retrieves detailed referrer analytics including top referrers and campaign data
// @Tags Analytics
// @Accept json
// @Produce json
// @Param shortCode path string true "Short code"
// @Param days query int false "Number of days to analyze" default(30)
// @Param limit query int false "Maximum number of results" default(20)
// @Success 200 {array} models.ReferrerStat
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/analytics/referrers/{shortCode} [get]
func (h *AdvancedAnalyticsHandler) GetReferrerAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get days parameter with default
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter (must be 1-365)"})
		return
	}

	// Get limit parameter with default
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter (must be 1-100)"})
		return
	}

	referrers, err := h.analyticsService.GetReferrerAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Referrer analytics not found"})
		return
	}

	// Apply limit
	if len(referrers) > limit {
		referrers = referrers[:limit]
	}

	c.JSON(http.StatusOK, referrers)
}

// GetClickHeatmap retrieves click heatmap data for visualization
// @Summary Get click heatmap data
// @Description Retrieves heatmap data showing clicks by hour and day
// @Tags Analytics
// @Accept json
// @Produce json
// @Param shortCode path string true "Short code"
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {array} models.HeatmapPoint
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/analytics/heatmap/{shortCode} [get]
func (h *AdvancedAnalyticsHandler) GetClickHeatmap(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get days parameter with default
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 90 { // Limit to 90 days for heatmap
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter (must be 1-90)"})
		return
	}

	timeAnalytics, err := h.analyticsService.GetTimeAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Heatmap data not found"})
		return
	}

	c.JSON(http.StatusOK, timeAnalytics.HeatmapData)
}

// GetMapData retrieves geographic map data for visualization
// @Summary Get geographic map data
// @Description Retrieves map points with coordinates and click counts for geographic visualization
// @Tags Analytics
// @Accept json
// @Produce json
// @Param shortCode path string true "Short code"
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {array} models.MapPoint
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/analytics/map/{shortCode} [get]
func (h *AdvancedAnalyticsHandler) GetMapData(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get days parameter with default
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter (must be 1-365)"})
		return
	}

	geoAnalytics, err := h.analyticsService.GetGeographicAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Map data not found"})
		return
	}

	c.JSON(http.StatusOK, geoAnalytics.MapData)
}

// GetHeatmapAnalytics handles heatmap analytics requests
func (h *AdvancedAnalyticsHandler) GetHeatmapAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get days parameter with default
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 90 { // Limit to 90 days for heatmap
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter (must be 1-90)"})
		return
	}

	timeAnalytics, err := h.analyticsService.GetTimeAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Heatmap data not found"})
		return
	}

	c.JSON(http.StatusOK, timeAnalytics.HeatmapData)
}

// GetMapAnalytics handles map analytics requests
func (h *AdvancedAnalyticsHandler) GetMapAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	// Get days parameter with default
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter (must be 1-365)"})
		return
	}

	geoAnalytics, err := h.analyticsService.GetGeographicAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Map data not found"})
		return
	}

	c.JSON(http.StatusOK, geoAnalytics.MapData)
}

// GetEnhancedReferrerAnalytics provides detailed referrer analytics with UTM tracking
func (h *AdvancedAnalyticsHandler) GetEnhancedReferrerAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	daysParam := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysParam)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	analytics, err := h.analyticsService.GetEnhancedReferrerAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve referrer analytics"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetUTMCampaignAnalytics provides UTM campaign performance analytics
func (h *AdvancedAnalyticsHandler) GetUTMCampaignAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	daysParam := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysParam)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	analytics, err := h.analyticsService.GetUTMCampaignAnalytics(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve UTM campaign analytics"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetReferrerInsights provides intelligent insights about referrer patterns
func (h *AdvancedAnalyticsHandler) GetReferrerInsights(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	daysParam := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysParam)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	insights, err := h.analyticsService.GetReferrerInsights(shortCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve referrer insights"})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// RegisterAdvancedAnalyticsRoutes registers all advanced analytics routes
func RegisterAdvancedAnalyticsRoutes(router *gin.RouterGroup, handler *AdvancedAnalyticsHandler) {
	analytics := router.Group("/analytics")
	{
		// Advanced analytics endpoints
		analytics.GET("/advanced/:shortCode", handler.GetAdvancedAnalytics)
		analytics.GET("/geographic/:shortCode", handler.GetGeographicAnalytics)
		analytics.GET("/time/:shortCode", handler.GetTimeAnalytics)
		analytics.GET("/device/:shortCode", handler.GetDeviceAnalytics)
		analytics.GET("/referrers/:shortCode", handler.GetReferrerAnalytics)
		analytics.GET("/heatmap/:shortCode", handler.GetClickHeatmap)
		analytics.GET("/map/:shortCode", handler.GetMapData)
	}
}

// RegisterProtectedAdvancedAnalyticsRoutes registers protected advanced analytics routes
func RegisterProtectedAdvancedAnalyticsRoutes(router *gin.RouterGroup, handler *AdvancedAnalyticsHandler, authService *services.AuthService) {
	protected := router.Group("/analytics")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		// Protected endpoints for detailed analytics (require authentication)
		protected.GET("/advanced/:shortCode/detailed", handler.GetAdvancedAnalytics)
		protected.GET("/geographic/:shortCode/detailed", handler.GetGeographicAnalytics)
		protected.GET("/time/:shortCode/detailed", handler.GetTimeAnalytics)
		protected.GET("/device/:shortCode/detailed", handler.GetDeviceAnalytics)
		protected.GET("/referrers/:shortCode/detailed", handler.GetReferrerAnalytics)
	}
}