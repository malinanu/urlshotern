# Script to fix admin routes handler references
$file = "D:\github\URLshorter\internal\routes\admin_routes.go"
$content = Get-Content $file -Raw

# URL Handlers
$content = $content -replace 'handler\.AdminListURLs\)', 'handler.AdminHandlers.URLHandlers.AdminListURLs)'
$content = $content -replace 'handler\.AdminGetURL\)', 'handler.AdminHandlers.URLHandlers.AdminGetURL)'
$content = $content -replace 'handler\.AdminUpdateURLStatus\)', 'handler.AdminHandlers.URLHandlers.AdminUpdateURLStatus)'
$content = $content -replace 'handler\.AdminDeleteURL\)', 'handler.AdminHandlers.URLHandlers.AdminDeleteURL)'
$content = $content -replace 'handler\.AdminTransferURL\)', 'handler.AdminHandlers.URLHandlers.AdminTransferURL)'
$content = $content -replace 'handler\.AdminExtendURLExpiration\)', 'handler.AdminHandlers.URLHandlers.AdminExtendExpiration)'
$content = $content -replace 'handler\.AdminGetFlaggedURLs\)', 'handler.AdminHandlers.URLHandlers.AdminGetFlaggedURLs)'
$content = $content -replace 'handler\.AdminFlagURL\)', 'handler.AdminHandlers.URLHandlers.AdminFlagURL)'
$content = $content -replace 'handler\.AdminUnflagURL\)', 'handler.AdminHandlers.URLHandlers.AdminUnflagURL)'

# Domain Handlers
$content = $content -replace 'handler\.AdminGetBlockedDomains\)', 'handler.AdminHandlers.DomainHandlers.AdminGetBlockedDomains)'
$content = $content -replace 'handler\.AdminBlockDomain\)', 'handler.AdminHandlers.DomainHandlers.AdminBlockDomain)'
$content = $content -replace 'handler\.AdminUnblockDomain\)', 'handler.AdminHandlers.DomainHandlers.AdminUnblockDomain)'

# Analytics Handlers
$content = $content -replace 'handler\.AdminDashboard\)', 'handler.AdminHandlers.AnalyticsHandlers.AdminGetDashboard)'
$content = $content -replace 'handler\.AdminUserAnalytics\)', 'handler.AdminHandlers.AnalyticsHandlers.AdminGetUserAnalytics)'
$content = $content -replace 'handler\.AdminURLAnalytics\)', 'handler.AdminHandlers.AnalyticsHandlers.AdminGetURLAnalytics)'
$content = $content -replace 'handler\.AdminClickAnalytics\)', 'handler.AdminHandlers.AnalyticsHandlers.AdminGetClickAnalytics)'
$content = $content -replace 'handler\.AdminRevenueAnalytics\)', 'handler.AdminHandlers.BillingHandlers.AdminGetRevenue)'
$content = $content -replace 'handler\.AdminGeographicAnalytics\)', 'handler.AdminHandlers.AnalyticsHandlers.AdminGetGeographicAnalytics)'
$content = $content -replace 'handler\.AdminDeviceAnalytics\)', 'handler.AdminHandlers.AnalyticsHandlers.AdminGetDeviceAnalytics)'
$content = $content -replace 'handler\.AdminTrendAnalytics\)', 'handler.AdminHandlers.AnalyticsHandlers.AdminGetTrendAnalytics)'
$content = $content -replace 'handler\.AdminTopURLs\)', 'handler.AdminHandlers.AnalyticsHandlers.AdminGetTopURLs)'
$content = $content -replace 'handler\.AdminExportReport\)', 'handler.AdminHandlers.AnalyticsHandlers.AdminExportReport)'

# Billing Handlers
$content = $content -replace 'handler\.AdminListSubscriptions\)', 'handler.AdminHandlers.BillingHandlers.AdminListSubscriptions)'
$content = $content -replace 'handler\.AdminGetSubscription\)', 'handler.AdminHandlers.BillingHandlers.AdminGetSubscription)'
$content = $content -replace 'handler\.AdminUpdateSubscription\)', 'handler.AdminHandlers.BillingHandlers.AdminUpdateSubscription)'
$content = $content -replace 'handler\.AdminCancelSubscription\)', 'handler.AdminHandlers.BillingHandlers.AdminCancelSubscription)'
$content = $content -replace 'handler\.AdminRefundSubscription\)', 'handler.AdminHandlers.BillingHandlers.AdminRefundSubscription)'
$content = $content -replace 'handler\.AdminListPlans', 'handler.AdminHandlers.BillingHandlers.AdminListPlans'
$content = $content -replace 'handler\.AdminCreatePlan\)', 'handler.AdminHandlers.BillingHandlers.AdminCreatePlan)'
$content = $content -replace 'handler\.AdminUpdatePlan\)', 'handler.AdminHandlers.BillingHandlers.AdminUpdatePlan)'
$content = $content -replace 'handler\.AdminDeletePlan\)', 'handler.AdminHandlers.BillingHandlers.AdminDeletePlan)'
$content = $content -replace 'handler\.AdminListWebhookEvents\)', 'handler.AdminHandlers.BillingHandlers.AdminListWebhookEvents)'
$content = $content -replace 'handler\.AdminRetryWebhook\)', 'handler.AdminHandlers.BillingHandlers.AdminRetryWebhook)'
$content = $content -replace 'handler\.AdminSyncPolar\)', 'handler.AdminHandlers.BillingHandlers.AdminSyncPolar)'

