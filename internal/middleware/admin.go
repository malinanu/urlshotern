package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

// AdminMiddleware provides admin-specific middleware functionality
type AdminMiddleware struct {
	dbManager   *storage.DatabaseManager
	auditService *services.AdminAuditService
	rbacService  *services.RBACService
}

// NewAdminMiddleware creates a new admin middleware instance
func NewAdminMiddleware(
	dbManager *storage.DatabaseManager,
	auditService *services.AdminAuditService,
	rbacService *services.RBACService,
) *AdminMiddleware {
	return &AdminMiddleware{
		dbManager:    dbManager,
		auditService: auditService,
		rbacService:  rbacService,
	}
}

// AdminDBContext middleware switches database context to use admin connection
// This should be applied to all admin routes
func (am *AdminMiddleware) AdminDBContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set admin database in context for handlers to use
		ctx := context.WithValue(c.Request.Context(), "db_type", "admin")
		ctx = context.WithValue(ctx, "db", am.dbManager.GetAdminDB())
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// ReadOnlyDBContext middleware switches database context to use read-only connection
// Useful for analytics endpoints
func (am *AdminMiddleware) ReadOnlyDBContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "db_type", "readonly")
		ctx = context.WithValue(ctx, "db", am.dbManager.GetReadOnlyDB())
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireAdminAccountType ensures user has an admin account type
// Checks account_type field (super_admin, system_admin, support_admin, billing_admin, content_admin)
func (am *AdminMiddleware) RequireAdminAccountType() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "invalid_user",
				Message: "Invalid user context",
			})
			c.Abort()
			return
		}

		// Check if account type is admin
		adminTypes := []string{"super_admin", "system_admin", "support_admin", "billing_admin", "content_admin"}
		isAdmin := false
		for _, adminType := range adminTypes {
			if u.AccountType == adminType {
				isAdmin = true
				break
			}
		}

		// Also check the legacy is_admin flag
		if !isAdmin && !u.IsAdmin {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "admin_access_required",
				Message: "Administrative access required",
			})
			c.Abort()
			return
		}

		// Set admin context
		c.Set("is_admin_user", true)
		c.Set("admin_type", u.AccountType)

		c.Next()
	}
}

// RequireSuperAdmin ensures user is a super administrator
func (am *AdminMiddleware) RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok || u.AccountType != "super_admin" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "super_admin_required",
				Message: "Super administrator access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSystemAdmin ensures user is system admin or higher
func (am *AdminMiddleware) RequireSystemAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "invalid_user",
				Message: "Invalid user context",
			})
			c.Abort()
			return
		}

		allowedTypes := []string{"super_admin", "system_admin"}
		isAllowed := false
		for _, allowedType := range allowedTypes {
			if u.AccountType == allowedType {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "system_admin_required",
				Message: "System administrator access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminPermission checks for admin-specific permissions
func (am *AdminMiddleware) RequireAdminPermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		// Get user to check account type
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "invalid_user",
				Message: "Invalid user context",
			})
			c.Abort()
			return
		}

		// Super admins have all permissions by default
		if u.AccountType == "super_admin" {
			c.Next()
			return
		}

		// Check if user has the permission
		hasPermission, err := am.rbacService.UserHasPermission(userID.(int64), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "permission_check_failed",
				Message: "Failed to verify permissions",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "permission_denied",
				Message: fmt.Sprintf("Required permission: %s", permission),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuditLog middleware automatically logs admin actions
func (am *AdminMiddleware) AuditLog(action string, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute the handler first
		c.Next()

		// After handler execution, log the action
		go am.logAuditAsync(c, action, resourceType)
	}
}

// logAuditAsync logs audit trail asynchronously
func (am *AdminMiddleware) logAuditAsync(c *gin.Context, action, resourceType string) {
	// Extract user information
	var userID *int64
	var username string

	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			userID = &u.ID
			username = u.Email
		}
	}

	if uid, exists := c.Get("user_id"); exists {
		id := uid.(int64)
		userID = &id
	}

	// Extract resource ID from URL params if available
	resourceID := c.Param("id")
	if resourceID == "" {
		resourceID = c.Param("shortCode")
	}
	if resourceID == "" {
		resourceID = c.Param("userId")
	}
	if resourceID == "" {
		resourceID = c.Param("roleId")
	}

	// Extract request details
	details := make(map[string]interface{})
	details["method"] = c.Request.Method
	details["path"] = c.Request.URL.Path
	details["query"] = c.Request.URL.RawQuery
	details["status_code"] = c.Writer.Status()

	// Determine status
	status := models.AuditStatusSuccess
	if c.Writer.Status() >= 400 {
		status = models.AuditStatusFailed
	}
	if c.Writer.Status() == http.StatusForbidden || c.Writer.Status() == http.StatusUnauthorized {
		status = models.AuditStatusUnauthorized
	}

	// Determine severity
	severity := models.AuditSeverityInfo
	if status == models.AuditStatusFailed && c.Writer.Status() >= 500 {
		severity = models.AuditSeverityWarning
	}
	if strings.Contains(action, "delete") || strings.Contains(action, "ban") {
		severity = models.AuditSeverityCritical
	}

	// Get session ID
	sessionID := ""
	if sid, exists := c.Get("session_id"); exists {
		sessionID = sid.(string)
	}

	// Create audit log request
	auditReq := models.CreateAuditLogRequest{
		UserID:       userID,
		Username:     username,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		SessionID:    sessionID,
		Status:       status,
		Severity:     severity,
	}

	// Log to database
	if err := am.auditService.LogAction(context.Background(), auditReq); err != nil {
		// Log error but don't fail the request
		LogError(c, err, "Failed to log audit trail")
	}
}

