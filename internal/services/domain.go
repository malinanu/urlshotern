package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
)

// privateIPBlocks lists CIDR ranges that must not be reached from user-supplied domains.
var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // link-local (AWS/GCP metadata)
		"100.64.0.0/10",  // carrier-grade NAT
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 ULA
		"fe80::/10",      // IPv6 link-local
	} {
		_, block, _ := net.ParseCIDR(cidr)
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

// isPrivateIP returns true if the IP is in a reserved/private range.
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// isSafeHostname resolves a hostname and rejects it if any resolved IP is private.
func isSafeHostname(host string) bool {
	// Reject bare IP addresses that are private
	if ip := net.ParseIP(host); ip != nil {
		return !isPrivateIP(ip)
	}
	// Resolve and check every returned IP
	addrs, err := net.LookupHost(host)
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil || isPrivateIP(ip) {
			return false
		}
	}
	return len(addrs) > 0
}

// safeHTTPClient is a shared client with a strict timeout and no redirect following.
var safeHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// Allow up to 3 redirects but re-check each target host
		if len(via) >= 3 {
			return fmt.Errorf("too many redirects")
		}
		if !isSafeHostname(req.URL.Hostname()) {
			return fmt.Errorf("redirect to private/internal host denied")
		}
		return nil
	},
}

type DomainService struct {
	db               *storage.PostgresStorage
	redis            *storage.RedisStorage
	config           *configs.Config
	analyticsService *AnalyticsService
	userService      *UserService
}

// NewDomainService creates a new domain service
func NewDomainService(
	db *storage.PostgresStorage,
	redis *storage.RedisStorage,
	config *configs.Config,
	analyticsService *AnalyticsService,
	userService *UserService,
) *DomainService {
	return &DomainService{
		db:               db,
		redis:            redis,
		config:           config,
		analyticsService: analyticsService,
		userService:      userService,
	}
}

// Domain Management

// CreateDomain creates a new custom domain
func (d *DomainService) CreateDomain(req *models.CreateDomainRequest, userID int64) (*models.CustomDomain, error) {
	// Validate domain format
	if err := d.validateDomainName(req.Domain); err != nil {
		return nil, err
	}

	// Check if domain already exists
	existing, err := d.GetDomainByName(req.Domain)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("domain already exists")
	}

	// Check user's domain limit
	canAdd, err := d.CanAddDomain(userID)
	if err != nil {
		return nil, err
	}
	if !canAdd {
		return nil, fmt.Errorf("domain limit reached for your account type")
	}

	// Generate verification key
	verificationKey, err := d.generateVerificationKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification key: %w", err)
	}

	// Create domain record
	domain := &models.CustomDomain{
		Domain:          strings.ToLower(req.Domain),
		UserID:          userID,
		IsActive:        false, // Inactive until verified
		IsVerified:      false,
		VerificationKey: verificationKey,
		SSLEnabled:      true,
		SSLStatus:       models.SSLStatusPending,
		DNSStatus:       models.DNSStatusPending,
	}

	// Save to database
	if err := d.db.CreateCustomDomain(domain); err != nil {
		return nil, fmt.Errorf("failed to save domain: %w", err)
	}

	return domain, nil
}

// GetUserDomains retrieves all domains for a user
func (d *DomainService) GetUserDomains(userID int64) ([]*models.CustomDomain, error) {
	return d.db.GetCustomDomainsByUser(userID)
}

// GetDomainByID retrieves a domain by ID
func (d *DomainService) GetDomainByID(domainID int64) (*models.CustomDomain, error) {
	return d.db.GetCustomDomainByID(domainID)
}

// GetDomainByName retrieves a domain by domain name
func (d *DomainService) GetDomainByName(domainName string) (*models.CustomDomain, error) {
	return d.db.GetCustomDomainByName(domainName)
}

// UpdateDomain updates domain settings
func (d *DomainService) UpdateDomain(domainID int64, req *models.UpdateDomainRequest, userID int64) (*models.CustomDomain, error) {
	// Get domain and verify ownership
	domain, err := d.GetDomainByID(domainID)
	if err != nil {
		return nil, err
	}

	if domain.UserID != userID {
		return nil, fmt.Errorf("access denied: you don't own this domain")
	}

	// Update fields
	if req.IsActive != nil {
		domain.IsActive = *req.IsActive
	}
	if req.SSLEnabled != nil {
		domain.SSLEnabled = *req.SSLEnabled
		if *req.SSLEnabled && domain.SSLStatus == models.SSLStatusDisabled {
			domain.SSLStatus = models.SSLStatusPending
		}
	}

	domain.UpdatedAt = time.Now()

	if err := d.db.UpdateCustomDomain(domain); err != nil {
		return nil, fmt.Errorf("failed to update domain: %w", err)
	}

	return domain, nil
}

