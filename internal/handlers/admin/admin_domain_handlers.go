package admin

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminDomainHandlers handles admin domain blocking operations
type AdminDomainHandlers struct {
	domainBlocklistService *services.DomainBlocklistService
	auditService           *services.AdminAuditService
}

// NewAdminDomainHandlers creates a new admin domain handlers instance
func NewAdminDomainHandlers(domainBlocklistService *services.DomainBlocklistService, auditService *services.AdminAuditService) *AdminDomainHandlers {
	return &AdminDomainHandlers{
		domainBlocklistService: domainBlocklistService,
		auditService:           auditService,
	}
}

// AdminGetBlockedDomains retrieves all blocked domains with pagination
// GET /api/v1/admin/domains/blocked
func (h *AdminDomainHandlers) AdminGetBlockedDomains(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	domains, total, err := h.domainBlocklistService.GetBlockedDomains(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve blocked domains",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains":     domains,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// AdminBlockDomain adds a domain to the blocklist
// POST /api/v1/admin/domains/block
func (h *AdminDomainHandlers) AdminBlockDomain(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req models.BlockDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.domainBlocklistService.BlockDomain(c.Request.Context(), req.Domain, req.Reason, adminID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "block_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.domain.block", "domain", nil, gin.H{
		"domain": req.Domain,
		"reason": req.Reason,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain blocked successfully",
	})
}

// AdminUnblockDomain removes a domain from the blocklist
// DELETE /api/v1/admin/domains/block/:domain
func (h *AdminDomainHandlers) AdminUnblockDomain(c *gin.Context) {
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_domain",
			Message: "Domain is required",
		})
		return
	}

	if err := h.domainBlocklistService.UnblockDomain(c.Request.Context(), domain); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "unblock_failed",
			Message: err.Error(),
		})
		return
	}

	// Audit log
	h.auditService.LogActionHelper(c.Request.Context(), adminID.(int64), "admin.domain.unblock", "domain", nil, gin.H{
		"domain": domain,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain unblocked successfully",
	})
}
