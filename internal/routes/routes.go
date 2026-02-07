package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/middleware"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	router *gin.Engine,
	handler *handlers.Handler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Add global middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware())

	// Handle HEAD requests by converting them to GET requests
	// This allows Next.js prefetching to work correctly
	router.Use(func(c *gin.Context) {
		if c.Request.Method == "HEAD" {
			c.Request.Method = "GET"
			c.Next()
			c.Request.Method = "HEAD"
		} else {
			c.Next()
		}
	})

	// Health check
	router.GET("/health", handler.HealthCheck)

	// Database connection pool statistics (for monitoring)
	router.GET("/health/db-pool", handler.DBPoolStats)

	// Goroutine error statistics (for monitoring)
	router.GET("/health/goroutines", handler.GoroutineStats)

	// Serve common static files BEFORE /:shortCode route to prevent them from being treated as short codes
	commonStaticFiles := []string{
		"/manifest.json", "/robots.txt", "/sitemap.xml",
		"/favicon.ico", "/sw.js", "/apple-touch-icon.png",
	}
	for _, file := range commonStaticFiles {
		filePath := file // Capture for closure
		router.GET(filePath, func(c *gin.Context) {
			staticPath := filepath.Join("./static", filePath)
			if _, err := os.Stat(staticPath); err == nil {
				c.File(staticPath)
			} else {
				c.JSON(404, gin.H{"error": "File not found"})
			}
		})
	}

	// Public routes (no authentication required)
	setupPublicRoutes(router, handler, authMiddleware)

	// Authentication routes
	setupAuthRoutes(router, handler.AuthHandlers, authMiddleware)

	// Protected routes (authentication required - supports both JWT and API keys)
	setupProtectedRoutes(router, handler, authMiddleware)

	// Admin routes (admin access required) - temporarily disabled
	// setupAdminRoutes(router, handler, authService)

	// Static file serving for frontend (must be last to not interfere with API routes)
	setupStaticRoutes(router)
}

