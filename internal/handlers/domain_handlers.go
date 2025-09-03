package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

type DomainHandlers struct {
	domainService *services.DomainService
	rbacService   *services.RBACService
}

// NewDomainHandlers creates new domain handlers
func NewDomainHandlers(domainService *services.DomainService, rbacService *services.RBACService) *DomainHandlers {
	return &DomainHandlers{
		domainService: domainService,
		rbacService:   rbacService,
	}
}

// Domain CRUD Operations

// CreateDomain adds a new custom domain
// POST /api/v1/domains
func (h *DomainHandlers) CreateDomain(c *gin.Context) {
	var req models.CreateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	domain, err := h.domainService.CreateDomain(&req, userID.(int64))
	if err != nil {
		if err.Error() == "domain already exists" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "domain_exists",
				Message: "Domain is already registered",
			})
			return
		}

		if err.Error() == "domain limit reached for your account type" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "domain_limit_reached",
				Message: "You have reached the domain limit for your account type",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "create_domain_failed",
			Message: "Failed to create domain",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Domain created successfully. Please verify domain ownership.",
		"domain":  domain,
	})
}

// GetUserDomains retrieves all domains for the current user
// GET /api/v1/domains
func (h *DomainHandlers) GetUserDomains(c *gin.Context) {
	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	domains, err := h.domainService.GetUserDomains(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_domains_failed",
			Message: "Failed to retrieve domains",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
		"count":   len(domains),
	})
}

// GetDomain retrieves a specific domain by ID
// GET /api/v1/domains/:id
func (h *DomainHandlers) GetDomain(c *gin.Context) {
	domainIDStr := c.Param("id")
	domainID, err := strconv.ParseInt(domainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_domain_id",
			Message: "Invalid domain ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	domain, err := h.domainService.GetDomainByID(domainID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "domain_not_found",
			Message: "Domain not found",
		})
		return
	}

	// Check ownership
	if domain.UserID != userID.(int64) && !h.hasSystemAccess(userID.(int64)) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "access_denied",
			Message: "You don't have access to this domain",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain": domain,
	})
}

// UpdateDomain updates domain settings
// PUT /api/v1/domains/:id
func (h *DomainHandlers) UpdateDomain(c *gin.Context) {
	domainIDStr := c.Param("id")
	domainID, err := strconv.ParseInt(domainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_domain_id",
			Message: "Invalid domain ID",
		})
		return
	}

	var req models.UpdateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	domain, err := h.domainService.UpdateDomain(domainID, &req, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: you don't own this domain" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "access_denied",
				Message: "You don't have access to this domain",
			})
			return
		}

		if err.Error() == "domain not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "domain_not_found",
				Message: "Domain not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_domain_failed",
			Message: "Failed to update domain",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain updated successfully",
		"domain":  domain,
	})
}

// DeleteDomain deletes a domain
// DELETE /api/v1/domains/:id
func (h *DomainHandlers) DeleteDomain(c *gin.Context) {
	domainIDStr := c.Param("id")
	domainID, err := strconv.ParseInt(domainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_domain_id",
			Message: "Invalid domain ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	err = h.domainService.DeleteDomain(domainID, userID.(int64))
	if err != nil {
		if err.Error() == "access denied: you don't own this domain" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "access_denied",
				Message: "You don't have access to this domain",
			})
			return
		}

		if err.Error() == "domain not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "domain_not_found",
				Message: "Domain not found",
			})
			return
		}

		if err.Error() == "cannot delete domain with active URLs" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "domain_in_use",
				Message: "Cannot delete domain with active URLs",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_domain_failed",
			Message: "Failed to delete domain",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain deleted successfully",
	})
}

// Domain Verification

// GetDomainVerification retrieves verification instructions for a domain
// GET /api/v1/domains/:id/verification
func (h *DomainHandlers) GetDomainVerification(c *gin.Context) {
	domainIDStr := c.Param("id")
	domainID, err := strconv.ParseInt(domainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_domain_id",
			Message: "Invalid domain ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Verify ownership
	domain, err := h.domainService.GetDomainByID(domainID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "domain_not_found",
			Message: "Domain not found",
		})
		return
	}

	if domain.UserID != userID.(int64) && !h.hasSystemAccess(userID.(int64)) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "access_denied",
			Message: "You don't have access to this domain",
		})
		return
	}

	verification, err := h.domainService.GetDomainVerificationInfo(domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_verification_failed",
			Message: "Failed to retrieve verification information",
		})
		return
	}

	c.JSON(http.StatusOK, verification)
}

