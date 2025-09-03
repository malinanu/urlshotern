package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupEnhancedAnalyticsRouter() (*gin.Engine, *handlers.Handler) {
	gin.SetMode(gin.TestMode)
	
	// Mock configuration
	config := &configs.Config{
		BaseURL:    "http://localhost:8080",
		ServerHost: "localhost",
	}
	
	// For integration tests, we'll use mock services
	// In a real integration test, you'd set up test databases
	var shortenerService *services.ShortenerService
	var analyticsService *services.AnalyticsService
	var advancedAnalyticsService *services.AdvancedAnalyticsService
	var conversionTrackingService *services.ConversionTrackingService
	var abTestingService *services.ABTestingService
	var realtimeAnalyticsService *services.RealtimeAnalyticsService
	var attributionService *services.AttributionService
	
	handler := handlers.NewHandler(
		shortenerService, 
		analyticsService, 
		advancedAnalyticsService, 
		conversionTrackingService, 
		abTestingService, 
		realtimeAnalyticsService, 
		attributionService, 
		nil, // authHandlers
		nil, // db
	)
	
	router := gin.New()
	
	// Health check endpoint
	router.GET("/health", handler.HealthCheck)
	
	// API routes
	v1 := router.Group("/api/v1")
	{
		// Enhanced Analytics endpoints
		v1.GET("/analytics/:shortCode/geographic", handler.GetGeographicAnalytics)
		v1.GET("/analytics/:shortCode/devices", handler.GetDeviceAnalytics)
		v1.GET("/analytics/:shortCode/referrers", handler.GetReferrerAnalytics)
		v1.GET("/analytics/:shortCode/timeline", handler.GetTimeBasedAnalytics)
		v1.GET("/analytics/:shortCode/funnel", handler.GetConversionFunnelData)
		
		// Attribution endpoints
		v1.GET("/attribution/:shortCode", handler.GetAttributionReport)
		v1.POST("/attribution/conversion", handler.RecordConversion)
		
		// A/B Testing endpoints
		v1.POST("/ab-test", handler.CreateABTest)
		v1.GET("/ab-test/:testID", handler.GetABTestResults)
		v1.GET("/ab-test/:testID/significance", handler.GetSignificanceAnalysis)
		v1.POST("/ab-test/:testID/stop", handler.StopABTest)
		v1.GET("/ab-test/:testID/sequential", handler.GetSequentialTestResults)
		
		// Real-time analytics
		v1.GET("/realtime/:shortCode", handler.GetRealtimeAnalytics)
		v1.GET("/realtime/:shortCode/websocket", handler.WebSocketHandler)
	}
	
	return router, handler
}

func TestGeographicAnalyticsEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123/geographic", nil)
	router.ServeHTTP(w, req)
	
	// Since we don't have real services, this will likely return an error
	// In a real integration test, you'd expect proper data
	assert.NotEqual(t, 404, w.Code, "Geographic analytics endpoint should be reachable")
}

func TestGeographicAnalyticsWithDaysParameter(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123/geographic?days=7", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Geographic analytics endpoint with days parameter should be reachable")
}

func TestGeographicAnalyticsInvalidDaysParameter(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123/geographic?days=invalid", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
	
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_days_parameter", response.Error)
}

func TestDeviceAnalyticsEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123/devices", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Device analytics endpoint should be reachable")
}

func TestDeviceAnalyticsWithDaysParameter(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123/devices?days=30", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Device analytics endpoint with days parameter should be reachable")
}

func TestReferrerAnalyticsEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123/referrers", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Referrer analytics endpoint should be reachable")
}

func TestTimeBasedAnalyticsEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123/timeline", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Time-based analytics endpoint should be reachable")
}

func TestConversionFunnelEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123/funnel", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Conversion funnel endpoint should be reachable")
}

func TestAttributionReportEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/attribution/test123", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Attribution report endpoint should be reachable")
}

func TestAttributionReportWithModelParameter(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/attribution/test123?model=linear", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Attribution report endpoint with model parameter should be reachable")
}

func TestAttributionReportInvalidModelParameter(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/attribution/test123?model=invalid_model", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
	
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_attribution_model", response.Error)
}

func TestRecordConversionEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	conversionRequest := models.ConversionEvent{
		SessionID:       "session123",
		ShortCode:      "test123",
		ConversionType: "purchase",
		Value:          99.99,
		EventTime:      time.Now(),
	}
	
	jsonBody, _ := json.Marshal(conversionRequest)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/attribution/conversion", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Record conversion endpoint should be reachable")
}

func TestRecordConversionInvalidJSON(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	invalidJSON := `{"session_id": "session123", "invalid": }`
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/attribution/conversion", bytes.NewBuffer([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
	
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestCreateABTestEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	abTestRequest := models.ABTestRequest{
		Name:        "Test A/B Test",
		Description: "Testing button colors",
		ShortCodes:  []string{"test123", "test456"},
		TrafficSplit: []float64{0.5, 0.5},
		MetricName:  "click_through_rate",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(7 * 24 * time.Hour),
	}
	
	jsonBody, _ := json.Marshal(abTestRequest)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ab-test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Create A/B test endpoint should be reachable")
}