// setupPublicRoutes configures public routes
func setupPublicRoutes(router *gin.Engine, handler *handlers.Handler, authMiddleware *middleware.AuthMiddleware) {
	// Get API key service from handler
	apiKeyService := handler.APIKeyHandlers.GetAPIKeyService()

	// URL shortening (supports both JWT and API key authentication)
	router.POST("/api/v1/shorten",
		middleware.FlexibleAuthMiddleware(authMiddleware, apiKeyService),
		middleware.RateLimitMiddleware(middleware.DefaultRateLimit),
		handler.ShortenURL,
	)

	// URL redirection (no auth required)
	router.GET("/:shortCode", handler.RedirectURL)

	// Public analytics (supports both JWT and API key authentication for enhanced features)
	router.GET("/api/v1/analytics/:shortCode",
		middleware.FlexibleAuthMiddleware(authMiddleware, apiKeyService),
		middleware.RateLimitMiddleware(middleware.DefaultRateLimit),
		handler.GetAnalytics,
	)

	// Public product listing for pricing page (no auth required)
	router.GET("/api/v1/products", handler.BillingHandlers.GetProducts)

	// Public CMS pages - temporarily disabled
	// router.GET("/api/v1/pages/:slug", handler.CMSHandlers.PublicPageHandler)
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(router *gin.Engine, authHandlers *handlers.AuthHandlers, authMiddleware *middleware.AuthMiddleware) {
	auth := router.Group("/api/v1/auth")
	
	// Apply auth-specific rate limiting
	auth.Use(middleware.RateLimitMiddleware(middleware.AuthRateLimit))
	
	// Registration and login
	auth.POST("/register", authHandlers.Register)
	auth.POST("/login", authHandlers.Login)
	auth.POST("/refresh", authHandlers.RefreshToken)
	
	// OTP routes (more restrictive rate limiting)
	otpGroup := auth.Group("/otp")
	otpGroup.Use(middleware.RateLimitMiddleware(middleware.OTPRateLimit))
	{
		otpGroup.POST("/send", authHandlers.SendOTP)
		otpGroup.POST("/verify", authHandlers.VerifyOTP)
		otpGroup.POST("/resend", authHandlers.ResendOTP)
	}
	
	// Email verification routes
	auth.GET("/verify-email", authHandlers.VerifyEmail)
	auth.POST("/resend-email", authHandlers.SendEmailVerification)
	
	// Protected auth routes (require authentication)
	protected := auth.Group("/")
	protected.Use(authMiddleware.RequireAuth())  // Re-enabled authentication middleware
	{
		protected.GET("/profile", authHandlers.GetProfile)
		protected.PUT("/profile", authHandlers.UpdateProfile)
		protected.GET("/preferences", authHandlers.GetPreferences)
		protected.PUT("/preferences", authHandlers.UpdatePreferences)
		protected.PUT("/change-password", authHandlers.ChangePassword)
		protected.POST("/logout", authHandlers.Logout)
		protected.POST("/send-otp", authHandlers.SendOTP) // For authenticated users
	}
}

// setupProtectedRoutes configures routes that require authentication
func setupProtectedRoutes(router *gin.Engine, handler *handlers.Handler, authMiddleware *middleware.AuthMiddleware) {
	// Get API key service from handler
	apiKeyService := handler.APIKeyHandlers.GetAPIKeyService()

	api := router.Group("/api/v1")
	// Use FlexibleAuthMiddleware to accept both JWT tokens and API keys
	api.Use(middleware.FlexibleAuthMiddleware(authMiddleware, apiKeyService))
	api.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimit))

	// Protected URL operations
	api.POST("/batch-shorten", handler.BatchShortenURLs)
	api.GET("/my-urls", handler.GetUserURLs)
	api.DELETE("/my-urls/:shortCode", handler.DeleteURL)
	api.PUT("/my-urls/:shortCode", handler.UpdateURL)
	
	// Enhanced analytics for authenticated users - temporarily disabled
	api.GET("/analytics/:shortCode/trends", handler.GetClickTrends)
	// api.GET("/analytics/:shortCode/advanced", handler.AdvancedAnalyticsHandlers.GetAdvancedAnalytics)
	// api.GET("/analytics/:shortCode/geographic", handler.AdvancedAnalyticsHandlers.GetGeographicAnalytics)
	// api.GET("/analytics/:shortCode/time", handler.AdvancedAnalyticsHandlers.GetTimeAnalytics)
	// api.GET("/analytics/:shortCode/device", handler.AdvancedAnalyticsHandlers.GetDeviceAnalytics)
	// api.GET("/analytics/:shortCode/referrers", handler.AdvancedAnalyticsHandlers.GetReferrerAnalytics)
	// api.GET("/analytics/:shortCode/referrers/enhanced", handler.AdvancedAnalyticsHandlers.GetEnhancedReferrerAnalytics)
	// api.GET("/analytics/:shortCode/utm-campaigns", handler.AdvancedAnalyticsHandlers.GetUTMCampaignAnalytics)
	// api.GET("/analytics/:shortCode/referrers/insights", handler.AdvancedAnalyticsHandlers.GetReferrerInsights)
	// api.GET("/analytics/:shortCode/heatmap", handler.AdvancedAnalyticsHandlers.GetHeatmapAnalytics)
	// api.GET("/analytics/:shortCode/map", handler.AdvancedAnalyticsHandlers.GetMapAnalytics)
	api.GET("/dashboard/stats", handler.GetUserDashboardStats)
	
	// Billing and subscription routes
	setupBillingRoutes(api, handler)
	
	// Conversion tracking routes
	setupConversionTrackingRoutes(api, handler)
	
	// A/B testing routes
	setupABTestingRoutes(api, handler)
	
	// Attribution routes
	setupAttributionRoutes(api, handler)
	
	// Real-time analytics WebSocket and stats
	api.GET("/realtime/ws", handler.RealtimeHandlers.HandleWebSocket)
	api.GET("/realtime/stats", handler.RealtimeHandlers.GetRealtimeStats)
	
	// User analytics routes
	setupUserAnalyticsRoutes(api, handler)
}

// setupAdminRoutes configures routes that require admin access - TEMPORARILY DISABLED
/*
func setupAdminRoutes(router *gin.Engine, handler *handlers.Handler, authService *services.AuthService) {
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(authService))
	admin.Use(middleware.AdminMiddleware())
	admin.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimit))
	
	// Admin user management
	admin.GET("/users/:id", handler.AuthHandlers.GetUserByID)
	admin.GET("/users", func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "User list endpoint not implemented"})
	})

	// Admin analytics
	admin.GET("/stats", func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Admin stats endpoint not implemented"})
	})
	
	// CMS management routes
	cms := admin.Group("/cms")
	{
		cms.POST("/pages", handler.CMSHandlers.CreateStaticPage)
		cms.GET("/pages", handler.CMSHandlers.ListStaticPages)
		cms.GET("/pages/:id", handler.CMSHandlers.GetStaticPageByID)
		cms.PUT("/pages/:id", handler.CMSHandlers.UpdateStaticPage)
		cms.DELETE("/pages/:id", handler.CMSHandlers.DeleteStaticPage)
		cms.GET("/pages/:id/revisions", handler.CMSHandlers.GetPageRevisions)
		cms.GET("/pages/:id/analytics", handler.CMSHandlers.GetPageAnalytics)
	}
}
*/

