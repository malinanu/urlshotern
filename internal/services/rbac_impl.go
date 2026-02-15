package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// RBACServiceImpl implements RBAC with real database operations
type RBACServiceImpl struct {
	db    *storage.PostgresStorage
	redis *storage.RedisStorage
}

// NewRBACServiceImpl creates a new RBAC service with database implementation
func NewRBACServiceImpl(db *storage.PostgresStorage, redis *storage.RedisStorage) *RBACServiceImpl {
	return &RBACServiceImpl{
		db:    db,
		redis: redis,
	}
}

// ============================================================================
// Core Permission Checking Methods (Most Critical - Used by Middleware)
// ============================================================================

// UserHasPermission checks if a user has a specific permission
func (s *RBACServiceImpl) UserHasPermission(userID int64, permissionName string) (bool, error) {
	// Check Redis cache first
	if cached, found := s.getCachedUserPermission(userID, permissionName); found {
		return cached, nil
	}

	// Query database
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN role_permissions rp ON ur.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE ur.user_id = $1
			  AND p.name = $2
			  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		)
	`

	var exists bool
	err := s.db.QueryRow(query, userID, permissionName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}

	// Cache the result for 5 minutes
	s.cacheUserPermission(userID, permissionName, exists)

	return exists, nil
}

// UserHasAnyPermission checks if a user has any of the specified permissions
func (s *RBACServiceImpl) UserHasAnyPermission(userID int64, permissions []string) (bool, error) {
	if len(permissions) == 0 {
		return false, nil
	}

	// Build query with IN clause for permissions
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN role_permissions rp ON ur.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE ur.user_id = $1
			  AND p.name = ANY($2)
			  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		)
	`

	var exists bool
	err := s.db.QueryRow(query, userID, permissions).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user permissions: %w", err)
	}

	return exists, nil
}

// UserHasRole checks if a user has a specific role (by ID)
func (s *RBACServiceImpl) UserHasRoleByID(userID, roleID int64) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles
			WHERE user_id = $1
			  AND role_id = $2
			  AND (expires_at IS NULL OR expires_at > NOW())
		)
	`

	var exists bool
	err := s.db.QueryRow(query, userID, roleID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user role: %w", err)
	}

	return exists, nil
}

// UserHasRole checks if a user has a specific role (by name)
func (s *RBACServiceImpl) UserHasRole(userID int64, roleName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1
			  AND r.name = $2
			  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		)
	`

	var exists bool
	err := s.db.QueryRow(query, userID, roleName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user role: %w", err)
	}

	return exists, nil
}

// UserHasAnyRole checks if a user has any of the specified roles
func (s *RBACServiceImpl) UserHasAnyRole(userID int64, roleNames []string) (bool, error) {
	if len(roleNames) == 0 {
		return false, nil
	}

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1
			  AND r.name = ANY($2)
			  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		)
	`

	var exists bool
	err := s.db.QueryRow(query, userID, roleNames).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user roles: %w", err)
	}

	return exists, nil
}

// GetUserPermissions retrieves all permissions for a user (through roles)
func (s *RBACServiceImpl) GetUserPermissions(userID int64) ([]*models.Permission, error) {
	// Check cache first
	if cached, found := s.getCachedUserPermissions(userID); found {
		return cached, nil
	}

	query := `
		SELECT DISTINCT p.id, p.name, p.display_name, p.description,
		       p.resource, p.action, p.is_system, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		ORDER BY p.resource, p.action
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
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
		permissions = append(permissions, perm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	// Cache for 5 minutes
	s.cacheUserPermissions(userID, permissions)

	return permissions, nil
}