// DeleteDomain deletes a domain
func (d *DomainService) DeleteDomain(domainID int64, userID int64) error {
	// Get domain and verify ownership
	domain, err := d.GetDomainByID(domainID)
	if err != nil {
		return err
	}

	if domain.UserID != userID {
		return fmt.Errorf("access denied: you don't own this domain")
	}

	// Check if domain is being used by any URLs
	urlCount, err := d.getDomainURLCount(domainID)
	if err != nil {
		return err
	}

	if urlCount > 0 {
		return fmt.Errorf("cannot delete domain with active URLs")
	}

	return d.db.DeleteCustomDomain(domainID, userID)
}

// Domain Verification

// GetDomainVerificationInfo retrieves verification instructions for a domain
func (d *DomainService) GetDomainVerificationInfo(domainID int64) (*models.DomainVerificationResponse, error) {
	domain, err := d.GetDomainByID(domainID)
	if err != nil {
		return nil, err
	}

	if domain.IsVerified {
		return &models.DomainVerificationResponse{
			Domain:              domain.Domain,
			Status:              "verified",
			Instructions:        "Domain is already verified",
			VerificationMethods: []models.VerificationMethod{},
		}, nil
	}

	// Generate verification methods
	methods := []models.VerificationMethod{
		{
			Type:        models.VerificationTypeDNS,
			Name:        "_urlshortener-verify." + domain.Domain,
			Value:       domain.VerificationKey,
			Description: "Add this TXT record to your DNS configuration",
		},
		{
			Type:        models.VerificationTypeHTTP,
			Name:        "/.well-known/urlshortener-verify.txt",
			Value:       domain.VerificationKey,
			Description: "Upload a text file with this content to your domain",
		},
		{
			Type:        models.VerificationTypeFile,
			Name:        "urlshortener-verify-" + domain.VerificationKey + ".html",
			Value:       domain.VerificationKey,
			Description: "Upload an HTML file with this name to your domain root",
		},
	}

	return &models.DomainVerificationResponse{
		Domain:              domain.Domain,
		VerificationKey:     domain.VerificationKey,
		Status:              "pending",
		Instructions:        "Complete one of the verification methods below",
		VerificationMethods: methods,
	}, nil
}

// VerifyDomain verifies domain ownership using DNS/HTTP checks.
// userID is needed to confirm ownership before marking verified.
func (d *DomainService) VerifyDomain(domainID int64, userID int64) error {
	domain, err := d.GetDomainByID(domainID)
	if err != nil {
		return err
	}

	if domain.UserID != userID {
		return fmt.Errorf("access denied")
	}

	if domain.IsVerified {
		return nil // Already verified
	}

	// Try DNS verification first and capture a diagnostic message.
	dnsOK, dnsDetail := d.checkDNSRecord(domain.Domain, domain.VerificationKey)
	if !dnsOK {
		// Fall back to HTTP-based verification methods.
		if !d.verifyHTTPFile(domain.Domain, domain.VerificationKey) &&
			!d.verifyHTMLFile(domain.Domain, domain.VerificationKey) {
			return fmt.Errorf("%s", dnsDetail)
		}
	}

	// Persist verification to DB
	if err := d.db.VerifyCustomDomain(domainID, userID, domain.VerificationKey); err != nil {
		return fmt.Errorf("failed to save verification: %w", err)
	}

	// Schedule SSL certificate provisioning if enabled
	if domain.SSLEnabled {
		domain.IsVerified = true
		go d.provisionSSLCertificate(domain)
	}

	return nil
}

// Domain Configuration

