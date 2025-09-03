package models

import (
	"time"
)

// CustomDomain represents a custom domain configuration
type CustomDomain struct {
	ID              int64      `json:"id" db:"id"`
	Domain          string     `json:"domain" db:"domain"`
	UserID          int64      `json:"user_id" db:"user_id"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	IsVerified      bool       `json:"is_verified" db:"is_verified"`
	VerificationKey string     `json:"-" db:"verification_key"` // Hidden from JSON
	SSLEnabled      bool       `json:"ssl_enabled" db:"ssl_enabled"`
	SSLStatus       string     `json:"ssl_status" db:"ssl_status"`
	DNSStatus       string     `json:"dns_status" db:"dns_status"`
	LastChecked     *time.Time `json:"last_checked" db:"last_checked"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// DomainVerification represents domain verification record
type DomainVerification struct {
	ID             int64     `json:"id" db:"id"`
	DomainID       int64     `json:"domain_id" db:"domain_id"`
	VerificationType string  `json:"verification_type" db:"verification_type"` // dns, http, file
	VerificationKey  string  `json:"verification_key" db:"verification_key"`
	VerificationValue string `json:"verification_value" db:"verification_value"`
	IsVerified     bool      `json:"is_verified" db:"is_verified"`
	VerifiedAt     *time.Time `json:"verified_at" db:"verified_at"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// SSLCertificate represents SSL certificate information
type SSLCertificate struct {
	ID         int64      `json:"id" db:"id"`
	DomainID   int64      `json:"domain_id" db:"domain_id"`
	Provider   string     `json:"provider" db:"provider"` // letsencrypt, custom, cloudflare
	Status     string     `json:"status" db:"status"`     // pending, active, expired, failed
	IssuedAt   *time.Time `json:"issued_at" db:"issued_at"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
	AutoRenew  bool       `json:"auto_renew" db:"auto_renew"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// DomainAnalytics represents analytics for a custom domain
type DomainAnalytics struct {
	DomainID      int64                `json:"domain_id"`
	Domain        string               `json:"domain"`
	TotalClicks   int64                `json:"total_clicks"`
	TotalURLs     int64                `json:"total_urls"`
	TopURLs       []DomainTopURL       `json:"top_urls"`
	ClickTrends   []ClickTrend         `json:"click_trends"`
	ReferrerStats []ReferrerStat       `json:"referrer_stats"`
	GeographicStats []CountryDetail    `json:"geographic_stats"`
	DeviceStats   []DeviceTypeStat     `json:"device_stats"`
	LastUpdated   time.Time            `json:"last_updated"`
}

// DomainTopURL represents top performing URLs for a domain
type DomainTopURL struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	Title       string `json:"title"`
	Clicks      int64  `json:"clicks"`
}

// Request/Response Models

// CreateDomainRequest represents a request to add a custom domain
type CreateDomainRequest struct {
	Domain string `json:"domain" validate:"required,fqdn"`
}

// UpdateDomainRequest represents a request to update domain settings
type UpdateDomainRequest struct {
	IsActive   *bool `json:"is_active,omitempty"`
	SSLEnabled *bool `json:"ssl_enabled,omitempty"`
}

// DomainVerificationResponse represents domain verification instructions
type DomainVerificationResponse struct {
	Domain            string                    `json:"domain"`
	VerificationMethods []VerificationMethod    `json:"verification_methods"`
	Status            string                    `json:"status"`
	Instructions      string                    `json:"instructions"`
}

// VerificationMethod represents a domain verification method
type VerificationMethod struct {
	Type        string `json:"type"`        // dns, http, file
	Name        string `json:"name"`        // Record name or file name
	Value       string `json:"value"`       // Record value or file content
	Description string `json:"description"` // Human-readable instructions
}

// DomainSettingsResponse represents domain configuration
type DomainSettingsResponse struct {
	*CustomDomain
	DNSRecords     []DNSRecord        `json:"dns_records"`
	SSL            *SSLCertificate    `json:"ssl_certificate,omitempty"`
	Analytics      *DomainAnalytics   `json:"analytics,omitempty"`
}

// DNSRecord represents required DNS configuration
type DNSRecord struct {
	Type        string `json:"type"`        // CNAME, A, TXT
	Name        string `json:"name"`        // Record name
	Value       string `json:"value"`       // Record value
	TTL         int    `json:"ttl"`         // Time to live
	Description string `json:"description"` // Purpose of the record
	IsRequired  bool   `json:"is_required"` // Whether this record is required
	IsConfigured bool  `json:"is_configured"` // Whether this record is properly configured
}

// DomainStatus represents the current status of a domain
type DomainStatus struct {
	Domain       string               `json:"domain"`
	IsActive     bool                 `json:"is_active"`
	IsVerified   bool                 `json:"is_verified"`
	DNSStatus    string               `json:"dns_status"`    // configured, misconfigured, pending
	SSLStatus    string               `json:"ssl_status"`    // active, pending, failed, disabled
	OverallStatus string              `json:"overall_status"` // active, pending, failed
	Checks       []DomainStatusCheck  `json:"checks"`
	LastChecked  time.Time            `json:"last_checked"`
	NextCheck    time.Time            `json:"next_check"`
}

// DomainStatusCheck represents individual domain health checks
type DomainStatusCheck struct {
	Name        string    `json:"name"`        // dns_resolution, ssl_certificate, http_response
	Status      string    `json:"status"`      // pass, fail, warning
	Message     string    `json:"message"`     // Detailed status message
	LastChecked time.Time `json:"last_checked"`
}

// DomainStats represents domain usage statistics
type DomainStats struct {
	TotalDomains    int64 `json:"total_domains"`
	ActiveDomains   int64 `json:"active_domains"`
	VerifiedDomains int64 `json:"verified_domains"`
	SSLDomains      int64 `json:"ssl_domains"`
	PendingDomains  int64 `json:"pending_domains"`
	FailedDomains   int64 `json:"failed_domains"`
}

// DomainUsage represents domain usage and limits
type DomainUsage struct {
	UserID        int64 `json:"user_id"`
	DomainsUsed   int   `json:"domains_used"`
	DomainsLimit  int   `json:"domains_limit"`
	CanAddDomain  bool  `json:"can_add_domain"`
	AccountType   string `json:"account_type"`
}

// Constants for domain status values
const (
	// Domain verification status
	DomainStatusPending    = "pending"
	DomainStatusActive     = "active"
	DomainStatusFailed     = "failed"
	DomainStatusSuspended  = "suspended"

	// DNS status values
	DNSStatusPending       = "pending"
	DNSStatusConfigured    = "configured"
	DNSStatusMisconfigured = "misconfigured"
	DNSStatusFailed        = "failed"

	// SSL status values
	SSLStatusDisabled      = "disabled"
	SSLStatusPending       = "pending"
	SSLStatusActive        = "active"
	SSLStatusExpired       = "expired"
	SSLStatusFailed        = "failed"
	SSLStatusRenewing      = "renewing"

	// Verification types
	VerificationTypeDNS    = "dns"
	VerificationTypeHTTP   = "http"
	VerificationTypeFile   = "file"

	// SSL providers
	SSLProviderLetsEncrypt = "letsencrypt"
	SSLProviderCloudflare  = "cloudflare"
	SSLProviderCustom      = "custom"

	// Check types
	CheckTypeDNS           = "dns_resolution"
	CheckTypeSSL           = "ssl_certificate"
	CheckTypeHTTP          = "http_response"
	CheckTypeConnectivity  = "connectivity"

	// Check status
	CheckStatusPass        = "pass"
	CheckStatusFail        = "fail"
	CheckStatusWarning     = "warning"
	CheckStatusPending     = "pending"
)

// Domain limits by account type
var DomainLimits = map[string]int{
	"free":       1,
	"premium":    10,
	"enterprise": 100,
}

// GetDomainLimit returns the domain limit for an account type
func GetDomainLimit(accountType string) int {
	if limit, exists := DomainLimits[accountType]; exists {
		return limit
	}
	return DomainLimits["free"] // Default to free tier limit
}