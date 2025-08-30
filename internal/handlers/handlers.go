package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	shortenerService *services.ShortenerService
	analyticsService *services.AnalyticsService
}

// NewHandler creates a new handler instance
func NewHandler(shortenerService *services.ShortenerService, analyticsService *services.AnalyticsService) *Handler {
	// Set Redis on analytics service if available
	return &Handler{
		shortenerService: shortenerService,
		analyticsService: analyticsService,
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "url-shortener",
		"timestamp": gin.H{},
	})
}

// ShortenURL handles URL shortening requests
func (h *Handler) ShortenURL(c *gin.Context) {
	var request models.ShortenRequest
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Get client IP
	clientIP := getClientIP(c)

	// Shorten the URL
	response, err := h.shortenerService.ShortenURL(&request, clientIP)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorType := "internal_error"

		// Handle specific service errors
		switch err {
		case services.ErrInvalidURL, services.ErrInvalidURLScheme, services.ErrSelfReferentialURL:
			statusCode = http.StatusBadRequest
			errorType = "invalid_url"
		case services.ErrInvalidCustomCodeLength, services.ErrInvalidCustomCodeCharacters:
			statusCode = http.StatusBadRequest
			errorType = "invalid_custom_code"
		case services.ErrReservedCustomCode, services.ErrCustomCodeAlreadyExists:
			statusCode = http.StatusConflict
			errorType = "custom_code_unavailable"
		}

		c.JSON(statusCode, models.ErrorResponse{
			Error:   errorType,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// RedirectURL handles URL redirection
func (h *Handler) RedirectURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	// Get original URL
	mapping, err := h.shortenerService.GetOriginalURL(shortCode)
	if err != nil {
		statusCode := http.StatusNotFound
		errorType := "not_found"
		
		if err == storage.ErrURLExpired {
			statusCode = http.StatusGone
			errorType = "url_expired"
		}

		c.JSON(statusCode, models.ErrorResponse{
			Error:   errorType,
			Message: err.Error(),
		})
		return
	}

	// Record the click for analytics
	clientIP := getClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")
	
	go func() {
		if err := h.shortenerService.RecordClick(shortCode, clientIP, userAgent, referrer); err != nil {
			// Log error but don't fail the redirect
			// In production, you'd use proper logging
		}
	}()

	// Perform the redirect
	c.Redirect(http.StatusMovedPermanently, mapping.OriginalURL)
}

// GetAnalytics handles analytics requests
func (h *Handler) GetAnalytics(c *gin.Context) {
	shortCode := c.Param("shortCode")
	
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	// Get days parameter (default to 30 days)
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_days_parameter",
			Message: "Days must be between 1 and 365",
		})
		return
	}

	// Get analytics data
	analytics, err := h.analyticsService.GetAnalytics(shortCode, days)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorType := "internal_error"
		
		if err == storage.ErrURLNotFound {
			statusCode = http.StatusNotFound
			errorType = "not_found"
		}

		c.JSON(statusCode, models.ErrorResponse{
			Error:   errorType,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetClickTrends handles click trends requests
func (h *Handler) GetClickTrends(c *gin.Context) {
	shortCode := c.Param("shortCode")
	period := c.DefaultQuery("period", "day")
	
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_short_code",
			Message: "Short code is required",
		})
		return
	}

	trends, err := h.analyticsService.GetClickTrends(shortCode, period)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_period",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_code": shortCode,
		"period":     period,
		"trends":     trends,
	})
}

// GetDashboardStats handles dashboard statistics requests
func (h *Handler) GetDashboardStats(c *gin.Context) {
	stats, err := h.analyticsService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// BatchShortenURLs handles batch URL shortening (for future extension)
func (h *Handler) BatchShortenURLs(c *gin.Context) {
	var requests []models.ShortenRequest
	
	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if len(requests) > 100 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "too_many_requests",
			Message: "Maximum 100 URLs can be processed in batch",
		})
		return
	}

	clientIP := getClientIP(c)
	var responses []models.ShortenResponse
	var errors []models.ErrorResponse

	for _, request := range requests {
		response, err := h.shortenerService.ShortenURL(&request, clientIP)
		if err != nil {
			errors = append(errors, models.ErrorResponse{
				Error:   "processing_error",
				Message: err.Error(),
			})
			continue
		}
		responses = append(responses, *response)
	}

	c.JSON(http.StatusOK, gin.H{
		"successful": len(responses),
		"failed":     len(errors),
		"results":    responses,
		"errors":     errors,
	})
}

// getClientIP extracts the real client IP from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for load balancers/proxies)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Get the first IP from the comma-separated list
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return c.ClientIP()
}