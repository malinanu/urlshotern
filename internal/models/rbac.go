package models

import (
	"time"
)

// Role represents a system role
type Role struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Description string    `json:"description" db:"description"`
	IsSystem    bool      `json:"is_system" db:"is_system"`     // System roles cannot be deleted
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Permission represents a system permission
type Permission struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Description string    `json:"description" db:"description"`
	Resource    string    `json:"resource" db:"resource"`       // e.g., "urls", "users", "analytics"
	Action      string    `json:"action" db:"action"`           // e.g., "create", "read", "update", "delete"
	IsSystem    bool      `json:"is_system" db:"is_system"`     // System permissions cannot be deleted
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	ID           int64     `json:"id" db:"id"`
	RoleID       int64     `json:"role_id" db:"role_id"`
	PermissionID int64     `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	ID        int64      `json:"id" db:"id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	RoleID    int64      `json:"role_id" db:"role_id"`
	AssignedBy *int64    `json:"assigned_by" db:"assigned_by"` // User ID who assigned this role
	ExpiresAt *time.Time `json:"expires_at" db:"expires_at"`   // Optional expiration
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// Team represents a team/organization
type Team struct {
	ID          int64      `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	DisplayName string     `json:"display_name" db:"display_name"`
	Description *string    `json:"description" db:"description"`
	OwnerID     int64      `json:"owner_id" db:"owner_id"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	Settings    *string    `json:"settings" db:"settings"` // JSON string for team settings
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TeamMember represents team membership
type TeamMember struct {
	ID        int64      `json:"id" db:"id"`
	TeamID    int64      `json:"team_id" db:"team_id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	RoleID    int64      `json:"role_id" db:"role_id"`     // Role within the team
	InvitedBy *int64     `json:"invited_by" db:"invited_by"`
	JoinedAt  *time.Time `json:"joined_at" db:"joined_at"`
	Status    string     `json:"status" db:"status"` // invited, active, suspended
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// TeamInvitation represents pending team invitations
type TeamInvitation struct {
	ID        int64      `json:"id" db:"id"`
	TeamID    int64      `json:"team_id" db:"team_id"`
	Email     string     `json:"email" db:"email"`
	RoleID    int64      `json:"role_id" db:"role_id"`
	InvitedBy int64      `json:"invited_by" db:"invited_by"`
	Token     string     `json:"-" db:"token"` // Hidden from JSON
	Status    string     `json:"status" db:"status"` // pending, accepted, rejected, expired
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// Extended models with relationships

// UserWithRoles represents a user with their roles
type UserWithRoles struct {
	*User
	Roles       []*Role       `json:"roles"`
	Permissions []*Permission `json:"permissions"`
	Teams       []*TeamInfo   `json:"teams"`
}

// TeamInfo represents team information for user context
type TeamInfo struct {
	*Team
	Role   *Role `json:"role"`
	Status string `json:"status"`
}

// RoleWithPermissions represents a role with its permissions
type RoleWithPermissions struct {
	*Role
	Permissions []*Permission `json:"permissions"`
}

// TeamWithMembers represents a team with its members
type TeamWithMembers struct {
	*Team
	Owner   *PublicUser    `json:"owner"`
	Members []*TeamMemberInfo `json:"members"`
}

// TeamMemberInfo represents team member information
type TeamMemberInfo struct {
	*PublicUser
	Role      *Role      `json:"role"`
	Status    string     `json:"status"`
	JoinedAt  *time.Time `json:"joined_at"`
	InvitedBy *PublicUser `json:"invited_by,omitempty"`
}

// RBAC Request/Response Models

// CreateRoleRequest represents role creation request
type CreateRoleRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=50"`
	DisplayName string  `json:"display_name" validate:"required,min=2,max=100"`
	Description string  `json:"description" validate:"max=500"`
	Permissions []int64 `json:"permissions"`
}

// UpdateRoleRequest represents role update request
type UpdateRoleRequest struct {
	DisplayName *string  `json:"display_name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	Permissions *[]int64 `json:"permissions,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

// AssignRoleRequest represents role assignment request
type AssignRoleRequest struct {
	UserID    int64      `json:"user_id" validate:"required"`
	RoleID    int64      `json:"role_id" validate:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CreatePermissionRequest represents permission creation request
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
	Resource    string `json:"resource" validate:"required,min=2,max=50"`
	Action      string `json:"action" validate:"required,min=2,max=50"`
}

// UpdatePermissionRequest represents permission update request
type UpdatePermissionRequest struct {
	DisplayName *string `json:"display_name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// CreateTeamRequest represents team creation request
type CreateTeamRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=50"`
	DisplayName string  `json:"display_name" validate:"required,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// UpdateTeamRequest represents team update request
type UpdateTeamRequest struct {
	DisplayName *string `json:"display_name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// InviteTeamMemberRequest represents team member invitation request
type InviteTeamMemberRequest struct {
	Email  string `json:"email" validate:"required,email"`
	RoleID int64  `json:"role_id" validate:"required"`
}

// UpdateTeamMemberRequest represents team member update request
type UpdateTeamMemberRequest struct {
	RoleID *int64  `json:"role_id,omitempty"`
	Status *string `json:"status,omitempty" validate:"omitempty,oneof=active suspended"`
}

// AcceptInvitationRequest represents invitation acceptance request
type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}

// RBAC Constants
const (
	// System Roles
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
	RoleModerator  = "moderator"
	RoleGuest      = "guest"

	// Team Roles
	RoleTeamOwner  = "team_owner"
	RoleTeamAdmin  = "team_admin"
	RoleTeamMember = "team_member"
	RoleTeamViewer = "team_viewer"

	// Resources
	ResourceUsers     = "users"
	ResourceURLs      = "urls"
	ResourceAnalytics = "analytics"
	ResourceTeams     = "teams"
	ResourceSettings  = "settings"
	ResourceSystem    = "system"

	// Actions
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionManage = "manage"
	ActionInvite = "invite"

	// Team Member Status
	StatusInvited   = "invited"
	StatusActive    = "active"
	StatusSuspended = "suspended"

	// Team Invitation Status
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusRejected = "rejected"
	InvitationStatusExpired  = "expired"
)

// Permission names (used for permission checking)
const (
	// User Management
	PermissionUsersCreate = "users.create"
	PermissionUsersRead   = "users.read"
	PermissionUsersUpdate = "users.update"
	PermissionUsersDelete = "users.delete"
	PermissionUsersManage = "users.manage"

	// URL Management
	PermissionURLsCreate = "urls.create"
	PermissionURLsRead   = "urls.read"
	PermissionURLsUpdate = "urls.update"
	PermissionURLsDelete = "urls.delete"
	PermissionURLsManage = "urls.manage"

	// Analytics
	PermissionAnalyticsRead   = "analytics.read"
	PermissionAnalyticsManage = "analytics.manage"

	// Team Management
	PermissionTeamsCreate = "teams.create"
	PermissionTeamsRead   = "teams.read"
	PermissionTeamsUpdate = "teams.update"
	PermissionTeamsDelete = "teams.delete"
	PermissionTeamsManage = "teams.manage"
	PermissionTeamsInvite = "teams.invite"

	// System Management
	PermissionSystemRead   = "system.read"
	PermissionSystemManage = "system.manage"

	// Settings Management
	PermissionSettingsRead   = "settings.read"
	PermissionSettingsUpdate = "settings.update"
	PermissionSettingsManage = "settings.manage"
)