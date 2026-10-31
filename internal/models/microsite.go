package models

import (
	"encoding/json"
	"time"
)

// Microsite statuses
const (
	MicrositeStatusDraft     = "draft"
	MicrositeStatusPublished = "published"
)

// Microsite block types (v1)
const (
	BlockTypeProfileHeader  = "profile_header"
	BlockTypeVCardAction    = "vcard_action"
	BlockTypeLinkList       = "link_list"
	BlockTypeContactButtons = "contact_buttons"
	BlockTypeSocialIcons    = "social_icons"
	BlockTypeText           = "text"
	BlockTypeImage          = "image"
	BlockTypeDivider        = "divider"
)

// Microsite telemetry event types
const (
	MicrositeEventView          = "view"
	MicrositeEventClick         = "click"
	MicrositeEventVCardDownload = "vcard_download"
	MicrositeEventCall          = "call"
	MicrositeEventSMS           = "sms"
	MicrositeEventEmail         = "email"
	MicrositeEventWhatsApp      = "whatsapp"
)

// Microsite is a user-owned micro-landing page served at /m/<slug>.
type Microsite struct {
	ID          int64             `json:"id"`
	UserID      int64             `json:"user_id"`
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Template    string            `json:"template"`
	Status      string            `json:"status"`
	Draft       MicrositeContent  `json:"draft"`
	Published   *MicrositeContent `json:"published,omitempty"`
	Version     int               `json:"version"`
	ShortCode   *string           `json:"short_code,omitempty"`
	PublishedAt *time.Time        `json:"published_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// MicrositeSummary is the list-view representation of a microsite.
type MicrositeSummary struct {
	ID          int64      `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Template    string     `json:"template"`
	Status      string     `json:"status"`
	ShortCode   *string    `json:"short_code,omitempty"`
	Views30d    int64      `json:"views_30d"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// MicrositeContent is the editable document: theme, ordered blocks, settings.
type MicrositeContent struct {
	Theme    MicrositeTheme    `json:"theme"`
	Blocks   []MicrositeBlock  `json:"blocks"`
	Settings MicrositeSettings `json:"settings"`
}

// MicrositeTheme is structured design data. CSS is generated server-side from
// validated values; raw CSS is never accepted.
type MicrositeTheme struct {
	Background   MicrositeBackground `json:"background"`
	CardFill     string              `json:"card_fill"`     // #rrggbb
	Accent       string              `json:"accent"`        // #rrggbb
	ButtonText   string              `json:"button_text"`   // #rrggbb, text on accent buttons
	Text         string              `json:"text"`          // #rrggbb
	Font         string              `json:"font"`          // system | serif | mono | rounded | humanist | geometric
	ButtonRadius string              `json:"button_radius"` // sharp | rounded | pill
	Shadow       string              `json:"shadow"`        // none | soft | strong
	AvatarBorder string              `json:"avatar_border"` // none | thin | thick
}

// MicrositeBackground describes the page background.
type MicrositeBackground struct {
	Type     string   `json:"type"`                // solid | linear | radial | image
	Color    string   `json:"color,omitempty"`     // solid / fallback color
	Stops    []string `json:"stops,omitempty"`     // 2-3 colors for gradients
	Angle    int      `json:"angle,omitempty"`     // 0-360, linear gradients
	ImageURL string   `json:"image_url,omitempty"` // image backgrounds
	Overlay  int      `json:"overlay,omitempty"`   // 0-80 (% darkening over image)
}

// MicrositeSettings holds page-level options.
type MicrositeSettings struct {
	HideBadge      bool   `json:"hide_badge"`     // plan-gated (remove "Made with Trunc")
	AllowIndexing  bool   `json:"allow_indexing"` // false → noindex
	SEOTitle       string `json:"seo_title,omitempty"`
	SEODescription string `json:"seo_description,omitempty"`
	OGImageURL     string `json:"og_image_url,omitempty"`
	FaviconURL     string `json:"favicon_url,omitempty"`
}

// MicrositeBlock is one polymorphic block. Data is validated and normalized per
// Type by the block registry before it is stored.
type MicrositeBlock struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	SchemaVersion int             `json:"schema_version"`
	Hidden        bool            `json:"hidden,omitempty"`
	Data          json.RawMessage `json:"data"`
}

// ---- Block data payloads -------------------------------------------------

// ProfileHeaderData is the identity block (avatar, name, title, company).
type ProfileHeaderData struct {
	AvatarURL      string `json:"avatar_url,omitempty"`
	DisplayName    string `json:"display_name"`
	Pronouns       string `json:"pronouns,omitempty"`
	Title          string `json:"title,omitempty"`
	Company        string `json:"company,omitempty"`
	CompanyLogoURL string `json:"company_logo_url,omitempty"`
	Bio            string `json:"bio,omitempty"`
	Location       string `json:"location,omitempty"`
	Verified       bool   `json:"verified,omitempty"`
}

// VCardActionData is the "Save Contact" button and the contact it exports.
type VCardActionData struct {
	ButtonText    string `json:"button_text"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name,omitempty"`
	Organization  string `json:"organization,omitempty"`
	Title         string `json:"title,omitempty"`
	PhoneMobile   string `json:"phone_mobile,omitempty"`
	PhoneWork     string `json:"phone_work,omitempty"`
	EmailWork     string `json:"email_work,omitempty"`
	EmailPersonal string `json:"email_personal,omitempty"`
	Website       string `json:"website,omitempty"`
	Street        string `json:"street,omitempty"`
	City          string `json:"city,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
	Note          string `json:"note,omitempty"`
	PhotoURL      string `json:"photo_url,omitempty"`
}

// LinkListData is a stack of link buttons.
type LinkListData struct {
	Items []LinkItem `json:"items"`
}

// LinkItem is one link button.
type LinkItem struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	URL          string `json:"url"`
	Highlight    bool   `json:"highlight,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// ContactButtonsData is a row of direct-communication buttons.
type ContactButtonsData struct {
	Items []ContactButton `json:"items"`
}

// ContactButton launches tel:/sms:/mailto:/wa.me.
type ContactButton struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"` // call | sms | email | whatsapp
	Value string `json:"value"`
	Label string `json:"label,omitempty"`
}