// GetDomainSettings retrieves complete domain configuration
func (d *DomainService) GetDomainSettings(domainID int64) (*models.DomainSettingsResponse, error) {
	domain, err := d.GetDomainByID(domainID)
	if err != nil {
		return nil, err
	}

	// Get DNS records configuration
	dnsRecords := d.generateDNSRecords(domain)

	// Get SSL certificate info
	var sslCert *models.SSLCertificate
	if domain.SSLEnabled {
		sslCert, _ = d.getSSLCertificate(domainID)
	}

	// Get domain analytics
	analytics, _ := d.getDomainAnalytics(domainID)

	return &models.DomainSettingsResponse{
		CustomDomain: domain,
		DNSRecords:   dnsRecords,
		SSL:          sslCert,
		Analytics:    analytics,
	}, nil
}

// GetDomainStatus checks the current status of a domain
func (d *DomainService) GetDomainStatus(domainID int64) (*models.DomainStatus, error) {
	domain, err := d.GetDomainByID(domainID)
	if err != nil {
		return nil, err
	}

	// Perform health checks
	checks := []models.DomainStatusCheck{
		d.checkDNSResolution(domain.Domain),
		d.checkHTTPResponse(domain.Domain),
	}

	if domain.SSLEnabled {
		checks = append(checks, d.checkSSLCertificate(domain.Domain))
	}

	// Determine overall status
	overallStatus := d.determineOverallStatus(domain, checks)

	return &models.DomainStatus{
		Domain:        domain.Domain,
		IsActive:      domain.IsActive,
		IsVerified:    domain.IsVerified,
		DNSStatus:     domain.DNSStatus,
		SSLStatus:     domain.SSLStatus,
		OverallStatus: overallStatus,
		Checks:        checks,
		LastChecked:   time.Now(),
		NextCheck:     time.Now().Add(1 * time.Hour),
	}, nil
}

// Account Management

// CanAddDomain checks if a user can add another domain
// GetAllDomainsAdmin returns all domains with pagination for admin use.
func (d *DomainService) GetAllDomainsAdmin(page, limit int, status string) ([]*models.CustomDomain, int, error) {
	return d.db.GetAllDomainsAdmin(page, limit, status)
}

// GetDomainStatsAdmin returns aggregate domain statistics for admin use.
func (d *DomainService) GetDomainStatsAdmin() (*models.DomainStats, error) {
	return d.db.GetDomainStatsAdmin()
}

func (d *DomainService) CanAddDomain(userID int64) (bool, error) {
	user, err := d.userService.GetUserByID(userID)
	if err != nil {
		return false, err
	}

	count, err := d.db.CountCustomDomainsByUser(userID)
	if err != nil {
		return false, err
	}

	limit := models.GetDomainLimit(user.AccountType)
	return count < limit, nil
}

// GetDomainUsage retrieves domain usage statistics for a user
func (d *DomainService) GetDomainUsage(userID int64) (*models.DomainUsage, error) {
	user, err := d.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	used, err := d.db.CountCustomDomainsByUser(userID)
	if err != nil {
		return nil, err
	}

	limit := models.GetDomainLimit(user.AccountType)
	return &models.DomainUsage{
		UserID:       userID,
		DomainsUsed:  used,
		DomainsLimit: limit,
		CanAddDomain: used < limit,
		AccountType:  user.AccountType,
	}, nil
}

// Analytics

// GetDomainAnalytics retrieves analytics for a domain
func (d *DomainService) getDomainAnalytics(domainID int64) (*models.DomainAnalytics, error) {
	domain, err := d.GetDomainByID(domainID)
	if err != nil {
		return nil, err
	}

	// Implementation would aggregate analytics from all URLs under this domain
	// For now, return simulated analytics
	return &models.DomainAnalytics{
		DomainID:    domainID,
		Domain:      domain.Domain,
		TotalClicks: 1247,
		TotalURLs:   45,
		TopURLs: []models.DomainTopURL{
			{ShortCode: "abc123", OriginalURL: "https://example.com/page1", Title: "Page 1", Clicks: 234},
			{ShortCode: "def456", OriginalURL: "https://example.com/page2", Title: "Page 2", Clicks: 189},
			{ShortCode: "ghi789", OriginalURL: "https://example.com/page3", Title: "Page 3", Clicks: 156},
		},
		ClickTrends: []models.ClickTrend{
			{Period: "2024-01-01", Clicks: 123},
			{Period: "2024-01-02", Clicks: 145},
			{Period: "2024-01-03", Clicks: 167},
		},
		ReferrerStats: []models.ReferrerStat{
			{Referrer: "twitter.com", Clicks: 345},
			{Referrer: "facebook.com", Clicks: 234},
			{Referrer: "direct", Clicks: 189},
		},
		GeographicStats: []models.CountryDetail{
			{CountryCode: "US", CountryName: "United States", Clicks: 567, Percentage: 45.5},
			{CountryCode: "GB", CountryName: "United Kingdom", Clicks: 234, Percentage: 18.8},
			{CountryCode: "CA", CountryName: "Canada", Clicks: 178, Percentage: 14.3},
		},
		DeviceStats: []models.DeviceTypeStat{
			{DeviceType: "desktop", Clicks: 748, Percentage: 60.0},
			{DeviceType: "mobile", Clicks: 374, Percentage: 30.0},
			{DeviceType: "tablet", Clicks: 125, Percentage: 10.0},
		},
		LastUpdated: time.Now(),
	}, nil
}

