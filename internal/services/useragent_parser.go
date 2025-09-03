package services

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/mssola/user_agent"
)

// UserAgentService handles user agent string parsing
type UserAgentService struct {
	// Add any configuration if needed
}

// DeviceInfo represents parsed device information from user agent
type DeviceInfo struct {
	DeviceType       string `json:"device_type"`       // mobile, tablet, desktop
	DeviceBrand      string `json:"device_brand"`      // Apple, Samsung, etc.
	DeviceModel      string `json:"device_model"`      // iPhone, Galaxy S21, etc.
	OSName           string `json:"os_name"`           // iOS, Android, Windows, macOS, Linux
	OSVersion        string `json:"os_version"`        // 14.5, 11, etc.
	BrowserName      string `json:"browser_name"`      // Chrome, Safari, Firefox, Edge
	BrowserVersion   string `json:"browser_version"`   // 91.0.4472.124
	BrowserEngine    string `json:"browser_engine"`    // WebKit, Gecko, Blink
	ScreenResolution string `json:"screen_resolution"` // 1920x1080 (if detectable)
	UserAgentHash    string `json:"user_agent_hash"`   // SHA-256 hash for uniqueness
	IsBot            bool   `json:"is_bot"`            // Whether this is a bot/crawler
	IsMobile         bool   `json:"is_mobile"`         // Whether this is a mobile device
	IsTablet         bool   `json:"is_tablet"`         // Whether this is a tablet
}

// BotInfo represents information about detected bots
type BotInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // search_engine, social_media, seo, monitoring, etc.
	Description string `json:"description"`
}

// NewUserAgentService creates a new user agent parsing service
func NewUserAgentService() *UserAgentService {
	return &UserAgentService{}
}

// ParseUserAgent parses a user agent string and extracts device information
func (u *UserAgentService) ParseUserAgent(userAgentString string) *DeviceInfo {
	if userAgentString == "" {
		return u.getDefaultDeviceInfo()
	}

	// Parse using the user_agent library
	ua := user_agent.New(userAgentString)

	deviceInfo := &DeviceInfo{
		UserAgentHash: u.hashUserAgent(userAgentString),
		IsMobile:      ua.Mobile(),
		IsBot:         ua.Bot(),
	}

	// Determine device type
	deviceInfo.DeviceType = u.getDeviceType(ua, userAgentString)
	deviceInfo.IsTablet = deviceInfo.DeviceType == "tablet"

	// Get operating system information
	osName := ua.OS()
	deviceInfo.OSName = u.normalizeOSName(osName)
	deviceInfo.OSVersion = "" // Version info not available in this method

	// Get browser information
	browserName, browserVersion := ua.Browser()
	deviceInfo.BrowserName = u.normalizeBrowserName(browserName)
	deviceInfo.BrowserVersion = browserVersion

	// Get browser engine
	deviceInfo.BrowserEngine = u.getBrowserEngine(ua, userAgentString)

	// Extract device brand and model
	deviceInfo.DeviceBrand, deviceInfo.DeviceModel = u.extractDeviceInfo(userAgentString, deviceInfo.OSName)

	// Try to extract screen resolution if available
	deviceInfo.ScreenResolution = u.extractScreenResolution(userAgentString)

	return deviceInfo
}

// ParseBatch parses multiple user agent strings
func (u *UserAgentService) ParseBatch(userAgentStrings []string) []*DeviceInfo {
	results := make([]*DeviceInfo, len(userAgentStrings))
	for i, uaString := range userAgentStrings {
		results[i] = u.ParseUserAgent(uaString)
	}
	return results
}

