package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// CampaignHandlers handles promotional campaign API endpoints.
type CampaignHandlers struct {
	svc *services.CampaignService
}

// NewCampaignHandlers creates a new CampaignHandlers.
func NewCampaignHandlers(svc *services.CampaignService) *CampaignHandlers {
	return &CampaignHandlers{svc: svc}
}

// CreateCampaign POST /api/v1/admin/campaigns
func (h *CampaignHandlers) CreateCampaign(c *gin.Context) {
	var req models.CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := middleware.GetUserID(c)
	var createdBy *int64
	if userID > 0 {
		createdBy = &userID
	}

	campaign, err := h.svc.CreateCampaign(&req, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, campaign)
}

// ListCampaigns GET /api/v1/admin/campaigns
func (h *CampaignHandlers) ListCampaigns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	campaigns, total, err := h.svc.ListCampaigns(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	c.JSON(http.StatusOK, models.CampaignListResponse{
		Campaigns:  campaigns,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// GetCampaign GET /api/v1/admin/campaigns/:id
func (h *CampaignHandlers) GetCampaign(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	campaign, err := h.svc.GetCampaign(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
		return
	}
	c.JSON(http.StatusOK, campaign)
}

// UpdateCampaign PUT /api/v1/admin/campaigns/:id
func (h *CampaignHandlers) UpdateCampaign(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	var req models.UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	campaign, err := h.svc.UpdateCampaign(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, campaign)
}

// DeleteCampaign DELETE /api/v1/admin/campaigns/:id
func (h *CampaignHandlers) DeleteCampaign(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	if err := h.svc.DeleteCampaign(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "campaign deleted"})
}

// ImportContacts POST /api/v1/admin/campaigns/:id/contacts
// Accepts JSON body (array) OR multipart CSV file field "file".
func (h *CampaignHandlers) ImportContacts(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}

	contentType := c.ContentType()
	var imported int

	switch {
	case containsStr(contentType, "multipart/form-data"):
		// CSV file upload
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing file field"})
			return
		}
		defer file.Close()
		imported, err = h.svc.ImportContactsCSV(id, file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

	default:
		// JSON array
		var rows []models.ContactImportRow
		if err := json.NewDecoder(c.Request.Body).Decode(&rows); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
			return
		}
		imported, err = h.svc.ImportContactsJSON(id, rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"imported": imported})
}

// SendCampaign POST /api/v1/admin/campaigns/:id/send
// Triggers an async campaign send and returns 202 immediately.
func (h *CampaignHandlers) SendCampaign(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	if err := h.svc.SendCampaign(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "Campaign send started"})
}

// GetCampaignStats GET /api/v1/admin/campaigns/:id/stats
func (h *CampaignHandlers) GetCampaignStats(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	stats, err := h.svc.GetCampaignStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// containsStr is a simple case-insensitive substring check.
func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
