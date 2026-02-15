package models

import "time"

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalUsers       int64  `json:"total_users"`
	ActiveUsers      int64  `json:"active_users"`
	TotalURLs        int64  `json:"total_urls"`
	TotalClicks      int64  `json:"total_clicks"`
	StorageUsed      string `json:"storage_used"` // e.g., "1.2 GB"
	DailySignups     int64  `json:"daily_signups"`
	DailyURLsCreated int64  `json:"daily_urls_created"`
	SystemStatus     string `json:"system_status"` // "healthy", "warning", "error"
}

// AdminUser represents user data for admin panel
type AdminUser struct {
	ID              int64      `json:"id,string"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	IsActive        bool       `json:"is_active"`
	IsEmailVerified bool       `json:"is_email_verified"`
	AccountType     string     `json:"account_type"`
	CreatedAt       time.Time  `json:"created_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	URLsCount       int64      `json:"urls_count"`
	TotalClicks     int64      `json:"total_clicks"`
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Users  []*AdminUser `json:"users"`
	Total  int64        `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

// RecentActivity represents an activity log entry
type RecentActivity struct {
	ID           int64     `json:"id"`
	UserID       *int64    `json:"user_id,omitempty"`
	UserName     string    `json:"user_name"`
	ActivityType string    `json:"activity_type"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	IPAddress    *string   `json:"ip_address,omitempty"`
}

// CMS Models

// CMSPage represents a CMS page for admin panel
type CMSPage struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	IsPublished bool      `json:"is_published"`
	AuthorID    *int64    `json:"author_id,omitempty"`
	ViewCount   int64     `json:"view_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CMSBlock represents a content block for admin panel
type CMSBlock struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Identifier string    `json:"identifier"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	IsActive   bool      `json:"is_active"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CMSTemplate represents a page template for admin panel
type CMSTemplate struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Admin User Management Models

// CreateUserAdminRequest represents admin request to create a new user
type CreateUserAdminRequest struct {
	Name               string  `json:"name" validate:"required,min=2,max=100"`
	Email              string  `json:"email" validate:"required,email"`
	Password           *string `json:"password,omitempty" validate:"omitempty,min=8"`
	Phone              *string `json:"phone,omitempty"`
	AccountType        string  `json:"account_type" validate:"required,oneof=free premium enterprise"`
	RoleIDs            []int64 `json:"role_ids" validate:"required,min=1"`
	SendInvite         bool    `json:"send_invite"`          // If true, send invitation email
	EmailVerified      bool    `json:"email_verified"`       // Admin can pre-verify email
	IsActive           bool    `json:"is_active"`            // User active status
	ForcePasswordReset bool    `json:"force_password_reset"` // Require password change on first login
}

// CreateUserAdminResponse represents response after admin creates a user
type CreateUserAdminResponse struct {
	User         *AdminUser `json:"user"`
	Roles        []*Role    `json:"roles"`
	InviteURL    *string    `json:"invite_url,omitempty"`    // Only if SendInvite=true
	TempPassword *string    `json:"temp_password,omitempty"` // Only if password was generated
}

// AssignUserRoleRequest represents role assignment request
type AssignUserRoleRequest struct {
	RoleID    int64      `json:"role_id" validate:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// GetUserRolesResponse represents user roles and permissions response
type GetUserRolesResponse struct {
	Roles       []*Role       `json:"roles"`
	Permissions []*Permission `json:"permissions"`
}

// UpdateAccountTypeRequest represents account type update request
type UpdateAccountTypeRequest struct {
	AccountType string `json:"account_type" validate:"required,oneof=free premium enterprise"`
}

// RBAC Audit Models

// RBACAuditEntry represents an RBAC audit log entry
type RBACAuditEntry struct {
	ID             int64     `json:"id,string"`
	UserID         *int64    `json:"user_id,omitempty"`
	TargetUserID   *int64    `json:"target_user_id,omitempty"`
	Action         string    `json:"action"`
	RoleID         *int64    `json:"role_id,omitempty"`
	RoleName       string    `json:"role_name,omitempty"`
	PermissionID   *int64    `json:"permission_id,omitempty"`
	AssignedBy     *int64    `json:"assigned_by,omitempty"`
	AssignedByName string    `json:"assigned_by_name,omitempty"`
	Details        *string   `json:"details,omitempty"` // JSON string
	IPAddress      *string   `json:"ip_address,omitempty"`
	UserAgent      *string   `json:"user_agent,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// RBACAuditFilters represents filters for audit log queries
type RBACAuditFilters struct {
	UserID       *int64     `json:"user_id,omitempty"`
	TargetUserID *int64     `json:"target_user_id,omitempty"`
	Action       *string    `json:"action,omitempty"`
	RoleID       *int64     `json:"role_id,omitempty"`
	AssignedBy   *int64     `json:"assigned_by,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}

// RBAC Audit Action Constants
const (
	AuditActionRoleAssigned    = "role_assigned"
	AuditActionRoleRemoved     = "role_removed"
	AuditActionUserCreated     = "user_created"
	AuditActionUserUpdated     = "user_updated"
	AuditActionUserDeleted     = "user_deleted"
	AuditActionUserActivated   = "user_activated"
	AuditActionUserDeactivated = "user_deactivated"
	AuditActionRoleCreated     = "role_created"
	AuditActionRoleUpdated     = "role_updated"
	AuditActionRoleDeleted     = "role_deleted"
	AuditActionPermModified    = "permission_modified"
)
