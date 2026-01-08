package admin

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminCMSHandlers handles admin CMS operations
type AdminCMSHandlers struct {
	cmsService   *services.CMSService
	mediaService *services.MediaService
	auditService *services.AdminAuditService
}

// NewAdminCMSHandlers creates a new admin CMS handlers instance
func NewAdminCMSHandlers(cmsService *services.CMSService, mediaService *services.MediaService, auditService *services.AdminAuditService) *AdminCMSHandlers {
	return &AdminCMSHandlers{
		cmsService:   cmsService,
		mediaService: mediaService,
		auditService: auditService,
	}
}

// AdminListPages lists all CMS pages
// GET /api/v1/admin/cms/pages
func (h *AdminCMSHandlers) AdminListPages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build query parameters
	query := &models.ListPagesQuery{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Optional filters
	if published := c.Query("published"); published != "" {
		isPublished := published == "true"
		query.Published = &isPublished
	}
	if template := c.Query("template"); template != "" {
		query.Template = &template
	}
	if search := c.Query("search"); search != "" {
		query.Search = &search
	}

	// List pages
	result, err := h.cmsService.ListStaticPages(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list pages",
		})
		return
	}

	totalPages := (result.Total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{
		"pages":       result.Data,
		"total":       result.Total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// AdminGetPage retrieves a specific CMS page
// GET /api/v1/admin/cms/pages/:id
func (h *AdminCMSHandlers) AdminGetPage(c *gin.Context) {
	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_page_id",
			Message: "Invalid page ID",
		})
		return
	}

	page, err := h.cmsService.GetStaticPageByID(pageID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "page_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page": page,
	})
}

// AdminCreatePage creates a new CMS page
// POST /api/v1/admin/cms/pages
func (h *AdminCMSHandlers) AdminCreatePage(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req models.CreateStaticPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Create page
	page, err := h.cmsService.CreateStaticPage(&req, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "cms.create_page", "page", &page.ID, map[string]interface{}{
		"slug":  page.Slug,
		"title": page.Title,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Page created successfully",
		"page":    page,
	})
}

// AdminUpdatePage updates a CMS page
// PUT /api/v1/admin/cms/pages/:id
func (h *AdminCMSHandlers) AdminUpdatePage(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_page_id",
			Message: "Invalid page ID",
		})
		return
	}

	var req models.UpdateStaticPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Update page
	page, err := h.cmsService.UpdateStaticPage(pageID, &req, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "cms.update_page", "page", &pageID, map[string]interface{}{
		"slug": page.Slug,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Page updated successfully",
		"page":    page,
	})
}

// AdminDeletePage deletes a CMS page
// DELETE /api/v1/admin/cms/pages/:id
func (h *AdminCMSHandlers) AdminDeletePage(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_page_id",
			Message: "Invalid page ID",
		})
		return
	}

	// Get page info before deleting for audit log
	page, err := h.cmsService.GetStaticPageByID(pageID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "page_not_found",
			Message: err.Error(),
		})
		return
	}

	// Delete page
	if err := h.cmsService.DeleteStaticPage(pageID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "cms.delete_page", "page", &pageID, map[string]interface{}{
		"slug":  page.Slug,
		"title": page.Title,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Page deleted successfully",
	})
}

// AdminPublishPage publishes a draft CMS page
// POST /api/v1/admin/cms/pages/:id/publish
func (h *AdminCMSHandlers) AdminPublishPage(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_page_id",
			Message: "Invalid page ID",
		})
		return
	}

	// Publish page by updating is_published to true
	isPublished := true
	req := &models.UpdateStaticPageRequest{
		IsPublished: &isPublished,
	}

	page, err := h.cmsService.UpdateStaticPage(pageID, req, adminID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "publish_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "cms.publish_page", "page", &pageID, map[string]interface{}{
		"slug":  page.Slug,
		"title": page.Title,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Page published successfully",
		"page":    page,
	})
}

// AdminListMedia lists all media files
// GET /api/v1/admin/cms/media
func (h *AdminCMSHandlers) AdminListMedia(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Optional file type filter
	fileType := c.Query("file_type")

	// Calculate offset
	offset := (page - 1) * limit

	// List media files
	files, total, err := h.mediaService.AdminListMediaFiles(fileType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list media files",
		})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{
		"media":       files,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// AdminUploadMedia uploads a new media file
// POST /api/v1/admin/cms/media
func (h *AdminCMSHandlers) AdminUploadMedia(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Get uploaded file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_file",
			Message: "No file uploaded",
		})
		return
	}

	// Validate file
	if err := h.mediaService.ValidateFile(fileHeader); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
		return
	}

	// Parse form data
	req := &models.MediaUploadRequest{
		IsPublic: c.PostForm("is_public") == "true",
	}
	if alt := c.PostForm("alt"); alt != "" {
		req.Alt = &alt
	}
	if caption := c.PostForm("caption"); caption != "" {
		req.Caption = &caption
	}

	// Upload file
	mediaFile, err := h.mediaService.UploadFile(adminID.(int64), fileHeader, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "upload_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "cms.upload_media", "media", &mediaFile.ID, map[string]interface{}{
		"filename":  mediaFile.Filename,
		"file_type": mediaFile.FileType,
		"file_size": mediaFile.FileSize,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Media uploaded successfully",
		"media":   mediaFile,
	})
}

// AdminDeleteMedia deletes a media file
// DELETE /api/v1/admin/cms/media/:id
func (h *AdminCMSHandlers) AdminDeleteMedia(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	mediaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_media_id",
			Message: "Invalid media ID",
		})
		return
	}

	// Delete media file
	if err := h.mediaService.AdminDeleteMediaFile(mediaID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "cms.delete_media", "media", &mediaID, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Media deleted successfully",
	})
}

// AdminGetPageVersions retrieves version history for a page
// GET /api/v1/admin/cms/pages/:id/versions
func (h *AdminCMSHandlers) AdminGetPageVersions(c *gin.Context) {
	pageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_page_id",
			Message: "Invalid page ID",
		})
		return
	}

	// Get limit (default 10, max 50)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	versions, err := h.cmsService.GetPageRevisions(pageID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve page versions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"versions": versions,
		"total":    len(versions),
	})
}
