package services

import (
	"fmt"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// GeoIPService handles geographic IP address lookups
type GeoIPService struct {
	cityDB      *geoip2.Reader
	countryDB   *geoip2.Reader
	mutex       sync.RWMutex
	initialized bool
}

// GeoIPResult represents the result of a GeoIP lookup
type GeoIPResult struct {
	CountryCode string   `json:"country_code"`
	CountryName string   `json:"country_name"`
	Region      string   `json:"region,omitempty"`
	City        string   `json:"city,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	TimeZone    string   `json:"timezone,omitempty"`
	ISP         string   `json:"isp,omitempty"`
}

// NewGeoIPService creates a new GeoIP service
func NewGeoIPService() *GeoIPService {
	return &GeoIPService{
		initialized: false,
	}
}

// Initialize loads the GeoIP databases
func (g *GeoIPService) Initialize(cityDBPath, countryDBPath string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Load city database (includes country data)
	if cityDBPath != "" {
		cityDB, err := geoip2.Open(cityDBPath)
		if err != nil {
			return fmt.Errorf("failed to open city GeoIP database: %w", err)
		}
		g.cityDB = cityDB
	}

	// Load country database (fallback if city is not available)
	if countryDBPath != "" {
		countryDB, err := geoip2.Open(countryDBPath)
		if err != nil {
			return fmt.Errorf("failed to open country GeoIP database: %w", err)
		}
		g.countryDB = countryDB
	}

	if g.cityDB == nil && g.countryDB == nil {
		return fmt.Errorf("at least one GeoIP database must be provided")
	}

	g.initialized = true
	return nil
}

// LookupIP performs a GeoIP lookup for the given IP address
func (g *GeoIPService) LookupIP(ipAddress string) (*GeoIPResult, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	if !g.initialized {
		return g.getFallbackResult(ipAddress), nil
	}

	// Parse the IP address
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return g.getFallbackResult(ipAddress), nil
	}

	// Skip private IP addresses
	if g.isPrivateIP(ip) {
		return g.getFallbackResult(ipAddress), nil
	}

	result := &GeoIPResult{}

	// Try city database first (more detailed)
	if g.cityDB != nil {
		if cityRecord, err := g.cityDB.City(ip); err == nil {
			result.CountryCode = cityRecord.Country.IsoCode
			result.CountryName = cityRecord.Country.Names["en"]
			result.TimeZone = cityRecord.Location.TimeZone

			// Region information
			if len(cityRecord.Subdivisions) > 0 {
				result.Region = cityRecord.Subdivisions[0].Names["en"]
			}

			// City information
			result.City = cityRecord.City.Names["en"]

			// Coordinates
			if cityRecord.Location.Latitude != 0 || cityRecord.Location.Longitude != 0 {
				lat := float64(cityRecord.Location.Latitude)
				lng := float64(cityRecord.Location.Longitude)
				result.Latitude = &lat
				result.Longitude = &lng
			}

			return result, nil
		}
	}

	// Fallback to country database
	if g.countryDB != nil {
		if countryRecord, err := g.countryDB.Country(ip); err == nil {
			result.CountryCode = countryRecord.Country.IsoCode
			result.CountryName = countryRecord.Country.Names["en"]
			return result, nil
		}
	}

	// Return fallback result if all lookups fail
	return g.getFallbackResult(ipAddress), nil
}

// LookupIPBatch performs batch GeoIP lookups for multiple IP addresses
func (g *GeoIPService) LookupIPBatch(ipAddresses []string) map[string]*GeoIPResult {
	results := make(map[string]*GeoIPResult)

	for _, ip := range ipAddresses {
		if result, err := g.LookupIP(ip); err == nil {
			results[ip] = result
		} else {
			results[ip] = g.getFallbackResult(ip)
		}
	}

	return results
}

// GetCountryName returns the full country name for a country code
func (g *GeoIPService) GetCountryName(countryCode string) string {
	countryNames := map[string]string{
		"US": "United States",
		"GB": "United Kingdom", 
		"DE": "Germany",
		"FR": "France",
		"CA": "Canada",
		"AU": "Australia",
		"JP": "Japan",
		"IN": "India",
		"BR": "Brazil",
		"ES": "Spain",
		"IT": "Italy",
		"NL": "Netherlands",
		"SE": "Sweden",
		"NO": "Norway",
		"DK": "Denmark",
		"FI": "Finland",
		"PL": "Poland",
		"RU": "Russia",
		"CN": "China",
		"KR": "South Korea",
		"MX": "Mexico",
		"AR": "Argentina",
		"CL": "Chile",
		"CO": "Colombia",
		"VE": "Venezuela",
		"PE": "Peru",
		"ZA": "South Africa",
		"EG": "Egypt",
		"NG": "Nigeria",
		"KE": "Kenya",
		"TH": "Thailand",
		"VN": "Vietnam",
		"MY": "Malaysia",
		"SG": "Singapore",
		"PH": "Philippines",
		"ID": "Indonesia",
		"TR": "Turkey",
		"SA": "Saudi Arabia",
		"AE": "United Arab Emirates",
		"IL": "Israel",
		"GR": "Greece",
		"PT": "Portugal",
		"BE": "Belgium",
		"AT": "Austria",
		"CH": "Switzerland",
		"IE": "Ireland",
		"CZ": "Czech Republic",
		"HU": "Hungary",
		"RO": "Romania",
		"BG": "Bulgaria",
		"HR": "Croatia",
		"SK": "Slovakia",
		"SI": "Slovenia",
		"LT": "Lithuania",
		"LV": "Latvia",
		"EE": "Estonia",
		"IS": "Iceland",
		"UA": "Ukraine",
		"BY": "Belarus",
		"RS": "Serbia",
		"BA": "Bosnia and Herzegovina",
		"MK": "North Macedonia",
		"AL": "Albania",
		"MT": "Malta",
		"CY": "Cyprus",
		"LU": "Luxembourg",
		"MC": "Monaco",
		"SM": "San Marino",
		"VA": "Vatican City",
		"AD": "Andorra",
		"LI": "Liechtenstein",
	}

	if name, exists := countryNames[countryCode]; exists {
		return name
	}
	return countryCode // Return country code if name not found
}

// IsGeoIPEnabled returns whether GeoIP service is initialized and ready
func (g *GeoIPService) IsGeoIPEnabled() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.initialized
}

// Close closes the GeoIP databases and releases resources
func (g *GeoIPService) Close() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	var errs []error

	if g.cityDB != nil {
		if err := g.cityDB.Close(); err != nil {
			errs = append(errs, err)
		}
		g.cityDB = nil
	}

	if g.countryDB != nil {
		if err := g.countryDB.Close(); err != nil {
			errs = append(errs, err)
		}
		g.countryDB = nil
	}

	g.initialized = false

	if len(errs) > 0 {
		return fmt.Errorf("errors closing GeoIP databases: %v", errs)
	}

	return nil
}

// Helper methods

func (g *GeoIPService) isPrivateIP(ip net.IP) bool {
	// Check if IP is in private ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	for _, cidr := range privateRanges {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipnet.Contains(ip) {
			return true
		}
	}

	return false
}

func (g *GeoIPService) getFallbackResult(ipAddress string) *GeoIPResult {
	// Return a basic result for unknown/private IPs
	return &GeoIPResult{
		CountryCode: "XX",
		CountryName: "Unknown",
		Region:      "",
		City:        "",
		Latitude:    nil,
		Longitude:   nil,
		TimeZone:    "UTC",
		ISP:         "Unknown",
	}
}

// GetStatistics returns basic statistics about GeoIP lookups
func (g *GeoIPService) GetStatistics() map[string]interface{} {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	stats := map[string]interface{}{
		"initialized":       g.initialized,
		"city_db_loaded":    g.cityDB != nil,
		"country_db_loaded": g.countryDB != nil,
	}

	if g.cityDB != nil {
		// Get metadata from city database
		metadata := g.cityDB.Metadata()
		stats["city_db_build_time"] = metadata.BuildEpoch
		stats["city_db_description"] = metadata.Description["en"]
	}

	if g.countryDB != nil {
		// Get metadata from country database
		metadata := g.countryDB.Metadata()
		stats["country_db_build_time"] = metadata.BuildEpoch
		stats["country_db_description"] = metadata.Description["en"]
	}

	return stats
}

// ValidateIPAddress validates if a string is a valid IP address
func (g *GeoIPService) ValidateIPAddress(ipAddress string) bool {
	ip := net.ParseIP(ipAddress)
	return ip != nil
}

// GetClientIPFromHeaders extracts the real client IP from HTTP headers
func (g *GeoIPService) GetClientIPFromHeaders(headers map[string]string, fallbackIP string) string {
	// Check common headers for real client IP
	headerKeys := []string{
		"CF-Connecting-IP",     // Cloudflare
		"X-Forwarded-For",      // Standard proxy header
		"X-Real-IP",            // Nginx
		"X-Client-IP",          // Apache
		"X-Forwarded",          // General forwarded
		"X-Cluster-Client-IP",  // Cluster environments
		"Forwarded-For",        // RFC 7239
		"Forwarded",           // RFC 7239
	}

	for _, key := range headerKeys {
		if value, exists := headers[key]; exists && value != "" {
			// X-Forwarded-For can contain multiple IPs, use the first one
			if key == "X-Forwarded-For" {
				ips := parseForwardedHeader(value)
				if len(ips) > 0 && g.ValidateIPAddress(ips[0]) {
					return ips[0]
				}
			} else if g.ValidateIPAddress(value) {
				return value
			}
		}
	}

	return fallbackIP
}

// parseForwardedHeader parses X-Forwarded-For header value
func parseForwardedHeader(value string) []string {
	var ips []string
	parts := splitAndTrim(value, ",")
	
	for _, part := range parts {
		// Remove port if present
		ip := splitAndTrim(part, ":")[0]
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	
	return ips
}

// splitAndTrim splits a string and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	var result []string
	parts := splitString(s, sep)
	
	for _, part := range parts {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}

// splitString splits a string by separator
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	
	var result []string
	start := 0
	
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	
	result = append(result, s[start:])
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)
	
	// Find start
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	
	// Find end
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	
	return s[start:end]
}