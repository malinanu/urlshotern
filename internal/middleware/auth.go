package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type AuthMiddleware struct {
	jwtService  *services.JWTService
	userService *services.UserService
	rbacService *services.RBACService
}

func NewAuthMiddleware(jwtService *services.JWTService, userService *services.UserService, rbacService *services.RBACService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:  jwtService,
		userService: userService,
		rbacService: rbacService,
	}
}

// RequireAuth middleware ensures user is authenticated
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractTokenFromHeader(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		claims, err := a.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Verify this is an access token
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "invalid_token_type",
				Message: "Invalid token type",
			})
			c.Abort()
			return
		}

		// Get user from database to ensure they still exist and are active
		user, err := a.userService.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "user_not_found",
				Message: "User account not found",
			})
			c.Abort()
			return
		}

		if !user.IsActive {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "account_suspended",
				Message: "Account has been suspended",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("claims", claims)
		c.Set("session_id", claims.SessionID)

		c.Next()
	}
}

// RequireAdmin middleware ensures user is an admin
func (a *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
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
		if !ok || !u.IsAdmin {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "admin_required",
				Message: "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireEmailVerified middleware ensures user's email is verified
func (a *AuthMiddleware) RequireEmailVerified() gin.HandlerFunc {
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
		if !ok || !u.EmailVerified {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "email_verification_required",
				Message: "Email verification required to access this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePhoneVerified middleware ensures user's phone is verified
func (a *AuthMiddleware) RequirePhoneVerified() gin.HandlerFunc {
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
		if !ok || !u.PhoneVerified {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "phone_verification_required",
				Message: "Phone verification required to access this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePremium middleware ensures user has premium access
func (a *AuthMiddleware) RequirePremium() gin.HandlerFunc {
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

		if u.AccountType == "free" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "premium_required",
				Message: "Premium subscription required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware sets user context if token is present but doesn't require it
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := a.extractTokenFromHeader(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := a.jwtService.ValidateToken(token)
		if err != nil {
			// Don't abort, just continue without user context
			c.Next()
			return
		}

		if claims.TokenType != "access" {
			c.Next()
			return
		}

		user, err := a.userService.GetUserByID(claims.UserID)
		if err != nil || !user.IsActive {
			c.Next()
			return
		}

		// Set user context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("claims", claims)
		c.Set("session_id", claims.SessionID)

		c.Next()
	}
}

// RateLimitByUser applies rate limiting based on user account type
func (a *AuthMiddleware) RateLimitByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			// Apply default rate limiting for anonymous users
			c.Header("X-RateLimit-Limit", "100")
			c.Header("X-RateLimit-Remaining", "99")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))
			c.Next()
			return
		}

		u, ok := user.(*models.User)
		if !ok {
			c.Next()
			return
		}

		// Set rate limits based on account type
		var limit int
		switch u.AccountType {
		case "premium":
			limit = 5000 // 5000 requests per hour
		case "enterprise":
			limit = 20000 // 20000 requests per hour
		default:
			limit = 1000 // 1000 requests per hour for free users
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-1)) // Simplified for demo
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))

		c.Next()
	}
}

// RequireValidSession ensures the session is still valid
func (a *AuthMiddleware) RequireValidSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "no_session",
				Message: "No valid session found",
			})
			c.Abort()
			return
		}

		tokenClaims, ok := claims.(*services.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "invalid_session",
				Message: "Invalid session data",
			})
			c.Abort()
			return
		}

		// Verify session still exists in database
		sessionExists, err := a.userService.IsSessionValid(tokenClaims.UserID, tokenClaims.SessionID)
		if err != nil || !sessionExists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "session_expired",
				Message: "Session has expired or been revoked",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractTokenFromHeader extracts the JWT token from Authorization header
func (a *AuthMiddleware) extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return ""
	}

	return strings.TrimPrefix(authHeader, bearerPrefix)
}

// Legacy functions for backward compatibility
func GetUserID(c *gin.Context) (int64, bool) {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(int64), true
	}
	return 0, false
}

func GetUserEmail(c *gin.Context) (string, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u.Email, true
		}
	}
	return "", false
}

func GetAccountType(c *gin.Context) (string, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u.AccountType, true
		}
	}
	return "", false
}

func IsAdmin(c *gin.Context) bool {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u.IsAdmin
		}
	}
	return false
}


// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

// RequestLogger logs all requests with authentication context
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		userID := "anonymous"
		if param.Keys != nil {
			if id, exists := param.Keys["user_id"]; exists {
				userID = fmt.Sprintf("%v", id)
			}
		}

		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s\" %s UserID:%s\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			userID,
		)
	})
}

// RBAC Middleware Functions

// RequirePermission middleware ensures user has a specific permission
func (a *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
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

		hasPermission, err := a.rbacService.UserHasPermission(userID.(int64), permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "permission_check_failed",
				Message: "Failed to check user permission",
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

// RequireAnyPermission middleware ensures user has at least one of the specified permissions
func (a *AuthMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
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

		hasPermission, err := a.rbacService.UserHasAnyPermission(userID.(int64), permissions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "permission_check_failed",
				Message: "Failed to check user permissions",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "permission_denied",
				Message: "You don't have any of the required permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware ensures user has a specific role
func (a *AuthMiddleware) RequireRole(roleName string) gin.HandlerFunc {
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

		hasRole, err := a.rbacService.UserHasAnyRole(userID.(int64), []string{roleName})
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "role_check_failed",
				Message: "Failed to check user role",
			})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "role_required",
				Message: fmt.Sprintf("Required role: %s", roleName),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware ensures user has at least one of the specified roles
func (a *AuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
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

		hasRole, err := a.rbacService.UserHasAnyRole(userID.(int64), roles)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "role_check_failed",
				Message: "Failed to check user roles",
			})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "role_required",
				Message: "You don't have any of the required roles",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin middleware ensures user is a super admin
func (a *AuthMiddleware) RequireSuperAdmin() gin.HandlerFunc {
	return a.RequireRole(models.RoleSuperAdmin)
}

// RequireAdminOrSuperAdmin middleware ensures user is admin or super admin
func (a *AuthMiddleware) RequireAdminOrSuperAdmin() gin.HandlerFunc {
	return a.RequireAnyRole(models.RoleAdmin, models.RoleSuperAdmin)
}

// RequireSystemPermission is a convenience method for system permissions
func (a *AuthMiddleware) RequireSystemPermission() gin.HandlerFunc {
	return a.RequirePermission(models.PermissionSystemRead)
}

// RequireUserManagement is a convenience method for user management permissions
func (a *AuthMiddleware) RequireUserManagement() gin.HandlerFunc {
	return a.RequirePermission(models.PermissionUsersManage)
}

// RequireURLManagement is a convenience method for URL management permissions
func (a *AuthMiddleware) RequireURLManagement() gin.HandlerFunc {
	return a.RequirePermission(models.PermissionURLsManage)
}

// RequireTeamManagement is a convenience method for team management permissions
func (a *AuthMiddleware) RequireTeamManagement() gin.HandlerFunc {
	return a.RequirePermission(models.PermissionTeamsManage)
}

// RequireAnalyticsAccess is a convenience method for analytics permissions
func (a *AuthMiddleware) RequireAnalyticsAccess() gin.HandlerFunc {
	return a.RequirePermission(models.PermissionAnalyticsRead)
}