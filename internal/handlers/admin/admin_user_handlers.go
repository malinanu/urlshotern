package admin

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminUserHandlers handles admin user management operations
type AdminUserHandlers struct {
	adminUserService *services.AdminUserService
	auditService     *services.AdminAuditService
}

// NewAdminUserHandlers creates a new admin user handlers instance
func NewAdminUserHandlers(adminUserService *services.AdminUserService, auditService *services.AdminAuditService) *AdminUserHandlers {
	return &AdminUserHandlers{
		adminUserService: adminUserService,
		auditService:     auditService,
	}
}

// AdminUserList lists all users with filtering and pagination
// GET /api/v1/admin/users
func (h *AdminUserHandlers) AdminUserList(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	filter := models.UserFilter{
		Email:       c.Query("email"),
		AccountType: c.Query("account_type"),
		Search:      c.Query("search"),
		SortBy:      c.DefaultQuery("sort_by", "created_at"),
		SortOrder:   c.DefaultQuery("sort_order", "desc"),
	}

	if verified := c.Query("verified"); verified != "" {
		v := verified == "true"
		filter.Verified = &v
	}

	// Use ListUsersWithStats to get users with their usage statistics
	users, total, err := h.adminUserService.ListUsersWithStats(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// AdminGetUser retrieves a specific user by ID
// GET /api/v1/admin/users/:id
func (h *AdminUserHandlers) AdminGetUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	user, err := h.adminUserService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "user_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// AdminGetUserDetailedStats retrieves comprehensive statistics for a specific user
// GET /api/v1/admin/users/:id/stats
func (h *AdminUserHandlers) AdminGetUserDetailedStats(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	stats, err := h.adminUserService.GetUserDetailedStats(c.Request.Context(), userID)
	if err != nil {
		log.Printf("ERROR: Failed to get user detailed stats for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "stats_fetch_failed",
			Message: "Failed to retrieve user statistics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// AdminUpdateUser updates user information
// PUT /api/v1/admin/users/:id
func (h *AdminUserHandlers) AdminUpdateUser(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	var req models.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	updatedUser, err := h.adminUserService.UpdateUser(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.update", "user", &userID, gin.H{
		"updated_fields": req,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    updatedUser,
	})
}

// AdminBanUser bans a user account
// POST /api/v1/admin/users/:id/ban
func (h *AdminUserHandlers) AdminBanUser(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	// Prevent self-ban
	if userID == adminID.(int64) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "self_action_forbidden",
			Message: "Cannot ban yourself",
		})
		return
	}

	var req models.BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.adminUserService.BanUser(c.Request.Context(), userID, req, adminID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "ban_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.ban", "user", &userID, gin.H{
		"reason":   req.Reason,
		"duration": req.Duration,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "User banned successfully",
	})
}

// AdminUnbanUser unbans a user account
// POST /api/v1/admin/users/:id/unban
func (h *AdminUserHandlers) AdminUnbanUser(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	if err := h.adminUserService.UnbanUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "unban_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.unban", "user", &userID, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "User unbanned successfully",
	})
}

// AdminVerifyUser forces email verification for a user
// POST /api/v1/admin/users/:id/verify
func (h *AdminUserHandlers) AdminVerifyUser(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	if err := h.adminUserService.ForceVerifyEmail(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "verification_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.force_verify", "user", &userID, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "User verified successfully",
	})
}

// AdminResetUserPassword sends a password reset email to a user
// POST /api/v1/admin/users/:id/reset-password
func (h *AdminUserHandlers) AdminResetUserPassword(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	if err := h.adminUserService.SendPasswordReset(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.password_reset", "user", &userID, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent successfully",
	})
}

// AdminDeleteUser soft deletes a user account
// DELETE /api/v1/admin/users/:id
func (h *AdminUserHandlers) AdminDeleteUser(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	// Prevent self-deletion
	if userID == adminID.(int64) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "self_action_forbidden",
			Message: "Cannot delete yourself",
		})
		return
	}

	if err := h.adminUserService.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "deletion_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.delete", "user", &userID, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// AdminImpersonateUser creates an impersonation session