// AdminRateLimitByAction applies stricter rate limiting for sensitive admin actions
func (am *AdminMiddleware) AdminRateLimitByAction() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Determine rate limit based on action
		var limit int
		var window string

		// Sensitive operations (delete, ban, etc.)
		if strings.Contains(path, "delete") || strings.Contains(path, "ban") {
			limit = 10
			window = "10 requests per minute"
		} else if strings.Contains(path, "export") {
			limit = 5
			window = "5 requests per minute"
		} else {
			// Regular admin operations
			limit = 100
			window = "100 requests per minute"
		}

		c.Header("X-Admin-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-Admin-RateLimit-Window", window)

		// TODO: Implement actual rate limiting with Redis
		// For now, just set headers

		c.Next()
	}
}

// RequireReAuthentication ensures user has recently authenticated
// Useful for highly sensitive operations
func (am *AdminMiddleware) RequireReAuthentication(maxAgeMinutes int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement re-authentication check
		// Check if last authentication was within maxAgeMinutes
		// For now, just pass through

		c.Next()
	}
}

// AdminSessionTimeout applies shorter session timeout for admin users
func (am *AdminMiddleware) AdminSessionTimeout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set header to indicate session timeout policy
		c.Header("X-Session-Timeout", "30") // 30 minutes for admin

		c.Next()
	}
}

// PreventSelfAction prevents admin from performing actions on their own account
func (am *AdminMiddleware) PreventSelfAction() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		// Get target user ID from URL params
		targetUserID := c.Param("id")
		if targetUserID == "" {
			targetUserID = c.Param("userId")
		}

		// Compare user IDs
		if fmt.Sprintf("%d", currentUserID.(int64)) == targetUserID {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "self_action_forbidden",
				Message: "Cannot perform this action on your own account",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminAccessLog logs admin panel access
func (am *AdminMiddleware) AdminAccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log admin panel access
		go func() {
			var userID *int64
			var username string

			if user, exists := c.Get("user"); exists {
				if u, ok := user.(*models.User); ok {
					userID = &u.ID
					username = u.Email
				}
			}

			details := map[string]interface{}{
				"method":       c.Request.Method,
				"path":         c.Request.URL.Path,
				"query":        c.Request.URL.RawQuery,
				"access_time":  c.Request.Header.Get("Date"),
			}

			auditReq := models.CreateAuditLogRequest{
				UserID:       userID,
				Username:     username,
				Action:       "admin.access",
				ResourceType: "admin_panel",
				Details:      details,
				IPAddress:    c.ClientIP(),
				UserAgent:    c.Request.UserAgent(),
				Status:       models.AuditStatusSuccess,
				Severity:     models.AuditSeverityInfo,
			}

			am.auditService.LogAction(context.Background(), auditReq)
		}()

		c.Next()
	}
}

// Convenience methods for common admin permission checks

// RequireUserManagementPermission checks for user management permissions
func (am *AdminMiddleware) RequireUserManagementPermission() gin.HandlerFunc {
	return am.RequireAdminPermission("admin.user.manage")
}

// RequireURLManagementPermission checks for URL management permissions
func (am *AdminMiddleware) RequireURLManagementPermission() gin.HandlerFunc {
	return am.RequireAdminPermission("admin.url.manage")
}

// RequireBillingManagementPermission checks for billing management permissions
func (am *AdminMiddleware) RequireBillingManagementPermission() gin.HandlerFunc {
	return am.RequireAdminPermission("admin.billing.manage")
}

// RequireSettingsManagementPermission checks for settings management permissions
func (am *AdminMiddleware) RequireSettingsManagementPermission() gin.HandlerFunc {
	return am.RequireAdminPermission("admin.settings.manage")
}

// RequireRBACManagementPermission checks for RBAC management permissions
func (am *AdminMiddleware) RequireRBACManagementPermission() gin.HandlerFunc {
	return am.RequireAdminPermission("admin.rbac.manage")
}

// RequireCMSManagementPermission checks for CMS management permissions
func (am *AdminMiddleware) RequireCMSManagementPermission() gin.HandlerFunc {
	return am.RequireAdminPermission("admin.cms.manage")
}

// RequireAuditLogViewPermission checks for audit log viewing permissions
func (am *AdminMiddleware) RequireAuditLogViewPermission() gin.HandlerFunc {
	return am.RequireAdminPermission("admin.audit.view")
}

// RequireMonitoringPermission checks for monitoring permissions
func (am *AdminMiddleware) RequireMonitoringPermission() gin.HandlerFunc {
	return am.RequireAdminPermission("admin.monitoring.view")
}
