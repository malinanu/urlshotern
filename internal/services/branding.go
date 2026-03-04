package services

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// BrandingStore is the minimal storage interface needed by BrandingService.
type BrandingStore interface {
	GetBrandSettings(userID int64) (*models.BrandSettings, error)
	SaveBrandSettings(b *models.BrandSettings) error
}

// BrandingService manages enterprise white-label brand settings.
type BrandingService struct {
	db BrandingStore
}

// NewBrandingService creates a new BrandingService.
func NewBrandingService(db *storage.PostgresStorage) *BrandingService {
	return &BrandingService{db: db}
}

var hexColorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// dangerousCSS matches patterns that could be used for CSS injection / XSS.
var dangerousCSS = regexp.MustCompile(
	`(?i)(expression\s*\(|javascript\s*:|vbscript\s*:|@import\b|behavior\s*:|` +
		`<\s*script|binding\s*:|moz-binding\s*:|data\s*:text|url\s*\(\s*['"]?\s*(?:javascript|data|vbscript))`,
)

// validateHTTPSURL ensures a URL is a non-empty absolute HTTPS URL pointing to a
// public (non-private) host. An empty string is allowed (field cleared).
func validateHTTPSURL(raw, fieldName string) error {
	if raw == "" {
		return nil
	}
	if len(raw) > 2048 {
		return fmt.Errorf("%s: URL too long", fieldName)
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("%s: invalid URL", fieldName)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("%s: only https:// URLs are allowed", fieldName)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("%s: missing host", fieldName)
	}
	// Block private/internal hosts (reuse the same helper from domain service)
	if !isSafeHostname(host) {
		return fmt.Errorf("%s: URL points to a private or internal host", fieldName)
	}
	return nil
}

// GetSettings returns brand settings for a user.
func (s *BrandingService) GetSettings(userID int64) (*models.BrandSettings, error) {
	return s.db.GetBrandSettings(userID)
}

// UpdateSettings validates and saves brand settings.
func (s *BrandingService) UpdateSettings(userID int64, req *models.UpdateBrandSettingsRequest) (*models.BrandSettings, error) {
	// Validate hex colors
	if req.PrimaryColor != "" && !hexColorRe.MatchString(req.PrimaryColor) {
		return nil, fmt.Errorf("invalid primary_color: must be a 6-digit hex color like #3b82f6")
	}
	if req.AccentColor != "" && !hexColorRe.MatchString(req.AccentColor) {
		return nil, fmt.Errorf("invalid accent_color: must be a 6-digit hex color like #10b981")
	}

	// Validate URLs — HTTPS only, public hosts only
	if err := validateHTTPSURL(req.LogoURL, "logo_url"); err != nil {
		return nil, err
	}
	if err := validateHTTPSURL(req.FaviconURL, "favicon_url"); err != nil {
		return nil, err
	}

	// Validate custom CSS — block dangerous patterns
	if req.CustomCSS != "" {
		if dangerousCSS.MatchString(req.CustomCSS) {
			return nil, fmt.Errorf("custom_css contains disallowed patterns (expressions, imports, or scripting)")
		}
		if len(req.CustomCSS) > 32768 { // 32 KB max
			return nil, fmt.Errorf("custom_css exceeds maximum allowed size (32 KB)")
		}
	}

	// Validate company name length
	if len(req.CompanyName) > 100 {
		return nil, fmt.Errorf("company_name must be 100 characters or fewer")
	}
	req.CompanyName = strings.TrimSpace(req.CompanyName)

	b := &models.BrandSettings{
		UserID:       userID,
		CompanyName:  req.CompanyName,
		LogoURL:      req.LogoURL,
		PrimaryColor: req.PrimaryColor,
		AccentColor:  req.AccentColor,
		CustomCSS:    req.CustomCSS,
		FaviconURL:   req.FaviconURL,
	}
	if b.PrimaryColor == "" {
		b.PrimaryColor = "#3b82f6"
	}
	if b.AccentColor == "" {
		b.AccentColor = "#10b981"
	}

	if err := s.db.SaveBrandSettings(b); err != nil {
		return nil, fmt.Errorf("save brand settings: %w", err)
	}
	return b, nil
}
