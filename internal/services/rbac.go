package services

import (
	"context"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type RBACService struct {
	db    storage.PostgresStorageInterface
	redis storage.RedisStorageInterface
}

// NewRBACService creates a new RBAC service
func NewRBACService(db storage.PostgresStorageInterface, redis storage.RedisStorageInterface) *RBACService {
	return &RBACService{
		db:    db,
		redis: redis,
	}
}

// Role Management

// CreateRole creates a new role
func (r *RBACService) CreateRole(req *models.CreateRoleRequest, createdBy int64) (*models.RoleWithPermissions, error) {
	// Check if role name already exists
	existingRole, err := r.GetRoleByName(req.Name)
	if err == nil && existingRole != nil {
		return nil, fmt.Errorf("role with name '%s' already exists", req.Name)
	}

	role := &models.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create role in database
	// Implementation would use r.db to insert role
	// For now, simulate ID assignment
	role.ID = time.Now().Unix() // Temporary ID generation

	// Assign permissions to role
	if len(req.Permissions) > 0 {
		err = r.AssignPermissionsToRole(role.ID, req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to assign permissions to role: %w", err)
		}
	}

	// Get role with permissions
	roleWithPerms, err := r.GetRoleWithPermissions(role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created role: %w", err)
	}

	return roleWithPerms, nil
}

// GetRoleByID retrieves a role by ID
func (r *RBACService) GetRoleByID(roleID int64) (*models.Role, error) {
	// Implementation would query database
	// For now, return simulated data based on constants
	switch roleID {
	case 1:
		return &models.Role{
			ID:          1,
			Name:        models.RoleSuperAdmin,
			DisplayName: "Super Administrator",
			Description: "Full system access",
			IsSystem:    true,
			IsActive:    true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		}, nil
	case 2:
		return &models.Role{
			ID:          2,
			Name:        models.RoleAdmin,
			DisplayName: "Administrator",
			Description: "Administrative access",
			IsSystem:    true,
			IsActive:    true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		}, nil
	case 3:
		return &models.Role{
			ID:          3,
			Name:        models.RoleUser,
			DisplayName: "User",
			Description: "Standard user access",
			IsSystem:    true,
			IsActive:    true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		}, nil
	default:
		return nil, fmt.Errorf("role not found")
	}
}

// GetRoleByName retrieves a role by name
func (r *RBACService) GetRoleByName(name string) (*models.Role, error) {
	// Implementation would query database by name
	switch name {
	case models.RoleSuperAdmin:
		return r.GetRoleByID(1)
	case models.RoleAdmin:
		return r.GetRoleByID(2)
	case models.RoleUser:
		return r.GetRoleByID(3)
	default:
		return nil, fmt.Errorf("role not found")
	}
}

// GetRoleWithPermissions retrieves a role with its permissions
func (r *RBACService) GetRoleWithPermissions(roleID int64) (*models.RoleWithPermissions, error) {
	role, err := r.GetRoleByID(roleID)
	if err != nil {
		return nil, err
	}

	permissions, err := r.GetRolePermissions(roleID)
	if err != nil {
		return nil, err
	}

	return &models.RoleWithPermissions{
		Role:        role,
		Permissions: permissions,
	}, nil
}

// ListRoles retrieves all roles with optional filtering
func (r *RBACService) ListRoles(activeOnly bool) ([]*models.Role, error) {
	// Implementation would query database with filtering
	roles := []*models.Role{
		{
			ID:          1,
			Name:        models.RoleSuperAdmin,
			DisplayName: "Super Administrator",
			Description: "Full system access",
			IsSystem:    true,
			IsActive:    true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:          2,
			Name:        models.RoleAdmin,
			DisplayName: "Administrator",
			Description: "Administrative access",
			IsSystem:    true,
			IsActive:    true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:          3,
			Name:        models.RoleUser,
			DisplayName: "User",
			Description: "Standard user access",
			IsSystem:    true,
			IsActive:    true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		},
	}

	if activeOnly {
		var activeRoles []*models.Role
		for _, role := range roles {
			if role.IsActive {
				activeRoles = append(activeRoles, role)
			}
		}
		return activeRoles, nil
	}

	return roles, nil
}

// UpdateRole updates an existing role
func (r *RBACService) UpdateRole(roleID int64, req *models.UpdateRoleRequest, updatedBy int64) (*models.Role, error) {
	role, err := r.GetRoleByID(roleID)
	if err != nil {
		return nil, err
	}

	if role.IsSystem {
		return nil, fmt.Errorf("cannot modify system role")
	}

	// Update fields if provided
	if req.DisplayName != nil {
		role.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	role.UpdatedAt = time.Now()

	// Update permissions if provided
	if req.Permissions != nil {
		err = r.AssignPermissionsToRole(roleID, *req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to update role permissions: %w", err)
		}
	}

	// Implementation would save to database
	return role, nil
}

// DeleteRole deletes a role (soft delete for system roles)
func (r *RBACService) DeleteRole(roleID int64, deletedBy int64) error {
	role, err := r.GetRoleByID(roleID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return fmt.Errorf("cannot delete system role")
	}

	// Check if role is assigned to any users
	users, err := r.GetUsersWithRole(roleID)
	if err != nil {
		return err
	}

	if len(users) > 0 {
		return fmt.Errorf("cannot delete role assigned to %d users", len(users))
	}

	// Implementation would delete from database
	return nil
}

// Permission Management

// CreatePermission creates a new permission
func (r *RBACService) CreatePermission(req *models.CreatePermissionRequest) (*models.Permission, error) {
	// Check if permission already exists
	existing, err := r.GetPermissionByName(req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("permission '%s' already exists", req.Name)
	}

	permission := &models.Permission{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
		IsSystem:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Implementation would save to database
	permission.ID = time.Now().Unix() // Temporary ID

	return permission, nil
}

// GetPermissionByName retrieves a permission by name
func (r *RBACService) GetPermissionByName(name string) (*models.Permission, error) {
	// Implementation would query database
	// For now, return simulated system permissions
	systemPermissions := r.getSystemPermissions()
	for _, perm := range systemPermissions {
		if perm.Name == name {
			return perm, nil
		}
	}
	return nil, fmt.Errorf("permission not found")
}

// ListPermissions retrieves all permissions
func (r *RBACService) ListPermissions() ([]*models.Permission, error) {
	// Implementation would query database
	return r.getSystemPermissions(), nil
}

// User Role Management

// AssignRoleToUser assigns a role to a user
func (r *RBACService) AssignRoleToUser(req *models.AssignRoleRequest, assignedBy int64) error {
	// Check if role exists
	role, err := r.GetRoleByID(req.RoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if user already has this role
	hasRole, err := r.UserHasRole(req.UserID, req.RoleID)
	if err != nil {
		return err
	}
	if hasRole {
		return fmt.Errorf("user already has role '%s'", role.Name)
	}

	userRole := &models.UserRole{
		UserID:     req.UserID,
		RoleID:     req.RoleID,
		AssignedBy: &assignedBy,
		ExpiresAt:  req.ExpiresAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Implementation would save to database
	userRole.ID = time.Now().Unix() // Temporary ID

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *RBACService) RemoveRoleFromUser(userID, roleID int64, removedBy int64) error {
	// Check if user has this role
	hasRole, err := r.UserHasRole(userID, roleID)
	if err != nil {
		return err
	}
	if !hasRole {
		return fmt.Errorf("user does not have this role")
	}

	// Implementation would remove from database
	return nil
}

// GetUserRoles retrieves all roles for a user
func (r *RBACService) GetUserRoles(userID int64) ([]*models.Role, error) {
	// Implementation would query database
	// For now, return default user role
	userRole, _ := r.GetRoleByName(models.RoleUser)
	return []*models.Role{userRole}, nil
}

// GetUserPermissions retrieves all permissions for a user (through roles)
func (r *RBACService) GetUserPermissions(userID int64) ([]*models.Permission, error) {
	roles, err := r.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	permissionMap := make(map[string]*models.Permission)
	
	for _, role := range roles {
		rolePermissions, err := r.GetRolePermissions(role.ID)
		if err != nil {
			continue
		}

		for _, perm := range rolePermissions {
			permissionMap[perm.Name] = perm
		}
	}

	permissions := make([]*models.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// Permission Checking

// UserHasRole checks if a user has a specific role
func (r *RBACService) UserHasRole(userID, roleID int64) (bool, error) {
	roles, err := r.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.ID == roleID {
			return true, nil
		}
	}

	return false, nil
}

// UserHasPermission checks if a user has a specific permission
func (r *RBACService) UserHasPermission(userID int64, permissionName string) (bool, error) {
	permissions, err := r.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.Name == permissionName {
			return true, nil
		}
	}

	return false, nil
}

// UserHasAnyRole checks if a user has any of the specified roles
func (r *RBACService) UserHasAnyRole(userID int64, roleNames []string) (bool, error) {
	userRoles, err := r.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	for _, userRole := range userRoles {
		for _, roleName := range roleNames {
			if userRole.Name == roleName {
				return true, nil
			}
		}
	}

	return false, nil
}

// UserHasAnyPermission checks if a user has any of the specified permissions
func (r *RBACService) UserHasAnyPermission(userID int64, permissionNames []string) (bool, error) {
	userPermissions, err := r.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, userPerm := range userPermissions {
		for _, permName := range permissionNames {
			if userPerm.Name == permName {
				return true, nil
			}
		}
	}

	return false, nil
}

// Helper Methods

// AssignPermissionsToRole assigns multiple permissions to a role
func (r *RBACService) AssignPermissionsToRole(roleID int64, permissionIDs []int64) error {
	// Implementation would manage role_permissions table
	for _, permID := range permissionIDs {
		// Check if permission exists
		_, err := r.GetPermissionByID(permID)
		if err != nil {
			return fmt.Errorf("permission %d not found", permID)
		}

		// Create role_permission relationship
		rolePermission := &models.RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
			CreatedAt:    time.Now(),
		}
		
		// Implementation would save to database
		rolePermission.ID = time.Now().Unix() // Temporary ID
	}

	return nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *RBACService) GetRolePermissions(roleID int64) ([]*models.Permission, error) {
	// Implementation would query role_permissions join
	// For now, return basic permissions based on role
	switch roleID {
	case 1: // Super Admin
		return r.getSystemPermissions(), nil
	case 2: // Admin
		return r.getAdminPermissions(), nil
	case 3: // User
		return r.getUserPermissions(), nil
	default:
		return []*models.Permission{}, nil
	}
}

// GetPermissionByID retrieves a permission by ID
func (r *RBACService) GetPermissionByID(permissionID int64) (*models.Permission, error) {
	permissions := r.getSystemPermissions()
	for _, perm := range permissions {
		if perm.ID == permissionID {
			return perm, nil
		}
	}
	return nil, fmt.Errorf("permission not found")
}

// GetUsersWithRole retrieves all users with a specific role
func (r *RBACService) GetUsersWithRole(roleID int64) ([]*models.User, error) {
	// Implementation would query user_roles join
	return []*models.User{}, nil
}

// System Permission Definitions

func (r *RBACService) getSystemPermissions() []*models.Permission {
	now := time.Now()
	return []*models.Permission{
		{ID: 1, Name: models.PermissionUsersCreate, DisplayName: "Create Users", Description: "Create new users", Resource: models.ResourceUsers, Action: models.ActionCreate, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Name: models.PermissionUsersRead, DisplayName: "Read Users", Description: "View user information", Resource: models.ResourceUsers, Action: models.ActionRead, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 3, Name: models.PermissionUsersUpdate, DisplayName: "Update Users", Description: "Modify user information", Resource: models.ResourceUsers, Action: models.ActionUpdate, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 4, Name: models.PermissionUsersDelete, DisplayName: "Delete Users", Description: "Delete users", Resource: models.ResourceUsers, Action: models.ActionDelete, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 5, Name: models.PermissionUsersManage, DisplayName: "Manage Users", Description: "Full user management", Resource: models.ResourceUsers, Action: models.ActionManage, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		
		{ID: 6, Name: models.PermissionURLsCreate, DisplayName: "Create URLs", Description: "Create short URLs", Resource: models.ResourceURLs, Action: models.ActionCreate, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 7, Name: models.PermissionURLsRead, DisplayName: "Read URLs", Description: "View URL information", Resource: models.ResourceURLs, Action: models.ActionRead, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 8, Name: models.PermissionURLsUpdate, DisplayName: "Update URLs", Description: "Modify URL information", Resource: models.ResourceURLs, Action: models.ActionUpdate, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 9, Name: models.PermissionURLsDelete, DisplayName: "Delete URLs", Description: "Delete URLs", Resource: models.ResourceURLs, Action: models.ActionDelete, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 10, Name: models.PermissionURLsManage, DisplayName: "Manage URLs", Description: "Full URL management", Resource: models.ResourceURLs, Action: models.ActionManage, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		
		{ID: 11, Name: models.PermissionAnalyticsRead, DisplayName: "Read Analytics", Description: "View analytics data", Resource: models.ResourceAnalytics, Action: models.ActionRead, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 12, Name: models.PermissionAnalyticsManage, DisplayName: "Manage Analytics", Description: "Full analytics management", Resource: models.ResourceAnalytics, Action: models.ActionManage, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		
		{ID: 13, Name: models.PermissionTeamsCreate, DisplayName: "Create Teams", Description: "Create new teams", Resource: models.ResourceTeams, Action: models.ActionCreate, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 14, Name: models.PermissionTeamsRead, DisplayName: "Read Teams", Description: "View team information", Resource: models.ResourceTeams, Action: models.ActionRead, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 15, Name: models.PermissionTeamsUpdate, DisplayName: "Update Teams", Description: "Modify team information", Resource: models.ResourceTeams, Action: models.ActionUpdate, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 16, Name: models.PermissionTeamsDelete, DisplayName: "Delete Teams", Description: "Delete teams", Resource: models.ResourceTeams, Action: models.ActionDelete, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 17, Name: models.PermissionTeamsManage, DisplayName: "Manage Teams", Description: "Full team management", Resource: models.ResourceTeams, Action: models.ActionManage, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 18, Name: models.PermissionTeamsInvite, DisplayName: "Invite Team Members", Description: "Invite users to teams", Resource: models.ResourceTeams, Action: models.ActionInvite, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		
		{ID: 19, Name: models.PermissionSystemRead, DisplayName: "Read System", Description: "View system information", Resource: models.ResourceSystem, Action: models.ActionRead, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 20, Name: models.PermissionSystemManage, DisplayName: "Manage System", Description: "Full system management", Resource: models.ResourceSystem, Action: models.ActionManage, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		
		{ID: 21, Name: models.PermissionSettingsRead, DisplayName: "Read Settings", Description: "View settings", Resource: models.ResourceSettings, Action: models.ActionRead, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 22, Name: models.PermissionSettingsUpdate, DisplayName: "Update Settings", Description: "Modify settings", Resource: models.ResourceSettings, Action: models.ActionUpdate, IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: 23, Name: models.PermissionSettingsManage, DisplayName: "Manage Settings", Description: "Full settings management", Resource: models.ResourceSettings, Action: models.ActionManage, IsSystem: true, CreatedAt: now, UpdatedAt: now},
	}
}

func (r *RBACService) getAdminPermissions() []*models.Permission {
	allPerms := r.getSystemPermissions()
	var adminPerms []*models.Permission
	
	// Admin gets most permissions except system management
	for _, perm := range allPerms {
		if perm.Resource != models.ResourceSystem || perm.Action == models.ActionRead {
			adminPerms = append(adminPerms, perm)
		}
	}
	
	return adminPerms
}

func (r *RBACService) getUserPermissions() []*models.Permission {
	allPerms := r.getSystemPermissions()
	var userPerms []*models.Permission
	
	// Standard users get basic URL and analytics read permissions
	allowedPerms := []string{
		models.PermissionURLsCreate,
		models.PermissionURLsRead,
		models.PermissionURLsUpdate,
		models.PermissionURLsDelete,
		models.PermissionAnalyticsRead,
		models.PermissionSettingsRead,
		models.PermissionSettingsUpdate,
	}
	
	for _, perm := range allPerms {
		for _, allowed := range allowedPerms {
			if perm.Name == allowed {
				userPerms = append(userPerms, perm)
				break
			}
		}
	}
	
	return userPerms
}