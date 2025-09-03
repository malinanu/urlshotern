package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// FreeGeoIPService handles geographic IP lookups using free services
type FreeGeoIPService struct {
	client      *http.Client
	cache       map[string]*GeoIPResult
	cacheMutex  sync.RWMutex
	cacheExpiry time.Duration
}

// IPAPIResponse represents response from ip-api.com (free service)
type IPAPIResponse struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
}

// NewFreeGeoIPService creates a new free GeoIP service
func NewFreeGeoIPService() *FreeGeoIPService {
	return &FreeGeoIPService{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cache:       make(map[string]*GeoIPResult),
		cacheExpiry: 24 * time.Hour, // Cache results for 24 hours
	}
}

// LookupIP performs a free GeoIP lookup using ip-api.com
// Note: ip-api.com allows 1000 requests per hour for non-commercial use
func (f *FreeGeoIPService) LookupIP(ipAddress string) (*GeoIPResult, error) {
	// Check cache first
	f.cacheMutex.RLock()
	if cached, exists := f.cache[ipAddress]; exists {
		f.cacheMutex.RUnlock()
		return cached, nil
	}
	f.cacheMutex.RUnlock()

	// Parse IP and check if it's private
	ip := net.ParseIP(ipAddress)
	if ip == nil || f.isPrivateIP(ip) {
		return f.getFallbackResult(ipAddress), nil
	}

	// Try multiple free services as fallbacks
	result := f.tryIPAPI(ipAddress)
	if result == nil {
		result = f.tryIPInfo(ipAddress)
	}
	if result == nil {
		result = f.tryIPStack(ipAddress) // Free tier available
	}
	if result == nil {
		result = f.getFallbackResult(ipAddress)
	}

	// Cache the result
	f.cacheMutex.Lock()
	f.cache[ipAddress] = result
	f.cacheMutex.Unlock()

	return result, nil
}

// tryIPAPI tries ip-api.com (completely free)
func (f *FreeGeoIPService) tryIPAPI(ipAddress string) *GeoIPResult {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ipAddress)
	
	resp, err := f.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var apiResp IPAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil
	}

	if apiResp.Status != "success" {
		return nil
	}

	result := &GeoIPResult{
		CountryCode: apiResp.CountryCode,
		CountryName: apiResp.Country,
		Region:      apiResp.RegionName,
		City:        apiResp.City,
		TimeZone:    apiResp.Timezone,
		ISP:         apiResp.ISP,
	}

	if apiResp.Lat != 0 || apiResp.Lon != 0 {
		result.Latitude = &apiResp.Lat
		result.Longitude = &apiResp.Lon
	}

	return result
}

// tryIPInfo tries ipinfo.io (50,000 requests/month free)
func (f *FreeGeoIPService) tryIPInfo(ipAddress string) *GeoIPResult {
	url := fmt.Sprintf("https://ipinfo.io/%s/json", ipAddress)
	
	resp, err := f.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	result := &GeoIPResult{
		CountryCode: f.getString(data, "country"),
		City:        f.getString(data, "city"),
		Region:      f.getString(data, "region"),
		TimeZone:    f.getString(data, "timezone"),
		ISP:         f.getString(data, "org"),
	}

	// Get country name from country code
	result.CountryName = f.getCountryName(result.CountryCode)

	// Parse location coordinates if available
	if loc := f.getString(data, "loc"); loc != "" {
		coords := strings.Split(loc, ",")
		if len(coords) == 2 {
			if lat, err := parseFloat(coords[0]); err == nil {
				if lng, err := parseFloat(coords[1]); err == nil {
					result.Latitude = &lat
					result.Longitude = &lng
				}
			}
		}
	}

	return result
}

// tryIPStack tries ipstack.com (free tier: 10,000 requests/month)
func (f *FreeGeoIPService) tryIPStack(ipAddress string) *GeoIPResult {
	// Note: You would need to sign up for a free API key at ipstack.com
	// For demo purposes, this is a placeholder implementation
	// url := fmt.Sprintf("http://api.ipstack.com/%s?access_key=YOUR_FREE_API_KEY", ipAddress)
	
	// For now, return nil to fall back to other services
	return nil
}

