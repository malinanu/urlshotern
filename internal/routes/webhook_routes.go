package routes

import (
	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupWebhookAdminRoutes configures webhook admin routes
func SetupWebhookAdminRoutes(router *gin.Engine, webhookAdminHandler *handlers.WebhookAdminHandler, authMiddleware *middleware.AuthMiddleware) {
	// Admin routes for webhook management
	admin := router.Group("/api/v1/admin/webhooks")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireAdmin())
	admin.Use(middleware.RateLimitMiddleware(middleware.DefaultRateLimit))
	{
		// Manual replay of failed webhook
		admin.POST("/:event_id/replay", webhookAdminHandler.ReplayWebhook)

		// Get dead letter queue
		admin.GET("/dead-letter", webhookAdminHandler.GetDeadLetterQueue)

		// Get webhook metrics
		admin.GET("/metrics", webhookAdminHandler.GetWebhookMetrics)
	}
}
