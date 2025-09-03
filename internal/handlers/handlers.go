package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	shortenerService       *services.ShortenerService
	analyticsService       *services.AnalyticsService
	advancedAnalyticsService *services.AdvancedAnalyticsService
	conversionTrackingService *services.ConversionTrackingService
	abTestingService       *services.ABTestingService
	realtimeService        *services.RealtimeAnalyticsService
	attributionService     *services.AttributionService
	AuthHandlers           *AuthHandlers
	AnalyticsHandlers      *AnalyticsHandlers
	ConversionHandlers     *ConversionTrackingHandler
	ABTestHandlers         *ABTestingHandler
	RealtimeHandlers       *RealtimeAnalyticsHandler
	BillingHandlers        *BillingHandler
	AdvancedAnalyticsHandlers *AdvancedAnalyticsHandler
	AttributionHandlers    *AttributionHandler
}

// NewHandler creates a new handler instance
func NewHandler(shortenerService *services.ShortenerService, analyticsService *services.AnalyticsService, advancedAnalyticsService *services.AdvancedAnalyticsService, conversionService *services.ConversionTrackingService, abTestService *services.ABTestingService, realtimeService *services.RealtimeAnalyticsService, attributionService *services.AttributionService, authHandlers *AuthHandlers, analyticsHandlers *AnalyticsHandlers, db *storage.PostgresStorage) *Handler {
	return &Handler{
		shortenerService:       shortenerService,
		analyticsService:       analyticsService,
		advancedAnalyticsService: advancedAnalyticsService,
		conversionTrackingService: conversionService,
		abTestingService:       abTestService,
		realtimeService:        realtimeService,
		attributionService:     attributionService,
		AuthHandlers:           authHandlers,
		AnalyticsHandlers:      analyticsHandlers,
		ConversionHandlers:     NewConversionTrackingHandler(conversionService),
		ABTestHandlers:         NewABTestingHandler(abTestService),
		RealtimeHandlers:       NewRealtimeAnalyticsHandler(realtimeService),
		BillingHandlers:        NewBillingHandler(nil), // Will need to fix this properly
		AdvancedAnalyticsHandlers: NewAdvancedAnalyticsHandler(advancedAnalyticsService),
		AttributionHandlers:    NewAttributionHandler(attributionService),
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

	// Get user ID if authenticated (optional)
	userID, _ := middleware.GetUserID(c)
	var userIDPtr *int64
	if userID > 0 {
		userIDPtr = &userID
	}

	// Shorten the URL
	response, err := h.shortenerService.ShortenURL(&request, clientIP, userIDPtr)
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

// GetAdvancedAnalytics handles advanced analytics requests with detailed breakdowns
func (h *Handler) GetAdvancedAnalytics(c *gin.Context) {
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

	// Get advanced analytics data
	analytics, err := h.analyticsService.GetAdvancedAnalytics(shortCode, days)
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
	
	// Get user ID if authenticated (optional)
	userID, _ := middleware.GetUserID(c)
	var userIDPtr *int64
	if userID > 0 {
		userIDPtr = &userID
	}
	
	var responses []models.ShortenResponse
	var errors []models.ErrorResponse

	for _, request := range requests {
		response, err := h.shortenerService.ShortenURL(&request, clientIP, userIDPtr)
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

// GetUserURLs handles requests for user's URL list
func (h *Handler) GetUserURLs(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	urls, total, err := h.shortenerService.GetUserURLs(userID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"urls":       urls,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// DeleteURL handles URL deletion requests
func (h *Handler) DeleteURL(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Short code is required",
		})
		return
	}

	err := h.shortenerService.DeleteUserURL(userID, shortCode)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == storage.ErrURLNotFound {
			statusCode = http.StatusNotFound
		} else if err == storage.ErrUnauthorized {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, models.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URL deleted successfully",
	})
}

// UpdateURL handles URL update requests
func (h *Handler) UpdateURL(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Short code is required",
		})
		return
	}

	var updateRequest models.UpdateURLRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	updatedURL, err := h.shortenerService.UpdateUserURL(userID, shortCode, &updateRequest)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == storage.ErrURLNotFound {
			statusCode = http.StatusNotFound
		} else if err == storage.ErrUnauthorized {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedURL)
}

// GetUserDashboardStats handles user dashboard statistics
func (h *Handler) GetUserDashboardStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	stats, err := h.analyticsService.GetUserDashboardStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
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