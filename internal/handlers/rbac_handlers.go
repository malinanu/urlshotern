package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type RBACHandlers struct {
	rbacService *services.RBACService
	userService *services.UserService
}

// NewRBACHandlers creates new RBAC handlers
func NewRBACHandlers(rbacService *services.RBACService, userService *services.UserService) *RBACHandlers {
	return &RBACHandlers{
		rbacService: rbacService,
		userService: userService,
	}
}

// Role Management Endpoints

// CreateRole creates a new role
// POST /api/v1/admin/roles
func (h *RBACHandlers) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	createdBy := userID.(int64)

	role, err := h.rbacService.CreateRole(&req, createdBy)
	if err != nil {
		if err.Error() == "role with name '"+req.Name+"' already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "role_exists",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "create_role_failed",
			Message: "Failed to create role",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Role created successfully",
		"role":    role,
	})
}

// GetRole retrieves a role by ID
// GET /api/v1/admin/roles/:id
func (h *RBACHandlers) GetRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	role, err := h.rbacService.GetRoleWithPermissions(roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "role_not_found",
			Message: "Role not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role": role,
	})
}

// ListRoles retrieves all roles
// GET /api/v1/admin/roles
func (h *RBACHandlers) ListRoles(c *gin.Context) {
	activeOnlyStr := c.Query("active_only")
	activeOnly := activeOnlyStr == "true"

	roles, err := h.rbacService.ListRoles(activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "list_roles_failed",
			Message: "Failed to retrieve roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"count": len(roles),
	})
}

// UpdateRole updates an existing role
// PUT /api/v1/admin/roles/:id
func (h *RBACHandlers) UpdateRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	updatedBy := userID.(int64)

	role, err := h.rbacService.UpdateRole(roleID, &req, updatedBy)
	if err != nil {
		if err.Error() == "cannot modify system role" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "system_role",
				Message: "Cannot modify system role",
			})
			return
		}

		if err.Error() == "role not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "role_not_found",
				Message: "Role not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_role_failed",
			Message: "Failed to update role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role updated successfully",
		"role":    role,
	})
}

// DeleteRole deletes a role
// DELETE /api/v1/admin/roles/:id
func (h *RBACHandlers) DeleteRole(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	deletedBy := userID.(int64)

	err = h.rbacService.DeleteRole(roleID, deletedBy)
	if err != nil {
		if err.Error() == "cannot delete system role" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "system_role",
				Message: "Cannot delete system role",
			})
			return
		}

		if err.Error() == "role not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "role_not_found",
				Message: "Role not found",
			})
			return
		}

		if err.Error()[:31] == "cannot delete role assigned to" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "role_in_use",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_role_failed",
			Message: "Failed to delete role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role deleted successfully",
	})
}

// Permission Management Endpoints

// CreatePermission creates a new permission
// POST /api/v1/admin/permissions
func (h *RBACHandlers) CreatePermission(c *gin.Context) {
	var req models.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	permission, err := h.rbacService.CreatePermission(&req)
	if err != nil {
		if err.Error() == "permission '"+req.Name+"' already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "permission_exists",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "create_permission_failed",
			Message: "Failed to create permission",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Permission created successfully",
		"permission": permission,
	})
}

// ListPermissions retrieves all permissions
// GET /api/v1/admin/permissions
func (h *RBACHandlers) ListPermissions(c *gin.Context) {
	permissions, err := h.rbacService.ListPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "list_permissions_failed",
			Message: "Failed to retrieve permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"count":       len(permissions),
	})
}

// User Role Management Endpoints

// AssignRoleToUser assigns a role to a user
// POST /api/v1/admin/users/:user_id/roles
func (h *RBACHandlers) AssignRoleToUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	var req models.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Override the user_id from the URL
	req.UserID = targetUserID

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	assignedBy := userID.(int64)

	err = h.rbacService.AssignRoleToUser(&req, assignedBy)
	if err != nil {
		if err.Error() == "role not found: role not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "role_not_found",
				Message: "Role not found",
			})
			return
		}

		if err.Error()[:22] == "user already has role" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "role_already_assigned",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "assign_role_failed",
			Message: "Failed to assign role to user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role assigned to user successfully",
	})
}

// RemoveRoleFromUser removes a role from a user
// DELETE /api/v1/admin/users/:user_id/roles/:role_id
func (h *RBACHandlers) RemoveRoleFromUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	removedBy := userID.(int64)

	err = h.rbacService.RemoveRoleFromUser(targetUserID, roleID, removedBy)
	if err != nil {
		if err.Error() == "user does not have this role" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "role_not_assigned",
				Message: "User does not have this role",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "remove_role_failed",
			Message: "Failed to remove role from user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role removed from user successfully",
	})
}

// GetUserRoles retrieves all roles for a user
// GET /api/v1/admin/users/:user_id/roles
func (h *RBACHandlers) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	roles, err := h.rbacService.GetUserRoles(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_user_roles_failed",
			Message: "Failed to retrieve user roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": targetUserID,
		"roles":   roles,
		"count":   len(roles),
	})
}

// GetUserPermissions retrieves all permissions for a user
// GET /api/v1/admin/users/:user_id/permissions
func (h *RBACHandlers) GetUserPermissions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	permissions, err := h.rbacService.GetUserPermissions(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_user_permissions_failed",
			Message: "Failed to retrieve user permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     targetUserID,
		"permissions": permissions,
		"count":       len(permissions),
	})
}

// Permission Check Endpoints (for debugging/testing)

// CheckUserPermission checks if a user has a specific permission
// GET /api/v1/admin/users/:user_id/permissions/:permission/check
func (h *RBACHandlers) CheckUserPermission(c *gin.Context) {
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	permission := c.Param("permission")
	if permission == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_permission",
			Message: "Permission name is required",
		})
		return
	}

	hasPermission, err := h.rbacService.UserHasPermission(targetUserID, permission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "permission_check_failed",
			Message: "Failed to check user permission",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":        targetUserID,
		"permission":     permission,
		"has_permission": hasPermission,
	})
}

// CheckUserRole checks if a user has a specific role
// GET /api/v1/admin/users/:user_id/roles/:role_id/check
func (h *RBACHandlers) CheckUserRole(c *gin.Context) {
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	roleIDStr := c.Param("role_id")
	roleID, err := strconv.ParseInt(roleIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	hasRole, err := h.rbacService.UserHasRole(targetUserID, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "role_check_failed",
			Message: "Failed to check user role",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  targetUserID,
		"role_id":  roleID,
		"has_role": hasRole,
	})
}