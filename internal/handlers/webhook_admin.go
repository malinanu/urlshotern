package handlers

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/workers"
	"github.com/gin-gonic/gin"
)

// WebhookAdminHandler handles admin operations for webhooks
type WebhookAdminHandler struct {
	retryWorker *workers.WebhookRetryWorker
}

// NewWebhookAdminHandler creates a new webhook admin handler
func NewWebhookAdminHandler(retryWorker *workers.WebhookRetryWorker) *WebhookAdminHandler {
	return &WebhookAdminHandler{
		retryWorker: retryWorker,
	}
}

// ReplayWebhook manually replays a failed webhook event
// POST /api/v1/admin/webhooks/:event_id/replay
func (h *WebhookAdminHandler) ReplayWebhook(c *gin.Context) {
	eventID := c.Param("event_id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
		return
	}

	err := h.retryWorker.RetryWebhookManually(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to replay webhook",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Webhook replayed successfully",
		"event_id": eventID,
	})
}

// GetDeadLetterQueue retrieves all webhooks in the dead letter queue
// GET /api/v1/admin/webhooks/dead-letter?limit=50
func (h *WebhookAdminHandler) GetDeadLetterQueue(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 500 {
		limit = 50
	}

	events, err := h.retryWorker.GetDeadLetterEvents(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve dead letter queue",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  len(events),
		"limit":  limit,
		"events": events,
	})
}

// GetWebhookMetrics returns webhook processing metrics
// GET /api/v1/admin/webhooks/metrics
func (h *WebhookAdminHandler) GetWebhookMetrics(c *gin.Context) {
	metrics, err := h.retryWorker.GetRetryMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve webhook metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
