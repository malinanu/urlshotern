package admin

import (
	"net/http"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminAuthHandlers handles admin authentication endpoints
type AdminAuthHandlers struct {
	authService  *services.AuthService
	auditService *services.AdminAuditService
	userService  *services.UserService
}

// NewAdminAuthHandlers creates a new admin auth handlers instance
func NewAdminAuthHandlers(
	authService *services.AuthService,
	auditService *services.AdminAuditService,
	userService *services.UserService,
) *AdminAuthHandlers {
	return &AdminAuthHandlers{
		authService:  authService,
		auditService: auditService,
		userService:  userService,
	}
}

// AdminLogin handles admin-specific login
// POST /api/v1/admin/auth/login
func (h *AdminAuthHandlers) AdminLogin(c *gin.Context) {
	var req models.LoginRequest

	// Validate request format
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get IP and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Authenticate user via AuthService with admin-specific login
	response, err := h.authService.AdminLogin(&req, ipAddress, userAgent)
	if err != nil {
		// Generic error message to prevent user enumeration
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Return successful authentication response
	c.JSON(http.StatusOK, response)
}

// AdminLogout handles admin logout (token invalidation)
// POST /api/v1/admin/auth/logout
func (h *AdminAuthHandlers) AdminLogout(c *gin.Context) {
	// Extract user from context (set by RequireAuth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user context",
		})
		return
	}

	// Get token from Authorization header
	token := c.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Invalidate token via auth service
	if err := h.authService.InvalidateToken(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to logout",
		})
		return
	}

	// Log logout event
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	h.authService.LogAuditEvent(
		&user.ID,
		"admin_logout",
		"authentication",
		"",
		map[string]interface{}{
			"email": user.Email,
		},
		ipAddress,
		userAgent,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// AdminRefreshToken handles token refresh for admins
// POST /api/v1/admin/auth/refresh
func (h *AdminAuthHandlers) AdminRefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Refresh token is required",
		})
		return
	}

	// Validate refresh token and generate new token pair
	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid refresh token",
		})
		return
	}

	// Verify the refreshed token is for an admin user
	user, err := h.userService.GetUserByID(response.User.ID)
	if err != nil || (!user.IsAdmin && !isAdminAccountType(user.AccountType)) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Admin access required",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// isAdminAccountType checks if the account type is an admin type
func isAdminAccountType(accountType string) bool {
	adminTypes := []string{
		"super_admin",
		"system_admin",
		"support_admin",
		"billing_admin",
		"content_admin",
	}
	for _, t := range adminTypes {
		if accountType == t {
			return true
		}
	}
	return false
}