// POST /api/v1/admin/users/:id/impersonate
func (h *AdminUserHandlers) AdminImpersonateUser(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	// Prevent self-impersonation
	if userID == adminID.(int64) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "self_action_forbidden",
			Message: "Cannot impersonate yourself",
		})
		return
	}

	// Create impersonation token
	token, expiresAt, err := h.adminUserService.CreateImpersonationToken(c.Request.Context(), userID, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "impersonation_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.impersonate", "user", &userID, nil)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Impersonation token created successfully",
		"token":       token,
		"token_type":  "Bearer",
		"expires_at":  expiresAt,
		"target_user": userID,
		"impersonated_by": adminID,
	})
}

// AdminUpdateUserLimits updates user rate limits and quotas
// PUT /api/v1/admin/users/:id/limits
func (h *AdminUserHandlers) AdminUpdateUserLimits(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	var req models.UpdateUserLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.adminUserService.UpdateUserLimits(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.update_limits", "user", &userID, gin.H{
		"limits": req,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "User limits updated successfully",
	})
}

// AdminUpdateUserPlan changes a user's subscription plan
// PUT /api/v1/admin/users/:id/plan
func (h *AdminUserHandlers) AdminUpdateUserPlan(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	var req models.UpdateUserPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.adminUserService.UpdateUserPlan(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.update_plan", "user", &userID, gin.H{
		"new_plan": req.PlanID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "User plan updated successfully",
	})
}

// AdminNotifyUser sends a notification to a user
// POST /api/v1/admin/users/:id/notify
func (h *AdminUserHandlers) AdminNotifyUser(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	var req models.NotifyUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.adminUserService.NotifyUser(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "notification_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.notify", "user", &userID, gin.H{
		"type":    req.Type,
		"subject": req.Subject,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification sent successfully",
	})
}

// AdminGetUserActivity retrieves user activity history
// GET /api/v1/admin/users/:id/activity
func (h *AdminUserHandlers) AdminGetUserActivity(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	activity, err := h.adminUserService.GetUserActivity(c.Request.Context(), userID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"days":     days,
		"activity": activity,
	})
}

// AdminGetUserURLs retrieves all URLs created by a user
// GET /api/v1/admin/users/:id/urls
func (h *AdminUserHandlers) AdminGetUserURLs(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	urls, total, err := h.adminUserService.GetUserURLs(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"urls":        urls,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// AdminExportUsers exports user data (bulk export)
// GET /api/v1/admin/users/export
func (h *AdminUserHandlers) AdminExportUsers(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Parse filters (same as list)
	filter := models.UserFilter{
		Email:       c.Query("email"),
		AccountType: c.Query("account_type"),
		Search:      c.Query("search"),
	}

	data, err := h.adminUserService.ExportUsers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "export_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.user.export", "users", nil, gin.H{
		"filter": filter,
	})

	// Return CSV data
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=users_export.csv")
	c.Data(http.StatusOK, "text/csv", data)
}

// AdminGetUserRoles retrieves roles assigned to a user
// GET /api/v1/admin/rbac/users/:userId/roles
func (h *AdminUserHandlers) AdminGetUserRoles(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	// TODO: Implement via RBACService
	_ = userID
	c.JSON(http.StatusOK, gin.H{
		"roles": []interface{}{},
	})
}

// AdminGetUserAuditTrail retrieves audit trail for a user
// GET /api/v1/admin/audit-logs/user/:userId
func (h *AdminUserHandlers) AdminGetUserAuditTrail(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	// TODO: Implement via AdminAuditService
	_ = userID
	c.JSON(http.StatusOK, gin.H{
		"audit_logs": []interface{}{},
	})
}
