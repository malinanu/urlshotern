package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
)

type DomainService struct {
	db              *storage.PostgresStorage
	redis           *storage.RedisStorage
	config          *configs.Config
	analyticsService *AnalyticsService
	userService     *UserService
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
		ID:              time.Now().Unix(), // Temporary ID generation
		Domain:          strings.ToLower(req.Domain),
		UserID:          userID,
		IsActive:        false, // Inactive until verified
		IsVerified:      false,
		VerificationKey: verificationKey,
		SSLEnabled:      true, // Default to SSL enabled
		SSLStatus:       models.SSLStatusPending,
		DNSStatus:       models.DNSStatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save to database
	// Implementation would save domain to database

	// Create verification records
	err = d.createVerificationRecords(domain.ID, verificationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create verification records: %w", err)
	}

	return domain, nil
}

// GetUserDomains retrieves all domains for a user
func (d *DomainService) GetUserDomains(userID int64) ([]*models.CustomDomain, error) {
	// Implementation would query database
	// For now, return simulated domains
	domains := []*models.CustomDomain{
		{
			ID:         1,
			Domain:     "short.example.com",
			UserID:     userID,
			IsActive:   true,
			IsVerified: true,
			SSLEnabled: true,
			SSLStatus:  models.SSLStatusActive,
			DNSStatus:  models.DNSStatusConfigured,
			CreatedAt:  time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
		},
	}

	return domains, nil
}

// GetDomainByID retrieves a domain by ID
func (d *DomainService) GetDomainByID(domainID int64) (*models.CustomDomain, error) {
	// Implementation would query database
	// For now, return simulated domain
	return &models.CustomDomain{
		ID:         domainID,
		Domain:     "short.example.com",
		UserID:     1,
		IsActive:   true,
		IsVerified: true,
		SSLEnabled: true,
		SSLStatus:  models.SSLStatusActive,
		DNSStatus:  models.DNSStatusConfigured,
		CreatedAt:  time.Now().Add(-7 * 24 * time.Hour),
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
	}, nil
}

// GetDomainByName retrieves a domain by domain name
func (d *DomainService) GetDomainByName(domainName string) (*models.CustomDomain, error) {
	// Implementation would query database by domain name
	// For now, return nil for non-existing domains
	if domainName == "short.example.com" {
		return &models.CustomDomain{
			ID:         1,
			Domain:     "short.example.com",
			UserID:     1,
			IsActive:   true,
			IsVerified: true,
			SSLEnabled: true,
			SSLStatus:  models.SSLStatusActive,
			DNSStatus:  models.DNSStatusConfigured,
			CreatedAt:  time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
		}, nil
	}
	
	return nil, fmt.Errorf("domain not found")
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

	// Save changes to database
	// Implementation would update database

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

	// Delete domain from database
	// Implementation would delete domain and related records

	return nil
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
			Domain:            domain.Domain,
			Status:            "verified",
			Instructions:      "Domain is already verified",
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
		Status:              "pending",
		Instructions:        "Complete one of the verification methods below",
		VerificationMethods: methods,
	}, nil
}

// VerifyDomain verifies domain ownership
func (d *DomainService) VerifyDomain(domainID int64) error {
	domain, err := d.GetDomainByID(domainID)
	if err != nil {
		return err
	}

	if domain.IsVerified {
		return nil // Already verified
	}

	// Try different verification methods
	verified := false

	// Try DNS verification
	if d.verifyDNSRecord(domain.Domain, domain.VerificationKey) {
		verified = true
	} else if d.verifyHTTPFile(domain.Domain, domain.VerificationKey) {
		// Try HTTP file verification
		verified = true
	} else if d.verifyHTMLFile(domain.Domain, domain.VerificationKey) {
		// Try HTML file verification
		verified = true
	}

	if !verified {
		return fmt.Errorf("domain verification failed")
	}

	// Update domain status
	domain.IsVerified = true
	domain.IsActive = true // Activate domain after verification
	domain.DNSStatus = models.DNSStatusConfigured
	domain.UpdatedAt = time.Now()

	// Save to database
	// Implementation would update database

	// Schedule SSL certificate provisioning if enabled
	if domain.SSLEnabled {
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
func (d *DomainService) CanAddDomain(userID int64) (bool, error) {
	// Get user account type
	user, err := d.userService.GetUserByID(userID)
	if err != nil {
		return false, err
	}

	// Count current domains
	domains, err := d.GetUserDomains(userID)
	if err != nil {
		return false, err
	}

	currentCount := len(domains)
	limit := models.GetDomainLimit(user.AccountType)

	return currentCount < limit, nil
}

// GetDomainUsage retrieves domain usage statistics for a user
func (d *DomainService) GetDomainUsage(userID int64) (*models.DomainUsage, error) {
	user, err := d.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	domains, err := d.GetUserDomains(userID)
	if err != nil {
		return nil, err
	}

	limit := models.GetDomainLimit(user.AccountType)
	used := len(domains)

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

// validateDomainName validates domain format
func (d *DomainService) validateDomainName(domain string) error {
	if len(domain) == 0 || len(domain) > 255 {
		return fmt.Errorf("invalid domain length")
	}

	if strings.Contains(domain, "..") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("invalid domain format")
	}

	// Basic domain validation (in production, use more robust validation)
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return fmt.Errorf("domain must have at least one subdomain and TLD")
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

func (d *DomainService) verifyDNSRecord(domain, key string) bool {
	// Check for TXT record at _urlshortener-verify subdomain
	txtRecords, err := net.LookupTXT("_urlshortener-verify." + domain)
	if err != nil {
		return false
	}

	for _, record := range txtRecords {
		if record == key {
			return true
		}
	}
	return false
}

func (d *DomainService) verifyHTTPFile(domain, key string) bool {
	// Check for verification file at /.well-known/urlshortener-verify.txt
	url := fmt.Sprintf("http://%s/.well-known/urlshortener-verify.txt", domain)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	// Read and compare content (simplified)
	// In production, would properly read and validate content
	return true
}

func (d *DomainService) verifyHTMLFile(domain, key string) bool {
	// Check for HTML file with verification key in filename
	url := fmt.Sprintf("http://%s/urlshortener-verify-%s.html", domain, key)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
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
	url := fmt.Sprintf("http://%s", domain)
	resp, err := http.Get(url)
	status := models.CheckStatusPass
	message := "HTTP response successful"
	
	if err != nil {
		status = models.CheckStatusFail
		message = fmt.Sprintf("HTTP request failed: %v", err)
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

