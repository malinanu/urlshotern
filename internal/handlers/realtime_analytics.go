package handlers

import (
	"net/http"

	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

type RealtimeAnalyticsHandler struct {
	realtimeService *services.RealtimeAnalyticsService
}

func NewRealtimeAnalyticsHandler(realtimeService *services.RealtimeAnalyticsService) *RealtimeAnalyticsHandler {
	return &RealtimeAnalyticsHandler{
		realtimeService: realtimeService,
	}
}

// HandleWebSocket handles WebSocket connections for real-time analytics
func (h *RealtimeAnalyticsHandler) HandleWebSocket(c *gin.Context) {
	// Upgrade the HTTP connection to WebSocket
	h.realtimeService.HandleWebSocket(c.Writer, c.Request)
}

// GetRealtimeStats returns real-time statistics about active connections
func (h *RealtimeAnalyticsHandler) GetRealtimeStats(c *gin.Context) {
	stats := gin.H{
		"active_clients":     h.realtimeService.GetActiveClients(),
		"active_subscriptions": h.realtimeService.GetActiveSubscriptions(),
	}
	
	c.JSON(http.StatusOK, stats)
}