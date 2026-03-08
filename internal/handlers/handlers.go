package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	shortenerService       *services.ShortenerService
	analyticsService       *services.AnalyticsService
	advancedAnalyticsService *services.AdvancedAnalyticsService
	conversionTrackingService *services.ConversionTrackingService
	abTestingService       *services.ABTestingService
	realtimeService        *services.RealtimeAnalyticsService
	attributionService     *services.AttributionService
	billingService         *services.BillingService  // Added for usage limit enforcement
	botDetector            *services.BotDetectionService
	db                     *storage.PostgresStorage  // Database storage for monitoring
	AuthHandlers           *AuthHandlers
	AnalyticsHandlers      *AnalyticsHandlers
	ConversionHandlers     *ConversionTrackingHandler
	ABTestHandlers         *ABTestingHandler
	RealtimeHandlers       *RealtimeAnalyticsHandler
	BillingHandlers        *BillingHandler
	APIKeyHandlers         *APIKeyHandler
	PolarWebhookHandler    *PolarWebhookHandler
	AdvancedAnalyticsHandlers *AdvancedAnalyticsHandler
	AttributionHandlers    *AttributionHandler
	CMSHandlers            *CMSHandlers
	AdminRBACHandlers      *AdminRBACHandlers  // RBAC user and role management
	AdminCouponHandlers    *AdminCouponHandlers
	ContactHandlers        *ContactHandlers
	AIInsightsHandlers     *AIInsightsHandler
	DomainHandlers         *DomainHandlers
	BrandingHandlers       *BrandingHandler
	TeamHandlers           *TeamHandlers
	GlobalSettingsHandlers *GlobalSettingsHandlers
}

// NewHandler creates a new handler instance
func NewHandler(shortenerService *services.ShortenerService, analyticsService *services.AnalyticsService, advancedAnalyticsService *services.AdvancedAnalyticsService, conversionService *services.ConversionTrackingService, abTestService *services.ABTestingService, realtimeService *services.RealtimeAnalyticsService, attributionService *services.AttributionService, authHandlers *AuthHandlers, analyticsHandlers *AnalyticsHandlers, db *storage.PostgresStorage, gormDB *gorm.DB, config *configs.Config, adminRBACHandlers *AdminRBACHandlers, botDetector *services.BotDetectionService, aiInsightsService *services.AIInsightsService, domainService *services.DomainService, brandingService *services.BrandingService, teamHandlers *TeamHandlers, globalSettingsService *services.GlobalSettingsService) *Handler {
	// Create API key service and handler
	apiKeyService := services.NewAPIKeyService(db)

	// Create CMS service and handler
	cmsService := services.NewCMSService(db)
	cmsHandlers := NewCMSHandlers(cmsService, apiKeyService)

	// Create billing service and handlers
	billingHandler := NewBillingHandler(gormDB, config)

	return &Handler{
		shortenerService:       shortenerService,
		analyticsService:       analyticsService,
		advancedAnalyticsService: advancedAnalyticsService,
		conversionTrackingService: conversionService,
		abTestingService:       abTestService,
		realtimeService:        realtimeService,
		attributionService:     attributionService,
		billingService:         billingHandler.billingService,  // Store billing service for usage limit checks
		botDetector:            botDetector,
		db:                     db,  // Store database for monitoring
		AuthHandlers:           authHandlers,
		AnalyticsHandlers:      analyticsHandlers,
		ConversionHandlers:     NewConversionTrackingHandler(conversionService),
		ABTestHandlers:         NewABTestingHandler(abTestService),
		RealtimeHandlers:       NewRealtimeAnalyticsHandler(realtimeService),
		BillingHandlers:        billingHandler,
		APIKeyHandlers:         NewAPIKeyHandler(apiKeyService),
		PolarWebhookHandler:    NewPolarWebhookHandler(gormDB, billingHandler.billingService),
		AdvancedAnalyticsHandlers: NewAdvancedAnalyticsHandler(advancedAnalyticsService),
		AttributionHandlers:    NewAttributionHandler(attributionService),
		CMSHandlers:            cmsHandlers,
		AdminRBACHandlers:      adminRBACHandlers,
		AdminCouponHandlers:    NewAdminCouponHandlers(gormDB, billingHandler.billingService.GetPolarService()),
		ContactHandlers:        NewContactHandlers(authHandlers.GetEmailService()),
		AIInsightsHandlers:     NewAIInsightsHandler(aiInsightsService),
		DomainHandlers:         NewDomainHandlers(domainService),
		BrandingHandlers:       NewBrandingHandler(brandingService),
		TeamHandlers:           teamHandlers,
		GlobalSettingsHandlers: func() *GlobalSettingsHandlers {
			if globalSettingsService != nil {
				return NewGlobalSettingsHandlers(globalSettingsService)
			}
			return nil
		}(),
	}
}