# RBAC Handlers
$content = $content -replace 'handler\.AdminListRoles', 'handler.AdminHandlers.RBACHandlers.AdminListRoles'
$content = $content -replace 'handler\.AdminGetRole', 'handler.AdminHandlers.RBACHandlers.AdminGetRole'
$content = $content -replace 'handler\.AdminCreateRole\)', 'handler.AdminHandlers.RBACHandlers.AdminCreateRole)'
$content = $content -replace 'handler\.AdminUpdateRole\)', 'handler.AdminHandlers.RBACHandlers.AdminUpdateRole)'
$content = $content -replace 'handler\.AdminDeleteRole\)', 'handler.AdminHandlers.RBACHandlers.AdminDeleteRole)'
$content = $content -replace 'handler\.AdminAssignPermissionToRole\)', 'handler.AdminHandlers.RBACHandlers.AdminAssignPermissionToRole)'
$content = $content -replace 'handler\.AdminRemovePermissionFromRole\)', 'handler.AdminHandlers.RBACHandlers.AdminRemovePermissionFromRole)'
$content = $content -replace 'handler\.AdminListPermissions', 'handler.AdminHandlers.RBACHandlers.AdminListPermissions'
$content = $content -replace 'handler\.AdminGetPermission', 'handler.AdminHandlers.RBACHandlers.AdminGetPermission'
$content = $content -replace 'handler\.AdminAssignRoleToUser\)', 'handler.AdminHandlers.RBACHandlers.AdminAssignRoleToUser)'
$content = $content -replace 'handler\.AdminRemoveRoleFromUser\)', 'handler.AdminHandlers.RBACHandlers.AdminRemoveRoleFromUser)'

# Settings Handlers
$content = $content -replace 'handler\.AdminGetSettings\)', 'handler.AdminHandlers.SettingsHandlers.AdminGetSettings)'
$content = $content -replace 'handler\.AdminGetSettingCategories\)', 'handler.AdminHandlers.SettingsHandlers.AdminGetSettingCategories)'
$content = $content -replace 'handler\.AdminGetSetting\)', 'handler.AdminHandlers.SettingsHandlers.AdminGetSetting)'
$content = $content -replace 'handler\.AdminUpdateSetting\)', 'handler.AdminHandlers.SettingsHandlers.AdminUpdateSetting)'
$content = $content -replace 'handler\.AdminBulkUpdateSettings\)', 'handler.AdminHandlers.SettingsHandlers.AdminBulkUpdateSettings)'
$content = $content -replace 'handler\.AdminResetSettings\)', 'handler.AdminHandlers.SettingsHandlers.AdminResetSettings)'

# CMS Handlers
$content = $content -replace 'handler\.AdminListCMSPages', 'handler.AdminHandlers.CMSHandlers.AdminListPages'
$content = $content -replace 'handler\.AdminGetCMSPage', 'handler.AdminHandlers.CMSHandlers.AdminGetPage'
$content = $content -replace 'handler\.AdminGetCMSPageVersions', 'handler.AdminHandlers.CMSHandlers.AdminGetPageVersions'
$content = $content -replace 'handler\.AdminCreateCMSPage\)', 'handler.AdminHandlers.CMSHandlers.AdminCreatePage)'
$content = $content -replace 'handler\.AdminUpdateCMSPage\)', 'handler.AdminHandlers.CMSHandlers.AdminUpdatePage)'
$content = $content -replace 'handler\.AdminDeleteCMSPage\)', 'handler.AdminHandlers.CMSHandlers.AdminDeletePage)'
$content = $content -replace 'handler\.AdminPublishCMSPage\)', 'handler.AdminHandlers.CMSHandlers.AdminPublishPage)'
$content = $content -replace 'handler\.AdminListMedia', 'handler.AdminHandlers.CMSHandlers.AdminListMedia'
$content = $content -replace 'handler\.AdminUploadMedia\)', 'handler.AdminHandlers.CMSHandlers.AdminUploadMedia)'
$content = $content -replace 'handler\.AdminDeleteMedia\)', 'handler.AdminHandlers.CMSHandlers.AdminDeleteMedia)'

