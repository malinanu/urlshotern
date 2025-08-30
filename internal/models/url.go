package models

import (
	"time"
)

// URLMapping represents a URL mapping in the database
type URLMapping struct {
	ID          int64     `json:"id" db:"id"`
	ShortCode   string    `json:"short_code" db:"short_code"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	ClickCount  int64     `json:"click_count" db:"click_count"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedByIP string    `json:"created_by_ip,omitempty" db:"created_by_ip"`
}

// ClickEvent represents a click tracking event
type ClickEvent struct {
	ID          int64     `json:"id" db:"id"`
	ShortCode   string    `json:"short_code" db:"short_code"`
	ClickedAt   time.Time `json:"clicked_at" db:"clicked_at"`
	IPAddress   string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   string    `json:"user_agent,omitempty" db:"user_agent"`
	Referrer    string    `json:"referrer,omitempty" db:"referrer"`
	CountryCode string    `json:"country_code,omitempty" db:"country_code"`
}

// ShortenRequest represents the request payload for shortening URLs
type ShortenRequest struct {
	URL       string     `json:"url" binding:"required,url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CustomCode string    `json:"custom_code,omitempty"`
}

// ShortenResponse represents the response for URL shortening
type ShortenResponse struct {
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// AnalyticsResponse represents analytics data for a short URL
type AnalyticsResponse struct {
	ShortCode    string    `json:"short_code"`
	OriginalURL  string    `json:"original_url"`
	TotalClicks  int64     `json:"total_clicks"`
	CreatedAt    time.Time `json:"created_at"`
	LastClickAt  *time.Time `json:"last_click_at,omitempty"`
	DailyClicks  []DailyClick `json:"daily_clicks,omitempty"`
	CountryStats []CountryStat `json:"country_stats,omitempty"`
}

// DailyClick represents click statistics per day
type DailyClick struct {
	Date   string `json:"date"`
	Clicks int64  `json:"clicks"`
}

// CountryStat represents click statistics per country
type CountryStat struct {
	CountryCode string `json:"country_code"`
	Clicks      int64  `json:"clicks"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// DashboardStats represents overall dashboard statistics
type DashboardStats struct {
	TotalURLs   int64    `json:"total_urls"`
	TotalClicks int64    `json:"total_clicks"`
	TopURLs     []TopURL `json:"top_urls"`
}

// TopURL represents a top performing URL
type TopURL struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	Clicks      int64  `json:"clicks"`
}

// ClickTrend represents click trends over time
type ClickTrend struct {
	Period string `json:"period"`
	Clicks int64  `json:"clicks"`
}

// ReferrerStat represents referrer statistics
type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Clicks   int64  `json:"clicks"`
}

// UserAgentStats represents user agent statistics
type UserAgentStats struct {
	Browsers []BrowserStat `json:"browsers"`
	Devices  []DeviceStat  `json:"devices"`
	OS       []OSStat      `json:"operating_systems"`
}

// BrowserStat represents browser statistics
type BrowserStat struct {
	Browser string `json:"browser"`
	Clicks  int64  `json:"clicks"`
}

// DeviceStat represents device statistics
type DeviceStat struct {
	Device string `json:"device"`
	Clicks int64  `json:"clicks"`
}

// OSStat represents operating system statistics
type OSStat struct {
	OS     string `json:"operating_system"`
	Clicks int64  `json:"clicks"`
}