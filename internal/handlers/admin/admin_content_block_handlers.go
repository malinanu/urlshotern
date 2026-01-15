package admin

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminContentBlockHandlers handles admin content block operations
type AdminContentBlockHandlers struct {
	contentBlockService *services.ContentBlockService
	auditService        *services.AdminAuditService
}

// NewAdminContentBlockHandlers creates a new admin content block handlers instance
func NewAdminContentBlockHandlers(contentBlockService *services.ContentBlockService, auditService *services.AdminAuditService) *AdminContentBlockHandlers {
	return &AdminContentBlockHandlers{
		contentBlockService: contentBlockService,
		auditService:        auditService,
	}
}

// AdminListContentBlocks lists all content blocks
// GET /api/v1/admin/content-blocks
func (h *AdminContentBlockHandlers) AdminListContentBlocks(c *gin.Context) {
	blockType := c.Query("type")
	activeOnly := c.Query("active") == "true"

	blocks, err := h.contentBlockService.ListContentBlocks(blockType, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list content blocks",
		})
		return
	}

	c.JSON(http.StatusOK, blocks)
}

// AdminGetContentBlock retrieves a specific content block
// GET /api/v1/admin/content-blocks/:id
func (h *AdminContentBlockHandlers) AdminGetContentBlock(c *gin.Context) {
	blockID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_block_id",
			Message: "Invalid content block ID",
		})
		return
	}

	block, err := h.contentBlockService.GetContentBlockByID(blockID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "block_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, block)
}

// AdminCreateContentBlock creates a new content block
// POST /api/v1/admin/content-blocks
func (h *AdminContentBlockHandlers) AdminCreateContentBlock(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req models.ContentBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	block, err := h.contentBlockService.CreateContentBlock(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "content_block.create", "content_block", &block.ID, map[string]interface{}{
		"identifier": block.Identifier,
		"name":       block.Name,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Content block created successfully",
		"block":   block,
	})
}

// AdminUpdateContentBlock updates a content block
// PUT /api/v1/admin/content-blocks/:id
func (h *AdminContentBlockHandlers) AdminUpdateContentBlock(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	blockID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_block_id",
			Message: "Invalid content block ID",
		})
		return
	}

	var req models.ContentBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	block, err := h.contentBlockService.UpdateContentBlock(blockID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "content_block.update", "content_block", &blockID, map[string]interface{}{
		"identifier": block.Identifier,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Content block updated successfully",
		"block":   block,
	})
}

// AdminDeleteContentBlock deletes a content block
// DELETE /api/v1/admin/content-blocks/:id
func (h *AdminContentBlockHandlers) AdminDeleteContentBlock(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	blockID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_block_id",
			Message: "Invalid content block ID",
		})
		return
	}

	// Get block info before deleting for audit log
	block, err := h.contentBlockService.GetContentBlockByID(blockID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "block_not_found",
			Message: err.Error(),
		})
		return
	}

	if err := h.contentBlockService.DeleteContentBlock(blockID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	// Log action
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "content_block.delete", "content_block", &blockID, map[string]interface{}{
		"identifier": block.Identifier,
		"name":       block.Name,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Content block deleted successfully",
	})
}

// AdminRenderContentBlock renders a content block with variable substitution
// GET /api/v1/admin/content-blocks/:id/render
func (h *AdminContentBlockHandlers) AdminRenderContentBlock(c *gin.Context) {
	blockID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_block_id",
			Message: "Invalid content block ID",
		})
		return
	}

	// Get the block first to get its identifier
	block, err := h.contentBlockService.GetContentBlockByID(blockID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "block_not_found",
			Message: err.Error(),
		})
		return
	}

	rendered, err := h.contentBlockService.RenderContentBlock(block.Identifier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "render_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": rendered,
	})
}
