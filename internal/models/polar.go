package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// PolarSubscription represents a subscription managed by Polar
type PolarSubscription struct {
	ID                  int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID              int64      `json:"user_id" gorm:"not null;index;uniqueIndex:idx_user_polar_sub"`
	PolarCustomerID     string     `json:"polar_customer_id" gorm:"not null;index"`
	PolarSubscriptionID *string    `json:"polar_subscription_id,omitempty" gorm:"index;uniqueIndex:idx_user_polar_sub"`
	PolarProductID      string     `json:"polar_product_id" gorm:"not null"`
	PolarPriceID        *string    `json:"polar_price_id,omitempty"`
	PolarCheckoutID     *string    `json:"polar_checkout_id,omitempty"`
	Status              string     `json:"status" gorm:"not null;index"` // active, cancelled, incomplete, incomplete_expired, trialing, past_due
	PlanType            string     `json:"plan_type" gorm:"not null"`    // free, premium, enterprise
	BillingCycle        *string    `json:"billing_cycle,omitempty"`      // month, year
	CurrentPeriodStart  *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd    *time.Time `json:"current_period_end,omitempty"`
	CancelAt            *time.Time `json:"cancel_at,omitempty"`
	CancelAtPeriodEnd   bool       `json:"cancel_at_period_end" gorm:"default:false"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	EndedAt             *time.Time `json:"ended_at,omitempty"`
	TrialStart          *time.Time `json:"trial_start,omitempty"`
	TrialEnd            *time.Time `json:"trial_end,omitempty"`
	Benefits            BenefitList `json:"benefits" gorm:"type:jsonb"`
	Metadata            JSONB       `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// PolarProduct represents a product/plan in Polar (cached locally)
type PolarProduct struct {
	ID               string       `json:"id" gorm:"primaryKey"`
	PolarProductID   string       `json:"polar_product_id" gorm:"uniqueIndex;not null"`
	Name             string       `json:"name" gorm:"not null"`
	Description      string       `json:"description"`
	PlanType         string       `json:"plan_type" gorm:"not null"` // free, premium, enterprise
	IsRecurring      bool         `json:"is_recurring" gorm:"default:true"`
	Prices           PriceList    `json:"prices" gorm:"type:jsonb"`
	Features         StringList   `json:"features" gorm:"type:jsonb"`
	Benefits         BenefitList  `json:"benefits" gorm:"type:jsonb"`
	OrganizationID   string       `json:"organization_id" gorm:"not null"`
	IsArchived       bool         `json:"is_archived" gorm:"default:false"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	LastSyncedAt     time.Time    `json:"last_synced_at" gorm:"autoUpdateTime"`
}

// PolarWebhookEvent stores webhook events for debugging and idempotency
type PolarWebhookEvent struct {
	ID               int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	PolarEventID     string    `json:"polar_event_id" gorm:"uniqueIndex;not null"`
	EventType        string    `json:"event_type" gorm:"not null;index"`
	Payload          JSONB     `json:"payload" gorm:"type:jsonb;not null"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty" gorm:"index"`
	ProcessingStatus string    `json:"processing_status" gorm:"default:'pending'"` // pending, processed, failed
	ErrorMessage     *string   `json:"error_message,omitempty"`
	RetryCount       int       `json:"retry_count" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// PolarPrice represents a price point for a product
type PolarPrice struct {
	ID              string  `json:"id"`
	PriceAmount     int     `json:"price_amount"`      // Amount in cents
	PriceCurrency   string  `json:"price_currency"`    // USD, EUR, etc.
	RecurringInterval string `json:"recurring_interval,omitempty"` // month, year
	IsArchived      bool    `json:"is_archived"`
}

// PolarBenefit represents a benefit granted to a subscription
type PolarBenefit struct {
	ID          string `json:"id"`
	Type        string `json:"type"`        // custom, articles, discord, github_repository, etc.
	Description string `json:"description"`
	Properties  JSONB  `json:"properties,omitempty"`
}

// Custom types for JSONB handling

// JSONB is a custom type for handling JSONB columns
type JSONB map[string]interface{}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	result := make(JSONB)
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// BenefitList is a list of Polar benefits
type BenefitList []PolarBenefit

// Scan implements the sql.Scanner interface
func (b *BenefitList) Scan(value interface{}) error {
	if value == nil {
		*b = []PolarBenefit{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	result := []PolarBenefit{}
	err := json.Unmarshal(bytes, &result)
	*b = result
	return err
}

// Value implements the driver.Valuer interface
func (b BenefitList) Value() (driver.Value, error) {
	if len(b) == 0 {
		return json.Marshal([]PolarBenefit{})
	}
	return json.Marshal(b)
}

// PriceList is a list of Polar prices
type PriceList []PolarPrice

// Scan implements the sql.Scanner interface
func (p *PriceList) Scan(value interface{}) error {
	if value == nil {
		*p = []PolarPrice{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	result := []PolarPrice{}
	err := json.Unmarshal(bytes, &result)
	*p = result
	return err
}

// Value implements the driver.Valuer interface
func (p PriceList) Value() (driver.Value, error) {
	if len(p) == 0 {
		return json.Marshal([]PolarPrice{})
	}
	return json.Marshal(p)
}

// StringList is a list of strings
type StringList []string

// Scan implements the sql.Scanner interface
func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	result := []string{}
	err := json.Unmarshal(bytes, &result)
	*s = result
	return err
}

// Value implements the driver.Valuer interface
func (s StringList) Value() (driver.Value, error) {
	if len(s) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

// Helper methods

// IsActive returns true if the subscription is currently active
func (ps *PolarSubscription) IsActive() bool {
	return ps.Status == "active" || ps.Status == "trialing"
}

// IsCancelled returns true if the subscription has been cancelled
func (ps *PolarSubscription) IsCancelled() bool {
	return ps.Status == "cancelled" || ps.CancelAtPeriodEnd
}

// IsExpired returns true if the subscription has ended
func (ps *PolarSubscription) IsExpired() bool {
	if ps.CurrentPeriodEnd == nil {
		return false
	}
	return time.Now().After(*ps.CurrentPeriodEnd) && !ps.IsActive()
}

// HasBenefit checks if the subscription includes a specific benefit type
func (ps *PolarSubscription) HasBenefit(benefitType string) bool {
	for _, benefit := range ps.Benefits {
		if benefit.Type == benefitType {
			return true
		}
	}
	return false
}

// TableName specifies the table name for PolarSubscription
func (PolarSubscription) TableName() string {
	return "polar_subscriptions"
}

// TableName specifies the table name for PolarProduct
func (PolarProduct) TableName() string {
	return "polar_products"
}

// TableName specifies the table name for PolarWebhookEvent
func (PolarWebhookEvent) TableName() string {
	return "polar_webhook_events"
}
