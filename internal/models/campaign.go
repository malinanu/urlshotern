package models

import "time"

// ─── DB entities ─────────────────────────────────────────────────────────────

// Campaign represents a promotional campaign.
type Campaign struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	AIContext   string    `json:"ai_context" db:"ai_context"`
	Channels    []string  `json:"channels" db:"channels"` // ["email","sms","whatsapp"]
	Status      string    `json:"status" db:"status"`     // draft | sending | done
	CreatedBy   *int64    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CampaignContact is a single recipient (shared across campaigns).
type CampaignContact struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     *string   `json:"email,omitempty" db:"email"`
	Phone     *string   `json:"phone,omitempty" db:"phone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CampaignSend tracks one send attempt per contact × campaign × channel.
type CampaignSend struct {
	ID         int64      `json:"id" db:"id"`
	CampaignID int64      `json:"campaign_id" db:"campaign_id"`
	ContactID  int64      `json:"contact_id" db:"contact_id"`
	Channel    string     `json:"channel" db:"channel"` // email | sms | whatsapp
	Status     string     `json:"status" db:"status"`   // pending | sent | failed
	Error      *string    `json:"error,omitempty" db:"error"`
	SentAt     *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	JoinedAt   *time.Time `json:"joined_at,omitempty" db:"joined_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// ─── Request / Response DTOs ──────────────────────────────────────────────────

// CreateCampaignRequest is the payload for POST /campaigns.
type CreateCampaignRequest struct {
	Name        string   `json:"name" binding:"required,min=1,max=255"`
	Description string   `json:"description"`
	AIContext   string   `json:"ai_context"`
	Channels    []string `json:"channels" binding:"required,min=1"`
}

// UpdateCampaignRequest is the payload for PUT /campaigns/:id.
type UpdateCampaignRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	AIContext   *string  `json:"ai_context,omitempty"`
	Channels    []string `json:"channels,omitempty"`
}

// ContactImportRow is one row in a contact import payload.
type ContactImportRow struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// CampaignStats summarises delivery and conversion for a campaign.
type CampaignStats struct {
	CampaignID int64 `json:"campaign_id"`
	Total      int64 `json:"total"`
	Sent       int64 `json:"sent"`
	Failed     int64 `json:"failed"`
	Joined     int64 `json:"joined"`
}

// CampaignListResponse wraps paginated campaign results.
type CampaignListResponse struct {
	Campaigns  []Campaign `json:"campaigns"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int64      `json:"total_pages"`
}
