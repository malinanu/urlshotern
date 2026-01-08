package services

import (
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type RBACService struct {
	db    *storage.PostgresStorage
	redis *storage.RedisStorage
}

// NewRBACService creates a new RBAC service
func NewRBACService(db *storage.PostgresStorage, redis *storage.RedisStorage) *RBACService {
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

	// Insert role into database
	query := `
		INSERT INTO roles (name, display_name, description, is_system, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	role := &models.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
		IsActive:    true,
	}

	err = r.db.QueryRow(query, role.Name, role.DisplayName, role.Description, role.IsSystem, role.IsActive).Scan(
		&role.ID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

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
	query := `
		SELECT id, name, display_name, description, is_system, is_active, created_at, updated_at
		FROM roles
		WHERE id = $1
	`

	role := &models.Role{}
	err := r.db.QueryRow(query, roleID).Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.IsSystem,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	return role, nil
}

// GetRoleByName retrieves a role by name
func (r *RBACService) GetRoleByName(name string) (*models.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, is_active, created_at, updated_at
		FROM roles
		WHERE name = $1
	`

	role := &models.Role{}
	err := r.db.QueryRow(query, name).Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.IsSystem,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	return role, nil
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
	query := `
		SELECT id, name, display_name, description, is_system, is_active, created_at, updated_at
		FROM roles
	`

	var args []interface{}
	if activeOnly {
		query += " WHERE is_active = $1"
		args = append(args, true)
	}

	query += " ORDER BY name"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.DisplayName,
			&role.Description,
			&role.IsSystem,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// ListRolesWithPagination retrieves roles with pagination and filtering
func (r *RBACService) ListRolesWithPagination(page, limit int, activeOnly bool) ([]*models.Role, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Build count query
	countQuery := "SELECT COUNT(*) FROM roles"
	var countArgs []interface{}
	if activeOnly {
		countQuery += " WHERE is_active = $1"
		countArgs = append(countArgs, true)
	}

	// Get total count
	var total int64
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// Build data query
	query := `
		SELECT id, name, display_name, description, is_system, is_active, created_at, updated_at
		FROM roles
	`

	var args []interface{}
	argNum := 1
	if activeOnly {
		query += fmt.Sprintf(" WHERE is_active = $%d", argNum)
		args = append(args, true)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY name LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.DisplayName,
			&role.Description,
			&role.IsSystem,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, total, nil
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

	// Build dynamic update query
	query := "UPDATE roles SET updated_at = NOW()"
	args := []interface{}{}
	argNum := 1

	if req.DisplayName != nil {
		query += fmt.Sprintf(", display_name = $%d", argNum)
		args = append(args, *req.DisplayName)
		argNum++
	}

	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argNum)
		args = append(args, *req.Description)
		argNum++
	}

	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argNum)
		args = append(args, *req.IsActive)
		argNum++
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, name, display_name, description, is_system, is_active, created_at, updated_at", argNum)
	args = append(args, roleID)

	// Execute update
	err = r.db.QueryRow(query, args...).Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.IsSystem,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Update permissions if provided
	if req.Permissions != nil {
		// Clear existing permissions
		_, err = r.db.Exec("DELETE FROM role_permissions WHERE role_id = $1", roleID)
		if err != nil {
			return nil, fmt.Errorf("failed to clear role permissions: %w", err)
		}

		// Assign new permissions
		err = r.AssignPermissionsToRole(roleID, *req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to update role permissions: %w", err)
		}
	}

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
	var userCount int
	err = r.db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE role_id = $1", roleID).Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to check user assignments: %w", err)
	}

	if userCount > 0 {
		return fmt.Errorf("cannot delete role assigned to %d users", userCount)
	}

	// Delete role permissions first (due to foreign key)
	_, err = r.db.Exec("DELETE FROM role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	// Delete role
	_, err = r.db.Exec("DELETE FROM roles WHERE id = $1", roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

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
	query := `
		SELECT id, name, display_name, description, resource, action, is_system, created_at
		FROM permissions
		ORDER BY resource, action
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		perm := &models.Permission{}
		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.DisplayName,
			&perm.Description,
			&perm.Resource,
			&perm.Action,
			&perm.IsSystem,
			&perm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		// Set UpdatedAt to CreatedAt since the table doesn't have updated_at
		perm.UpdatedAt = perm.CreatedAt
		permissions = append(permissions, perm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
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

	// Insert user_role relationship
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at, expires_at)
		VALUES ($1, $2, $3, NOW(), $4)
	`

	_, err = r.db.Exec(query, req.UserID, req.RoleID, assignedBy, req.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

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

	// Delete user_role relationship
	query := "DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2"
	result, err := r.db.Exec(query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("role not assigned to this user")
	}

	return nil
}

// GetUserRoles retrieves all roles for a user
func (r *RBACService) GetUserRoles(userID int64) ([]*models.Role, error) {
	fmt.Printf("DEBUG: Getting roles for user %d\n", userID)

	query := `
		SELECT r.id, r.name, r.display_name, r.description, r.is_system, r.is_active, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		ORDER BY r.name
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		fmt.Printf("DEBUG: Error querying user roles: %v\n", err)
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.DisplayName,
			&role.Description,
			&role.IsSystem,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("DEBUG: Error scanning role: %v\n", err)
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		fmt.Printf("DEBUG: Found role: id=%d, name=%s\n", role.ID, role.Name)
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		fmt.Printf("DEBUG: Error iterating roles: %v\n", err)
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	fmt.Printf("DEBUG: User %d has %d roles\n", userID, len(roles))
	return roles, nil
}

// GetUserPermissions retrieves all permissions for a user (through roles)
func (r *RBACService) GetUserPermissions(userID int64) ([]*models.Permission, error) {
	fmt.Printf("DEBUG: GetUserPermissions called for user %d\n", userID)

	roles, err := r.GetUserRoles(userID)
	if err != nil {
		fmt.Printf("DEBUG: Error getting user roles in GetUserPermissions: %v\n", err)
		return nil, err
	}

	permissionMap := make(map[string]*models.Permission)

	for _, role := range roles {
		fmt.Printf("DEBUG: Getting permissions for role %d (%s)\n", role.ID, role.Name)
		rolePermissions, err := r.GetRolePermissions(role.ID)
		if err != nil {
			fmt.Printf("DEBUG: Error getting role permissions for role %d: %v\n", role.ID, err)
			continue
		}

		fmt.Printf("DEBUG: Found %d permissions for role %d\n", len(rolePermissions), role.ID)
		for _, perm := range rolePermissions {
			fmt.Printf("DEBUG: Adding permission: %s\n", perm.Name)
			permissionMap[perm.Name] = perm
		}
	}

	permissions := make([]*models.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	fmt.Printf("DEBUG: Total unique permissions: %d\n", len(permissions))
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
	fmt.Printf("DEBUG: Checking if user %d has permission '%s'\n", userID, permissionName)

	permissions, err := r.GetUserPermissions(userID)
	if err != nil {
		fmt.Printf("DEBUG: Error getting user permissions: %v\n", err)
		return false, err
	}

	fmt.Printf("DEBUG: User %d has %d total permissions\n", userID, len(permissions))
	for _, perm := range permissions {
		fmt.Printf("DEBUG: User permission: %s\n", perm.Name)
		if perm.Name == permissionName {
			fmt.Printf("DEBUG: Permission MATCH found for %s\n", permissionName)
			return true, nil
		}
	}

	fmt.Printf("DEBUG: Permission '%s' NOT found for user %d\n", permissionName, userID)
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
	// Verify role exists
	_, err := r.GetRoleByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	for _, permID := range permissionIDs {
		// Check if permission exists
		_, err := r.GetPermissionByID(permID)
		if err != nil {
			return fmt.Errorf("permission %d not found: %w", permID, err)
		}

		// Insert role_permission relationship (ignore if already exists)
		query := `
			INSERT INTO role_permissions (role_id, permission_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`
		_, err = r.db.Exec(query, roleID, permID)
		if err != nil {
			return fmt.Errorf("failed to assign permission %d to role: %w", permID, err)
		}
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *RBACService) RemovePermissionFromRole(roleID, permissionID int64) error {
	// Verify role exists
	role, err := r.GetRoleByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if role.IsSystem {
		return fmt.Errorf("cannot modify permissions of system role")
	}

	// Delete role_permission relationship
	query := "DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2"
	result, err := r.db.Exec(query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("permission not assigned to this role")
	}

	return nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *RBACService) GetRolePermissions(roleID int64) ([]*models.Permission, error) {
	query := `
		SELECT p.id, p.name, p.display_name, p.description, p.resource, p.action, p.is_system, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`

	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		perm := &models.Permission{}
		err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.DisplayName,
			&perm.Description,
			&perm.Resource,
			&perm.Action,
			&perm.IsSystem,
			&perm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		// Set UpdatedAt to CreatedAt since the table doesn't have updated_at
		perm.UpdatedAt = perm.CreatedAt
		permissions = append(permissions, perm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// GetPermissionByID retrieves a permission by ID
func (r *RBACService) GetPermissionByID(permissionID int64) (*models.Permission, error) {
	query := `
		SELECT id, name, display_name, description, resource, action, is_system, created_at
		FROM permissions
		WHERE id = $1
	`

	perm := &models.Permission{}
	err := r.db.QueryRow(query, permissionID).Scan(
		&perm.ID,
		&perm.Name,
		&perm.DisplayName,
		&perm.Description,
		&perm.Resource,
		&perm.Action,
		&perm.IsSystem,
		&perm.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	// Set UpdatedAt to CreatedAt since the table doesn't have updated_at
	perm.UpdatedAt = perm.CreatedAt

	return perm, nil
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
