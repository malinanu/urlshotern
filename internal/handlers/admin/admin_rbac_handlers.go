package admin

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminRBACHandlers handles admin RBAC (Role-Based Access Control) operations
type AdminRBACHandlers struct {
	rbacService  *services.RBACService
	auditService *services.AdminAuditService
}

// NewAdminRBACHandlers creates a new admin RBAC handlers instance
func NewAdminRBACHandlers(rbacService *services.RBACService, auditService *services.AdminAuditService) *AdminRBACHandlers {
	return &AdminRBACHandlers{
		rbacService:  rbacService,
		auditService: auditService,
	}
}

// AdminListRoles lists all roles
// GET /api/v1/admin/rbac/roles
func (h *AdminRBACHandlers) AdminListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	activeOnly := c.DefaultQuery("active_only", "false") == "true"

	roles, total, err := h.rbacService.ListRolesWithPagination(page, limit, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles":       roles,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// AdminGetRole retrieves a specific role
// GET /api/v1/admin/rbac/roles/:id
func (h *AdminRBACHandlers) AdminGetRole(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	roleWithPerms, err := h.rbacService.GetRoleWithPermissions(roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "role_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role": roleWithPerms,
	})
}

// AdminCreateRole creates a new role
// POST /api/v1/admin/rbac/roles
func (h *AdminRBACHandlers) AdminCreateRole(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	role, err := h.rbacService.CreateRole(&req, adminID.(int64))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "role with name '"+req.Name+"' already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "role_creation_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	roleIDPtr := &role.Role.ID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "role.create", "role", roleIDPtr, map[string]interface{}{
		"role_name": role.Role.Name,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Role created successfully",
		"role":    role,
	})
}

// AdminUpdateRole updates a role
// PUT /api/v1/admin/rbac/roles/:id
func (h *AdminRBACHandlers) AdminUpdateRole(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
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
			Message: err.Error(),
		})
		return
	}

	role, err := h.rbacService.UpdateRole(roleID, &req, adminID.(int64))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "role not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot modify system role" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "role_update_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	roleIDPtr := &roleID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "role.update", "role", roleIDPtr, map[string]interface{}{
		"role_name": role.Name,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Role updated successfully",
		"role":    role,
	})
}

// AdminDeleteRole deletes a role
// DELETE /api/v1/admin/rbac/roles/:id
func (h *AdminRBACHandlers) AdminDeleteRole(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	if err := h.rbacService.DeleteRole(roleID, adminID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "role not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot delete system role" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "role_deletion_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	roleIDPtr := &roleID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "role.delete", "role", roleIDPtr, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Role deleted successfully",
	})
}

// AdminListPermissions lists all permissions
// GET /api/v1/admin/rbac/permissions
func (h *AdminRBACHandlers) AdminListPermissions(c *gin.Context) {
	permissions, err := h.rbacService.ListPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"total":       len(permissions),
	})
}

// AdminGetPermission retrieves a specific permission
// GET /api/v1/admin/rbac/permissions/:id
func (h *AdminRBACHandlers) AdminGetPermission(c *gin.Context) {
	permissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_permission_id",
			Message: "Invalid permission ID",
		})
		return
	}

	permission, err := h.rbacService.GetPermissionByID(permissionID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "permission_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permission": permission,
	})
}

// AdminAssignPermissionToRole assigns a permission to a role
// POST /api/v1/admin/rbac/roles/:roleId/permissions/:permissionId
func (h *AdminRBACHandlers) AdminAssignPermissionToRole(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	permissionID, err := strconv.ParseInt(c.Param("permissionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_permission_id",
			Message: "Invalid permission ID",
		})
		return
	}

	if err := h.rbacService.AssignPermissionsToRole(roleID, []int64{permissionID}); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "role not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot modify permissions of system role" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "permission_assignment_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	roleIDPtr := &roleID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "role.assign_permission", "role", roleIDPtr, map[string]interface{}{
		"permission_id": permissionID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Permission assigned to role successfully",
	})
}

// AdminRemovePermissionFromRole removes a permission from a role
// DELETE /api/v1/admin/rbac/roles/:roleId/permissions/:permissionId
func (h *AdminRBACHandlers) AdminRemovePermissionFromRole(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	permissionID, err := strconv.ParseInt(c.Param("permissionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_permission_id",
			Message: "Invalid permission ID",
		})
		return
	}

	if err := h.rbacService.RemovePermissionFromRole(roleID, permissionID); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "role not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot modify permissions of system role" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "permission not assigned to this role" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "permission_removal_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	roleIDPtr := &roleID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "role.remove_permission", "role", roleIDPtr, map[string]interface{}{
		"permission_id": permissionID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Permission removed from role successfully",
	})
}

// AdminGetUserRoles retrieves all roles assigned to a user
// GET /api/v1/admin/rbac/users/:userId/roles
func (h *AdminRBACHandlers) AdminGetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	roles, err := h.rbacService.GetUserRoles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get user roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"total": len(roles),
	})
}

// AdminAssignRoleToUser assigns a role to a user
// POST /api/v1/admin/rbac/users/:userId/roles/:roleId
func (h *AdminRBACHandlers) AdminAssignRoleToUser(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	req := &models.AssignRoleRequest{
		UserID: userID,
		RoleID: roleID,
	}

	if err := h.rbacService.AssignRoleToUser(req, adminID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "role not found" || err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "role_assignment_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	userIDPtr := &userID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "user.assign_role", "user", userIDPtr, map[string]interface{}{
		"role_id": roleID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Role assigned to user successfully",
	})
}

// AdminRemoveRoleFromUser removes a role from a user
// DELETE /api/v1/admin/rbac/users/:userId/roles/:roleId
func (h *AdminRBACHandlers) AdminRemoveRoleFromUser(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_role_id",
			Message: "Invalid role ID",
		})
		return
	}

	if err := h.rbacService.RemoveRoleFromUser(userID, roleID, adminID.(int64)); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "role assignment not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error:   "role_removal_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	userIDPtr := &userID
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "user.remove_role", "user", userIDPtr, map[string]interface{}{
		"role_id": roleID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Role removed from user successfully",
	})
}
