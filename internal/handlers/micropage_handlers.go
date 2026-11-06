package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

type MicroPageHandlers struct {
	service *services.MicroPageService
}

func NewMicroPageHandlers(service *services.MicroPageService) *MicroPageHandlers {
	return &MicroPageHandlers{service: service}
}

func (h *MicroPageHandlers) parseUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			return id, true
		}
	}
	return 0, false
}

// CreateMicroPage creates a new micro-page for the authenticated user
func (h *MicroPageHandlers) CreateMicroPage(c *gin.Context) {
	userID, ok := h.parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req models.CreateMicroPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	page, err := h.service.CreateMicroPage(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, page)
}

// ListUserMicroPages returns all micro-pages for the authenticated user
func (h *MicroPageHandlers) ListUserMicroPages(c *gin.Context) {
	userID, ok := h.parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	pages, total, err := h.service.GetUserMicroPages(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve micro pages", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pages": pages,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetMicroPage retrieves a single micro-page by ID for editing
func (h *MicroPageHandlers) GetMicroPage(c *gin.Context) {
	userID, ok := h.parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid micro page ID"})
		return
	}

	page, err := h.service.GetMicroPageByID(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, page)
}

// UpdateMicroPage updates an existing micro-page
func (h *MicroPageHandlers) UpdateMicroPage(c *gin.Context) {
	userID, ok := h.parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid micro page ID"})
		return
	}

	var req models.UpdateMicroPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	page, err := h.service.UpdateMicroPage(id, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, page)
}

// DeleteMicroPage removes an existing micro-page
func (h *MicroPageHandlers) DeleteMicroPage(c *gin.Context) {
	userID, ok := h.parseUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid micro page ID"})
		return
	}

	if err := h.service.DeleteMicroPage(id, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Micro page deleted successfully"})
}

// GetPublicMicroPage returns public data for visitors
func (h *MicroPageHandlers) GetPublicMicroPage(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	page, err := h.service.GetMicroPageBySlug(slug, true)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Micro page not found"})
		return
	}

	if !page.IsPublished {
		// If unauthenticated or not owner, return 404
		currentUserID, ok := h.parseUserID(c)
		if !ok || currentUserID != page.UserID {
			c.JSON(http.StatusNotFound, gin.H{"error": "Micro page is not published"})
			return
		}
	}

	c.JSON(http.StatusOK, page)
}

// DownloadVCard handles downloading the .vcf file for a micro-page
func (h *MicroPageHandlers) DownloadVCard(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	vcfContent, filename, err := h.service.GenerateVCF(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Micro page not found"})
		return
	}

	c.Header("Content-Type", "text/vcard; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Cache-Control", "no-cache")
	c.String(http.StatusOK, vcfContent)
}

// TrackMicroLinkClick tracks click metrics on an individual link button
func (h *MicroPageHandlers) TrackMicroLinkClick(c *gin.Context) {
	slug := c.Param("slug")
	linkID := c.Param("linkId")
	if slug == "" || linkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug and linkId are required"})
		return
	}

	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	ref := c.Request.Referer()
	country := c.GetHeader("CF-IPCountry")

	_ = h.service.TrackLinkClick(slug, linkID, ip, ua, ref, country)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
