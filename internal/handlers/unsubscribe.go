package handlers

import (
	"errors"
	"net/http"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// Unsubscribe handles the per-recipient unsubscribe link from email footers (CAN-SPAM).
// It is public and needs no login; the token itself proves which address it is for.
// POST only, so link-scanning mail filters can't unsubscribe people by prefetching.
func (h *ContactHandlers) Unsubscribe(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": services.ErrInvalidUnsubscribeToken.Error()})
		return
	}

	addr, err := h.emailService.Unsubscribe(req.Token)
	if errors.Is(err, services.ErrInvalidUnsubscribeToken) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		middleware.LogError(c, err, "Unsubscribe failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "We couldn't process your request. Please try again later."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"email":   addr,
		"message": "You've been unsubscribed from Trunc marketing emails.",
	})
}