// GetUserRoles retrieves all roles for a user
func (s *RBACServiceImpl) GetUserRoles(userID int64) ([]*models.Role, error) {
	query := `
		SELECT r.id, r.name, r.display_name, r.description,
		       r.is_system, r.is_active, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		ORDER BY r.name
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
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
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// ============================================================================
// Role Management
// ============================================================================

// GetRoleByID retrieves a role by ID
func (s *RBACServiceImpl) GetRoleByID(roleID int64) (*models.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, is_active,
		       created_at, updated_at
		FROM roles
		WHERE id = $1
	`

	role := &models.Role{}
	err := s.db.QueryRow(query, roleID).Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.IsSystem,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// GetRoleByName retrieves a role by name
func (s *RBACServiceImpl) GetRoleByName(name string) (*models.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, is_active,
		       created_at, updated_at
		FROM roles
		WHERE name = $1
	`

	role := &models.Role{}
	err := s.db.QueryRow(query, name).Scan(
		&role.ID,
		&role.Name,
		&role.DisplayName,
		&role.Description,
		&role.IsSystem,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// ListRoles retrieves all roles with optional filtering
func (s *RBACServiceImpl) ListRoles(activeOnly bool) ([]*models.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, is_active,
		       created_at, updated_at
		FROM roles
		WHERE ($1 = false OR is_active = true)
		ORDER BY is_system DESC, name
	`

	rows, err := s.db.Query(query, activeOnly)
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

// CreateRole creates a new custom role
func (s *RBACServiceImpl) CreateRole(req *models.CreateRoleRequest, createdBy int64) (*models.RoleWithPermissions, error) {
	// Check if role name already exists
	existing, err := s.GetRoleByName(req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("role with name '%s' already exists", req.Name)
	}

	// Insert role
	query := `
		INSERT INTO roles (name, display_name, description, is_system, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, false, true, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	role := &models.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
		IsActive:    true,
	}

	err = s.db.QueryRow(query, req.Name, req.DisplayName, req.Description).Scan(
		&role.ID,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Assign permissions if provided
	if len(req.Permissions) > 0 {
		err = s.AssignPermissionsToRole(role.ID, req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to assign permissions to role: %w", err)
		}
	}

	// Get role with permissions
	permissions, err := s.GetRolePermissions(role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return &models.RoleWithPermissions{
		Role:        role,
		Permissions: permissions,
	}, nil
}

// UpdateRole updates an existing role
func (s *RBACServiceImpl) UpdateRole(roleID int64, req *models.UpdateRoleRequest, updatedBy int64) (*models.Role, error) {
	role, err := s.GetRoleByID(roleID)
	if err != nil {
		return nil, err
	}

	if role.IsSystem {
		return nil, fmt.Errorf("cannot modify system role")
	}

	// Build dynamic update query
	query := "UPDATE roles SET updated_at = NOW()"
	args := []interface{}{roleID}
	argPos := 2

	if req.DisplayName != nil {
		query += fmt.Sprintf(", display_name = $%d", argPos)
		args = append(args, *req.DisplayName)
		argPos++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argPos)
		args = append(args, *req.Description)
		argPos++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argPos)
		args = append(args, *req.IsActive)
		argPos++
	}

	query += " WHERE id = $1 RETURNING updated_at"

	err = s.db.QueryRow(query, args...).Scan(&role.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Update permissions if provided
	if req.Permissions != nil {
		// Clear existing permissions
		_, err = s.db.Exec("DELETE FROM role_permissions WHERE role_id = $1", roleID)
		if err != nil {
			return nil, fmt.Errorf("failed to clear role permissions: %w", err)
		}

		// Assign new permissions
		if len(*req.Permissions) > 0 {
			err = s.AssignPermissionsToRole(roleID, *req.Permissions)
			if err != nil {
				return nil, fmt.Errorf("failed to assign permissions: %w", err)
			}
		}
	}

	// Refresh role data
	return s.GetRoleByID(roleID)
}

// DeleteRole deletes a custom role
func (s *RBACServiceImpl) DeleteRole(roleID int64, deletedBy int64) error {
	role, err := s.GetRoleByID(roleID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return fmt.Errorf("cannot delete system role")
	}

	// Check if role is assigned to any users
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM user_roles WHERE role_id = $1", roleID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check role assignments: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete role assigned to %d users", count)
	}

	// Delete role (cascade will delete role_permissions)
	_, err = s.db.Exec("DELETE FROM roles WHERE id = $1", roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// ============================================================================
// Permission Management
// ============================================================================

// ListPermissions retrieves all permissions
func (s *RBACServiceImpl) ListPermissions() ([]*models.Permission, error) {
	query := `
		SELECT id, name, display_name, description, resource, action,
		       is_system, created_at
		FROM permissions
		ORDER BY resource, action
	`

	rows, err := s.db.Query(query)
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
		permissions = append(permissions, perm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// GetRolePermissions retrieves all permissions for a role
func (s *RBACServiceImpl) GetRolePermissions(roleID int64) ([]*models.Permission, error) {
	query := `
		SELECT p.id, p.name, p.display_name, p.description,
		       p.resource, p.action, p.is_system, p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`

	rows, err := s.db.Query(query, roleID)
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
		permissions = append(permissions, perm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// AssignPermissionsToRole assigns multiple permissions to a role
func (s *RBACServiceImpl) AssignPermissionsToRole(roleID int64, permissionIDs []int64) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	// Prepare batch insert
	query := "INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES "
	args := []interface{}{}

	for i, permID := range permissionIDs {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($1, $%d, NOW())", i+2)
		args = append(args, permID)
	}

	query += " ON CONFLICT (role_id, permission_id) DO NOTHING"
	args = append([]interface{}{roleID}, args...)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to assign permissions to role: %w", err)
	}

	return nil
}

// ============================================================================
// User Role Assignment
// ============================================================================

// AssignRoleToUser assigns a role to a user
func (s *RBACServiceImpl) AssignRoleToUser(userID, roleID, assignedBy int64, expiresAt *time.Time) error {
	// Check if role exists
	_, err := s.GetRoleByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if user already has this role
	hasRole, err := s.UserHasRoleByID(userID, roleID)
	if err != nil {
		return err
	}
	if hasRole {
		return fmt.Errorf("user already has this role")
	}

	// Insert user role
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at, expires_at)
		VALUES ($1, $2, $3, NOW(), $4)
	`

	_, err = s.db.Exec(query, userID, roleID, assignedBy, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	// Invalidate user permission cache
	s.invalidateUserCache(userID)

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (s *RBACServiceImpl) RemoveRoleFromUser(userID, roleID, removedBy int64) error {
	// Check if user has this role
	hasRole, err := s.UserHasRoleByID(userID, roleID)
	if err != nil {
		return err
	}
	if !hasRole {
		return fmt.Errorf("user does not have this role")
	}

	query := "DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2"
	_, err = s.db.Exec(query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	// Invalidate user permission cache
	s.invalidateUserCache(userID)

	return nil
}

// CanAssignRole checks if a user with assignerRole can assign targetRole
func (s *RBACServiceImpl) CanAssignRole(assignerRole, targetRole string) bool {
	return models.CanAssignRole(assignerRole, targetRole)
}

// ============================================================================
// Redis Caching Methods
// ============================================================================

func (s *RBACServiceImpl) getCachedUserPermission(userID int64, permissionName string) (bool, bool) {
	if s.redis == nil {
		return false, false
	}

	cacheKey := fmt.Sprintf("user:%d:permission:%s", userID, permissionName)
	val, err := s.redis.Get(cacheKey)
	if err != nil {
		return false, false
	}

	return val == "1", true
}

func (s *RBACServiceImpl) cacheUserPermission(userID int64, permissionName string, hasPermission bool) {
	if s.redis == nil {
		return
	}

	cacheKey := fmt.Sprintf("user:%d:permission:%s", userID, permissionName)
	val := "0"
	if hasPermission {
		val = "1"
	}

	_ = s.redis.Set(cacheKey, val, 5*time.Minute)
}

func (s *RBACServiceImpl) getCachedUserPermissions(userID int64) ([]*models.Permission, bool) {
	if s.redis == nil {
		return nil, false
	}

	cacheKey := fmt.Sprintf("user:%d:permissions", userID)
	val, err := s.redis.Get(cacheKey)
	if err != nil {
		return nil, false
	}

	var permissions []*models.Permission
	if err := json.Unmarshal([]byte(val), &permissions); err != nil {
		return nil, false
	}

	return permissions, true
}

func (s *RBACServiceImpl) cacheUserPermissions(userID int64, permissions []*models.Permission) {
	if s.redis == nil {
		return
	}

	cacheKey := fmt.Sprintf("user:%d:permissions", userID)
	data, err := json.Marshal(permissions)
	if err != nil {
		return
	}

	_ = s.redis.Set(cacheKey, string(data), 5*time.Minute)
}

func (s *RBACServiceImpl) invalidateUserCache(userID int64) {
	if s.redis == nil {
		return
	}

	// Delete all permission caches for this user
	permKey := fmt.Sprintf("user:%d:permissions", userID)
	_ = s.redis.Del(permKey)

	// Note: Individual permission caches will expire naturally
}
