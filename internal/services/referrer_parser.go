package services

import (
	"net/url"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

// ReferrerParsingService handles parsing referrer URLs and extracting UTM parameters
type ReferrerParsingService struct {
	// Known referrer patterns for categorization
	socialMediaDomains map[string]string
	searchEngines      map[string]string
	emailProviders     map[string]string
}

// UTMParameters represents parsed UTM parameters
type UTMParameters struct {
	Source     string `json:"utm_source"`
	Medium     string `json:"utm_medium"`
	Campaign   string `json:"utm_campaign"`
	Term       string `json:"utm_term"`
	Content    string `json:"utm_content"`
}

// ParsedReferrer represents a parsed referrer with categorization
type ParsedReferrer struct {
	OriginalReferrer string         `json:"original_referrer"`
	Domain           string         `json:"domain"`
	Category         string         `json:"category"` // social, search, email, direct, other
	Platform         string         `json:"platform"` // specific platform name
	UTMParams        *UTMParameters `json:"utm_params,omitempty"`
	IsOrganic        bool           `json:"is_organic"`
	SearchQuery      string         `json:"search_query,omitempty"`
}

// NewReferrerParsingService creates a new referrer parsing service
func NewReferrerParsingService() *ReferrerParsingService {
	return &ReferrerParsingService{
		socialMediaDomains: map[string]string{
			"facebook.com":    "Facebook",
			"fb.com":          "Facebook",
			"m.facebook.com":  "Facebook",
			"twitter.com":     "Twitter/X",
			"x.com":           "Twitter/X",
			"t.co":            "Twitter/X",
			"instagram.com":   "Instagram",
			"linkedin.com":    "LinkedIn",
			"youtube.com":     "YouTube",
			"youtu.be":        "YouTube",
			"tiktok.com":      "TikTok",
			"snapchat.com":    "Snapchat",
			"pinterest.com":   "Pinterest",
			"reddit.com":      "Reddit",
			"discord.gg":      "Discord",
			"telegram.org":    "Telegram",
			"whatsapp.com":    "WhatsApp",
			"wa.me":           "WhatsApp",
		},
		searchEngines: map[string]string{
			"google.com":     "Google",
			"google.co.uk":   "Google",
			"google.ca":      "Google",
			"google.de":      "Google",
			"google.fr":      "Google",
			"bing.com":       "Bing",
			"yahoo.com":      "Yahoo",
			"duckduckgo.com": "DuckDuckGo",
			"yandex.com":     "Yandex",
			"baidu.com":      "Baidu",
			"ask.com":        "Ask",
		},
		emailProviders: map[string]string{
			"mail.google.com": "Gmail",
			"outlook.com":     "Outlook",
			"yahoo.com":       "Yahoo Mail",
			"aol.com":         "AOL Mail",
			"mail.yahoo.com":  "Yahoo Mail",
		},
	}
}

// ParseReferrer parses a referrer URL and extracts meaningful information
func (r *ReferrerParsingService) ParseReferrer(referrerURL string) *ParsedReferrer {
	if referrerURL == "" {
		return &ParsedReferrer{
			Category: "direct",
			Platform: "Direct Traffic",
		}
	}

	parsed := &ParsedReferrer{
		OriginalReferrer: referrerURL,
		IsOrganic:        true,
	}

	// Parse the URL
	u, err := url.Parse(referrerURL)
	if err != nil {
		parsed.Category = "other"
		parsed.Platform = "Unknown"
		return parsed
	}

	parsed.Domain = strings.ToLower(u.Host)
	
	// Remove www. prefix for matching
	domain := strings.TrimPrefix(parsed.Domain, "www.")
	
	// Parse UTM parameters
	parsed.UTMParams = r.parseUTMParameters(u.Query())
	if parsed.UTMParams != nil && parsed.UTMParams.hasAnyParam() {
		parsed.IsOrganic = false
	}

	// Categorize referrer
	parsed.Category, parsed.Platform = r.categorizeReferrer(domain)
	
	// Extract search query if from search engine
	if parsed.Category == "search" {
		parsed.SearchQuery = r.extractSearchQuery(u)
	}

	return parsed
}

// parseUTMParameters extracts UTM parameters from URL query
func (r *ReferrerParsingService) parseUTMParameters(values url.Values) *UTMParameters {
	utm := &UTMParameters{
		Source:   values.Get("utm_source"),
		Medium:   values.Get("utm_medium"),
		Campaign: values.Get("utm_campaign"),
		Term:     values.Get("utm_term"),
		Content:  values.Get("utm_content"),
	}

	// Return nil if no UTM parameters are present
	if !utm.hasAnyParam() {
		return nil
	}

	return utm
}

// hasAnyParam checks if UTM parameters has any non-empty values
func (u *UTMParameters) hasAnyParam() bool {
	return u.Source != "" || u.Medium != "" || u.Campaign != "" || u.Term != "" || u.Content != ""
}

// categorizeReferrer categorizes a referrer domain
func (r *ReferrerParsingService) categorizeReferrer(domain string) (category, platform string) {
	// Check social media
	if platform, exists := r.socialMediaDomains[domain]; exists {
		return "social", platform
	}

	// Check search engines
	if platform, exists := r.searchEngines[domain]; exists {
		return "search", platform
	}

	// Check email providers
	if platform, exists := r.emailProviders[domain]; exists {
		return "email", platform
	}

	// Check for common patterns
	if r.isNewsWebsite(domain) {
		return "news", r.extractDomainName(domain)
	}

	if r.isForumWebsite(domain) {
		return "forum", r.extractDomainName(domain)
	}

	// Default to other
	return "other", r.extractDomainName(domain)
}

// extractSearchQuery extracts search query from search engine URLs
func (r *ReferrerParsingService) extractSearchQuery(u *url.URL) string {
	values := u.Query()
	
	// Google uses 'q' parameter
	if query := values.Get("q"); query != "" {
		return query
	}
	
	// Bing uses 'q' parameter
	if query := values.Get("q"); query != "" {
		return query
	}
	
	// Yahoo uses 'p' parameter
	if query := values.Get("p"); query != "" {
		return query
	}
	
	// DuckDuckGo uses 'q' parameter
	if query := values.Get("q"); query != "" {
		return query
	}

	return ""
}

// isNewsWebsite checks if domain is likely a news website
func (r *ReferrerParsingService) isNewsWebsite(domain string) bool {
	newsPatterns := []string{
		"news", "bbc", "cnn", "reuters", "bloomberg", "techcrunch",
		"hackernews", "ycombinator", "medium", "substack",
	}

	for _, pattern := range newsPatterns {
		if strings.Contains(domain, pattern) {
			return true
		}
	}

	return false
}

// isForumWebsite checks if domain is likely a forum
func (r *ReferrerParsingService) isForumWebsite(domain string) bool {
	forumPatterns := []string{
		"reddit", "stackoverflow", "stackexchange", "quora",
		"forum", "discuss", "community",
	}

	for _, pattern := range forumPatterns {
		if strings.Contains(domain, pattern) {
			return true
		}
	}

	return false
}

// extractDomainName extracts a readable name from domain
func (r *ReferrerParsingService) extractDomainName(domain string) string {
	// Remove TLD for cleaner name
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		name := parts[len(parts)-2]
		// Capitalize first letter
		if len(name) > 0 {
			return strings.ToUpper(name[:1]) + name[1:]
		}
		return name
	}
	return domain
}

