package handlers

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AttributionHandler handles attribution-related endpoints
type AttributionHandler struct {
	attributionService *services.AttributionService
}

// NewAttributionHandler creates a new attribution handler
func NewAttributionHandler(attributionService *services.AttributionService) *AttributionHandler {
	return &AttributionHandler{
		attributionService: attributionService,
	}
}

// RecordTouchpoint records a touchpoint in the customer journey
func (h *AttributionHandler) RecordTouchpoint(c *gin.Context) {
	var touchpoint models.AttributionTouchpoint
	
	if err := c.ShouldBindJSON(&touchpoint); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Set touchpoint time if not provided
	if touchpoint.TouchpointTime.IsZero() {
		touchpoint.TouchpointTime = touchpoint.CreatedAt
	}

	err := h.attributionService.RecordTouchpoint(&touchpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "touchpoint_error",
			Message: "Failed to record touchpoint",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Touchpoint recorded successfully",
		"touchpoint_id": touchpoint.ID,
	})
}

// GetConversionJourney retrieves the complete customer journey for a conversion
func (h *AttributionHandler) GetConversionJourney(c *gin.Context) {
	conversionID := c.Param("conversionId")
	if conversionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Conversion ID is required",
		})
		return
	}

	journey, err := h.attributionService.GetConversionJourney(conversionID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "journey_not_found",
			Message: "Conversion journey not found",
		})
		return
	}

	c.JSON(http.StatusOK, journey)
}

// CalculateAttribution calculates attribution for a conversion using specified model
func (h *AttributionHandler) CalculateAttribution(c *gin.Context) {
	conversionID := c.Param("conversionId")
	if conversionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Conversion ID is required",
		})
		return
	}

	modelStr := c.Query("model")
	if modelStr == "" {
		modelStr = "linear" // Default to linear attribution
	}

	model := services.AttributionModel(modelStr)

	// Validate attribution model
	validModels := map[string]bool{
		"first_touch":     true,
		"last_touch":      true,
		"linear":          true,
		"time_decay":      true,
		"position_based":  true,
		"data_driven":     true,
	}

	if !validModels[modelStr] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_model",
			Message: "Invalid attribution model. Valid models: first_touch, last_touch, linear, time_decay, position_based, data_driven",
		})
		return
	}

	touchpointValues, err := h.attributionService.CalculateAttribution(conversionID, model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "attribution_error",
			Message: "Failed to calculate attribution",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversion_id":      conversionID,
		"attribution_model":  modelStr,
		"touchpoint_values":  touchpointValues,
		"total_touchpoints":  len(touchpointValues),
	})
}

// GetAttributionReport generates a comprehensive attribution report for a conversion
func (h *AttributionHandler) GetAttributionReport(c *gin.Context) {
	conversionID := c.Param("conversionId")
	if conversionID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Conversion ID is required",
		})
		return
	}

	report, err := h.attributionService.GetAttributionReport(conversionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "report_error",
			Message: "Failed to generate attribution report",
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetChannelAttribution provides attribution analysis by marketing channel
func (h *AttributionHandler) GetChannelAttribution(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Short code is required",
		})
		return
	}

	// Get query parameters
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	modelStr := c.DefaultQuery("model", "linear")
	model := services.AttributionModel(modelStr)

	// Validate attribution model
	validModels := map[string]bool{
		"first_touch":     true,
		"last_touch":      true,
		"linear":          true,
		"time_decay":      true,
		"position_based":  true,
		"data_driven":     true,
	}

	if !validModels[modelStr] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_model",
			Message: "Invalid attribution model",
		})
		return
	}

	channelAttribution, err := h.attributionService.GetChannelAttribution(shortCode, days, model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "channel_attribution_error",
			Message: "Failed to get channel attribution",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_code":         shortCode,
		"attribution_model":  modelStr,
		"days":              days,
		"channel_attribution": channelAttribution,
	})
}

// CompareAttributionModels compares different attribution models for a short code
func (h *AttributionHandler) CompareAttributionModels(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Short code is required",
		})
		return
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	models := []services.AttributionModel{
		services.FirstTouchAttribution,
		services.LastTouchAttribution,
		services.LinearAttribution,
		services.TimeDecayAttribution,
		services.PositionBasedAttribution,
		services.DataDrivenAttribution,
	}

	comparison := make(map[string]interface{})
	
	for _, model := range models {
		channelData, err := h.attributionService.GetChannelAttribution(shortCode, days, model)
		if err != nil {
			continue // Skip failed models
		}
		
		// Calculate total attributed value for this model
		totalValue := 0.0
		for _, channel := range channelData {
			if value, exists := channel.AttributionValue[string(model)]; exists {
				totalValue += value
			}
		}
		
		comparison[string(model)] = gin.H{
			"total_attributed_value": totalValue,
			"top_channels":          channelData[:min(len(channelData), 5)],
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"short_code":        shortCode,
		"days":             days,
		"model_comparison":  comparison,
		"recommendation":   "Use position_based for multi-touch journeys, linear for equal weighting",
	})
}

// GetAttributionInsights provides insights about attribution patterns
func (h *AttributionHandler) GetAttributionInsights(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Short code is required",
		})
		return
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	// This would implement comprehensive attribution insights
	// For now, return basic insights structure
	insights := gin.H{
		"short_code": shortCode,
		"days":      days,
		"insights": gin.H{
			"avg_touchpoints_per_conversion": 0.0,
			"avg_time_to_conversion_hours":   0.0,
			"most_effective_model":          "linear",
			"top_converting_channels":       []string{},
			"attribution_distribution":     gin.H{},
		},
	}

	c.JSON(http.StatusOK, insights)
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}