// VerifyDomain attempts to verify domain ownership
// POST /api/v1/domains/:id/verify
func (h *DomainHandlers) VerifyDomain(c *gin.Context) {
	domainIDStr := c.Param("id")
	domainID, err := strconv.ParseInt(domainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_domain_id",
			Message: "Invalid domain ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Verify ownership
	domain, err := h.domainService.GetDomainByID(domainID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "domain_not_found",
			Message: "Domain not found",
		})
		return
	}

	if domain.UserID != userID.(int64) && !h.hasSystemAccess(userID.(int64)) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "access_denied",
			Message: "You don't have access to this domain",
		})
		return
	}

	err = h.domainService.VerifyDomain(domainID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "verification_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Domain verified successfully",
	})
}

// Domain Configuration

// GetDomainSettings retrieves complete domain configuration
// GET /api/v1/domains/:id/settings
func (h *DomainHandlers) GetDomainSettings(c *gin.Context) {
	domainIDStr := c.Param("id")
	domainID, err := strconv.ParseInt(domainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_domain_id",
			Message: "Invalid domain ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Verify ownership
	domain, err := h.domainService.GetDomainByID(domainID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "domain_not_found",
			Message: "Domain not found",
		})
		return
	}

	if domain.UserID != userID.(int64) && !h.hasSystemAccess(userID.(int64)) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "access_denied",
			Message: "You don't have access to this domain",
		})
		return
	}

	settings, err := h.domainService.GetDomainSettings(domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_settings_failed",
			Message: "Failed to retrieve domain settings",
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetDomainStatus checks the current status of a domain
// GET /api/v1/domains/:id/status
func (h *DomainHandlers) GetDomainStatus(c *gin.Context) {
	domainIDStr := c.Param("id")
	domainID, err := strconv.ParseInt(domainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_domain_id",
			Message: "Invalid domain ID",
		})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	// Verify ownership
	domain, err := h.domainService.GetDomainByID(domainID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "domain_not_found",
			Message: "Domain not found",
		})
		return
	}

	if domain.UserID != userID.(int64) && !h.hasSystemAccess(userID.(int64)) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "access_denied",
			Message: "You don't have access to this domain",
		})
		return
	}

	status, err := h.domainService.GetDomainStatus(domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_status_failed",
			Message: "Failed to retrieve domain status",
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Account Management

// GetDomainUsage retrieves domain usage and limits for the user
// GET /api/v1/domains/usage
func (h *DomainHandlers) GetDomainUsage(c *gin.Context) {
	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	usage, err := h.domainService.GetDomainUsage(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "get_usage_failed",
			Message: "Failed to retrieve domain usage",
		})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// Admin Operations

// GetAllDomains retrieves all domains (admin only)
// GET /api/v1/admin/domains
func (h *DomainHandlers) GetAllDomains(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// This would be implemented with proper database queries
	// For now, return simulated data
	domains := []*models.CustomDomain{}
	total := 0

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
		"pagination": gin.H{
			"current_page":   page,
			"total_pages":    totalPages,
			"total_items":    total,
			"items_per_page": limit,
			"has_next":       hasNext,
			"has_previous":   hasPrev,
		},
		"filters": gin.H{
			"status": status,
		},
	})
}

// GetDomainStats retrieves domain statistics (admin only)
// GET /api/v1/admin/domains/stats
func (h *DomainHandlers) GetDomainStats(c *gin.Context) {
	// This would aggregate statistics from database
	// For now, return simulated stats
	stats := &models.DomainStats{
		TotalDomains:    157,
		ActiveDomains:   143,
		VerifiedDomains: 139,
		SSLDomains:      135,
		PendingDomains:  8,
		FailedDomains:   6,
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// Helper Methods

// hasSystemAccess checks if user has system-level access to all domains
func (h *DomainHandlers) hasSystemAccess(userID int64) bool {
	hasPermission, _ := h.rbacService.UserHasPermission(userID, models.PermissionSystemManage)
	return hasPermission
}