# Audit Handlers
$content = $content -replace 'handler\.AdminListAuditLogs', 'handler.AdminHandlers.AuditHandlers.AdminListAuditLogs'
$content = $content -replace 'handler\.AdminGetAuditLog', 'handler.AdminHandlers.AuditHandlers.AdminGetAuditLog'
$content = $content -replace 'handler\.AdminGetUserAuditTrail', 'handler.AdminHandlers.AuditHandlers.AdminGetUserAuditTrail'
$content = $content -replace 'handler\.AdminGetResourceAuditTrail', 'handler.AdminHandlers.AuditHandlers.AdminGetResourceAuditTrail'
$content = $content -replace 'handler\.AdminGetCriticalActions', 'handler.AdminHandlers.AuditHandlers.AdminGetCriticalActions'
$content = $content -replace 'handler\.AdminGetAuditStats', 'handler.AdminHandlers.AuditHandlers.AdminGetAuditStatistics'
$content = $content -replace 'handler\.AdminExportAuditLogs\)', 'handler.AdminHandlers.AuditHandlers.AdminExportAuditLogs)'
$content = $content -replace 'handler\.AdminCleanupAuditLogs\)', 'handler.AdminHandlers.AuditHandlers.AdminCleanupAuditLogs)'

# Monitoring Handlers
$content = $content -replace 'handler\.AdminSystemHealth', 'handler.AdminHandlers.MonitoringHandlers.AdminGetSystemHealth'
$content = $content -replace 'handler\.AdminDatabaseHealth', 'handler.AdminHandlers.MonitoringHandlers.AdminGetDatabaseHealth'
$content = $content -replace 'handler\.AdminRedisHealth', 'handler.AdminHandlers.MonitoringHandlers.AdminGetRedisHealth'
$content = $content -replace 'handler\.AdminServicesHealth', 'handler.AdminHandlers.MonitoringHandlers.AdminGetServicesHealth'
$content = $content -replace 'handler\.AdminSystemMetrics', 'handler.AdminHandlers.MonitoringHandlers.AdminGetSystemMetrics'
$content = $content -replace 'handler\.AdminPerformanceMetrics', 'handler.AdminHandlers.MonitoringHandlers.AdminGetPerformanceMetrics'
$content = $content -replace 'handler\.AdminListAlerts', 'handler.AdminHandlers.MonitoringHandlers.AdminListAlerts'
$content = $content -replace 'handler\.AdminGetAlert', 'handler.AdminHandlers.MonitoringHandlers.AdminGetAlert'
$content = $content -replace 'handler\.AdminAcknowledgeAlert\)', 'handler.AdminHandlers.MonitoringHandlers.AdminAcknowledgeAlert)'
$content = $content -replace 'handler\.AdminResolveAlert\)', 'handler.AdminHandlers.MonitoringHandlers.AdminResolveAlert)'
$content = $content -replace 'handler\.AdminGetErrorLogs', 'handler.AdminHandlers.MonitoringHandlers.AdminGetErrorLogs'

# API Key Handlers
$content = $content -replace 'handler\.AdminListAllAPIKeys\)', 'handler.AdminHandlers.APIKeyHandlers.AdminListAPIKeys)'
$content = $content -replace 'handler\.AdminGetAPIKey\)', 'handler.AdminHandlers.APIKeyHandlers.AdminGetAPIKey)'
$content = $content -replace 'handler\.AdminRevokeAPIKey\)', 'handler.AdminHandlers.APIKeyHandlers.AdminRevokeAPIKey)'
$content = $content -replace 'handler\.AdminDeleteAPIKey\)', 'handler.AdminHandlers.APIKeyHandlers.AdminDeleteAPIKey)'
$content = $content -replace 'handler\.AdminGetAPIKeyUsage\)', 'handler.AdminHandlers.APIKeyHandlers.AdminGetAPIKeyUsage)'
$content = $content -replace 'handler\.AdminCreateSystemAPIKey\)', 'handler.AdminHandlers.APIKeyHandlers.AdminCreateSystemAPIKey)'

# User handlers that weren't replaced yet
$content = $content -replace 'handler\.AdminUpdateUserLimits,', 'handler.AdminHandlers.UserHandlers.AdminUpdateUserLimits,'
$content = $content -replace 'handler\.AdminUpdateUserPlan,', 'handler.AdminHandlers.UserHandlers.AdminUpdateUserPlan,'
$content = $content -replace 'handler\.AdminGetUserActivity,', 'handler.AdminHandlers.UserHandlers.AdminGetUserActivity,'
$content = $content -replace 'handler\.AdminGetUserURLs,', 'handler.AdminHandlers.UserHandlers.AdminGetUserURLs,'
$content = $content -replace 'handler\.AdminGetUserRoles,', 'handler.AdminHandlers.RBACHandlers.AdminGetUserRoles,'

# Save the file
$content | Set-Content $file -NoNewline

Write-Host "Routes file updated successfully!"