// DetectBot detects if a user agent is a bot and provides bot information
func (u *UserAgentService) DetectBot(userAgentString string) *BotInfo {
	if userAgentString == "" {
		return nil
	}

	ua := user_agent.New(userAgentString)
	if !ua.Bot() {
		return nil
	}

	// Common bot patterns
	botPatterns := map[string]*BotInfo{
		"Googlebot":        {"Googlebot", "search_engine", "Google's web crawler"},
		"Bingbot":          {"Bingbot", "search_engine", "Microsoft Bing's web crawler"},
		"Slurp":            {"Yahoo! Slurp", "search_engine", "Yahoo's web crawler"},
		"DuckDuckBot":      {"DuckDuckBot", "search_engine", "DuckDuckGo's web crawler"},
		"Baiduspider":      {"Baiduspider", "search_engine", "Baidu's web crawler"},
		"YandexBot":        {"YandexBot", "search_engine", "Yandex's web crawler"},
		"facebookexternalhit": {"Facebook Bot", "social_media", "Facebook's link preview crawler"},
		"Twitterbot":       {"Twitterbot", "social_media", "Twitter's link preview crawler"},
		"LinkedInBot":      {"LinkedInBot", "social_media", "LinkedIn's link preview crawler"},
		"WhatsApp":         {"WhatsApp Bot", "social_media", "WhatsApp's link preview crawler"},
		"TelegramBot":      {"TelegramBot", "social_media", "Telegram's link preview crawler"},
		"DiscordBot":       {"DiscordBot", "social_media", "Discord's link preview crawler"},
		"Applebot":         {"Applebot", "search_engine", "Apple's web crawler for Siri and Spotlight"},
		"AhrefsBot":        {"AhrefsBot", "seo", "Ahrefs SEO tool crawler"},
		"MJ12bot":          {"MJ12bot", "seo", "Majestic SEO crawler"},
		"SemrushBot":       {"SemrushBot", "seo", "SEMrush SEO tool crawler"},
		"DotBot":           {"DotBot", "seo", "Moz SEO tool crawler"},
		"UptimeRobot":      {"UptimeRobot", "monitoring", "Website uptime monitoring service"},
		"Pingdom":          {"Pingdom", "monitoring", "Website monitoring service"},
		"StatusCake":       {"StatusCake", "monitoring", "Website monitoring service"},
		"Postman":          {"Postman", "development", "API development and testing tool"},
		"curl":             {"cURL", "development", "Command-line HTTP client"},
		"wget":             {"wget", "development", "Command-line web retrieval tool"},
		"python-requests":  {"Python Requests", "development", "Python HTTP library"},
	}

	userAgentLower := strings.ToLower(userAgentString)
	
	for pattern, botInfo := range botPatterns {
		if strings.Contains(userAgentLower, strings.ToLower(pattern)) {
			return botInfo
		}
	}

	// Generic bot detection
	return &BotInfo{
		Name:        "Unknown Bot",
		Type:        "unknown",
		Description: "Unidentified bot or crawler",
	}
}

// Helper methods

func (u *UserAgentService) getDeviceType(ua *user_agent.UserAgent, userAgentString string) string {
	// Check for tablet first (more specific than mobile)
	if u.isTablet(userAgentString) {
		return "tablet"
	}
	
	if ua.Mobile() {
		return "mobile"
	}
	
	return "desktop"
}

func (u *UserAgentService) isTablet(userAgentString string) bool {
	userAgentLower := strings.ToLower(userAgentString)
	tabletKeywords := []string{
		"ipad", "tablet", "kindle", "playbook", "nexus 7", "nexus 9", "nexus 10",
		"xoom", "sch-i800", "galaxy tab", "gt-p1000", "gt-p1010", "gt-p7510",
		"fonepad", "transformer", "tf101", "tf201", "tf300t", "tf700t", "tf701t",
		"surface", "sm-t", "sm-p", "pixel c",
	}
	
	for _, keyword := range tabletKeywords {
		if strings.Contains(userAgentLower, keyword) {
			return true
		}
	}
	
	return false
}

func (u *UserAgentService) normalizeOSName(osName string) string {
	osName = strings.ToLower(osName)
	
	if strings.Contains(osName, "windows") {
		return "Windows"
	} else if strings.Contains(osName, "mac") || strings.Contains(osName, "darwin") {
		return "macOS"
	} else if strings.Contains(osName, "ios") {
		return "iOS"
	} else if strings.Contains(osName, "android") {
		return "Android"
	} else if strings.Contains(osName, "linux") {
		return "Linux"
	} else if strings.Contains(osName, "ubuntu") {
		return "Ubuntu"
	} else if strings.Contains(osName, "fedora") {
		return "Fedora"
	} else if strings.Contains(osName, "centos") {
		return "CentOS"
	} else if osName == "" {
		return "Unknown"
	}
	
	return strings.Title(osName)
}

func (u *UserAgentService) normalizeBrowserName(browserName string) string {
	browserName = strings.ToLower(browserName)
	
	if strings.Contains(browserName, "chrome") {
		return "Chrome"
	} else if strings.Contains(browserName, "firefox") {
		return "Firefox"
	} else if strings.Contains(browserName, "safari") {
		return "Safari"
	} else if strings.Contains(browserName, "edge") {
		return "Edge"
	} else if strings.Contains(browserName, "opera") {
		return "Opera"
	} else if strings.Contains(browserName, "internet explorer") || strings.Contains(browserName, "ie") {
		return "Internet Explorer"
	} else if strings.Contains(browserName, "brave") {
		return "Brave"
	} else if strings.Contains(browserName, "vivaldi") {
		return "Vivaldi"
	} else if browserName == "" {
		return "Unknown"
	}
	
	return strings.Title(browserName)
}

func (u *UserAgentService) getBrowserEngine(ua *user_agent.UserAgent, userAgentString string) string {
	userAgentLower := strings.ToLower(userAgentString)
	
	if strings.Contains(userAgentLower, "webkit") {
		if strings.Contains(userAgentLower, "blink") || strings.Contains(userAgentLower, "chrome") {
			return "Blink"
		}
		return "WebKit"
	} else if strings.Contains(userAgentLower, "gecko") {
		return "Gecko"
	} else if strings.Contains(userAgentLower, "trident") {
		return "Trident"
	} else if strings.Contains(userAgentLower, "edgehtml") {
		return "EdgeHTML"
	} else if strings.Contains(userAgentLower, "presto") {
		return "Presto"
	}
	
	return "Unknown"
}

