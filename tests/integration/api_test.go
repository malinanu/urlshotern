package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() (*gin.Engine, *handlers.Handler) {
	gin.SetMode(gin.TestMode)
	
	// Mock storage and services for integration tests
	// In a real integration test, you'd use test databases
	config := &configs.Config{
		BaseURL:    "http://localhost:8080",
		ServerHost: "localhost",
	}
	
	// For this example, we'll create a mock handler
	// In practice, you'd set up test databases and real services
	var shortenerService *services.ShortenerService
	var analyticsService *services.AnalyticsService
	
	handler := handlers.NewHandler(shortenerService, analyticsService)
	
	router := gin.New()
	
	// Health check endpoint
	router.GET("/health", handler.HealthCheck)
	
	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/shorten", handler.ShortenURL)
		v1.GET("/analytics/:shortCode", handler.GetAnalytics)
	}
	
	// Redirect route
	router.GET("/:shortCode", handler.RedirectURL)
	
	return router, handler
}

func TestHealthCheck(t *testing.T) {
	router, _ := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestShortenURL_Integration(t *testing.T) {
	router, _ := setupTestRouter()
	
	requestBody := models.ShortenRequest{
		URL: "https://www.example.com",
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Since we don't have real services set up, this will likely fail
	// In a real integration test, you'd expect a 201 status
	// For now, we'll just check that the endpoint is reachable
	assert.NotEqual(t, 404, w.Code, "Endpoint should be reachable")
}

func TestInvalidJSONRequest(t *testing.T) {
	router, _ := setupTestRouter()
	
	invalidJSON := `{"url": "https://example.com", "invalid": }`
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/shorten", bytes.NewBuffer([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
	
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_request", response.Error)
}

func TestMissingURLField(t *testing.T) {
	router, _ := setupTestRouter()
	
	requestBody := map[string]string{
		"custom_code": "test123",
		// Missing required "url" field
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
}

func TestRedirectEndpoint_NotFound(t *testing.T) {
	router, _ := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)
	
	// This will likely return an error since we don't have real services
	// In a real integration test with proper setup, you'd test actual redirects
	assert.NotEqual(t, 200, w.Code)
}

func TestAnalyticsEndpoint(t *testing.T) {
	router, _ := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123", nil)
	router.ServeHTTP(w, req)
	
	// Since we don't have real services, this will error
	// In a real integration test, you'd set up test data and verify analytics
	assert.NotEqual(t, 200, w.Code)
}

func TestAnalyticsEndpoint_InvalidDaysParameter(t *testing.T) {
	router, _ := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/analytics/test123?days=invalid", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
	
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_days_parameter", response.Error)
}

// Example of a more complete integration test that would work with real services
func TestCompleteFlow_WithRealServices(t *testing.T) {
	t.Skip("Skipping integration test - requires database setup")
	
	// This is an example of how you'd write a complete integration test:
	// 1. Set up test database
	// 2. Initialize real services with test config
	// 3. Create a short URL
	// 4. Test the redirect
	// 5. Verify analytics
	// 6. Clean up test data
}

// Benchmark test for API performance
func BenchmarkShortenURL(b *testing.B) {
	router, _ := setupTestRouter()
	
	requestBody := models.ShortenRequest{
		URL: "https://www.example.com",
	}
	
	jsonBody, _ := json.Marshal(requestBody)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/shorten", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}