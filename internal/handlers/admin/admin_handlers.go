package admin

import (
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// AdminHandlers aggregates all admin handler groups
type AdminHandlers struct {
	UserHandlers       *AdminUserHandlers
	URLHandlers        *AdminURLHandlers
	DomainHandlers     *AdminDomainHandlers
	AnalyticsHandlers  *AdminAnalyticsHandlers
	BillingHandlers    *AdminBillingHandlers
	RBACHandlers       *AdminRBACHandlers
	SettingsHandlers   *AdminSettingsHandlers
	CMSHandlers        *AdminCMSHandlers
	AuditHandlers      *AdminAuditHandlers
	MonitoringHandlers *AdminMonitoringHandlers
	APIKeyHandlers     *AdminAPIKeyHandlers
}

// NewAdminHandlers creates a new admin handlers instance with all dependencies
func NewAdminHandlers(
	dbManager *storage.DatabaseManager,
	redisStorage *storage.RedisStorage,
	adminUserService *services.AdminUserService,
	adminURLService *services.AdminURLService,
	domainBlocklistService *services.DomainBlocklistService,
	adminAnalyticsService *services.AdminAnalyticsService,
	monitoringService *services.MonitoringService,
	adminAuditService *services.AdminAuditService,
	rbacService *services.RBACService,
	billingService *services.BillingService,
	globalSettingsService *services.GlobalSettingsService,
	cmsService *services.CMSService,
	mediaService *services.MediaService,
	apiKeyService *services.APIKeyService,
) *AdminHandlers {
	return &AdminHandlers{
		UserHandlers:       NewAdminUserHandlers(adminUserService, adminAuditService),
		URLHandlers:        NewAdminURLHandlers(adminURLService, adminAuditService),
		DomainHandlers:     NewAdminDomainHandlers(domainBlocklistService, adminAuditService),
		AnalyticsHandlers:  NewAdminAnalyticsHandlers(adminAnalyticsService),
		BillingHandlers:    NewAdminBillingHandlers(billingService, adminAuditService),
		RBACHandlers:       NewAdminRBACHandlers(rbacService, adminAuditService),
		SettingsHandlers:   NewAdminSettingsHandlers(globalSettingsService, adminAuditService),
		CMSHandlers:        NewAdminCMSHandlers(cmsService, mediaService, adminAuditService),
		AuditHandlers:      NewAdminAuditHandlers(adminAuditService),
		MonitoringHandlers: NewAdminMonitoringHandlers(monitoringService),
		APIKeyHandlers:     NewAdminAPIKeyHandlers(apiKeyService, adminAuditService),
	}
}