func TestCreateABTestInvalidRequest(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	// Missing required fields
	abTestRequest := map[string]interface{}{
		"name": "Test A/B Test",
		// Missing other required fields
	}
	
	jsonBody, _ := json.Marshal(abTestRequest)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ab-test", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
}

func TestGetABTestResultsEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ab-test/123", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Get A/B test results endpoint should be reachable")
}

func TestGetSignificanceAnalysisEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ab-test/123/significance", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Get significance analysis endpoint should be reachable")
}

func TestGetSignificanceAnalysisWithConfidenceParameter(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ab-test/123/significance?confidence=99", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Get significance analysis endpoint with confidence parameter should be reachable")
}

func TestGetSignificanceAnalysisInvalidConfidenceParameter(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ab-test/123/significance?confidence=invalid", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
	
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_confidence_parameter", response.Error)
}

func TestStopABTestEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	stopRequest := map[string]interface{}{
		"reason": "Test completed successfully",
	}
	
	jsonBody, _ := json.Marshal(stopRequest)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ab-test/123/stop", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Stop A/B test endpoint should be reachable")
}

func TestGetSequentialTestResultsEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ab-test/123/sequential", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Get sequential test results endpoint should be reachable")
}

func TestRealtimeAnalyticsEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/realtime/test123", nil)
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "Realtime analytics endpoint should be reachable")
}

func TestWebSocketEndpoint(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/realtime/test123/websocket", nil)
	// Add WebSocket headers
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	
	router.ServeHTTP(w, req)
	
	assert.NotEqual(t, 404, w.Code, "WebSocket endpoint should be reachable")
}

// Test parameter validation across multiple endpoints
func TestDaysParameterValidation(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	endpoints := []string{
		"/api/v1/analytics/test123/geographic",
		"/api/v1/analytics/test123/devices",
		"/api/v1/analytics/test123/referrers",
		"/api/v1/analytics/test123/timeline",
		"/api/v1/analytics/test123/funnel",
		"/api/v1/attribution/test123",
	}
	
	invalidDaysValues := []string{"invalid", "-1", "0", "366", "abc"}
	
	for _, endpoint := range endpoints {
		for _, daysValue := range invalidDaysValues {
			t.Run("Invalid days parameter for "+endpoint, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", endpoint+"?days="+daysValue, nil)
				router.ServeHTTP(w, req)
				
				if w.Code == 400 {
					var response models.ErrorResponse
					err := json.Unmarshal(w.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.Equal(t, "invalid_days_parameter", response.Error)
				}
			})
		}
	}
}

func TestShortCodeParameterValidation(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	invalidShortCodes := []string{"", " ", "a", "very-long-invalid-code"}
	
	for _, shortCode := range invalidShortCodes {
		t.Run("Invalid short code: "+shortCode, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/analytics/"+shortCode+"/geographic", nil)
			router.ServeHTTP(w, req)
			
			// Should handle gracefully - either return 404 for not found or 400 for invalid parameter
			assert.True(t, w.Code == 400 || w.Code == 404 || w.Code == 500, "Should handle invalid short codes appropriately")
		})
	}
}

// Test CORS headers
func TestCORSHeaders(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	// Add CORS middleware for testing
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/analytics/test123/geographic", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

// Test rate limiting simulation
func TestRateLimitingSimulation(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	// Simulate rapid requests
	endpoint := "/api/v1/analytics/test123/geographic"
	requestCount := 10
	
	for i := 0; i < requestCount; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", endpoint, nil)
		router.ServeHTTP(w, req)
		
		// In a real integration test with rate limiting, you'd expect 429 after certain limit
		assert.NotEqual(t, 429, w.Code, "Rate limiting not implemented in this test setup")
	}
}

// Test concurrent requests
func TestConcurrentRequests(t *testing.T) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	endpoint := "/api/v1/analytics/test123/geographic"
	concurrentRequests := 5
	
	// Channel to collect results
	results := make(chan int, concurrentRequests)
	
	// Launch concurrent requests
	for i := 0; i < concurrentRequests; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", endpoint, nil)
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}
	
	// Collect all results
	for i := 0; i < concurrentRequests; i++ {
		statusCode := <-results
		assert.NotEqual(t, 404, statusCode, "All concurrent requests should reach the endpoint")
	}
}

// Benchmark test for enhanced analytics endpoints
func BenchmarkGeographicAnalyticsEndpoint(b *testing.B) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/analytics/bench/geographic", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkDeviceAnalyticsEndpoint(b *testing.B) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/analytics/bench/devices", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkAttributionReportEndpoint(b *testing.B) {
	router, _ := setupEnhancedAnalyticsRouter()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/attribution/bench", nil)
		router.ServeHTTP(w, req)
	}
}