// setupConversionTrackingRoutes configures conversion tracking routes
func setupConversionTrackingRoutes(api *gin.RouterGroup, handler *handlers.Handler) {
	conversion := api.Group("/conversions")
	
	// Conversion goal management
	conversion.POST("/goals", handler.ConversionHandlers.CreateConversionGoal)
	conversion.GET("/goals", handler.ConversionHandlers.GetConversionGoals)
	conversion.GET("/goals/:goalId", handler.ConversionHandlers.GetConversionGoal)
	conversion.PUT("/goals/:goalId", handler.ConversionHandlers.UpdateConversionGoal)
	conversion.DELETE("/goals/:goalId", handler.ConversionHandlers.DeleteConversionGoal)
	
	// Conversion tracking
	conversion.POST("/track", handler.ConversionHandlers.TrackConversion)
	conversion.GET("/goals/:goalId/stats", handler.ConversionHandlers.GetConversionStats)
	conversion.GET("/goals/:goalId/attribution", handler.ConversionHandlers.GetAttributionReport)
}

// setupABTestingRoutes configures A/B testing routes
func setupABTestingRoutes(api *gin.RouterGroup, handler *handlers.Handler) {
	abtest := api.Group("/ab-tests")
	
	// A/B test management
	abtest.POST("/", handler.ABTestHandlers.CreateABTest)
	abtest.GET("/", handler.ABTestHandlers.GetABTests)
	abtest.GET("/:testId", handler.ABTestHandlers.GetABTest)
	abtest.POST("/:testId/start", handler.ABTestHandlers.StartABTest)
	abtest.POST("/:testId/stop", handler.ABTestHandlers.StopABTest)
	
	// A/B test participation and results
	abtest.GET("/:testId/variant", handler.ABTestHandlers.GetABTestVariant)
	abtest.POST("/:testId/conversion", handler.ABTestHandlers.RecordConversion)
	abtest.GET("/:testId/results", handler.ABTestHandlers.GetABTestResults)
	
	// Enhanced statistical analysis
	abtest.GET("/:testId/sequential", handler.ABTestHandlers.GetSequentialTestResults)
	abtest.GET("/:testId/power", handler.ABTestHandlers.GetPowerAnalysis)
	abtest.GET("/sample-size-calculator", handler.ABTestHandlers.GetSampleSizeCalculator)
}

// setupAttributionRoutes configures attribution and multi-touch analytics routes
func setupAttributionRoutes(api *gin.RouterGroup, handler *handlers.Handler) {
	attribution := api.Group("/attribution")
	
	// Touchpoint management
	attribution.POST("/touchpoints", handler.AttributionHandlers.RecordTouchpoint)
	
	// Conversion journey analysis
	attribution.GET("/journey/:conversionId", handler.AttributionHandlers.GetConversionJourney)
	attribution.GET("/report/:conversionId", handler.AttributionHandlers.GetAttributionReport)
	attribution.POST("/calculate/:conversionId", handler.AttributionHandlers.CalculateAttribution)
	
	// Channel attribution analysis
	attribution.GET("/channels/:shortCode", handler.AttributionHandlers.GetChannelAttribution)
	attribution.GET("/compare/:shortCode", handler.AttributionHandlers.CompareAttributionModels)
	attribution.GET("/insights/:shortCode", handler.AttributionHandlers.GetAttributionInsights)
}

// setupBillingRoutes configures billing and subscription routes
func setupBillingRoutes(api *gin.RouterGroup, handler *handlers.Handler) {
	billing := api.Group("/billing")
	
	// Subscription management
	billing.GET("/subscription", handler.BillingHandlers.GetSubscription)
	billing.GET("/plans", handler.BillingHandlers.GetPlans)
	billing.GET("/usage", handler.BillingHandlers.GetUsage)
	
	// Checkout and payment
	billing.POST("/checkout", handler.BillingHandlers.CreateCheckoutSession)
	billing.POST("/cancel-subscription", handler.BillingHandlers.CancelSubscription)
	billing.POST("/reactivate-subscription", handler.BillingHandlers.ReactivateSubscription)
	billing.GET("/portal", handler.BillingHandlers.GetCustomerPortalURL)
	billing.GET("/products", handler.BillingHandlers.GetProducts)

	// Admin: Manual product sync (for debugging)
	billing.POST("/sync-products", handler.BillingHandlers.SyncProducts)

	// Payment methods
	billing.POST("/payment-method", handler.BillingHandlers.UpdatePaymentMethod)
	billing.GET("/history", handler.BillingHandlers.GetBillingHistory)

	// Webhooks (no auth required for Polar webhooks)
	api.POST("/webhooks/polar", handler.PolarWebhookHandler.HandleWebhook)
	
	// Feature-protected routes
	customDomain := api.Group("/custom-domains")
	customDomain.Use(handler.BillingHandlers.CheckFeatureAccess("custom_domain"))
	{
		customDomain.POST("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Custom domain management - coming soon"})
		})
		customDomain.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "List custom domains - coming soon"})
		})
	}
	
	// API key management routes - ENABLED for professional API integration
	// Removed billing feature gate - API keys available to all authenticated users
	api.POST("/api-keys", handler.APIKeyHandlers.CreateAPIKey)
	api.GET("/api-keys", handler.APIKeyHandlers.ListAPIKeys)
	api.GET("/api-keys/:id", handler.APIKeyHandlers.GetAPIKey)
	api.PUT("/api-keys/:id/revoke", handler.APIKeyHandlers.RevokeAPIKey)
	api.DELETE("/api-keys/:id", handler.APIKeyHandlers.DeleteAPIKey)
	api.GET("/api-keys/:id/stats", handler.APIKeyHandlers.GetAPIKeyStats)
}

