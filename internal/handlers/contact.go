package handlers

import (
	"net/http"

	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// ContactHandlers handles contact form endpoints
type ContactHandlers struct {
	emailService *services.EmailService
}

// NewContactHandlers creates a new contact handlers instance
func NewContactHandlers(emailService *services.EmailService) *ContactHandlers {
	return &ContactHandlers{
		emailService: emailService,
	}
}

// SubmitContactForm handles contact form submissions
func (h *ContactHandlers) SubmitContactForm(c *gin.Context) {
	var req struct {
		Name    string `json:"name" binding:"required,min=2,max=100"`
		Email   string `json:"email" binding:"required,email"`
		Subject string `json:"subject"`
		Message string `json:"message" binding:"required,min=10,max=5000"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request. Please check your input.",
		})
		return
	}

	err := h.emailService.SendContactFormEmail(req.Name, req.Email, req.Subject, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send your message. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thank you for contacting us! We'll get back to you soon.",
	})
}