// Enhanced models for referrer tracking
type EnhancedReferrerData struct {
	ShortCode        string         `json:"short_code"`
	OriginalReferrer string         `json:"original_referrer"`
	ParsedReferrer   *ParsedReferrer `json:"parsed_referrer"`
	IPAddress        string         `json:"ip_address"`
	UserAgent        string         `json:"user_agent"`
	Timestamp        time.Time      `json:"timestamp"`
}

// ReferrerAnalytics represents aggregated referrer analytics
type ReferrerAnalytics struct {
	ShortCode      string                    `json:"short_code"`
	TotalClicks    int64                     `json:"total_clicks"`
	Referrers      []models.ReferrerStat     `json:"referrers"`
	Categories     []ReferrerCategoryStat    `json:"categories"`
	Campaigns      []CampaignStat            `json:"campaigns"`
	TopDomains     []DomainStat              `json:"top_domains"`
	UTMBreakdown   UTMBreakdown              `json:"utm_breakdown"`
	OrganicVsPaid  OrganicVsPaidStat         `json:"organic_vs_paid"`
}

// ReferrerCategoryStat represents statistics by referrer category
type ReferrerCategoryStat struct {
	Category   string  `json:"category"`
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

// CampaignStat represents UTM campaign statistics
type CampaignStat struct {
	Campaign   string  `json:"campaign"`
	Source     string  `json:"source"`
	Medium     string  `json:"medium"`
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

// DomainStat represents top referring domains
type DomainStat struct {
	Domain     string  `json:"domain"`
	Platform   string  `json:"platform"`
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

// UTMBreakdown represents breakdown of UTM parameters
type UTMBreakdown struct {
	Sources   []UTMParamStat `json:"sources"`
	Mediums   []UTMParamStat `json:"mediums"`
	Campaigns []UTMParamStat `json:"campaigns"`
}

// UTMParamStat represents statistics for UTM parameter values
type UTMParamStat struct {
	Value      string  `json:"value"`
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

// OrganicVsPaidStat represents organic vs paid traffic breakdown
type OrganicVsPaidStat struct {
	OrganicClicks   int64   `json:"organic_clicks"`
	PaidClicks      int64   `json:"paid_clicks"`
	OrganicPercent  float64 `json:"organic_percent"`
	PaidPercent     float64 `json:"paid_percent"`
}

// ProcessReferrerData processes and stores referrer data
func (r *ReferrerParsingService) ProcessReferrerData(shortCode, referrerURL, ipAddress, userAgent string) (*EnhancedReferrerData, error) {
	parsedReferrer := r.ParseReferrer(referrerURL)
	
	enhancedData := &EnhancedReferrerData{
		ShortCode:        shortCode,
		OriginalReferrer: referrerURL,
		ParsedReferrer:   parsedReferrer,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Timestamp:        time.Now(),
	}

	return enhancedData, nil
}

// GetReferrerInsights provides insights about referrer patterns
func (r *ReferrerParsingService) GetReferrerInsights(referrers []models.ReferrerStat) map[string]interface{} {
	insights := make(map[string]interface{})
	
	// Analyze top referrer categories
	categoryMap := make(map[string]int64)
	totalClicks := int64(0)
	
	for _, ref := range referrers {
		parsed := r.ParseReferrer(ref.Referrer)
		categoryMap[parsed.Category] += ref.Clicks
		totalClicks += ref.Clicks
	}
	
	// Find dominant referrer category
	var dominantCategory string
	var maxClicks int64
	for category, clicks := range categoryMap {
		if clicks > maxClicks {
			maxClicks = clicks
			dominantCategory = category
		}
	}
	
	insights["dominant_category"] = dominantCategory
	insights["dominant_category_percentage"] = float64(maxClicks) / float64(totalClicks) * 100
	insights["total_categories"] = len(categoryMap)
	
	// Analyze organic vs paid
	organicClicks := categoryMap["search"] + categoryMap["social"] + categoryMap["direct"]
	paidClicks := totalClicks - organicClicks
	
	if totalClicks > 0 {
		insights["organic_percentage"] = float64(organicClicks) / float64(totalClicks) * 100
		insights["paid_percentage"] = float64(paidClicks) / float64(totalClicks) * 100
	}
	
	return insights
}