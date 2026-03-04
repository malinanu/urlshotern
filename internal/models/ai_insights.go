package models

import "time"

// AIInsight represents a generated AI insight for a short link.
type AIInsight struct {
	ID           int64      `json:"id" db:"id"`
	ShortCode    string     `json:"short_code" db:"short_code"`
	UserID       *int64     `json:"user_id,omitempty" db:"user_id"`
	InsightType  string     `json:"insight_type" db:"insight_type"`
	Content      string     `json:"content" db:"content"`
	DataSnapshot []byte     `json:"data_snapshot,omitempty" db:"data_snapshot"`
	GeneratedAt  time.Time  `json:"generated_at" db:"generated_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	IsRead       bool       `json:"is_read" db:"is_read"`
}
