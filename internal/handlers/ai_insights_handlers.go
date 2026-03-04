package handlers

import (
	"net/http"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/gin-gonic/gin"
)

// AIInsightsHandler handles AI-powered insight requests (enterprise-gated).
type AIInsightsHandler struct {
	service *services.AIInsightsService
}

func NewAIInsightsHandler(svc *services.AIInsightsService) *AIInsightsHandler {
	return &AIInsightsHandler{service: svc}
}

func (h *AIInsightsHandler) GetSummary(c *gin.Context) {
	h.respond(c, func(shortCode string, userID int64) (*models.AIInsight, error) {
		return h.service.GetSummary(shortCode, userID)
	})
}

func (h *AIInsightsHandler) GetAnomalies(c *gin.Context) {
	h.respond(c, func(shortCode string, userID int64) (*models.AIInsight, error) {
		return h.service.GetAnomalies(shortCode, userID)
	})
}

func (h *AIInsightsHandler) GetRecommendations(c *gin.Context) {
	h.respond(c, func(shortCode string, userID int64) (*models.AIInsight, error) {
		return h.service.GetRecommendations(shortCode, userID)
	})
}

func (h *AIInsightsHandler) GetForecast(c *gin.Context) {
	h.respond(c, func(shortCode string, userID int64) (*models.AIInsight, error) {
		return h.service.GetForecast(shortCode, userID)
	})
}

func (h *AIInsightsHandler) GetAudienceProfile(c *gin.Context) {
	h.respond(c, func(shortCode string, userID int64) (*models.AIInsight, error) {
		return h.service.GetAudienceProfile(shortCode, userID)
	})
}

func (h *AIInsightsHandler) respond(c *gin.Context, fn func(string, int64) (*models.AIInsight, error)) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: "shortCode is required"})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "authentication_required", Message: "Authentication required"})
		return
	}

	insight, err := fn(shortCode, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "insight_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, insight)
}