// GetBillingService exposes the billing service for plan gating middleware.
func (h *Handler) GetBillingService() *services.BillingService {
	return h.billingService
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "url-shortener",
		"timestamp": gin.H{},
	})
}

// DBPoolStats handles database connection pool statistics requests
func (h *Handler) DBPoolStats(c *gin.Context) {
	stats := h.db.GetPoolStats()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// GoroutineStats handles goroutine error statistics requests
func (h *Handler) GoroutineStats(c *gin.Context) {
	stats := utils.GetGoroutineStats()

	// Get recent errors if requested
	limitStr := c.DefaultQuery("limit", "10")
	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	recentErrors := utils.GetRecentGoroutineErrors(limit)

	// Format errors for response
	formattedErrors := make([]gin.H, len(recentErrors))
	for i, err := range recentErrors {
		formattedErrors[i] = gin.H{
			"timestamp": err.Timestamp.Format(time.RFC3339),
			"operation": err.Operation,
			"error":     err.Error.Error(),
			"recovered": err.Recovered,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"stats":         stats,
			"recent_errors": formattedErrors,
		},
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

		// Check usage limits for authenticated users
		canCreate, message, err := h.billingService.CanCreateURL(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "usage_check_failed",
				Message: "Failed to verify usage limits",
			})
			return
		}
		if !canCreate {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "usage_limit_exceeded",
				Message: message,
			})
			return
		}
	} else {
		// Anonymous user: enforce 5 free URLs per IP
		count, err := h.db.CountURLsByIP(clientIP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to check URL limit",
			})
			return
		}
		if count >= 5 {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "anonymous_limit_reached",
				Message: "You've reached the free limit of 5 URLs. Please create an account to shorten more URLs.",
			})
			return
		}
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

	// Check if request came in via a custom domain
	var mapping *models.URLMapping
	var err error
	if domainHost, exists := c.Get("custom_domain_host"); exists {
		mapping, err = h.shortenerService.GetOriginalURLByDomain(shortCode, domainHost.(string))
	} else {
		mapping, err = h.shortenerService.GetOriginalURL(shortCode)
	}

	if err != nil {
		// Browser request → redirect to 404 page
		if strings.Contains(c.GetHeader("Accept"), "text/html") {
			c.Redirect(http.StatusFound, "/not-found")
			return
		}
		// API/curl request → keep JSON response
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

	// Password-protected link: redirect to password entry page
	if mapping.PasswordHash != nil {
		c.Redirect(http.StatusFound, "/p/"+shortCode)
		return
	}

	// Record the click for analytics
	clientIP := getClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")
	acceptLang := c.GetHeader("Accept-Language")
	accept := c.GetHeader("Accept")

	isBot := h.botDetector.IsBot(userAgent, acceptLang, accept, clientIP, shortCode)

	utils.Go("record_click", func() error {
		return h.shortenerService.RecordClick(shortCode, clientIP, userAgent, referrer, isBot)
	})

	// If OG metadata is set, serve an HTML page so social crawlers read OG tags
	if mapping.OGTitle != nil || mapping.OGImage != nil || mapping.OGDescription != nil {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(buildOGPage(mapping)))
		return
	}

	// Fast path: plain redirect
	c.Redirect(http.StatusMovedPermanently, mapping.OriginalURL)
}