func (u *UserAgentService) extractDeviceInfo(userAgentString, osName string) (brand, model string) {
	userAgentLower := strings.ToLower(userAgentString)
	
	// iOS devices
	if osName == "iOS" {
		if strings.Contains(userAgentLower, "iphone") {
			return "Apple", u.extractiPhoneModel(userAgentString)
		} else if strings.Contains(userAgentLower, "ipad") {
			return "Apple", u.extractiPadModel(userAgentString)
		} else if strings.Contains(userAgentLower, "ipod") {
			return "Apple", "iPod Touch"
		}
		return "Apple", "Unknown iOS Device"
	}
	
	// Android devices
	if osName == "Android" {
		return u.extractAndroidDevice(userAgentString)
	}
	
	return "", ""
}

func (u *UserAgentService) extractiPhoneModel(userAgentString string) string {
	// Extract iPhone model from user agent
	re := regexp.MustCompile(`iPhone OS (\d+)_(\d+)`)
	if matches := re.FindStringSubmatch(userAgentString); len(matches) > 2 {
		return fmt.Sprintf("iPhone (iOS %s.%s)", matches[1], matches[2])
	}
	return "iPhone"
}

func (u *UserAgentService) extractiPadModel(userAgentString string) string {
	// Extract iPad model from user agent
	re := regexp.MustCompile(`OS (\d+)_(\d+)`)
	if matches := re.FindStringSubmatch(userAgentString); len(matches) > 2 {
		return fmt.Sprintf("iPad (iOS %s.%s)", matches[1], matches[2])
	}
	return "iPad"
}

func (u *UserAgentService) extractAndroidDevice(userAgentString string) (brand, model string) {
	// Common Android device patterns
	devicePatterns := map[string]string{
		"samsung":  "Samsung",
		"huawei":   "Huawei",
		"xiaomi":   "Xiaomi",
		"oppo":     "OPPO",
		"vivo":     "Vivo",
		"oneplus":  "OnePlus",
		"lg":       "LG",
		"sony":     "Sony",
		"htc":      "HTC",
		"motorola": "Motorola",
		"google":   "Google",
		"nokia":    "Nokia",
	}
	
	userAgentLower := strings.ToLower(userAgentString)
	
	for pattern, brandName := range devicePatterns {
		if strings.Contains(userAgentLower, pattern) {
			// Try to extract more specific model information
			model = u.extractAndroidModel(userAgentString, pattern)
			if model == "" {
				model = "Unknown " + brandName + " Device"
			}
			return brandName, model
		}
	}
	
	return "Unknown", "Android Device"
}

func (u *UserAgentService) extractAndroidModel(userAgentString, brand string) string {
	// This is a simplified model extraction - in practice you'd have more sophisticated patterns
	userAgentLower := strings.ToLower(userAgentString)
	
	// Look for model patterns after the brand name
	re := regexp.MustCompile(brand + `[^;)]*([A-Z0-9\-]+)`)
	if matches := re.FindStringSubmatch(userAgentLower); len(matches) > 1 {
		return matches[1]
	}
	
	return ""
}

func (u *UserAgentService) extractScreenResolution(userAgentString string) string {
	// Try to extract screen resolution if available in user agent
	re := regexp.MustCompile(`(\d{3,4})[x×](\d{3,4})`)
	if matches := re.FindStringSubmatch(userAgentString); len(matches) > 2 {
		return fmt.Sprintf("%sx%s", matches[1], matches[2])
	}
	return ""
}

func (u *UserAgentService) hashUserAgent(userAgent string) string {
	hash := sha256.Sum256([]byte(userAgent))
	return fmt.Sprintf("%x", hash)
}

func (u *UserAgentService) getDefaultDeviceInfo() *DeviceInfo {
	return &DeviceInfo{
		DeviceType:       "Unknown",
		DeviceBrand:      "",
		DeviceModel:      "",
		OSName:           "Unknown",
		OSVersion:        "",
		BrowserName:      "Unknown",
		BrowserVersion:   "",
		BrowserEngine:    "Unknown",
		ScreenResolution: "",
		UserAgentHash:    u.hashUserAgent(""),
		IsBot:            false,
		IsMobile:         false,
		IsTablet:         false,
	}
}

// GetStatistics returns statistics about parsed user agents
func (u *UserAgentService) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"service": "User Agent Parser",
		"library": "mssola/user_agent",
		"features": []string{
			"Device type detection",
			"OS and browser parsing",
			"Bot detection",
			"Mobile and tablet detection",
			"Brand and model extraction",
		},
	}
}