// LookupIPBatch performs batch lookups with rate limiting for free services
func (f *FreeGeoIPService) LookupIPBatch(ipAddresses []string) map[string]*GeoIPResult {
	results := make(map[string]*GeoIPResult)
	
	// Implement rate limiting for free services (max 45 requests per minute for ip-api)
	rateLimiter := time.NewTicker(1500 * time.Millisecond) // ~40 requests per minute
	defer rateLimiter.Stop()

	for i, ip := range ipAddresses {
		if i > 0 {
			<-rateLimiter.C // Wait for rate limiter
		}
		
		if result, err := f.LookupIP(ip); err == nil {
			results[ip] = result
		} else {
			results[ip] = f.getFallbackResult(ip)
		}
	}

	return results
}

// Helper methods
func (f *FreeGeoIPService) isPrivateIP(ip net.IP) bool {
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

func (f *FreeGeoIPService) getFallbackResult(ipAddress string) *GeoIPResult {
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

func (f *FreeGeoIPService) getCountryName(countryCode string) string {
	countryNames := map[string]string{
		"US": "United States", "GB": "United Kingdom", "DE": "Germany", "FR": "France",
		"CA": "Canada", "AU": "Australia", "JP": "Japan", "IN": "India", "BR": "Brazil",
		"ES": "Spain", "IT": "Italy", "NL": "Netherlands", "SE": "Sweden", "NO": "Norway",
		"DK": "Denmark", "FI": "Finland", "PL": "Poland", "RU": "Russia", "CN": "China",
		"KR": "South Korea", "MX": "Mexico", "AR": "Argentina", "CL": "Chile", "CO": "Colombia",
		"VE": "Venezuela", "PE": "Peru", "ZA": "South Africa", "EG": "Egypt", "NG": "Nigeria",
		"KE": "Kenya", "TH": "Thailand", "VN": "Vietnam", "MY": "Malaysia", "SG": "Singapore",
		"PH": "Philippines", "ID": "Indonesia", "TR": "Turkey", "SA": "Saudi Arabia",
		"AE": "United Arab Emirates", "IL": "Israel", "GR": "Greece", "PT": "Portugal",
		"BE": "Belgium", "AT": "Austria", "CH": "Switzerland", "IE": "Ireland", "CZ": "Czech Republic",
	}
	
	if name, exists := countryNames[countryCode]; exists {
		return name
	}
	return countryCode
}

func (f *FreeGeoIPService) getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// parseFloat is a simple float parser
func parseFloat(s string) (float64, error) {
	var result float64
	var sign float64 = 1
	i := 0
	
	// Handle negative numbers
	if len(s) > 0 && s[0] == '-' {
		sign = -1
		i = 1
	} else if len(s) > 0 && s[0] == '+' {
		i = 1
	}
	
	// Parse integer part
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		result = result*10 + float64(s[i]-'0')
		i++
	}
	
	// Parse decimal part
	if i < len(s) && s[i] == '.' {
		i++
		decimal := float64(0)
		factor := float64(0.1)
		
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			decimal += float64(s[i]-'0') * factor
			factor *= 0.1
			i++
		}
		result += decimal
	}
	
	return result * sign, nil
}

// ClearCache clears the IP lookup cache
func (f *FreeGeoIPService) ClearCache() {
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()
	f.cache = make(map[string]*GeoIPResult)
}

// GetCacheSize returns the current cache size
func (f *FreeGeoIPService) GetCacheSize() int {
	f.cacheMutex.RLock()
	defer f.cacheMutex.RUnlock()
	return len(f.cache)
}

// IsRateLimited checks if we should rate limit requests
func (f *FreeGeoIPService) IsRateLimited() bool {
	// Simple rate limiting: allow max 900 requests per hour (safe margin for ip-api)
	return false // Implement actual rate limiting based on your needs
}