package routes

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	router *gin.Engine,
	handler *handlers.Handler,
	authMiddleware *middleware.AuthMiddleware,
	config *configs.Config,
) {
	// Add global middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware())

	// Custom domain middleware — detects requests via custom hostnames
	baseHost := extractBaseHost(config.BaseURL)
	router.Use(func(c *gin.Context) {
		host := c.Request.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host != baseHost && host != "localhost" && host != "" {
			c.Set("custom_domain_host", host)
		}
		c.Next()
	})

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

	// Email service health check (for monitoring)
	router.GET("/health/email", handler.AuthHandlers.EmailHealthCheck)

	// Serve common static files BEFORE /:shortCode route to prevent them from being treated as short codes
	commonStaticFiles := []string{
		"/manifest.json", "/robots.txt", "/sitemap.xml",
		"/favicon.ico", "/sw.js", "/apple-touch-icon.png",
		"/google26c27f314cce952a.html",
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

	// Admin routes (admin access required) - controlled by environment variable
	if config.EnableAdminRoutes {
		setupAdminRoutes(router, handler, authMiddleware)
	}

	// Static file serving for frontend (must be last to not interfere with API routes)
	setupStaticRoutes(router)
}

// setupPublicRoutes configures public routes
func setupPublicRoutes(router *gin.Engine, handler *handlers.Handler, authMiddleware *middleware.AuthMiddleware) {
	// Get API key service from handler
	apiKeyService := handler.APIKeyHandlers.GetAPIKeyService()

	// URL shortening (anonymous users get 5 free URLs, authenticated users use billing limits)
	// Password check for protected links (public — no auth required)
	router.POST("/api/v1/check-password/:shortCode", handler.CheckPassword)

	router.POST("/api/v1/shorten",
		authMiddleware.OptionalAuth(),
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

	// Public CMS pages
	router.GET("/api/v1/pages/:slug", handler.CMSHandlers.GetStaticPage)

	// Contact form
	router.POST("/api/v1/contact",
		middleware.RateLimitMiddleware(middleware.DefaultRateLimit),
		handler.ContactHandlers.SubmitContactForm,
	)

	// Public site settings (for frontend SEO metadata generation)
	if handler.GlobalSettingsHandlers != nil {
		router.GET("/api/v1/public/settings", handler.GlobalSettingsHandlers.GetPublicSettings)
	}
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(router *gin.Engine, authHandlers *handlers.AuthHandlers, authMiddleware *middleware.AuthMiddleware) {
	auth := router.Group("/api/v1/auth")
	
	// Apply auth-specific rate limiting
	auth.Use(middleware.RateLimitMiddleware(middleware.AuthRateLimit))
	
	// Google OAuth (before rate limit since these are redirects)
	auth.GET("/google", authHandlers.GoogleLogin)
	auth.GET("/google/callback", authHandlers.GoogleCallback)

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

	// Password reset routes
	auth.POST("/forgot-password", authHandlers.ForgotPassword)
	auth.POST("/reset-password", authHandlers.ResetPassword)

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
		protected.DELETE("/delete-account", authHandlers.DeleteAccount)
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
	
	// Enhanced analytics for authenticated users
	billing := handler.GetBillingService()
	api.GET("/analytics/:shortCode/trends", handler.GetClickTrends)
	api.GET("/analytics/:shortCode/advanced",           handler.AdvancedAnalyticsHandlers.GetAdvancedAnalytics)
	api.GET("/analytics/:shortCode/device",             handler.AdvancedAnalyticsHandlers.GetDeviceAnalytics)
	api.GET("/analytics/:shortCode/referrers",          handler.AdvancedAnalyticsHandlers.GetReferrerAnalytics)
	// Pro+ features
	api.GET("/analytics/:shortCode/geographic",
		middleware.CheckPlanFeature(billing, middleware.FeatureCityGeo),
		handler.AdvancedAnalyticsHandlers.GetGeographicAnalytics)
	api.GET("/analytics/:shortCode/time",
		middleware.CheckPlanFeature(billing, middleware.FeatureTimeSeries),
		handler.AdvancedAnalyticsHandlers.GetTimeAnalytics)
	api.GET("/analytics/:shortCode/heatmap",
		middleware.CheckPlanFeature(billing, middleware.FeatureHeatmap),
		handler.AdvancedAnalyticsHandlers.GetHeatmapAnalytics)
	api.GET("/analytics/:shortCode/utm-campaigns",
		middleware.CheckPlanFeature(billing, middleware.FeatureUTM),
		handler.AdvancedAnalyticsHandlers.GetUTMCampaignAnalytics)
	api.GET("/analytics/:shortCode/referrers/enhanced",
		middleware.CheckPlanFeature(billing, middleware.FeatureUTM),
		handler.AdvancedAnalyticsHandlers.GetEnhancedReferrerAnalytics)
	api.GET("/analytics/:shortCode/referrers/insights",
		middleware.CheckPlanFeature(billing, middleware.FeatureUTM),
		handler.AdvancedAnalyticsHandlers.GetReferrerInsights)
	api.GET("/analytics/:shortCode/map",
		middleware.CheckPlanFeature(billing, middleware.FeatureLiveMap),
		handler.AdvancedAnalyticsHandlers.GetMapAnalytics)
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

	// OG metadata for individual URLs (authenticated)
	api.PATCH("/urls/:shortCode/og", handler.UpdateURLOGMetadata)

	// White-label branding (enterprise only)
	api.GET("/branding",
		middleware.CheckPlanFeature(billing, middleware.FeatureBranding),
		handler.BrandingHandlers.GetBranding)
	api.PUT("/branding",
		middleware.CheckPlanFeature(billing, middleware.FeatureBranding),
		handler.BrandingHandlers.UpdateBranding)

	// Ultra (enterprise) — AI Insights
	api.GET("/insights/:shortCode/summary",
		middleware.CheckPlanFeature(billing, middleware.FeatureAISummary),
		handler.AIInsightsHandlers.GetSummary)
	api.GET("/insights/:shortCode/anomalies",
		middleware.CheckPlanFeature(billing, middleware.FeatureAIAnomaly),
		handler.AIInsightsHandlers.GetAnomalies)
	api.GET("/insights/:shortCode/recommendations",
		middleware.CheckPlanFeature(billing, middleware.FeatureAIRecommend),
		handler.AIInsightsHandlers.GetRecommendations)
	api.GET("/insights/:shortCode/forecast",
		middleware.CheckPlanFeature(billing, middleware.FeatureForecast),
		handler.AIInsightsHandlers.GetForecast)
	api.GET("/insights/:shortCode/audience",
		middleware.CheckPlanFeature(billing, middleware.FeatureAudience),
		handler.AIInsightsHandlers.GetAudienceProfile)
}

// setupAdminRoutes configures routes that require admin access
func setupAdminRoutes(router *gin.Engine, handler *handlers.Handler, authMiddleware *middleware.AuthMiddleware) {
	admin := router.Group("/api/v1/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireAdmin())
	admin.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimit))

	// Admin analytics
	admin.GET("/stats", handler.AuthHandlers.GetAdminStats)
	admin.GET("/activity", handler.AuthHandlers.GetRecentActivity)

	// Admin user management (existing)
	admin.GET("/users", handler.AuthHandlers.GetUserList)
	admin.GET("/users/:id", handler.AuthHandlers.GetUserByID)
	admin.PUT("/users/:id/status", handler.AuthHandlers.ToggleUserStatus)
	admin.DELETE("/users/:id", handler.AuthHandlers.DeleteUserAdmin)

	// ============================================================================
	// RBAC Admin User Management (NEW)
	// ============================================================================
	if handler.AdminRBACHandlers != nil {
		// User creation and role management
		admin.POST("/users/create",
			authMiddleware.RequirePermissionMiddleware(models.PermissionUsersCreate),
			handler.AdminRBACHandlers.CreateUserAdmin)

		admin.POST("/users/:id/roles/assign",
			authMiddleware.RequirePermissionMiddleware(models.PermissionUsersRolesAssign),
			authMiddleware.RequireCanAssignRole(),
			handler.AdminRBACHandlers.AssignUserRole)

		admin.DELETE("/users/:id/roles/:roleId",
			authMiddleware.RequirePermissionMiddleware(models.PermissionUsersRolesRemove),
			handler.AdminRBACHandlers.RemoveUserRole)

		admin.GET("/users/:id/roles",
			authMiddleware.RequirePermissionMiddleware(models.PermissionUsersRead),
			handler.AdminRBACHandlers.GetUserRoles)

		admin.PUT("/users/:id/account-type",
			authMiddleware.RequirePermissionMiddleware(models.PermissionUsersUpdate),
			handler.AdminRBACHandlers.UpdateAccountType)

		// RBAC Management
		rbac := admin.Group("/rbac")
		{
			// Roles
			rbac.GET("/roles",
				authMiddleware.RequirePermissionMiddleware(models.PermissionRBACRolesRead),
				handler.AdminRBACHandlers.ListRoles)

			rbac.GET("/assignable-roles",
				authMiddleware.RequirePermissionMiddleware(models.PermissionRBACRolesRead),
				handler.AdminRBACHandlers.GetAssignableRoles)

			rbac.GET("/roles/:id",
				authMiddleware.RequirePermissionMiddleware(models.PermissionRBACRolesRead),
				handler.AdminRBACHandlers.GetRole)

			rbac.POST("/roles",
				authMiddleware.RequirePermissionMiddleware(models.PermissionRBACRolesCreate),
				handler.AdminRBACHandlers.CreateCustomRole)

			rbac.PUT("/roles/:id",
				authMiddleware.RequirePermissionMiddleware(models.PermissionRBACRolesUpdate),
				handler.AdminRBACHandlers.UpdateRole)

			rbac.DELETE("/roles/:id",
				authMiddleware.RequirePermissionMiddleware(models.PermissionRBACRolesDelete),
				handler.AdminRBACHandlers.DeleteRole)

			// Permissions
			rbac.GET("/permissions",
				authMiddleware.RequirePermissionMiddleware(models.PermissionRBACPermRead),
				handler.AdminRBACHandlers.ListPermissions)

			// Audit Log
			rbac.GET("/audit",
				authMiddleware.RequirePermissionMiddleware(models.PermissionRBACPermRead),
				handler.AdminRBACHandlers.GetRBACAuditLog)
		}
	}

	// ============================================================================
	// Admin Coupons (NEW)
	// ============================================================================
	if handler.AdminCouponHandlers != nil {
		coupons := admin.Group("/coupons")
		{
			coupons.GET("", handler.AdminCouponHandlers.ListCoupons)
			coupons.POST("", handler.AdminCouponHandlers.CreateCoupon)
			coupons.DELETE("/:id", handler.AdminCouponHandlers.DeleteCoupon)
			coupons.PUT("/:id/toggle", handler.AdminCouponHandlers.ToggleCoupon)
			coupons.POST("/:id/sync-polar", handler.AdminCouponHandlers.SyncToPolar)
		}
	}

	// ============================================================================
	// CMS Routes (UPDATED with permission-based middleware)
	// ============================================================================
	cms := admin.Group("/cms")
	{
		// Pages CRUD - permission-based access control
		cms.GET("/pages",
			authMiddleware.RequirePermissionMiddleware(models.PermissionCMSPagesRead),
			handler.CMSHandlers.ListStaticPages)

		cms.POST("/pages",
			authMiddleware.RequirePermissionMiddleware(models.PermissionCMSPagesCreate),
			handler.CMSHandlers.CreateStaticPage)

		cms.PUT("/pages/:id",
			authMiddleware.RequirePermissionMiddleware(models.PermissionCMSPagesUpdate),
			handler.CMSHandlers.UpdateStaticPage)

		cms.GET("/pages/:id",
			authMiddleware.RequirePermissionMiddleware(models.PermissionCMSPagesRead),
			handler.CMSHandlers.GetStaticPageByID)

		cms.DELETE("/pages/:id",
			authMiddleware.RequirePermissionMiddleware(models.PermissionCMSPagesDelete),
			handler.CMSHandlers.DeleteStaticPage)

		// NEW: Publish/Unpublish endpoint
		cms.PUT("/pages/:id/publish",
			authMiddleware.RequirePermissionMiddleware(models.PermissionCMSPagesPublish),
			handler.CMSHandlers.PublishStaticPage)

		// Blocks - permission-based access control
		cms.GET("/blocks",
			authMiddleware.RequirePermissionMiddleware(models.PermissionCMSBlocksRead),
			handler.AuthHandlers.GetCMSBlocks)

		// Templates
		cms.GET("/templates",
			authMiddleware.RequirePermissionMiddleware(models.PermissionCMSPagesRead),
			handler.AuthHandlers.GetCMSTemplates)
	}

	// Content blocks CRUD endpoints (for admin panel)
	admin.GET("/content-blocks", handler.AuthHandlers.GetCMSBlocks)
	admin.POST("/content-blocks", handler.AuthHandlers.CreateCMSBlock)
	admin.PUT("/content-blocks/:id", handler.AuthHandlers.UpdateCMSBlock)
	admin.DELETE("/content-blocks/:id", handler.AuthHandlers.DeleteCMSBlock)
	admin.GET("/content-blocks/:id/render", handler.AuthHandlers.RenderCMSBlock)

	// Page templates CRUD endpoints (for admin panel)
	admin.GET("/page-templates", handler.AuthHandlers.ListAdminPageTemplates)
	admin.POST("/page-templates", handler.AuthHandlers.CreateAdminPageTemplate)
	admin.GET("/page-templates/:id", handler.AuthHandlers.GetAdminPageTemplate)
	admin.PUT("/page-templates/:id", handler.AuthHandlers.UpdateAdminPageTemplate)
	admin.DELETE("/page-templates/:id", handler.AuthHandlers.DeleteAdminPageTemplate)
	admin.GET("/page-templates/:id/render", handler.AuthHandlers.RenderAdminPageTemplate)

	// SEO routes
	seo := admin.Group("/seo")
	{
		// Existing SEO monitoring endpoints
		seo.GET("/settings", handler.AuthHandlers.GetSEOSettings)
		seo.GET("/metrics", handler.AuthHandlers.GetSEOMetrics)
		seo.GET("/recommendations", handler.AuthHandlers.GetSEORecommendations)

		// SEO Analysis endpoints
		seo.GET("/analyses", handler.AuthHandlers.GetSEOAnalyses)
		seo.POST("/analyze", handler.AuthHandlers.AnalyzeSEOURL)

		// Sitemap endpoints
		seo.GET("/sitemap", handler.AuthHandlers.GetSitemap)
		seo.POST("/sitemap/generate", handler.AuthHandlers.GenerateSitemap)

		// Meta tags endpoints
		seo.GET("/meta-tags", handler.AuthHandlers.GetMetaTags)
		seo.POST("/meta-tags", handler.AuthHandlers.CreateMetaTag)
		seo.DELETE("/meta-tags/:id", handler.AuthHandlers.DeleteMetaTag)
	}

	// Custom domain admin routes
	admin.GET("/domains", handler.DomainHandlers.GetAllDomains)
	admin.GET("/domains/stats", handler.DomainHandlers.GetDomainStats)

	// Site-wide settings (SEO, GA, etc.)
	if handler.GlobalSettingsHandlers != nil {
		admin.GET("/settings", handler.GlobalSettingsHandlers.GetSettings)
		admin.PUT("/settings", handler.GlobalSettingsHandlers.BulkUpdateSettings)
		admin.GET("/settings/:key", handler.GlobalSettingsHandlers.GetSetting)
		admin.PUT("/settings/:key", handler.GlobalSettingsHandlers.UpdateSetting)
	}
}

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
	
	// Discount Validation
	billing.GET("/coupons/validate", handler.BillingHandlers.ValidateCouponHandler)
	
	// Checkout and payment
	billing.POST("/checkout", handler.BillingHandlers.CreateCheckoutSession)
	billing.POST("/cancel-subscription", handler.BillingHandlers.CancelSubscription)
	billing.POST("/reactivate-subscription", handler.BillingHandlers.ReactivateSubscription)
	billing.GET("/portal", handler.BillingHandlers.GetCustomerPortalURL)
	billing.GET("/products", handler.BillingHandlers.GetProducts)

	// Admin: Manual product sync (for debugging)
	billing.POST("/sync-products", handler.BillingHandlers.SyncProducts)

	// Sync subscription data from Polar (fixes stale billing_cycle after monthly→yearly upgrade)
	billing.POST("/refresh-subscription", handler.BillingHandlers.RefreshSubscription)

	// Payment methods
	billing.POST("/payment-method", handler.BillingHandlers.UpdatePaymentMethod)
	billing.GET("/history", handler.BillingHandlers.GetBillingHistory)

	// Webhooks (no auth required for Polar webhooks)
	api.POST("/webhooks/polar", handler.PolarWebhookHandler.HandleWebhook)
	
	// Custom domain management (requires custom_domain feature)
	customDomain := api.Group("/custom-domains")
	customDomain.Use(handler.BillingHandlers.CheckFeatureAccess("custom_domain"))
	{
		customDomain.POST("/", handler.DomainHandlers.CreateDomain)
		customDomain.GET("/", handler.DomainHandlers.GetUserDomains)
		customDomain.GET("/usage", handler.DomainHandlers.GetDomainUsage)
		customDomain.GET("/:id", handler.DomainHandlers.GetDomain)
		customDomain.PUT("/:id", handler.DomainHandlers.UpdateDomain)
		customDomain.DELETE("/:id", handler.DomainHandlers.DeleteDomain)
		customDomain.GET("/:id/verification", handler.DomainHandlers.GetDomainVerification)
		customDomain.POST("/:id/verify", handler.DomainHandlers.VerifyDomain)
		customDomain.GET("/:id/settings", handler.DomainHandlers.GetDomainSettings)
		customDomain.GET("/:id/status", handler.DomainHandlers.GetDomainStatus)
	}

	// Team collaboration routes (Pro+ feature)
	teams := api.Group("/teams")
	teams.Use(handler.BillingHandlers.CheckFeatureAccess("custom_domain")) // Pro+ gate (same tier as custom domains)
	{
		teams.POST("/", handler.TeamHandlers.CreateTeam)
		teams.GET("/", handler.TeamHandlers.ListUserTeams)
		teams.GET("/:id", handler.TeamHandlers.GetTeam)
		teams.PUT("/:id", handler.TeamHandlers.UpdateTeam)
		teams.DELETE("/:id", handler.TeamHandlers.DeleteTeam)
		teams.POST("/:id/invite", handler.TeamHandlers.InviteTeamMember)
		teams.GET("/:id/invitations", handler.TeamHandlers.ListTeamInvitations)
		teams.DELETE("/:id/invitations/:invitationID", handler.TeamHandlers.CancelTeamInvitation)
		teams.PUT("/:id/members/:memberID", handler.TeamHandlers.UpdateTeamMember)
		teams.DELETE("/:id/members/:memberID", handler.TeamHandlers.RemoveTeamMember)
		teams.POST("/:id/leave", handler.TeamHandlers.LeaveTeam)
		teams.POST("/:id/transfer-ownership", handler.TeamHandlers.TransferTeamOwnership)
		teams.GET("/:id/links", handler.TeamHandlers.GetTeamLinks)
	}
	// Team invitation acceptance (public — token-based)
	api.POST("/teams/invitations/accept", handler.TeamHandlers.AcceptTeamInvitation)
	api.POST("/teams/invitations/reject", handler.TeamHandlers.RejectTeamInvitation)

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
		"/billing", "/pages", "/custom-domains",
		"/forgot-password", "/reset-password",
		"/teams", "/ab-tests",
		"/team-invite",
		"/not-found",
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
			subPath := c.Param("path")
			fullPath := routePath + subPath

			// Try to serve the specific HTML file for this sub-route
			htmlPath := filepath.Join("./static", fullPath+".html")
			if _, err := os.Stat(htmlPath); err == nil {
				c.File(htmlPath)
				return
			}

			// Try nested index.html
			nestedPath := filepath.Join("./static", fullPath, "index.html")
			if _, err := os.Stat(nestedPath); err == nil {
				c.File(nestedPath)
				return
			}

			// Try route's main HTML file (e.g. /admin.html)
			routeHtmlPath := filepath.Join("./static", routePath+".html")
			if _, err := os.Stat(routeHtmlPath); err == nil {
				c.File(routeHtmlPath)
				return
			}

			// Fallback to root index.html
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
				if strings.Contains(c.GetHeader("Accept"), "text/html") {
					c.Redirect(http.StatusFound, "/not-found")
					return
				}
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

// extractBaseHost parses a base URL and returns just the hostname (without port).
func extractBaseHost(baseURL string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	return parsed.Hostname()
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