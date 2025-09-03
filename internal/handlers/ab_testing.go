package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

type ABTestingHandler struct {
	abTestService *services.ABTestingService
}

func NewABTestingHandler(abTestService *services.ABTestingService) *ABTestingHandler {
	return &ABTestingHandler{
		abTestService: abTestService,
	}
}

// CreateABTest creates a new A/B test
func (h *ABTestingHandler) CreateABTest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request models.CreateABTestRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate required fields
	if request.TestName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test name is required"})
		return
	}

	if request.TestType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test type is required"})
		return
	}

	if len(request.Variants) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least 2 variants are required"})
		return
	}

	// Validate variants
	totalTraffic := 0
	hasControl := false
	for _, variant := range request.Variants {
		if variant.VariantName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Variant name is required"})
			return
		}
		if variant.ShortCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required for variant"})
			return
		}
		totalTraffic += variant.TrafficAllocation
		if variant.IsControl {
			hasControl = true
		}
	}

	if totalTraffic != 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Traffic allocation must sum to 100%"})
		return
	}

	if !hasControl {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one variant must be marked as control"})
		return
	}

	test, err := h.abTestService.CreateABTest(userID.(int64), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create A/B test"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "A/B test created successfully",
		"test":    test,
	})
}

// GetABTests retrieves all A/B tests for a user
func (h *ABTestingHandler) GetABTests(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// This would need to be implemented in the service
	c.JSON(http.StatusOK, gin.H{
		"tests": []models.ABTest{}, // Placeholder
	})
}

// GetABTest retrieves a specific A/B test
func (h *ABTestingHandler) GetABTest(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	testIDStr := c.Param("testId")
	_, err := strconv.ParseInt(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	// This would need to be implemented in the service
	c.JSON(http.StatusNotFound, gin.H{"error": "A/B test not found"})
}

// StartABTest starts an A/B test
func (h *ABTestingHandler) StartABTest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	testIDStr := c.Param("testId")
	testID, err := strconv.ParseInt(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	err = h.abTestService.StartABTest(testID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start A/B test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "A/B test started successfully",
	})
}

// StopABTest stops an A/B test
func (h *ABTestingHandler) StopABTest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	testIDStr := c.Param("testId")
	testID, err := strconv.ParseInt(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	err = h.abTestService.StopABTest(testID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop A/B test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "A/B test stopped successfully",
	})
}

// GetABTestVariant gets the appropriate variant for a session
func (h *ABTestingHandler) GetABTestVariant(c *gin.Context) {
	testIDStr := c.Param("testId")
	testID, err := strconv.ParseInt(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		// Generate session ID from IP and User-Agent if not provided
		sessionID = c.ClientIP() + c.GetHeader("User-Agent")
	}

	variant, err := h.abTestService.GetABTestVariant(testID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get A/B test variant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"variant": variant,
	})
}

