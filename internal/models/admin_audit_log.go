package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

var (
	// ErrInvalidJSONB is returned when JSONB data cannot be processed
	ErrInvalidJSONB = errors.New("invalid JSONB data")
)

// AdminAuditLog represents an admin action audit log entry
type AdminAuditLog struct {
	ID            int64           `json:"id" db:"id"`
	UserID        *int64          `json:"user_id,omitempty" db:"user_id"`
	Username      string          `json:"username" db:"username"`
	Action        string          `json:"action" db:"action"`
	ResourceType  string          `json:"resource_type,omitempty" db:"resource_type"`
	ResourceID    string          `json:"resource_id,omitempty" db:"resource_id"`
	Details       JSONB           `json:"details,omitempty" db:"details"`
	Changes       *AuditLogChange `json:"changes,omitempty" db:"changes"`
	IPAddress     string          `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     string          `json:"user_agent,omitempty" db:"user_agent"`
	SessionID     string          `json:"session_id,omitempty" db:"session_id"`
	Status        string          `json:"status" db:"status"` // success, failed, unauthorized
	ErrorMessage  string          `json:"error_message,omitempty" db:"error_message"`
	Severity      string          `json:"severity" db:"severity"` // info, warning, critical
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// AuditLogChange represents before/after values for an audit log entry
type AuditLogChange struct {
	Before map[string]interface{} `json:"before,omitempty"`
	After  map[string]interface{} `json:"after,omitempty"`
}

// Value implements the driver.Valuer interface for AuditLogChange
func (a AuditLogChange) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for AuditLogChange
func (a *AuditLogChange) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return ErrInvalidJSONB
	}

	return json.Unmarshal(b, a)
}

// CreateAuditLogRequest represents a request to create an audit log entry
type CreateAuditLogRequest struct {
	UserID       *int64                 `json:"user_id,omitempty"`
	Username     string                 `json:"username"`
	Action       string                 `json:"action" binding:"required"`
	ResourceType string                 `json:"resource_type,omitempty"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Changes      *AuditLogChange        `json:"changes,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	Status       string                 `json:"status,omitempty"`       // defaults to "success"
	ErrorMessage string                 `json:"error_message,omitempty"`
	Severity     string                 `json:"severity,omitempty"` // defaults to "info"
}

// AuditLogFilter represents filter options for querying audit logs
type AuditLogFilter struct {
	UserID       *int64    `json:"user_id,omitempty"`
	Username     string    `json:"username,omitempty"`
	Action       string    `json:"action,omitempty"`
	ResourceType string    `json:"resource_type,omitempty"`
	ResourceID   string    `json:"resource_id,omitempty"`
	Status       string    `json:"status,omitempty"`
	Severity     string    `json:"severity,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	StartDate    time.Time `json:"start_date,omitempty"`
	EndDate      time.Time `json:"end_date,omitempty"`
	Limit        int       `json:"limit,omitempty"`
	Offset       int       `json:"offset,omitempty"`
}

// AuditLogListResponse represents a paginated list of audit logs
type AuditLogListResponse struct {
	Logs       []AdminAuditLog `json:"logs"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// Audit action constants
const (
	// User Management Actions
	AuditActionUserView          = "user.view"
	AuditActionUserCreate        = "user.create"
	AuditActionUserUpdate        = "user.update"
	AuditActionUserDelete        = "user.delete"
	AuditActionUserBan           = "user.ban"
	AuditActionUserUnban         = "user.unban"
	AuditActionUserVerify        = "user.verify"
	AuditActionUserImpersonate   = "user.impersonate"
	AuditActionUserPasswordReset = "user.password_reset"

	// URL Management Actions
	AuditActionURLView     = "url.view"
	AuditActionURLCreate   = "url.create"
	AuditActionURLUpdate   = "url.update"
	AuditActionURLDelete   = "url.delete"
	AuditActionURLDisable  = "url.disable"
	AuditActionURLEnable   = "url.enable"
	AuditActionURLFlag     = "url.flag"
	AuditActionURLUnflag   = "url.unflag"
	AuditActionURLTransfer = "url.transfer"

	// RBAC Actions
	AuditActionRoleCreate = "role.create"
	AuditActionRoleUpdate = "role.update"
	AuditActionRoleDelete = "role.delete"
	AuditActionRoleAssign = "role.assign"
	AuditActionRoleRevoke = "role.revoke"

	AuditActionPermissionCreate = "permission.create"
	AuditActionPermissionUpdate = "permission.update"
	AuditActionPermissionDelete = "permission.delete"
	AuditActionPermissionAssign = "permission.assign"
	AuditActionPermissionRevoke = "permission.revoke"

	// System Settings Actions
	AuditActionSettingView   = "setting.view"
	AuditActionSettingUpdate = "setting.update"
	AuditActionSettingReset  = "setting.reset"

	// Billing Actions
	AuditActionSubscriptionView   = "subscription.view"
	AuditActionSubscriptionUpdate = "subscription.update"
	AuditActionSubscriptionCancel = "subscription.cancel"
	AuditActionSubscriptionRefund = "subscription.refund"

	// Domain Actions
	AuditActionDomainBlock   = "domain.block"
	AuditActionDomainUnblock = "domain.unblock"

	// Analytics Actions
	AuditActionAnalyticsView   = "analytics.view"
	AuditActionAnalyticsExport = "analytics.export"

	// CMS Actions
	AuditActionCMSPageCreate = "cms.page.create"
	AuditActionCMSPageUpdate = "cms.page.update"
	AuditActionCMSPageDelete = "cms.page.delete"

	// System Actions
	AuditActionSystemHealthView = "system.health.view"
	AuditActionSystemAlertAck   = "system.alert.acknowledge"
	AuditActionSystemAlertResolve = "system.alert.resolve"
)

// Audit status constants
const (
	AuditStatusSuccess      = "success"
	AuditStatusFailed       = "failed"
	AuditStatusUnauthorized = "unauthorized"
)

// Audit severity constants
const (
	AuditSeverityInfo     = "info"
	AuditSeverityWarning  = "warning"
	AuditSeverityCritical = "critical"
)

// Admin-specific resource type constants (in addition to those in user_analytics.go)
const (
	ResourceTypeRole         = "role"
	ResourceTypePermission   = "permission"
	ResourceTypeSetting      = "setting"
	ResourceTypeSubscription = "subscription"
	ResourceTypePage         = "cms_page"
	ResourceTypeAlert        = "system_alert"
)
