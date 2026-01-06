package models

// Admin Permission Constants
const (
	// User Management Permissions
	PermissionAdminUserView        = "admin.user.view"
	PermissionAdminUserManage      = "admin.user.manage"
	PermissionAdminUserImpersonate = "admin.user.impersonate"

	// URL Management Permissions
	PermissionAdminURLView   = "admin.url.view"
	PermissionAdminURLManage = "admin.url.manage"
	PermissionAdminURLFlag   = "admin.url.flag"

	// Analytics Permissions
	PermissionAdminAnalyticsView   = "admin.analytics.view"
	PermissionAdminAnalyticsExport = "admin.analytics.export"

	// Billing Permissions
	PermissionAdminBillingView   = "admin.billing.view"
	PermissionAdminBillingManage = "admin.billing.manage"

	// RBAC Permissions
	PermissionAdminRBACView   = "admin.rbac.view"
	PermissionAdminRBACManage = "admin.rbac.manage"

	// Settings Permissions
	PermissionAdminSettingsView   = "admin.settings.view"
	PermissionAdminSettingsManage = "admin.settings.manage"

	// CMS Permissions
	PermissionAdminCMSView   = "admin.cms.view"
	PermissionAdminCMSManage = "admin.cms.manage"

	// Audit Log Permissions
	PermissionAdminAuditView = "admin.audit.view"

	// Monitoring Permissions
	PermissionAdminMonitoringView   = "admin.monitoring.view"
	PermissionAdminMonitoringManage = "admin.monitoring.manage"
)

// Admin-Specific Role Constants (in addition to existing roles in rbac.go)
const (
	RoleSystemAdmin  = "system_admin"
	RoleSupportAdmin = "support_admin"
	RoleBillingAdmin = "billing_admin"
	RoleContentAdmin = "content_admin"
	RolePremium      = "premium"
)
