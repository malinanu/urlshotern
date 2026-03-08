package models

import (
	"time"

	"github.com/lib/pq"
)

// Coupon represents a discount code for subscriptions
type Coupon struct {
	ID                 int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Code               string         `json:"code" gorm:"uniqueIndex;not null;type:varchar(50)"`
	Name               string         `json:"name" gorm:"not null;type:varchar(100)"`
	Description        string         `json:"description" gorm:"type:text"`
	DiscountType       string         `json:"discount_type" gorm:"not null;type:varchar(20)"` // percentage, fixed
	DiscountValue      int            `json:"discount_value" gorm:"not null"` // number of months, percentage (1-100), or fixed amount in cents
	Currency           string         `json:"currency" gorm:"default:'usd';type:varchar(3)"`
	MaxRedemptions     *int           `json:"max_redemptions,omitempty"`
	CurrentRedemptions int            `json:"current_redemptions" gorm:"default:0"`
	ValidFrom          time.Time      `json:"valid_from" gorm:"default:CURRENT_TIMESTAMP"`
	ValidUntil         *time.Time     `json:"valid_until,omitempty"`
	ApplicablePlans    pq.StringArray `json:"applicable_plans,omitempty" gorm:"type:text[]"`
	IsActive           bool           `json:"is_active" gorm:"default:true;index"`
	PolarDiscountID    *string        `json:"polar_discount_id,omitempty" gorm:"type:varchar(255);index"`
	CreatedAt          time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// CouponRedemption represents a user's usage of a coupon
type CouponRedemption struct {
	ID             int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	CouponID       int64      `json:"coupon_id" gorm:"not null;index"`
	UserID         int64      `json:"user_id" gorm:"not null;index"`
	SubscriptionID *int64     `json:"subscription_id,omitempty"`
	RedeemedAt     time.Time  `json:"redeemed_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships
	Coupon         *Coupon    `json:"coupon,omitempty" gorm:"foreignKey:CouponID"`
	User           *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for Coupon
func (Coupon) TableName() string {
	return "coupons"
}

// TableName specifies the table name for CouponRedemption
func (CouponRedemption) TableName() string {
	return "coupon_redemptions"
}
