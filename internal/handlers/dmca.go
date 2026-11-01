package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/services/email"
	"github.com/gin-gonic/gin"
)

// SubmitDMCANotice accepts a copyright takedown notice (17 U.S.C. §512(c)(3)) from /dmca
// and forwards it to the Designated Agent.
func (h *ContactHandlers) SubmitDMCANotice(c *gin.Context) {
	var req struct {
		Name            string `json:"name" binding:"required,min=2,max=200"`
		Company         string `json:"company" binding:"max=200"`
		Role            string `json:"role" binding:"required,oneof=owner agent"`
		Email           string `json:"email" binding:"required,email,max=320"`
		Phone           string `json:"phone" binding:"max=50"`
		Address         string `json:"address" binding:"required,min=5,max=500"`
		CopyrightedWork string `json:"copyrighted_work" binding:"required,min=10,max=5000"`
		OriginalURL     string `json:"original_url" binding:"max=2000"`
		InfringingURLs  string `json:"infringing_urls" binding:"required,min=5,max=5000"`
		GoodFaith       bool   `json:"good_faith"`
		Accuracy        bool   `json:"accuracy"`
		Signature       string `json:"signature" binding:"required,min=2,max=200"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Please complete all required fields of the notice.",
		})
		return
	}
	// Both sworn statements are statutory elements of a valid notice (§512(c)(3)(A)(v)-(vi))
	if !req.GoodFaith || !req.Accuracy {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "A DMCA notice must include both the good-faith and the accuracy statements.",
		})
		return
	}

	notice := email.DMCANotice{
		ReferenceID:     newDMCAReference(),
		Name:            strings.TrimSpace(req.Name),
		Company:         strings.TrimSpace(req.Company),
		Role:            req.Role,
		Email:           strings.TrimSpace(req.Email),
		Phone:           strings.TrimSpace(req.Phone),
		Address:         strings.TrimSpace(req.Address),
		CopyrightedWork: strings.TrimSpace(req.CopyrightedWork),
		OriginalURL:     strings.TrimSpace(req.OriginalURL),
		InfringingURLs:  strings.TrimSpace(req.InfringingURLs),
		Signature:       strings.TrimSpace(req.Signature),
		SubmittedAt:     time.Now(),
		IPAddress:       c.ClientIP(),
	}

	if err := h.emailService.SendDMCANotice(notice); err != nil {
		middleware.LogError(c, err, "Failed to deliver DMCA notice")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "We couldn't submit your notice right now. Please email it to " + email.DMCAAgentEmail + " instead.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"reference_id": notice.ReferenceID,
		"message":      "Your notice was sent to our Designated Agent. We emailed you a receipt.",
	})
}

func newDMCAReference() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "DMCA-" + time.Now().UTC().Format("20060102") + "-" + strings.ToUpper(hex.EncodeToString(b))
}