// setupStaticRoutes configures static file serving for the frontend
func setupStaticRoutes(router *gin.Engine) {
	// Serve static files from the /static directory
	router.Static("/static", "./static")

	// Serve frontend assets (_next directory for Next.js)
	router.StaticFS("/_next", http.Dir("./static/_next"))

	// Serve index.html for root path
	rootHandler := func(c *gin.Context) {
		indexPath := filepath.Join("./static", "index.html")
		c.File(indexPath)
	}
	router.GET("/", rootHandler)
	router.HEAD("/", rootHandler)

	// All frontend routes that need static HTML
	frontendRoutes := []string{
		"/features", "/pricing", "/api-docs",
		"/about", "/contact", "/terms", "/privacy",
		"/dashboard", "/login", "/register",
		"/analytics", "/profile", "/settings",
		"/bulk", "/conversions", "/api-keys", "/admin",
		"/billing",
	}

	for _, route := range frontendRoutes {
		routePath := route // Local copy to avoid closure issues

		routeHandler := func(c *gin.Context) {
			// Try to serve the specific HTML file for this route (e.g., /features.html)
			htmlPath := filepath.Join("./static", routePath+".html")
			if _, err := os.Stat(htmlPath); err == nil {
				c.File(htmlPath)
				return
			}
			// Try nested index.html (e.g., /features/index.html)
			nestedPath := filepath.Join("./static", routePath, "index.html")
			if _, err := os.Stat(nestedPath); err == nil {
				c.File(nestedPath)
				return
			}
			// Fallback to root index.html for client-side routing
			indexPath := filepath.Join("./static", "index.html")
			c.File(indexPath)
		}
		router.GET(routePath, routeHandler)
		router.HEAD(routePath, routeHandler)

		// Handle sub-routes with /*path
		subRouteHandler := func(c *gin.Context) {
			indexPath := filepath.Join("./static", "index.html")
			c.File(indexPath)
		}
		router.GET(routePath+"/*path", subRouteHandler)
		router.HEAD(routePath+"/*path", subRouteHandler)
	}

	// Catch-all route handler (MUST be registered LAST)
	router.NoRoute(func(c *gin.Context) {
		// Only handle non-API routes
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") &&
		   !strings.HasPrefix(c.Request.URL.Path, "/health") {
			// Check if it's a short code (single segment, no extension)
			pathSegments := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")
			if len(pathSegments) == 1 && !strings.Contains(pathSegments[0], ".") {
				// Let short code 404 naturally
				c.JSON(404, gin.H{"error": "Short URL not found"})
				return
			}

			// Serve index.html for frontend routes (SPA fallback)
			indexPath := filepath.Join("./static", "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				c.File(indexPath)
				return
			}

			c.JSON(404, gin.H{"error": "Page not found"})
		} else {
			// API routes get JSON 404
			c.JSON(404, gin.H{"error": "API endpoint not found"})
		}
	})
}

// setupUserAnalyticsRoutes configures user analytics routes
func setupUserAnalyticsRoutes(api *gin.RouterGroup, handler *handlers.Handler) {
	analytics := api.Group("/user-analytics")
	
	// User analytics summary and metrics
	analytics.GET("/summary", handler.AnalyticsHandlers.GetUserAnalyticsSummary)
	analytics.GET("/engagement", handler.AnalyticsHandlers.GetUserEngagementMetrics)
	analytics.GET("/dashboard/:period", handler.AnalyticsHandlers.GetDashboardAnalytics)
	
	// Activity logging and tracking
	analytics.GET("/activity-log", handler.AnalyticsHandlers.GetUserActivityLog)
	analytics.POST("/log-activity", handler.AnalyticsHandlers.LogUserActivity)
	
	// Session management
	analytics.POST("/start-session", handler.AnalyticsHandlers.StartSession)
	analytics.POST("/end-session", handler.AnalyticsHandlers.EndSession)
}