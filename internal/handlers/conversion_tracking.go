package handlers

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

type ConversionTrackingHandler struct {
	conversionService *services.ConversionTrackingService
}

func NewConversionTrackingHandler(conversionService *services.ConversionTrackingService) *ConversionTrackingHandler {
	return &ConversionTrackingHandler{
		conversionService: conversionService,
	}
}

// CreateConversionGoal creates a new conversion goal
func (h *ConversionTrackingHandler) CreateConversionGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request models.CreateConversionGoalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate required fields
	if request.GoalName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Goal name is required"})
		return
	}

	if request.GoalType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Goal type is required"})
		return
	}

	// Validate goal type
	validTypes := map[string]bool{
		"url_visit":    true,
		"custom_event": true,
		"form_submit":  true,
		"purchase":     true,
	}

	if !validTypes[request.GoalType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal type"})
		return
	}

	goal, err := h.conversionService.CreateConversionGoal(userID.(int64), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversion goal"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Conversion goal created successfully",
		"goal":    goal,
	})
}

// GetConversionGoals retrieves all conversion goals for a user
func (h *ConversionTrackingHandler) GetConversionGoals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goals, err := h.conversionService.GetUserConversionGoals(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve conversion goals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"goals": goals,
	})
}

// GetConversionGoal retrieves a specific conversion goal
func (h *ConversionTrackingHandler) GetConversionGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goalIDStr := c.Param("goalId")
	goalID, err := strconv.ParseInt(goalIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	goal, err := h.conversionService.GetConversionGoal(goalID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversion goal not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"goal": goal,
	})
}

// UpdateConversionGoal updates a conversion goal
func (h *ConversionTrackingHandler) UpdateConversionGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goalIDStr := c.Param("goalId")
	goalID, err := strconv.ParseInt(goalIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	var request models.UpdateConversionGoalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	goal, err := h.conversionService.UpdateConversionGoal(goalID, userID.(int64), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update conversion goal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conversion goal updated successfully",
		"goal":    goal,
	})
}

// DeleteConversionGoal deletes a conversion goal
func (h *ConversionTrackingHandler) DeleteConversionGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goalIDStr := c.Param("goalId")
	goalID, err := strconv.ParseInt(goalIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	err = h.conversionService.DeleteConversionGoal(goalID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversion goal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conversion goal deleted successfully",
	})
}

// TrackConversion tracks a conversion event
func (h *ConversionTrackingHandler) TrackConversion(c *gin.Context) {
	var request models.ConversionTrackingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate required fields
	if request.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	if request.GoalID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Goal ID is required"})
		return
	}

	conversion, err := h.conversionService.TrackConversionFromRequest(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track conversion"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Conversion tracked successfully",
		"conversion": conversion,
	})
}

// GetConversionStats retrieves conversion statistics for a goal
func (h *ConversionTrackingHandler) GetConversionStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goalIDStr := c.Param("goalId")
	goalID, err := strconv.ParseInt(goalIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	// Get date range from query parameters
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter"})
		return
	}

	stats, err := h.conversionService.GetConversionStatsByGoal(goalID, userID.(int64), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve conversion stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetAttributionReport retrieves attribution report for conversions
func (h *ConversionTrackingHandler) GetAttributionReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	goalIDStr := c.Param("goalId")
	goalID, err := strconv.ParseInt(goalIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	// Get attribution model from query parameter
	model := c.DefaultQuery("model", "last_click")
	validModels := map[string]bool{
		"first_click": true,
		"last_click":  true,
		"linear":      true,
	}

	if !validModels[model] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attribution model"})
		return
	}

	// Get date range from query parameters
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter"})
		return
	}

	report, err := h.conversionService.GetAttributionReport(goalID, userID.(int64), model, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attribution report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}