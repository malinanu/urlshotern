package routes

import (
	"net/http"
	"path/filepath"

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
	
	// Health check
	router.GET("/health", handler.HealthCheck)
	
	// Public routes (no authentication required)
	setupPublicRoutes(router, handler, authMiddleware)

	// Authentication routes
	setupAuthRoutes(router, handler.AuthHandlers, authMiddleware)

	// Protected routes (authentication required)
	setupProtectedRoutes(router, handler, authMiddleware)
	
	// Admin routes (admin access required) - temporarily disabled
	// setupAdminRoutes(router, handler, authService)
	
	// Static file serving for frontend (must be last to not interfere with API routes)
	setupStaticRoutes(router)
}

// setupPublicRoutes configures public routes
func setupPublicRoutes(router *gin.Engine, handler *handlers.Handler, authMiddleware *middleware.AuthMiddleware) {
	// URL shortening (with optional auth) - temporarily disabled auth middleware
	router.POST("/api/v1/shorten",
		// middleware.OptionalAuthMiddleware(authService),
		middleware.RateLimitMiddleware(middleware.DefaultRateLimit),
		handler.ShortenURL,
	)
	
	// URL redirection (no auth required)
	router.GET("/:shortCode", handler.RedirectURL)
	
	// Public analytics (if URL is public)
	router.GET("/api/v1/analytics/:shortCode", 
		middleware.RateLimitMiddleware(middleware.DefaultRateLimit),
		handler.GetAnalytics,
	)
	
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
	// protected.Use(authMiddleware.RequireAuth())  // Temporarily disabled
	{
		protected.GET("/profile", authHandlers.GetProfile)
		protected.PUT("/profile", authHandlers.UpdateProfile)
		protected.POST("/change-password", authHandlers.ChangePassword)
		protected.POST("/logout", authHandlers.Logout)
		protected.POST("/send-otp", authHandlers.SendOTP) // For authenticated users
	}
}

// setupProtectedRoutes configures routes that require authentication
func setupProtectedRoutes(router *gin.Engine, handler *handlers.Handler, authMiddleware *middleware.AuthMiddleware) {
	api := router.Group("/api/v1")
	// api.Use(authMiddleware.RequireAuth())  // Temporarily disabled due to auth setup issues
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
		// TODO: Implement user list
		c.JSON(200, gin.H{"message": "Admin user list - coming soon"})
	})
	
	// Admin analytics
	admin.GET("/stats", func(c *gin.Context) {
		// TODO: Implement admin stats
		c.JSON(200, gin.H{"message": "Admin stats - coming soon"})
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
	
	// Payment methods
	billing.POST("/payment-method", handler.BillingHandlers.UpdatePaymentMethod)
	billing.GET("/history", handler.BillingHandlers.GetBillingHistory)
	
	// Webhooks (no auth required for Stripe webhooks)
	api.POST("/webhooks/stripe", handler.BillingHandlers.StripeWebhook)
	
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
	
	// API access protected routes - temporarily disabled
	/*
	apiAccess := api.Group("/api-keys")
	apiAccess.Use(handler.BillingHandlers.CheckFeatureAccess("api_access"))
	{
		apiAccess.POST("/", handler.CMSHandlers.CreateAPIKey)
		apiAccess.GET("/", handler.CMSHandlers.ListAPIKeys)
		apiAccess.GET("/:id", handler.CMSHandlers.GetAPIKey)
		apiAccess.PUT("/:id/revoke", handler.CMSHandlers.RevokeAPIKey)
		apiAccess.DELETE("/:id", handler.CMSHandlers.DeleteAPIKey)
		apiAccess.GET("/:id/stats", handler.CMSHandlers.GetAPIKeyStats)
	}
	*/
}

// setupStaticRoutes configures static file serving for the frontend
func setupStaticRoutes(router *gin.Engine) {
	// Serve static files from the /static directory
	router.Static("/static", "./static")
	
	// Serve frontend assets (_next directory for Next.js)
	router.StaticFS("/_next", http.Dir("./static/_next"))
	
	// Serve index.html for root path
	router.GET("/", func(c *gin.Context) {
		indexPath := filepath.Join("./static", "index.html")
		c.File(indexPath)
	})
	
	// Serve index.html for common frontend routes
	frontendRoutes := []string{"/dashboard", "/login", "/register", "/analytics", "/profile", "/settings"}
	for _, route := range frontendRoutes {
		router.GET(route, func(c *gin.Context) {
			indexPath := filepath.Join("./static", "index.html")
			c.File(indexPath)
		})
		// Also handle sub-routes with /*path
		router.GET(route+"/*path", func(c *gin.Context) {
			indexPath := filepath.Join("./static", "index.html")
			c.File(indexPath)
		})
	}
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