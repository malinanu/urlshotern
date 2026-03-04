package models

import "time"

// BrandSettings holds enterprise white-label customisation for a user account.
type BrandSettings struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	CompanyName  string    `json:"company_name" db:"company_name"`
	LogoURL      string    `json:"logo_url" db:"logo_url"`
	PrimaryColor string    `json:"primary_color" db:"primary_color"`
	AccentColor  string    `json:"accent_color" db:"accent_color"`
	CustomCSS    string    `json:"custom_css" db:"custom_css"`
	FaviconURL   string    `json:"favicon_url" db:"favicon_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateBrandSettingsRequest is the JSON body for PUT /api/v1/branding.
type UpdateBrandSettingsRequest struct {
	CompanyName  string `json:"company_name"`
	LogoURL      string `json:"logo_url"`
	PrimaryColor string `json:"primary_color"`
	AccentColor  string `json:"accent_color"`
	CustomCSS    string `json:"custom_css"`
	FaviconURL   string `json:"favicon_url"`
}