// SocialIconsData is a grid of social profile icons.
type SocialIconsData struct {
	Items []SocialLink `json:"items"`
}

// SocialLink is one social profile.
type SocialLink struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

// TextData is a plain-text paragraph (no HTML).
type TextData struct {
	Text  string `json:"text"`
	Align string `json:"align,omitempty"` // left | center
}

// ImageData is a single image, optionally linked.
type ImageData struct {
	URL     string `json:"url"`
	Alt     string `json:"alt,omitempty"`
	LinkURL string `json:"link_url,omitempty"`
}

// DividerData is a separator.
type DividerData struct {
	Style string `json:"style,omitempty"` // line | space
}

// ---- Requests / responses -------------------------------------------------

// CreateMicrositeRequest creates a page from a template.
type CreateMicrositeRequest struct {
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	Template string `json:"template"`
}

// UpdateMicrositeRequest saves the draft. Version must match the stored
// version (optimistic concurrency); a mismatch returns 409.
type UpdateMicrositeRequest struct {
	Version int               `json:"version"`
	Title   *string           `json:"title,omitempty"`
	Slug    *string           `json:"slug,omitempty"`
	Content *MicrositeContent `json:"content,omitempty"`
}

// MicrositeTemplateInfo describes a starter template.
type MicrositeTemplateInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// MicrositeEntitlements are the plan-derived capabilities of a page owner.
type MicrositeEntitlements struct {
	MaxPages       int  `json:"max_pages"` // -1 = unlimited
	CanRemoveBadge bool `json:"can_remove_badge"`
	AnalyticsDays  int  `json:"analytics_days"`
	UsedPages      int  `json:"used_pages"`
	CanCreateMore  bool `json:"can_create_more"`
}

// MicrositeEvent is one telemetry row.
type MicrositeEvent struct {
	MicrositeID      int64
	EventType        string
	BlockID          string
	ItemID           string
	VisitorHash      string
	IsUnique         bool
	IsBot            bool
	Country          string
	DeviceType       string
	OS               string
	Browser          string
	ReferrerDomain   string
	ReferrerCategory string
	UTMSource        string
	UTMMedium        string
	UTMCampaign      string
}

// MicrositeAsset is an uploaded, re-encoded image.
type MicrositeAsset struct {
	ID          int64     `json:"id"`
	MicrositeID *int64    `json:"microsite_id,omitempty"`
	URL         string    `json:"url"`
	MimeType    string    `json:"mime_type"`
	SizeBytes   int64     `json:"size_bytes"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	CreatedAt   time.Time `json:"created_at"`
}

// MicrositeAnalytics is the Track-tab report.
type MicrositeAnalytics struct {
	Days             int                    `json:"days"`
	Views            int64                  `json:"views"`
	UniqueVisitors   int64                  `json:"unique_visitors"`
	LinkClicks       int64                  `json:"link_clicks"`
	VCardDownloads   int64                  `json:"vcard_downloads"`
	ContactActions   map[string]int64       `json:"contact_actions"`
	CTR              float64                `json:"ctr"`               // (link clicks + contact actions) / views * 100
	ContactSaveRate  float64                `json:"contact_save_rate"` // vcard downloads / unique views * 100
	Daily            []MicrositeDailyPoint  `json:"daily"`
	TopReferrers     []MicrositeCount       `json:"top_referrers"`
	ReferrerSources  []MicrositeCount       `json:"referrer_sources"`
	Devices          []MicrositeCount       `json:"devices"`
	OperatingSystems []MicrositeCount       `json:"operating_systems"`
	Countries        []MicrositeCount       `json:"countries"`
	Blocks           []MicrositeBlockClicks `json:"blocks"`
}

// MicrositeDailyPoint is one day in the time series.
type MicrositeDailyPoint struct {
	Date           string `json:"date"`
	Views          int64  `json:"views"`
	UniqueVisitors int64  `json:"unique_visitors"`
	Clicks         int64  `json:"clicks"`
	VCardDownloads int64  `json:"vcard_downloads"`
}

// MicrositeCount is a labelled count.
type MicrositeCount struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// MicrositeBlockClicks is the click distribution for one block item.
type MicrositeBlockClicks struct {
	BlockID   string `json:"block_id"`
	ItemID    string `json:"item_id,omitempty"`
	EventType string `json:"event_type"`
	Label     string `json:"label"`
	Count     int64  `json:"count"`
}