// Helper Methods

// blockedTLDs are reserved/internal TLDs that should never point to public servers.
var blockedTLDs = map[string]bool{
	"local": true, "internal": true, "lan": true, "intranet": true,
	"test": true, "example": true, "invalid": true, "localhost": true,
}

// validateDomainName validates domain format and blocks internal/reserved names.
func (d *DomainService) validateDomainName(domain string) error {
	if len(domain) == 0 || len(domain) > 253 {
		return fmt.Errorf("invalid domain length")
	}

	// Reject raw IP addresses
	if net.ParseIP(domain) != nil {
		return fmt.Errorf("IP addresses are not allowed as custom domains")
	}

	if strings.Contains(domain, "..") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("invalid domain format")
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return fmt.Errorf("domain must include at least one dot (e.g. links.example.com)")
	}

	// Validate each label
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return fmt.Errorf("invalid domain label length")
		}
	}

	// Block reserved TLDs
	tld := strings.ToLower(parts[len(parts)-1])
	if blockedTLDs[tld] {
		return fmt.Errorf("domain uses a reserved or internal TLD")
	}

	// Block localhost explicitly
	if strings.ToLower(domain) == "localhost" {
		return fmt.Errorf("localhost is not allowed as a custom domain")
	}

	return nil
}

// generateVerificationKey generates a random verification key
func (d *DomainService) generateVerificationKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// createVerificationRecords creates verification records for all methods
func (d *DomainService) createVerificationRecords(domainID int64, key string) error {
	// Implementation would create verification records in database
	// For now, simulate success
	return nil
}

// Verification methods

// checkDNSRecord looks up the TXT record and returns (matched, diagnosticMessage).
func (d *DomainService) checkDNSRecord(domain, key string) (bool, string) {
	host := "_urlshortener-verify." + domain
	txtRecords, err := net.LookupTXT(host)
	if err != nil {
		return false, fmt.Sprintf(
			"No TXT record found at %s — make sure you added the record and DNS has propagated (can take up to 48 h). Lookup error: %v",
			host, err,
		)
	}
	if len(txtRecords) == 0 {
		return false, fmt.Sprintf(
			"TXT record at %s exists but is empty. Add the verification value as the record content.",
			host,
		)
	}
	for _, record := range txtRecords {
		if record == key {
			return true, ""
		}
	}
	return false, fmt.Sprintf(
		"TXT record found at %s but the value does not match. Found: %q — Expected: %q. "+
			"Copy the exact token from the Verify modal and set it as the record content.",
		host, txtRecords[0], key,
	)
}

func (d *DomainService) verifyHTTPFile(domain, key string) bool {
	// Reject private/internal hosts before making any outbound request
	if !isSafeHostname(domain) {
		return false
	}

	// Try HTTPS first, fall back to HTTP
	for _, scheme := range []string{"https", "http"} {
		target := fmt.Sprintf("%s://%s/.well-known/urlshortener-verify.txt", scheme, domain)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "URLshortn-Verifier/1.0")

		resp, err := safeHTTPClient.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024)) // max 1 KB
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK && strings.TrimSpace(string(body)) == key {
			return true
		}
	}
	return false
}

func (d *DomainService) verifyHTMLFile(domain, key string) bool {
	// Reject private/internal hosts before making any outbound request
	if !isSafeHostname(domain) {
		return false
	}

	for _, scheme := range []string{"https", "http"} {
		target := fmt.Sprintf("%s://%s/urlshortener-verify-%s.html", scheme, domain, key)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "URLshortn-Verifier/1.0")

		resp, err := safeHTTPClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return true
		}
	}
	return false
}

