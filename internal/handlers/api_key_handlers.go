package handlers

import (
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIKeyHandler handles API key management operations
type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// GetAPIKeyService returns the API key service instance
func (h *APIKeyHandler) GetAPIKeyService() *services.APIKeyService {
	return h.apiKeyService
}

// CreateAPIKey creates a new API key
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req models.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	response, err := h.apiKeyService.CreateAPIKey(userID.(int64), &req)
	if err != nil {
		// Log the detailed error
		println("API Key Creation Error:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ListAPIKeys lists API keys for the authenticated user
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	query := &models.ListAPIKeysQuery{
		Limit:  20,
		Offset: 0,
	}

	// Parse query parameters
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			query.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			query.Offset = o
		}
	}

	if active := c.Query("active"); active != "" {
		if a, err := strconv.ParseBool(active); err == nil {
			query.Active = &a
		}
	}

	if expired := c.Query("expired"); expired != "" {
		if e, err := strconv.ParseBool(expired); err == nil {
			query.Expired = &e
		}
	}

	if search := c.Query("search"); search != "" {
		query.Search = &search
	}

	query.SortBy = c.DefaultQuery("sort_by", "created_at")
	query.SortOrder = c.DefaultQuery("sort_order", "desc")

	result, err := h.apiKeyService.ListAPIKeys(userID.(int64), query)
	if err != nil {
		// Log the detailed error
		println("API Key List Error:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list API keys", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAPIKey retrieves a specific API key
func (h *APIKeyHandler) GetAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	apiKey, err := h.apiKeyService.GetAPIKey(id, userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// RevokeAPIKey deactivates an API key
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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

// GetAPIKeyStats retrieves usage statistics for an API key
func (h *APIKeyHandler) GetAPIKeyStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
		return
	}

	stats, err := h.apiKeyService.GetAPIKeyStats(id, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get API key stats", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
