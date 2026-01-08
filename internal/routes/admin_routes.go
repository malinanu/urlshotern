package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
)

// SetupAdminRoutes configures all admin routes with proper authentication and audit logging
func SetupAdminRoutes(
	router *gin.Engine,
	handler *handlers.Handler,
	authMiddleware *middleware.AuthMiddleware,
	adminMiddleware *middleware.AdminMiddleware,
) {
	// Admin routes group
	admin := router.Group("/api/v1/admin")

	// Apply middleware stack
	admin.Use(authMiddleware.RequireAuth())                    // Must be authenticated
	admin.Use(adminMiddleware.RequireAdminAccountType())       // Must have admin account type
	admin.Use(adminMiddleware.AdminDBContext())                // Use admin database connection
	admin.Use(adminMiddleware.AdminSessionTimeout())           // Enforce shorter session timeout
	admin.Use(adminMiddleware.AdminRateLimitByAction())        // Admin-specific rate limiting
	admin.Use(adminMiddleware.AdminAccessLog())                // Log all admin panel access

	// ================================================================
	// ADMIN DASHBOARD
	// ================================================================
	// Root endpoint returns dashboard overview
	admin.GET("", handler.AdminHandlers.DashboardHandlers.GetDashboardOverview)

	// Additional dashboard stats endpoint
	admin.GET("/stats", handler.AdminHandlers.DashboardHandlers.AdminDashboardStats)

	// ================================================================
	// ADMIN USER MANAGEMENT
	// ================================================================
	users := admin.Group("/users")
	users.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminUserView))
	{
		// List and search users
		users.GET("",
			adminMiddleware.AuditLog(models.AuditActionUserView, models.ResourceTypeUser),
			handler.AdminHandlers.UserHandlers.AdminUserList)

		// Get specific user details
		users.GET("/:id",
			adminMiddleware.AuditLog(models.AuditActionUserView, models.ResourceTypeUser),
			handler.AdminHandlers.UserHandlers.AdminGetUser)

		// User management operations (require manage permission)
		usersManage := users.Group("")
		usersManage.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminUserManage))
		{
			// Update user
			usersManage.PUT("/:id",
				adminMiddleware.PreventSelfAction(),
				adminMiddleware.AuditLog(models.AuditActionUserUpdate, models.ResourceTypeUser),
				handler.AdminHandlers.UserHandlers.AdminUpdateUser)

			// Ban user
			usersManage.POST("/:id/ban",
				adminMiddleware.PreventSelfAction(),
				adminMiddleware.RequireSuperAdmin(), // Only super admin can ban
				adminMiddleware.AuditLog(models.AuditActionUserBan, models.ResourceTypeUser),
				handler.AdminHandlers.UserHandlers.AdminBanUser)

			// Unban user
			usersManage.DELETE("/:id/ban",
				adminMiddleware.RequireSuperAdmin(),
				adminMiddleware.AuditLog(models.AuditActionUserUnban, models.ResourceTypeUser),
				handler.AdminHandlers.UserHandlers.AdminUnbanUser)

			// Force email verification
			usersManage.PUT("/:id/verify",
				adminMiddleware.AuditLog(models.AuditActionUserVerify, models.ResourceTypeUser),
				handler.AdminHandlers.UserHandlers.AdminVerifyUser)

			// Force password reset
			usersManage.POST("/:id/reset-password",
				adminMiddleware.PreventSelfAction(),
				adminMiddleware.AuditLog(models.AuditActionUserPasswordReset, models.ResourceTypeUser),
				handler.AdminHandlers.UserHandlers.AdminResetUserPassword)

			// Delete user (soft delete)
			usersManage.DELETE("/:id",
				adminMiddleware.PreventSelfAction(),
				adminMiddleware.RequireSuperAdmin(), // Only super admin can delete
				adminMiddleware.AuditLog(models.AuditActionUserDelete, models.ResourceTypeUser),
				handler.AdminHandlers.UserHandlers.AdminDeleteUser)

			// Impersonate user (for support)
			usersManage.POST("/:id/impersonate",
				adminMiddleware.PreventSelfAction(),
				adminMiddleware.RequireAdminPermission(models.PermissionAdminUserImpersonate),
				adminMiddleware.AuditLog(models.AuditActionUserImpersonate, models.ResourceTypeUser),
				handler.AdminHandlers.UserHandlers.AdminImpersonateUser)

			// Update user limits
			usersManage.PUT("/:id/limits",
				adminMiddleware.AuditLog(models.AuditActionUserUpdate, models.ResourceTypeUser),
				handler.AdminHandlers.UserHandlers.AdminUpdateUserLimits)

			// Change user subscription
			usersManage.PUT("/:id/plan",
				adminMiddleware.RequireAdminPermission(models.PermissionAdminBillingManage),
				adminMiddleware.AuditLog(models.AuditActionSubscriptionUpdate, models.ResourceTypeSubscription),
				handler.AdminHandlers.UserHandlers.AdminUpdateUserPlan)

			// Send notification to user
			usersManage.POST("/:id/notify",
				adminMiddleware.AuditLog("user.notify", models.ResourceTypeUser),
				handler.AdminHandlers.UserHandlers.AdminNotifyUser)
		}

		// User activity and history
		users.GET("/:id/activity",
			adminMiddleware.AuditLog(models.AuditActionUserView, models.ResourceTypeUser),
			handler.AdminHandlers.UserHandlers.AdminGetUserActivity)

		users.GET("/:id/urls",
			adminMiddleware.AuditLog(models.AuditActionURLView, models.ResourceTypeURL),
			handler.AdminHandlers.UserHandlers.AdminGetUserURLs)

		// Bulk operations (super admin only)
		users.POST("/bulk-export",
			adminMiddleware.RequireSuperAdmin(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsExport, "user_export"),
			handler.AdminHandlers.UserHandlers.AdminExportUsers)
	}

	// ================================================================
	// ADMIN URL MANAGEMENT
	// ================================================================
	urls := admin.Group("/urls")
	urls.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminURLView))
	{
		// List all URLs
		urls.GET("",
			adminMiddleware.AuditLog(models.AuditActionURLView, models.ResourceTypeURL),
			handler.AdminHandlers.URLHandlers.AdminListURLs)

		// Get specific URL details
		urls.GET("/:shortCode",
			adminMiddleware.AuditLog(models.AuditActionURLView, models.ResourceTypeURL),
			handler.AdminHandlers.URLHandlers.AdminGetURL)

		// URL management operations (require manage permission)
		urlsManage := urls.Group("")
		urlsManage.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminURLManage))
		{
			// Update URL status
			urlsManage.PUT("/:shortCode/status",
				adminMiddleware.AuditLog(models.AuditActionURLUpdate, models.ResourceTypeURL),
				handler.AdminHandlers.URLHandlers.AdminUpdateURLStatus)

			// Delete URL
			urlsManage.DELETE("/:shortCode",
				adminMiddleware.AuditLog(models.AuditActionURLDelete, models.ResourceTypeURL),
				handler.AdminHandlers.URLHandlers.AdminDeleteURL)

			// Transfer URL ownership
			urlsManage.PUT("/:shortCode/owner",
				adminMiddleware.RequireSystemAdmin(),
				adminMiddleware.AuditLog(models.AuditActionURLTransfer, models.ResourceTypeURL),
				handler.AdminHandlers.URLHandlers.AdminTransferURL)

			// Extend URL expiration
			urlsManage.PUT("/:shortCode/expiration",
				adminMiddleware.AuditLog(models.AuditActionURLUpdate, models.ResourceTypeURL),
				handler.AdminHandlers.URLHandlers.AdminExtendExpiration)
		}

		// Flagged URLs management
		urls.GET("/flagged",
			adminMiddleware.AuditLog("url.flag.view", models.ResourceTypeURL),
			handler.AdminHandlers.URLHandlers.AdminGetFlaggedURLs)

		urls.POST("/:shortCode/flag",
			adminMiddleware.RequireAdminPermission(models.PermissionAdminURLFlag),
			adminMiddleware.AuditLog(models.AuditActionURLFlag, models.ResourceTypeURL),
			handler.AdminHandlers.URLHandlers.AdminFlagURL)

		urls.DELETE("/:shortCode/flag",
			adminMiddleware.RequireAdminPermission(models.PermissionAdminURLManage),
			adminMiddleware.AuditLog(models.AuditActionURLUnflag, models.ResourceTypeURL),
			handler.AdminHandlers.URLHandlers.AdminUnflagURL)
	}

	// ================================================================
	// DOMAIN MANAGEMENT
	// ================================================================
	domains := admin.Group("/domains")
	domains.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminURLManage))
	{
		// List blocked domains
		domains.GET("/blocked",
			adminMiddleware.AuditLog("domain.blocked.view", models.ResourceTypeDomain),
			handler.AdminHandlers.DomainHandlers.AdminGetBlockedDomains)

		// Block domain
		domains.POST("/block",
			adminMiddleware.RequireSystemAdmin(),
			adminMiddleware.AuditLog(models.AuditActionDomainBlock, models.ResourceTypeDomain),
			handler.AdminHandlers.DomainHandlers.AdminBlockDomain)

		// Unblock domain
		domains.DELETE("/block/:domain",
			adminMiddleware.RequireSystemAdmin(),
			adminMiddleware.AuditLog(models.AuditActionDomainUnblock, models.ResourceTypeDomain),
			handler.AdminHandlers.DomainHandlers.AdminUnblockDomain)
	}

	// ================================================================
	// ANALYTICS & REPORTING
	// ================================================================
	analytics := admin.Group("/analytics")
	analytics.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminAnalyticsView))
	{
		// Real-time dashboard
		analytics.GET("/dashboard",
			adminMiddleware.ReadOnlyDBContext(), // Use read-only connection for analytics
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "dashboard"),
			handler.AdminHandlers.AnalyticsHandlers.AdminGetDashboard)

		// User analytics
		analytics.GET("/users",
			adminMiddleware.ReadOnlyDBContext(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "users"),
			handler.AdminHandlers.AnalyticsHandlers.AdminGetUserAnalytics)

		// URL analytics
		analytics.GET("/urls",
			adminMiddleware.ReadOnlyDBContext(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "urls"),
			handler.AdminHandlers.AnalyticsHandlers.AdminGetURLAnalytics)

		// Click analytics
		analytics.GET("/clicks",
			adminMiddleware.ReadOnlyDBContext(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "clicks"),
			handler.AdminHandlers.AnalyticsHandlers.AdminGetClickAnalytics)

		// Revenue analytics (requires billing permission)
		analytics.GET("/revenue",
			adminMiddleware.RequireAdminPermission(models.PermissionAdminBillingView),
			adminMiddleware.ReadOnlyDBContext(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "revenue"),
			handler.AdminHandlers.BillingHandlers.AdminGetRevenue)

		// Geographic analytics
		analytics.GET("/geography",
			adminMiddleware.ReadOnlyDBContext(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "geography"),
			handler.AdminHandlers.AnalyticsHandlers.AdminGetGeographicAnalytics)

		// Device/browser analytics
		analytics.GET("/devices",
			adminMiddleware.ReadOnlyDBContext(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "devices"),
			handler.AdminHandlers.AnalyticsHandlers.AdminGetDeviceAnalytics)

		// Trend analytics
		analytics.GET("/trends",
			adminMiddleware.ReadOnlyDBContext(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "trends"),
			handler.AdminHandlers.AnalyticsHandlers.AdminGetTrendAnalytics)

		// Top performing URLs
		analytics.GET("/top-urls",
			adminMiddleware.ReadOnlyDBContext(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "top_urls"),
			handler.AdminHandlers.AnalyticsHandlers.AdminGetTopURLs)

		// Export report (requires export permission)
		analytics.POST("/export",
			adminMiddleware.RequireAdminPermission(models.PermissionAdminAnalyticsExport),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsExport, "report"),
			handler.AdminHandlers.AnalyticsHandlers.AdminExportReport)
	}

	// ================================================================
	// BILLING & SUBSCRIPTION ADMINISTRATION
	// ================================================================
	billing := admin.Group("/billing")
	billing.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminBillingView))
	{
		// List all subscriptions
		billing.GET("/subscriptions",
			adminMiddleware.AuditLog(models.AuditActionSubscriptionView, models.ResourceTypeSubscription),
			handler.AdminHandlers.BillingHandlers.AdminListSubscriptions)

		// Get subscription details
		billing.GET("/subscriptions/:id",
			adminMiddleware.AuditLog(models.AuditActionSubscriptionView, models.ResourceTypeSubscription),
			handler.AdminHandlers.BillingHandlers.AdminGetSubscription)

		// Subscription management (requires manage permission)
		billingManage := billing.Group("")
		billingManage.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminBillingManage))
		{
			// Update subscription
			billingManage.PUT("/subscriptions/:id",
				adminMiddleware.AuditLog(models.AuditActionSubscriptionUpdate, models.ResourceTypeSubscription),
				handler.AdminHandlers.BillingHandlers.AdminUpdateSubscription)

			// Cancel subscription
			billingManage.POST("/subscriptions/:id/cancel",
				adminMiddleware.AuditLog(models.AuditActionSubscriptionCancel, models.ResourceTypeSubscription),
				handler.AdminHandlers.BillingHandlers.AdminCancelSubscription)

			// Process refund
			billingManage.POST("/subscriptions/:id/refund",
				adminMiddleware.RequireSuperAdmin(),
				adminMiddleware.AuditLog(models.AuditActionSubscriptionRefund, models.ResourceTypeSubscription),
				handler.AdminHandlers.BillingHandlers.AdminRefundSubscription)
		}

		// Plan management (super admin only)
		plans := billing.Group("/plans")
		plans.Use(adminMiddleware.RequireSuperAdmin())
		{
			plans.GET("", handler.AdminHandlers.BillingHandlers.AdminListPlans)
			plans.POST("",
				adminMiddleware.AuditLog("plan.create", "plan"),
				handler.AdminHandlers.BillingHandlers.AdminCreatePlan)
			plans.PUT("/:id",
				adminMiddleware.AuditLog("plan.update", "plan"),
				handler.AdminHandlers.BillingHandlers.AdminUpdatePlan)
			plans.DELETE("/:id",
				adminMiddleware.AuditLog("plan.delete", "plan"),
				handler.AdminHandlers.BillingHandlers.AdminDeletePlan)
		}

		// Revenue analytics
		billing.GET("/revenue",
			adminMiddleware.ReadOnlyDBContext(),
			adminMiddleware.AuditLog(models.AuditActionAnalyticsView, "revenue"),
			handler.AdminHandlers.BillingHandlers.AdminGetRevenue)

		// Webhook management
		billing.GET("/webhooks",
			adminMiddleware.AuditLog("webhook.view", "webhook"),
			handler.AdminHandlers.BillingHandlers.AdminListWebhookEvents)

		billing.POST("/webhooks/:id/retry",
			adminMiddleware.RequireSystemAdmin(),
			adminMiddleware.AuditLog("webhook.retry", "webhook"),
			handler.AdminHandlers.BillingHandlers.AdminRetryWebhook)

		// Sync with Polar
		billing.POST("/sync",
			adminMiddleware.RequireSystemAdmin(),
			adminMiddleware.AuditLog("billing.sync", "polar"),
			handler.AdminHandlers.BillingHandlers.AdminSyncPolar)
	}

	// ================================================================
	// RBAC MANAGEMENT
	// ================================================================
	rbac := admin.Group("/rbac")
	rbac.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminRBACView))
	{
		// Roles
		roles := rbac.Group("/roles")
		{
			roles.GET("", handler.AdminHandlers.RBACHandlers.AdminListRoles)
			roles.GET("/:id", handler.AdminHandlers.RBACHandlers.AdminGetRole)

			// Role management (requires manage permission)
			rolesManage := roles.Group("")
			rolesManage.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminRBACManage))
			{
				rolesManage.POST("",
					adminMiddleware.RequireSuperAdmin(),
					adminMiddleware.AuditLog(models.AuditActionRoleCreate, models.ResourceTypeRole),
					handler.AdminHandlers.RBACHandlers.AdminCreateRole)

				rolesManage.PUT("/:id",
					adminMiddleware.RequireSuperAdmin(),
					adminMiddleware.AuditLog(models.AuditActionRoleUpdate, models.ResourceTypeRole),
					handler.AdminHandlers.RBACHandlers.AdminUpdateRole)

				rolesManage.DELETE("/:id",
					adminMiddleware.RequireSuperAdmin(),
					adminMiddleware.AuditLog(models.AuditActionRoleDelete, models.ResourceTypeRole),
					handler.AdminHandlers.RBACHandlers.AdminDeleteRole)

				// Assign permissions to role
				rolesManage.POST("/:id/permissions",
					adminMiddleware.RequireSuperAdmin(),
					adminMiddleware.AuditLog(models.AuditActionPermissionAssign, models.ResourceTypeRole),
					handler.AdminHandlers.RBACHandlers.AdminAssignPermissionToRole)

				rolesManage.DELETE("/:id/permissions/:permId",
					adminMiddleware.RequireSuperAdmin(),
					adminMiddleware.AuditLog(models.AuditActionPermissionRevoke, models.ResourceTypeRole),
					handler.AdminHandlers.RBACHandlers.AdminRemovePermissionFromRole)
			}
		}

		// Permissions
		permissions := rbac.Group("/permissions")
		{
			permissions.GET("", handler.AdminHandlers.RBACHandlers.AdminListPermissions)
			permissions.GET("/:id", handler.AdminHandlers.RBACHandlers.AdminGetPermission)
		}

		// User role assignments
		userRoles := rbac.Group("/users/:userId/roles")
		userRoles.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminRBACManage))
		{
			userRoles.GET("", handler.AdminHandlers.UserHandlers.AdminGetUserRoles)

			userRoles.POST("",
				adminMiddleware.RequireSystemAdmin(),
				adminMiddleware.AuditLog(models.AuditActionRoleAssign, models.ResourceTypeRole),
				handler.AdminHandlers.RBACHandlers.AdminAssignRoleToUser)

			userRoles.DELETE("/:roleId",
				adminMiddleware.RequireSystemAdmin(),
				adminMiddleware.AuditLog(models.AuditActionRoleRevoke, models.ResourceTypeRole),
				handler.AdminHandlers.RBACHandlers.AdminRemoveRoleFromUser)
		}
	}

	// ================================================================
	// SYSTEM SETTINGS
	// ================================================================
	settings := admin.Group("/settings")
	settings.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminSettingsView))
	{
		// Get all settings or by category
		settings.GET("",
			adminMiddleware.AuditLog(models.AuditActionSettingView, models.ResourceTypeSetting),
			handler.AdminHandlers.SettingsHandlers.AdminGetSettings)

		settings.GET("/categories", handler.AdminHandlers.SettingsHandlers.AdminGetSettingCategories)

		settings.GET("/:key",
			adminMiddleware.AuditLog(models.AuditActionSettingView, models.ResourceTypeSetting),
			handler.AdminHandlers.SettingsHandlers.AdminGetSetting)

		// Update settings (requires manage permission)
		settingsManage := settings.Group("")
		settingsManage.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminSettingsManage))
		{
			settingsManage.PUT("/:key",
				adminMiddleware.RequireSystemAdmin(),
				adminMiddleware.AuditLog(models.AuditActionSettingUpdate, models.ResourceTypeSetting),
				handler.AdminHandlers.SettingsHandlers.AdminUpdateSetting)

			settingsManage.PUT("/bulk",
				adminMiddleware.RequireSystemAdmin(),
				adminMiddleware.AuditLog(models.AuditActionSettingUpdate, models.ResourceTypeSetting),
				handler.AdminHandlers.SettingsHandlers.AdminBulkUpdateSettings)

			settingsManage.POST("/reset",
				adminMiddleware.RequireSuperAdmin(),
				adminMiddleware.AuditLog(models.AuditActionSettingReset, models.ResourceTypeSetting),
				handler.AdminHandlers.SettingsHandlers.AdminResetSettings)
		}
	}

	// ================================================================
	// CMS MANAGEMENT
	// ================================================================
	cms := admin.Group("/cms")
	cms.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminCMSView))
	{
		// Pages
		pages := cms.Group("/pages")
		{
			pages.GET("", handler.AdminHandlers.CMSHandlers.AdminListPages)
			pages.GET("/:id", handler.AdminHandlers.CMSHandlers.AdminGetPage)
			pages.GET("/:id/versions", handler.AdminHandlers.CMSHandlers.AdminGetPageVersions)

			// Page management (requires manage permission)
			pagesManage := pages.Group("")
			pagesManage.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminCMSManage))
			{
				pagesManage.POST("",
					adminMiddleware.AuditLog(models.AuditActionCMSPageCreate, models.ResourceTypePage),
					handler.AdminHandlers.CMSHandlers.AdminCreatePage)

				pagesManage.PUT("/:id",
					adminMiddleware.AuditLog(models.AuditActionCMSPageUpdate, models.ResourceTypePage),
					handler.AdminHandlers.CMSHandlers.AdminUpdatePage)

				pagesManage.DELETE("/:id",
					adminMiddleware.AuditLog(models.AuditActionCMSPageDelete, models.ResourceTypePage),
					handler.AdminHandlers.CMSHandlers.AdminDeletePage)

				pagesManage.POST("/:id/publish",
					adminMiddleware.AuditLog("cms.page.publish", models.ResourceTypePage),
					handler.AdminHandlers.CMSHandlers.AdminPublishPage)
			}
		}

		// Media
		media := cms.Group("/media")
		media.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminCMSManage))
		{
			media.GET("", handler.AdminHandlers.CMSHandlers.AdminListMedia)
			media.POST("",
				adminMiddleware.AuditLog("cms.media.upload", "media"),
				handler.AdminHandlers.CMSHandlers.AdminUploadMedia)
			media.DELETE("/:id",
				adminMiddleware.AuditLog("cms.media.delete", "media"),
				handler.AdminHandlers.CMSHandlers.AdminDeleteMedia)
		}
	}

	// ================================================================
	// AUDIT LOGS
	// ================================================================
	auditLogs := admin.Group("/audit-logs")
	auditLogs.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminAuditView))
	{
		// List audit logs with filters
		auditLogs.GET("", handler.AdminHandlers.AuditHandlers.AdminListAuditLogs)

		// Get specific audit log
		auditLogs.GET("/:id", handler.AdminHandlers.AuditHandlers.AdminGetAuditLog)

		// User audit trail
		auditLogs.GET("/user/:userId", handler.AdminHandlers.UserHandlers.AdminGetUserAuditTrail)

		// Resource audit trail
		auditLogs.GET("/resource/:type/:id", handler.AdminHandlers.AuditHandlers.AdminGetResourceAuditTrail)

		// Recent critical actions
		auditLogs.GET("/critical", handler.AdminHandlers.AuditHandlers.AdminGetCriticalActions)

		// Audit statistics
		auditLogs.GET("/stats", handler.AdminHandlers.AuditHandlers.AdminGetAuditStatistics)

		// Export audit logs (super admin only)
		auditLogs.POST("/export",
			adminMiddleware.RequireSuperAdmin(),
			handler.AdminHandlers.AuditHandlers.AdminExportAuditLogs)

		// Cleanup old logs (super admin only)
		auditLogs.DELETE("/cleanup",
			adminMiddleware.RequireSuperAdmin(),
			adminMiddleware.AuditLog("audit.cleanup", "audit_logs"),
			handler.AdminHandlers.AuditHandlers.AdminCleanupAuditLogs)
	}

	// ================================================================
	// SYSTEM MONITORING
	// ================================================================
	monitoring := admin.Group("/monitoring")
	monitoring.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminMonitoringView))
	{
		// System health
		monitoring.GET("/health", handler.AdminHandlers.MonitoringHandlers.AdminGetSystemHealth)
		monitoring.GET("/health/database", handler.AdminHandlers.MonitoringHandlers.AdminGetDatabaseHealth)
		monitoring.GET("/health/redis", handler.AdminHandlers.MonitoringHandlers.AdminGetRedisHealth)
		monitoring.GET("/health/services", handler.AdminHandlers.MonitoringHandlers.AdminGetServicesHealth)

		// System metrics
		monitoring.GET("/metrics", handler.AdminHandlers.MonitoringHandlers.AdminGetSystemMetrics)
		monitoring.GET("/metrics/performance", handler.AdminHandlers.MonitoringHandlers.AdminGetPerformanceMetrics)

		// Alerts
		alerts := monitoring.Group("/alerts")
		{
			alerts.GET("", handler.AdminHandlers.MonitoringHandlers.AdminListAlerts)
			alerts.GET("/:id", handler.AdminHandlers.MonitoringHandlers.AdminGetAlert)

			// Alert management (requires manage permission)
			alertsManage := alerts.Group("")
			alertsManage.Use(adminMiddleware.RequireAdminPermission(models.PermissionAdminMonitoringManage))
			{
				alertsManage.POST("/:id/acknowledge",
					adminMiddleware.AuditLog(models.AuditActionSystemAlertAck, models.ResourceTypeAlert),
					handler.AdminHandlers.MonitoringHandlers.AdminAcknowledgeAlert)

				alertsManage.POST("/:id/resolve",
					adminMiddleware.AuditLog(models.AuditActionSystemAlertResolve, models.ResourceTypeAlert),
					handler.AdminHandlers.MonitoringHandlers.AdminResolveAlert)
			}
		}

		// Error logs
		monitoring.GET("/errors", handler.AdminHandlers.MonitoringHandlers.AdminGetErrorLogs)
	}

	// ================================================================
	// API KEY ADMINISTRATION
	// ================================================================
	apiKeys := admin.Group("/api-keys")
	apiKeys.Use(adminMiddleware.RequireSystemAdmin())
	{
		// List all API keys across users
		apiKeys.GET("",
			adminMiddleware.AuditLog("api_key.view", "api_key"),
			handler.AdminHandlers.APIKeyHandlers.AdminListAPIKeys)

		// Get API key details
		apiKeys.GET("/:id",
			adminMiddleware.AuditLog("api_key.view", "api_key"),
			handler.AdminHandlers.APIKeyHandlers.AdminGetAPIKey)

		// Revoke API key
		apiKeys.PUT("/:id/revoke",
			adminMiddleware.AuditLog("api_key.revoke", "api_key"),
			handler.AdminHandlers.APIKeyHandlers.AdminRevokeAPIKey)

		// Delete API key
		apiKeys.DELETE("/:id",
			adminMiddleware.RequireSuperAdmin(),
			adminMiddleware.AuditLog("api_key.delete", "api_key"),
			handler.AdminHandlers.APIKeyHandlers.AdminDeleteAPIKey)

		// API key usage stats
		apiKeys.GET("/:id/usage",
			adminMiddleware.ReadOnlyDBContext(),
			handler.AdminHandlers.APIKeyHandlers.AdminGetAPIKeyUsage)

		// Create system API key
		apiKeys.POST("/system",
			adminMiddleware.RequireSuperAdmin(),
			adminMiddleware.AuditLog("api_key.system.create", "api_key"),
			handler.AdminHandlers.APIKeyHandlers.AdminCreateSystemAPIKey)
	}
}
