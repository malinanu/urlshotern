package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type AdminRBACHandlers struct {
	adminService *services.AdminService
	rbacService  *services.RBACServiceImpl
	validator    *validator.Validate
}

// NewAdminRBACHandlers creates new admin RBAC handlers
func NewAdminRBACHandlers(adminService *services.AdminService, rbacService *services.RBACServiceImpl) *AdminRBACHandlers {
	return &AdminRBACHandlers{
		adminService: adminService,
		rbacService:  rbacService,
		validator:    validator.New(),
	}
}

// ============================================================================
// Admin User Management Handlers
// ============================================================================

// CreateUserAdmin handles admin user creation
// POST /api/v1/admin/users/create
func (h *AdminRBACHandlers) CreateUserAdmin(c *gin.Context) {
	var req models.CreateUserAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user from context
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user := currentUser.(*models.User)

	// Create user
	response, err := h.adminService.CreateUserAsAdmin(&req, user.ID)
	if err != nil {
		middleware.LogError(c, err, "Failed to create user")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Admin user created successfully")
	c.JSON(http.StatusCreated, response)
}

// AssignUserRole assigns a role to a user
// POST /api/v1/admin/users/:id/roles/assign
func (h *AdminRBACHandlers) AssignUserRole(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.AssignUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user := currentUser.(*models.User)

	// Assign role
	err = h.adminService.AssignUserRole(userID, req.RoleID, user.ID, req.ExpiresAt)
	if err != nil {
		middleware.LogError(c, err, "Failed to assign role")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Role assigned successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Role assigned successfully",
		"user_id": userID,
		"role_id": req.RoleID,
	})
}

// RemoveUserRole removes a role from a user
// DELETE /api/v1/admin/users/:id/roles/:roleId
func (h *AdminRBACHandlers) RemoveUserRole(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	// Get current user
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user := currentUser.(*models.User)

	// Remove role
	err = h.adminService.RemoveUserRole(userID, roleID, user.ID)
	if err != nil {
		middleware.LogError(c, err, "Failed to remove role")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Role removed successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Role removed successfully",
		"user_id": userID,
		"role_id": roleID,
	})
}

// GetUserRoles retrieves user's roles and permissions
// GET /api/v1/admin/users/:id/roles
func (h *AdminRBACHandlers) GetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	response, err := h.adminService.GetUserRoles(userID)
	if err != nil {
		middleware.LogError(c, err, "Failed to get user roles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateAccountType updates user's account type
// PUT /api/v1/admin/users/:id/account-type
func (h *AdminRBACHandlers) UpdateAccountType(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateAccountTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user := currentUser.(*models.User)

	// Update account type
	err = h.adminService.UpdateAccountType(userID, req.AccountType, user.ID)
	if err != nil {
		middleware.LogError(c, err, "Failed to update account type")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Account type updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"message":      "Account type updated successfully",
		"user_id":      userID,
		"account_type": req.AccountType,
	})
}

// GetRBACAuditLog retrieves RBAC audit logs
// GET /api/v1/admin/rbac/audit
func (h *AdminRBACHandlers) GetRBACAuditLog(c *gin.Context) {
	// Parse query parameters
	filters := &models.RBACAuditFilters{
		Limit:  20,
		Offset: 0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			filters.UserID = &userID
		}
	}

	if targetUserIDStr := c.Query("target_user_id"); targetUserIDStr != "" {
		if targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64); err == nil {
			filters.TargetUserID = &targetUserID
		}
	}

	if action := c.Query("action"); action != "" {
		filters.Action = &action
	}

	if roleIDStr := c.Query("role_id"); roleIDStr != "" {
		if roleID, err := strconv.ParseInt(roleIDStr, 10, 64); err == nil {
			filters.RoleID = &roleID
		}
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = &endDate
		}
	}

	// Get audit logs
	entries, err := h.adminService.GetRBACAuditLog(filters)
	if err != nil {
		middleware.LogError(c, err, "Failed to get audit log")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"limit":   filters.Limit,
		"offset":  filters.Offset,
	})
}

