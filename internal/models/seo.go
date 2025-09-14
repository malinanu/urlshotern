package models

import "time"

// MetaTag represents a meta tag for SEO
type MetaTag struct {
	ID        int64     `json:"id" db:"id"`
	PageID    *int64    `json:"page_id" db:"page_id"`
	Name      string    `json:"name" db:"name"`
	Content   string    `json:"content" db:"content"`
	Property  *string   `json:"property,omitempty" db:"property"`
	CreatedBy int64     `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedBy *int64    `json:"updated_by" db:"updated_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PageSEO represents SEO-specific data for a page
type PageSEO struct {
	ID              int64     `json:"id" db:"id"`
	PageID          int64     `json:"page_id" db:"page_id"`
	MetaTitle       string    `json:"meta_title" db:"meta_title"`
	MetaDescription string    `json:"meta_description" db:"meta_description"`
	MetaKeywords    string    `json:"meta_keywords" db:"meta_keywords"`
	CanonicalURL    string    `json:"canonical_url" db:"canonical_url"`
	OGTitle         string    `json:"og_title" db:"og_title"`
	OGDescription   string    `json:"og_description" db:"og_description"`
	OGImage         string    `json:"og_image" db:"og_image"`
	TwitterCard     string    `json:"twitter_card" db:"twitter_card"`
	TwitterTitle    string    `json:"twitter_title" db:"twitter_title"`
	TwitterDesc     string    `json:"twitter_description" db:"twitter_description"`
	TwitterImage    string    `json:"twitter_image" db:"twitter_image"`
	SchemaMarkup    string    `json:"schema_markup" db:"schema_markup"`
	NoIndex         bool      `json:"no_index" db:"no_index"`
	NoFollow        bool      `json:"no_follow" db:"no_follow"`
	CreatedBy       int64     `json:"created_by" db:"created_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedBy       int64     `json:"updated_by" db:"updated_by"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// URLRedirect represents a URL redirect for SEO purposes
type URLRedirect struct {
	ID           int64     `json:"id" db:"id"`
	SourceURL    string    `json:"source_url" db:"source_url"`
	TargetURL    string    `json:"target_url" db:"target_url"`
	RedirectType int       `json:"redirect_type" db:"redirect_type"` // 301, 302, etc.
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedBy    int64     `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedBy    *int64    `json:"updated_by" db:"updated_by"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// SEOAnalysis represents an SEO analysis result
type SEOAnalysisDB struct {
	ID               int64     `json:"id" db:"id"`
	PageID           *int64    `json:"page_id" db:"page_id"`
	URL              string    `json:"url" db:"url"`
	Title            string    `json:"title" db:"title"`
	MetaDescription  string    `json:"meta_description" db:"meta_description"`
	H1Tags           string    `json:"h1_tags" db:"h1_tags"`           // JSON array
	H2Tags           string    `json:"h2_tags" db:"h2_tags"`           // JSON array
	ImageAltCount    int       `json:"image_alt_count" db:"image_alt_count"`
	WordCount        int       `json:"word_count" db:"word_count"`
	InternalLinks    int       `json:"internal_links" db:"internal_links"`
	ExternalLinks    int       `json:"external_links" db:"external_links"`
	MobileFriendly   bool      `json:"mobile_friendly" db:"mobile_friendly"`
	LoadingSpeed     float64   `json:"loading_speed" db:"loading_speed"`
	SSLEnabled       bool      `json:"ssl_enabled" db:"ssl_enabled"`
	SchemaMarkup     string    `json:"schema_markup" db:"schema_markup"` // JSON object
	Score            int       `json:"score" db:"score"`
	Issues           string    `json:"issues" db:"issues"`                     // JSON array
	Recommendations  string    `json:"recommendations" db:"recommendations"`   // JSON array
	CreatedBy        int64     `json:"created_by" db:"created_by"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// RobotsTxt represents the robots.txt configuration
type RobotsTxt struct {
	ID        int64     `json:"id" db:"id"`
	Content   string    `json:"content" db:"content"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedBy int64     `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedBy int64     `json:"updated_by" db:"updated_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}