// RecordConversion records a conversion for an A/B test
func (h *ABTestingHandler) RecordConversion(c *gin.Context) {
	testIDStr := c.Param("testId")
	testID, err := strconv.ParseInt(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	var request struct {
		VariantID       int64   `json:"variant_id" binding:"required"`
		SessionID       string  `json:"session_id" binding:"required"`
		ConversionValue float64 `json:"conversion_value"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err = h.abTestService.RecordConversion(testID, request.VariantID, request.SessionID, request.ConversionValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record conversion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Conversion recorded successfully",
	})
}

// GetABTestResults gets the statistical results for an A/B test
func (h *ABTestingHandler) GetABTestResults(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	testIDStr := c.Param("testId")
	testID, err := strconv.ParseInt(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	results, err := h.abTestService.GetABTestResults(testID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get A/B test results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// GetSequentialTestResults gets sequential testing results for early stopping
func (h *ABTestingHandler) GetSequentialTestResults(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	testIDStr := c.Param("testId")
	testID, err := strconv.ParseInt(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	results, err := h.abTestService.CalculateSequentialTest(testID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate sequential test results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sequential_test": results,
	})
}

// GetSampleSizeCalculator provides sample size recommendations
func (h *ABTestingHandler) GetSampleSizeCalculator(c *gin.Context) {
	// Get query parameters
	baselineRateStr := c.DefaultQuery("baseline_rate", "0.1")
	minEffectStr := c.DefaultQuery("min_effect", "0.05") // 5% relative improvement
	confidenceStr := c.DefaultQuery("confidence", "95")
	powerStr := c.DefaultQuery("power", "80")

	baselineRate, err := strconv.ParseFloat(baselineRateStr, 64)
	if err != nil || baselineRate <= 0 || baselineRate >= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid baseline rate (must be between 0 and 1)"})
		return
	}

	minEffect, err := strconv.ParseFloat(minEffectStr, 64)
	if err != nil || minEffect <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid minimum effect"})
		return
	}

	confidence, err := strconv.ParseFloat(confidenceStr, 64)
	if err != nil || confidence <= 0 || confidence >= 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid confidence level (must be between 0 and 100)"})
		return
	}

	power, err := strconv.ParseFloat(powerStr, 64)
	if err != nil || power <= 0 || power >= 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid power level (must be between 0 and 100)"})
		return
	}

	// Calculate sample size recommendation
	sampleSize := h.abTestService.RecommendSampleSize(baselineRate, minEffect, confidence, power)
	
	// Calculate test duration estimates
	estimatedDaysLow := calculateTestDuration(sampleSize, baselineRate*0.8)  // Conservative estimate
	estimatedDaysHigh := calculateTestDuration(sampleSize, baselineRate*1.2) // Optimistic estimate

	c.JSON(http.StatusOK, gin.H{
		"sample_size_per_variant": sampleSize,
		"total_sample_size":      sampleSize * 2, // Assuming 2 variants
		"baseline_rate":          baselineRate,
		"minimum_effect":         minEffect,
		"confidence_level":       confidence,
		"power":                  power,
		"estimated_duration": gin.H{
			"low_traffic_days":  estimatedDaysLow,
			"high_traffic_days": estimatedDaysHigh,
		},
		"assumptions": []string{
			"50/50 traffic split between variants",
			"Two-sided test",
			"Normal approximation to binomial",
		},
	})
}

// calculateTestDuration estimates test duration based on traffic
func calculateTestDuration(sampleSize int64, dailyConversions float64) int {
	if dailyConversions <= 0 {
		return 0
	}
	
	// Estimate daily sessions based on conversion rate
	// Assuming a reasonable conversion rate to session ratio
	dailySessions := dailyConversions / 0.05 // Assuming 5% conversion rate
	
	if dailySessions <= 0 {
		return 0
	}
	
	days := float64(sampleSize) / dailySessions
	return int(math.Ceil(days))
}

// GetPowerAnalysis provides power analysis for existing tests
func (h *ABTestingHandler) GetPowerAnalysis(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	testIDStr := c.Param("testId")
	testID, err := strconv.ParseInt(testIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
		return
	}

	// Get current test results
	results, err := h.abTestService.GetABTestResults(testID, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get A/B test results"})
		return
	}

	if len(results.VariantResults) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Need at least 2 variants for power analysis"})
		return
	}

	// Find control and best variant
	var control, best *models.VariantResult
	for _, result := range results.VariantResults {
		if result.IsControl {
			control = result
		} else if best == nil || result.ConversionRate > best.ConversionRate {
			best = result
		}
	}

	if control == nil || best == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Control or variant not found"})
		return
	}

	// Calculate observed effect size
	observedEffect := (best.ConversionRate - control.ConversionRate) / control.ConversionRate

	// Calculate current power
	currentPower := h.calculatePower(control.ConversionRate, best.ConversionRate, int64(control.Sessions), int64(best.Sessions), 95.0)
	
	// Calculate power for different effect sizes
	powerCurve := make(map[string]float64)
	effectSizes := []float64{0.01, 0.02, 0.05, 0.1, 0.15, 0.2}
	
	for _, effect := range effectSizes {
		testRate := control.ConversionRate * (1 + effect)
		if testRate <= 1.0 {
			power := h.calculatePower(control.ConversionRate, testRate, int64(control.Sessions), int64(best.Sessions), 95.0)
			powerCurve[fmt.Sprintf("%.1f%%", effect*100)] = power
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"current_power":    currentPower,
		"observed_effect":  observedEffect,
		"power_curve":      powerCurve,
		"sample_sizes": gin.H{
			"control": control.Sessions,
			"variant": best.Sessions,
		},
		"conversion_rates": gin.H{
			"control": control.ConversionRate,
			"variant": best.ConversionRate,
		},
	})
}

// calculatePower calculates statistical power for given parameters
func (h *ABTestingHandler) calculatePower(p1, p2 float64, n1, n2 int64, confidence float64) float64 {
	if p1 <= 0 || p1 >= 1 || p2 <= 0 || p2 >= 1 || n1 == 0 || n2 == 0 {
		return 0
	}

	// Calculate effect size (Cohen's h)
	phi1 := 2 * math.Asin(math.Sqrt(p1))
	phi2 := 2 * math.Asin(math.Sqrt(p2))
	d := phi2 - phi1

	// Standard error
	se := math.Sqrt(1.0/float64(n1) + 1.0/float64(n2))

	// Critical value for two-tailed test (approximate for 95% confidence)
	zAlpha := 1.96

	// Non-centrality parameter
	delta := d / se

	// Power calculation (approximate)
	zBeta := delta - zAlpha
	power := 1 - normalCDF(-zBeta) // Using standard normal CDF

	if power < 0 {
		power = 0
	}
	if power > 1 {
		power = 1
	}

	return power
}

// Helper function for normal CDF (simplified)
func normalCDF(z float64) float64 {
	return 0.5 * (1 + math.Erf(z/math.Sqrt(2)))
}

// calculateSampleSize calculates recommended sample size per variant
func (h *ABTestingHandler) calculateSampleSize(baselineRate, minEffect, confidence, power float64) int64 {
	if baselineRate <= 0 || baselineRate >= 1 {
		return 0
	}
	
	alpha := 1 - (confidence / 100)
	beta := 1 - (power / 100)
	
	// Approximate inverse normal function
	zAlpha := h.normalInverse(1 - alpha/2)
	zBeta := h.normalInverse(1 - beta)
	
	p1 := baselineRate
	p2 := baselineRate * (1 + minEffect)
	
	if p2 >= 1 {
		p2 = 0.99 // Cap at 99%
	}
	
	pooledP := (p1 + p2) / 2
	
	numerator := math.Pow(zAlpha*math.Sqrt(2*pooledP*(1-pooledP)) + zBeta*math.Sqrt(p1*(1-p1)+p2*(1-p2)), 2)
	denominator := math.Pow(p2-p1, 2)
	
	if denominator == 0 {
		return 0
	}
	
	sampleSize := numerator / denominator
	
	return int64(math.Ceil(sampleSize))
}

// normalInverse calculates the inverse normal distribution (approximate)
func (h *ABTestingHandler) normalInverse(p float64) float64 {
	if p <= 0.5 {
		return -h.normalInverse(1 - p)
	}
	
	// Beasley-Springer-Moro approximation
	c0 := 2.515517
	c1 := 0.802853
	c2 := 0.010328
	d1 := 1.432788
	d2 := 0.189269
	d3 := 0.001308
	
	t := math.Sqrt(-2 * math.Log(1-p))
	
	return t - (c0+c1*t+c2*t*t)/(1+d1*t+d2*t*t+d3*t*t*t)
}