// buildOGPage generates a minimal HTML page with OG meta tags and a JS redirect.
// Social crawlers (Facebook, Twitter, Slack) read the tags; real users get an
// instant redirect via the <script> block.
func buildOGPage(m *models.URLMapping) string {
	title := m.OriginalURL
	if m.OGTitle != nil && *m.OGTitle != "" {
		title = *m.OGTitle
	}
	description := ""
	if m.OGDescription != nil {
		description = *m.OGDescription
	}
	image := ""
	if m.OGImage != nil {
		image = *m.OGImage
	}
	card := "summary"
	if m.TwitterCard != nil && *m.TwitterCard != "" {
		card = *m.TwitterCard
	}

	escaped := strings.NewReplacer(`"`, `&quot;`, `<`, `&lt;`, `>`, `&gt;`, `&`, `&amp;`).Replace
	dest := strings.ReplaceAll(m.OriginalURL, `"`, `%22`)

	return `<!DOCTYPE html><html><head>` +
		`<meta charset="utf-8">` +
		`<title>` + escaped(title) + `</title>` +
		`<meta property="og:title" content="` + escaped(title) + `">` +
		`<meta property="og:description" content="` + escaped(description) + `">` +
		`<meta property="og:image" content="` + escaped(image) + `">` +
		`<meta property="og:url" content="` + escaped(m.OriginalURL) + `">` +
		`<meta name="twitter:card" content="` + escaped(card) + `">` +
		`<meta name="twitter:title" content="` + escaped(title) + `">` +
		`<meta name="twitter:description" content="` + escaped(description) + `">` +
		`<meta name="twitter:image" content="` + escaped(image) + `">` +
		`<script>window.location.href="` + dest + `";</script>` +
		`</head><body><a href="` + escaped(m.OriginalURL) + `">Click here if not redirected.</a></body></html>`
}

// UpdateURLOGMetadata handles PATCH /api/v1/urls/:shortCode/og
func (h *Handler) UpdateURLOGMetadata(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "authentication_required", Message: "Authentication required"})
		return
	}

	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: "Short code is required"})
		return
	}

	var req models.UpdateOGRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	if err := h.db.UpdateURLOGMetadata(shortCode, userID, &req); err != nil {
		if err == storage.ErrURLNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "URL not found or not owned by you"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OG metadata updated successfully"})
}

// CheckPassword verifies a link password and returns the destination URL
func (h *Handler) CheckPassword(c *gin.Context) {
	shortCode := c.Param("shortCode")
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password required"})
		return
	}

	mapping, err := h.shortenerService.GetOriginalURL(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	if mapping.PasswordHash == nil {
		// No password needed — just return the URL
		c.JSON(http.StatusOK, gin.H{"redirect_url": mapping.OriginalURL})
		return
	}

	if err := h.shortenerService.CheckPassword(*mapping.PasswordHash, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
		return
	}

	// Record click after successful password verification
	clientIP := getClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")
	acceptLang := c.GetHeader("Accept-Language")
	accept := c.GetHeader("Accept")
	isBot := h.botDetector.IsBot(userAgent, acceptLang, accept, clientIP, shortCode)
	utils.Go("record_click", func() error {
		return h.shortenerService.RecordClick(shortCode, clientIP, userAgent, referrer, isBot)
	})

	c.JSON(http.StatusOK, gin.H{"redirect_url": mapping.OriginalURL})
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
			Error:   "authentication_required",
			Message: "Authentication required to access this resource",
		})
		c.Abort()
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
			Error:   "authentication_required",
			Message: "Authentication required to access this resource",
		})
		c.Abort()
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
			Error:   "authentication_required",
			Message: "Authentication required to access this resource",
		})
		c.Abort()
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
			Error:   "authentication_required",
			Message: "Authentication required to access this resource",
		})
		c.Abort()
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