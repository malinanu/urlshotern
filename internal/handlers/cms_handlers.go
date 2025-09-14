package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type CMSHandlers struct {
	cmsService    *services.CMSService
	apiKeyService *services.APIKeyService
}

func NewCMSHandlers(cmsService *services.CMSService, apiKeyService *services.APIKeyService) *CMSHandlers {
	return &CMSHandlers{
		cmsService:    cmsService,
		apiKeyService: apiKeyService,
	}
}

// Static Page Handlers

// CreateStaticPage creates a new static page
func (h *CMSHandlers) CreateStaticPage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateStaticPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.Slug == "" || req.Title == "" || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug, title, and content are required"})
		return
	}

	// Set default template if not provided
	if req.Template == "" {
		req.Template = "default"
	}

	page, err := h.cmsService.CreateStaticPage(&req, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create page", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"page": page})
}

// GetStaticPage gets a static page by slug
func (h *CMSHandlers) GetStaticPage(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	page, err := h.cmsService.GetStaticPageBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}

	// Record page visit (for public pages)
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")
	ipAddress := c.ClientIP()
	
	go func() {
		_ = h.cmsService.RecordPageVisit(page.ID, ipAddress, userAgent, referrer)
	}()

	c.JSON(http.StatusOK, gin.H{"page": page})
}

// GetStaticPageByID gets a static page by ID (admin only)
func (h *CMSHandlers) GetStaticPageByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page ID"})
		return
	}

	page, err := h.cmsService.GetStaticPageByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"page": page})
}

// UpdateStaticPage updates a static page
func (h *CMSHandlers) UpdateStaticPage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page ID"})
		return
	}

	var req models.UpdateStaticPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	page, err := h.cmsService.UpdateStaticPage(id, &req, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update page", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"page": page})
}

// DeleteStaticPage deletes a static page
func (h *CMSHandlers) DeleteStaticPage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page ID"})
		return
	}

	err = h.cmsService.DeleteStaticPage(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete page", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Page deleted successfully"})
}

// ListStaticPages lists static pages with filtering
func (h *CMSHandlers) ListStaticPages(c *gin.Context) {
	query := &models.ListPagesQuery{
		SortBy:    "created_at",
		SortOrder: "desc",
		Limit:     50,
		Offset:    0,
	}

	// Parse query parameters
	if published := c.Query("published"); published != "" {
		if published == "true" {
			query.Published = &[]bool{true}[0]
		} else if published == "false" {
			query.Published = &[]bool{false}[0]
		}
	}

	if template := c.Query("template"); template != "" {
		query.Template = &template
	}

	if authorIDStr := c.Query("author_id"); authorIDStr != "" {
		if authorID, err := strconv.ParseInt(authorIDStr, 10, 64); err == nil {
			query.AuthorID = &authorID
		}
	}

	if search := c.Query("search"); search != "" {
		query.Search = &search
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		query.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		query.SortOrder = sortOrder
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			query.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	result, err := h.cmsService.ListStaticPages(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pages", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPageRevisions gets revisions for a page
func (h *CMSHandlers) GetPageRevisions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page ID"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	revisions, err := h.cmsService.GetPageRevisions(id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get page revisions", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"revisions": revisions})
}

// GetPageAnalytics gets analytics for a page
func (h *CMSHandlers) GetPageAnalytics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page ID"})
		return
	}

	analytics, err := h.cmsService.GetPageAnalytics(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get page analytics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}

// API Key Handlers

// CreateAPIKey creates a new API key
func (h *CMSHandlers) CreateAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	response, err := h.apiKeyService.CreateAPIKey(userID.(int64), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListAPIKeys lists API keys for the authenticated user
func (h *CMSHandlers) ListAPIKeys(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	query := &models.ListAPIKeysQuery{
		SortBy:    "created_at",
		SortOrder: "desc",
		Limit:     50,
		Offset:    0,
	}

	// Parse query parameters
	if active := c.Query("active"); active != "" {
		if active == "true" {
			query.Active = &[]bool{true}[0]
		} else if active == "false" {
			query.Active = &[]bool{false}[0]
		}
	}

	if expired := c.Query("expired"); expired != "" {
		if expired == "true" {
			query.Expired = &[]bool{true}[0]
		} else if expired == "false" {
			query.Expired = &[]bool{false}[0]
		}
	}

	if search := c.Query("search"); search != "" {
		query.Search = &search
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		query.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		query.SortOrder = sortOrder
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			query.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	result, err := h.apiKeyService.ListAPIKeys(userID.(int64), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list API keys", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAPIKey gets a specific API key
func (h *CMSHandlers) GetAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	apiKey, err := h.apiKeyService.GetAPIKey(id, userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_key": apiKey})
}

// RevokeAPIKey revokes an API key
func (h *CMSHandlers) RevokeAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	err = h.apiKeyService.RevokeAPIKey(id, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke API key", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}

// DeleteAPIKey permanently deletes an API key
func (h *CMSHandlers) DeleteAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	err = h.apiKeyService.DeleteAPIKey(id, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}

// GetAPIKeyStats gets usage statistics for an API key
func (h *CMSHandlers) GetAPIKeyStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	stats, err := h.apiKeyService.GetAPIKeyStats(id, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get API key stats", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// PublicPageHandler serves public static pages
func (h *CMSHandlers) PublicPageHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page slug is required"})
		return
	}

	page, err := h.cmsService.GetStaticPageBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}

	// Only serve published pages to public
	if !page.IsPublished {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}

	// Record page visit
	userAgent := c.GetHeader("User-Agent")
	referrer := c.GetHeader("Referer")
	ipAddress := c.ClientIP()
	
	go func() {
		_ = h.cmsService.RecordPageVisit(page.ID, ipAddress, userAgent, referrer)
	}()

	// Return page data for frontend to render
	c.JSON(http.StatusOK, gin.H{
		"page": gin.H{
			"title":            page.Title,
			"content":          page.Content,
			"meta_description": page.MetaDescription,
			"meta_keywords":    page.MetaKeywords,
			"slug":             page.Slug,
		},
	})
}