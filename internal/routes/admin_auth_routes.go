package routes

import (
	"github.com/URLshorter/url-shortener/internal/handlers/admin"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupAdminAuthRoutes configures admin authentication and dashboard routes
func SetupAdminAuthRoutes(
	router *gin.Engine,
	adminHandlers *admin.AdminHandlers,
	adminAuthHandlers *admin.AdminAuthHandlers,
	authMiddleware *middleware.AuthMiddleware,
	adminMiddleware *middleware.AdminMiddleware,
) {
	// Public admin auth routes (no authentication required)
	adminAuth := router.Group("/api/v1/admin/auth")
	// TODO: Add rate limiting middleware for admin login
	// adminAuth.Use(middleware.RateLimitMiddleware(middleware.AdminAuthRateLimit))
	{
		adminAuth.POST("/login", adminAuthHandlers.AdminLogin)
		adminAuth.POST("/refresh", adminAuthHandlers.AdminRefreshToken)
	}

	// Protected admin auth routes (require authentication)
	adminAuthProtected := router.Group("/api/v1/admin/auth")
	adminAuthProtected.Use(authMiddleware.RequireAuth())
	adminAuthProtected.Use(adminMiddleware.RequireAdminAccountType())
	{
		adminAuthProtected.POST("/logout", adminAuthHandlers.AdminLogout)
	}

	// Note: Dashboard endpoints are now in SetupAdminRoutes to avoid route conflicts
}
