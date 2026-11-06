package models

import (
	"time"
)

// MicroPage represents a digital vCard & micro landing page in the database
type MicroPage struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Slug        string    `json:"slug" db:"slug"`
	Title       string    `json:"title" db:"title"`
	Bio         string    `json:"bio" db:"bio"`
	AvatarURL   string    `json:"avatar_url" db:"avatar_url"`
	CoverURL    string    `json:"cover_url" db:"cover_url"`
	Theme       string    `json:"theme" db:"theme"` // executive-obsidian, california-minimalist, cyber-neon, nordic-sage, swiss-grid, sunset-glass, cobalt-steel, editorial-monocle, neo-brutalist, aura-holographic
	CustomCSS   string    `json:"custom_css" db:"custom_css"`
	Config      string    `json:"config" db:"config"`           // JSON string for layout & visual preferences
	VCardData   string    `json:"vcard_data" db:"vcard_data"`   // JSON string of VCardContact
	SocialLinks string    `json:"social_links" db:"social_links"` // JSON array of MicroSocialLink
	Links       string    `json:"links" db:"links"`             // JSON array of MicroLinkItem
	IsPublished bool      `json:"is_published" db:"is_published"`
	ViewCount   int64     `json:"view_count" db:"view_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// VCardContact represents structured contact information for generating vCard 3.0
type VCardContact struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Organization string `json:"organization"`
	Title        string `json:"title"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	CellPhone    string `json:"cell_phone"`
	WorkPhone    string `json:"work_phone"`
	Address      string `json:"address"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	Website      string `json:"website"`
	Note         string `json:"note"`
}

// MicroSocialLink represents a social profile icon link
type MicroSocialLink struct {
	Platform string `json:"platform"` // twitter, linkedin, github, instagram, youtube, whatsapp, telegram, email, phone, website
	URL      string `json:"url"`
	Label    string `json:"label,omitempty"`
}

// MicroLinkItem represents an interactive action button or link
type MicroLinkItem struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	URL           string `json:"url"`
	Icon          string `json:"icon,omitempty"`
	Description   string `json:"description,omitempty"`
	IsHighlighted bool   `json:"is_highlighted,omitempty"`
	IsActive      bool   `json:"is_active"`
	SortOrder     int    `json:"sort_order"`
	ClickCount    int64  `json:"click_count,omitempty"`
}

// MicroPageResponse is the client-facing fully-parsed structure
type MicroPageResponse struct {
	ID          int64             `json:"id"`
	UserID      int64             `json:"user_id"`
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Bio         string            `json:"bio"`
	AvatarURL   string            `json:"avatar_url"`
	CoverURL    string            `json:"cover_url"`
	Theme       string            `json:"theme"`
	CustomCSS   string            `json:"custom_css"`
	Config      map[string]any    `json:"config"`
	VCard       *VCardContact     `json:"vcard"`
	SocialLinks []MicroSocialLink `json:"social_links"`
	Links       []MicroLinkItem   `json:"links"`
	IsPublished bool              `json:"is_published"`
	ViewCount   int64             `json:"view_count"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CreateMicroPageRequest represents the payload to create a new micro-page
type CreateMicroPageRequest struct {
	Slug        string            `json:"slug" binding:"required"`
	Title       string            `json:"title" binding:"required"`
	Bio         string            `json:"bio"`
	AvatarURL   string            `json:"avatar_url"`
	CoverURL    string            `json:"cover_url"`
	Theme       string            `json:"theme"`
	CustomCSS   string            `json:"custom_css"`
	Config      map[string]any    `json:"config"`
	VCard       *VCardContact     `json:"vcard"`
	SocialLinks []MicroSocialLink `json:"social_links"`
	Links       []MicroLinkItem   `json:"links"`
	IsPublished bool              `json:"is_published"`
}

// UpdateMicroPageRequest represents the payload to update an existing micro-page
type UpdateMicroPageRequest struct {
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Bio         *string           `json:"bio"`
	AvatarURL   *string           `json:"avatar_url"`
	CoverURL    *string           `json:"cover_url"`
	Theme       string            `json:"theme"`
	CustomCSS   *string           `json:"custom_css"`
	Config      map[string]any    `json:"config"`
	VCard       *VCardContact     `json:"vcard"`
	SocialLinks []MicroSocialLink `json:"social_links"`
	Links       []MicroLinkItem   `json:"links"`
	IsPublished *bool             `json:"is_published"`
}