// ============================================================================
// RBAC Management Handlers
// ============================================================================

// ListRoles lists all roles
// GET /api/v1/admin/rbac/roles
func (h *AdminRBACHandlers) ListRoles(c *gin.Context) {
	activeOnly := c.Query("active_only") == "true"

	roles, err := h.rbacService.ListRoles(activeOnly)
	if err != nil {
		middleware.LogError(c, err, "Failed to list roles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
	})
}

// GetAssignableRoles returns roles that the current user can assign based on role hierarchy
// GET /api/v1/admin/rbac/assignable-roles
func (h *AdminRBACHandlers) GetAssignableRoles(c *gin.Context) {
	// Get current user's roles from context
	userRolesInterface, exists := c.Get("user_roles")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user roles not found in context"})
		return
	}

	// userRoles is a slice of role names ([]string)
	userRoles, ok := userRolesInterface.([]string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user roles format"})
		return
	}

	// Get assignable role names based on hierarchy
	assignableRoleNames := make(map[string]bool)
	for _, role := range userRoles {
		if assignable, ok := models.RoleAssignmentHierarchy[role]; ok {
			for _, r := range assignable {
				assignableRoleNames[r] = true
			}
		}
	}

	// Convert map to slice
	roleNamesList := make([]string, 0, len(assignableRoleNames))
	for name := range assignableRoleNames {
		roleNamesList = append(roleNamesList, name)
	}

	// Fetch full role details from database
	assignableRoles, err := h.rbacService.GetRolesByNames(roleNamesList)
	if err != nil {
		middleware.LogError(c, err, "Failed to fetch assignable roles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignable roles"})
		return
	}

	c.JSON(http.StatusOK, assignableRoles)
}

// GetRole retrieves a specific role with its permissions
// GET /api/v1/admin/rbac/roles/:id
func (h *AdminRBACHandlers) GetRole(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	role, err := h.rbacService.GetRoleByID(roleID)
	if err != nil {
		middleware.LogError(c, err, "Failed to get role")
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	permissions, err := h.rbacService.GetRolePermissions(roleID)
	if err != nil {
		middleware.LogError(c, err, "Failed to get role permissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role":        role,
		"permissions": permissions,
	})
}

// CreateCustomRole creates a new custom role
// POST /api/v1/admin/rbac/roles
func (h *AdminRBACHandlers) CreateCustomRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user := currentUser.(*models.User)

	// Create role
	roleWithPerms, err := h.rbacService.CreateRole(&req, user.ID)
	if err != nil {
		middleware.LogError(c, err, "Failed to create role")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Custom role created successfully")
	c.JSON(http.StatusCreated, roleWithPerms)
}

// UpdateRole updates an existing role
// PUT /api/v1/admin/rbac/roles/:id
func (h *AdminRBACHandlers) UpdateRole(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user := currentUser.(*models.User)

	// Update role
	role, err := h.rbacService.UpdateRole(roleID, &req, user.ID)
	if err != nil {
		middleware.LogError(c, err, "Failed to update role")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Role updated successfully")
	c.JSON(http.StatusOK, role)
}

// DeleteRole deletes a custom role
// DELETE /api/v1/admin/rbac/roles/:id
func (h *AdminRBACHandlers) DeleteRole(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	// Get current user
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	user := currentUser.(*models.User)

	// Delete role
	err = h.rbacService.DeleteRole(roleID, user.ID)
	if err != nil {
		middleware.LogError(c, err, "Failed to delete role")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	middleware.LogInfo(c, "Role deleted successfully")
	c.JSON(http.StatusOK, gin.H{
		"message": "Role deleted successfully",
		"role_id": roleID,
	})
}

// ListPermissions lists all available permissions
// GET /api/v1/admin/rbac/permissions
func (h *AdminRBACHandlers) ListPermissions(c *gin.Context) {
	permissions, err := h.rbacService.ListPermissions()
	if err != nil {
		middleware.LogError(c, err, "Failed to list permissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
	})
}