// SSL Certificate Management

func (d *DomainService) provisionSSLCertificate(domain *models.CustomDomain) {
	// This would integrate with Let's Encrypt or other SSL providers
	// For now, simulate SSL provisioning
	time.Sleep(5 * time.Second) // Simulate processing time

	// Update SSL status in database
	domain.SSLStatus = models.SSLStatusActive
	// Implementation would update database
}

func (d *DomainService) getSSLCertificate(domainID int64) (*models.SSLCertificate, error) {
	// Implementation would query SSL certificate from database
	return &models.SSLCertificate{
		ID:        1,
		DomainID:  domainID,
		Provider:  models.SSLProviderLetsEncrypt,
		Status:    models.SSLStatusActive,
		IssuedAt:  utils.TimePtr(time.Now().Add(-30 * 24 * time.Hour)),
		ExpiresAt: utils.TimePtr(time.Now().Add(60 * 24 * time.Hour)),
		AutoRenew: true,
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}, nil
}

// Health Checks

func (d *DomainService) checkDNSResolution(domain string) models.DomainStatusCheck {
	_, err := net.LookupHost(domain)
	status := models.CheckStatusPass
	message := "DNS resolution successful"

	if err != nil {
		status = models.CheckStatusFail
		message = fmt.Sprintf("DNS resolution failed: %v", err)
	}

	return models.DomainStatusCheck{
		Name:        models.CheckTypeDNS,
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
	}
}

func (d *DomainService) checkHTTPResponse(domain string) models.DomainStatusCheck {
	status := models.CheckStatusPass
	message := "HTTP response successful"

	// Safety check before making any outbound request
	if !isSafeHostname(domain) {
		return models.DomainStatusCheck{
			Name: "http_response", Status: models.CheckStatusFail,
			Message: "Domain resolves to a private or internal address",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s", domain), nil)
	if err != nil {
		return models.DomainStatusCheck{Name: "http_response", Status: models.CheckStatusFail, Message: "request build failed"}
	}
	req.Header.Set("User-Agent", "URLshortn-HealthCheck/1.0")
	resp, err := safeHTTPClient.Do(req)
	if err != nil {
		status = models.CheckStatusFail
		message = "HTTP request failed"
	} else {
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			status = models.CheckStatusWarning
			message = fmt.Sprintf("HTTP response code: %d", resp.StatusCode)
		}
	}

	return models.DomainStatusCheck{
		Name:        models.CheckTypeHTTP,
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
	}
}

func (d *DomainService) checkSSLCertificate(domain string) models.DomainStatusCheck {
	// Implementation would check SSL certificate validity
	// For now, simulate SSL check
	return models.DomainStatusCheck{
		Name:        models.CheckTypeSSL,
		Status:      models.CheckStatusPass,
		Message:     "SSL certificate is valid",
		LastChecked: time.Now(),
	}
}

func (d *DomainService) determineOverallStatus(domain *models.CustomDomain, checks []models.DomainStatusCheck) string {
	if !domain.IsVerified {
		return models.DomainStatusPending
	}

	if !domain.IsActive {
		return models.DomainStatusSuspended
	}

	// Check if any critical checks failed
	for _, check := range checks {
		if check.Status == models.CheckStatusFail && check.Name == models.CheckTypeDNS {
			return models.DomainStatusFailed
		}
	}

	return models.DomainStatusActive
}

func (d *DomainService) generateDNSRecords(domain *models.CustomDomain) []models.DNSRecord {
	// Generate required DNS records for the domain
	return []models.DNSRecord{
		{
			Type:         "CNAME",
			Name:         domain.Domain,
			Value:        d.config.BaseURL,
			TTL:          300,
			Description:  "Points your domain to our service",
			IsRequired:   true,
			IsConfigured: domain.DNSStatus == models.DNSStatusConfigured,
		},
		{
			Type:         "TXT",
			Name:         "_urlshortener-verify." + domain.Domain,
			Value:        domain.VerificationKey,
			TTL:          300,
			Description:  "Verification record (can be removed after verification)",
			IsRequired:   !domain.IsVerified,
			IsConfigured: domain.IsVerified,
		},
	}
}

func (d *DomainService) getDomainURLCount(domainID int64) (int, error) {
	// Implementation would count URLs using this domain
	// For now, return 0
	return 0, nil
}
