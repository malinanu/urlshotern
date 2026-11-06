package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
)

// Request body limits for microsite endpoints
const (
	micrositeMaxJSONBody  = 256 * 1024
	micrositeMaxEventBody = 1024
)

// MicrositeHandlers serves the page builder API and the public /m/<slug> pages.
type MicrositeHandlers struct {
	sites       *services.MicrositeService
	analytics   *services.MicrositeAnalyticsService
	assets      *services.MicrositeAssetService
	billing     *services.BillingService
	shortener   *services.ShortenerService
	botDetector *services.BotDetectionService
	uaParser    *services.UserAgentService
	refParser   *services.ReferrerParsingService
	baseURL     string
	uploadRoot  string
	forms             *services.MicrositeFormService // nil: form blocks render but submissions are refused
	microPages        *services.MicroPageService
	microPageHandlers *MicroPageHandlers
}

// SetForms wires the form submission service.
func (h *MicrositeHandlers) SetForms(forms *services.MicrositeFormService) { h.forms = forms }

// SetMicroPageHandlers wires the micro-page service and handlers for /m/:slug fallback.
func (h *MicrositeHandlers) SetMicroPageHandlers(microPages *services.MicroPageService, handlers *MicroPageHandlers) {
	h.microPages = microPages
	h.microPageHandlers = handlers
}

// NewMicrositeHandlers wires the microsite handlers. Any of billing,
// shortener, botDetector may be nil (features degrade gracefully).
func NewMicrositeHandlers(
	sites *services.MicrositeService,
	analytics *services.MicrositeAnalyticsService,
	assets *services.MicrositeAssetService,
	billing *services.BillingService,
	shortener *services.ShortenerService,
	botDetector *services.BotDetectionService,
	baseURL string,
	uploadRoot string,
) *MicrositeHandlers {
	if billing != nil {
		sites.SetEntitlementsFunc(MicrositeEntitlementsFromBilling(billing))
	}
	return &MicrositeHandlers{
		sites:       sites,
		analytics:   analytics,
		assets:      assets,
		billing:     billing,
		shortener:   shortener,
		botDetector: botDetector,
		uaParser:    services.NewUserAgentService(),
		refParser:   services.NewReferrerParsingService(),
		baseURL:     strings.TrimRight(baseURL, "/"),
		uploadRoot:  uploadRoot,
	}
}

// OwnershipMiddleware requires the caller to own the page in the :id param.
func (h *MicrositeHandlers) OwnershipMiddleware() gin.HandlerFunc {
	return middleware.RequireMicrositeOwnership(h.sites, services.ErrMicrositesUnavailable)
}

// MicrositeEntitlementsFromBilling derives page capabilities from the owner's
// subscription. Errors fall back to free-plan capabilities.
func MicrositeEntitlementsFromBilling(billing *services.BillingService) services.MicrositeEntitlementsFunc {
	return func(userID int64) models.MicrositeEntitlements {
		planType, limits, err := billing.GetPagePlanLimits(userID)
		if err != nil {
			planType, limits = "free", services.PredefinedPlans["free"].Limits
		}
		return models.MicrositeEntitlements{
			MaxPages:       limits.MaxPages,
			CanRemoveBadge: middleware.PlanHasFeature(planType, middleware.FeatureRemoveBranding),
			CanUseForms:    middleware.PlanHasFeature(planType, middleware.FeatureMicrositeForms),
			AnalyticsDays:  limits.Analytics,
		}
	}
}

func (h *MicrositeHandlers) entitlementsWithUsage(userID int64) models.MicrositeEntitlements {
	ent := h.sites.Entitlements(userID)
	if h.billing != nil {
		if used, err := h.billing.CountUserPages(userID); err == nil {
			ent.UsedPages = used
		}
	}
	ent.CanCreateMore = ent.MaxPages == -1 || ent.UsedPages < ent.MaxPages
	return ent
}

func (h *MicrositeHandlers) publicURL(slug string) string {
	return h.baseURL + "/m/" + slug
}

type micrositeResponse struct {
	*models.Microsite
	PublicURL string `json:"public_url"`
}

func (h *MicrositeHandlers) wrap(m *models.Microsite) micrositeResponse {
	return micrositeResponse{Microsite: m, PublicURL: h.publicURL(m.Slug)}
}

// ListMicrosites handles GET /api/v1/microsites
func (h *MicrositeHandlers) ListMicrosites(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	sites, err := h.sites.List(userID)
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"microsites":   sites,
		"entitlements": h.entitlementsWithUsage(userID),
		"base_url":     h.baseURL + "/m/",
	})
}

// ListMicrositeTemplates handles GET /api/v1/microsites/templates
func (h *MicrositeHandlers) ListMicrositeTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"templates": services.MicrositeTemplates})
}

// PreviewMicrositeTemplate handles GET /api/v1/microsites/templates/:template/preview.
// It renders a template's sample content for the create dialog, using the
// caller's name and time zone when given.
func (h *MicrositeHandlers) PreviewMicrositeTemplate(c *gin.Context) {
	id := c.Param("template")
	if !services.IsMicrositeTemplate(id) {
		writeMicrositeError(c, services.ErrMicrositeInvalidTmpl)
		return
	}
	name := c.Query("name")
	if len([]rune(name)) > 80 {
		name = string([]rune(name)[:80])
	}
	content := services.BuildMicrositeTemplate(id, strings.TrimSpace(name), c.Query("tz"))
	html, err := services.RenderMicrosite(content, services.MicrositeRenderOptions{
		BaseURL:   h.baseURL,
		Title:     name,
		ShowBadge: true,
		Preview:   true,
	})
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.Header("Cache-Control", "private, max-age=300")
	c.JSON(http.StatusOK, gin.H{"html": string(html)})
}

// CreateMicrosite handles POST /api/v1/microsites
func (h *MicrositeHandlers) CreateMicrosite(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req models.CreateMicrositeRequest
	if !bindMicrositeJSON(c, &req) {
		return
	}
	if h.billing != nil {
		ok, msg, err := h.billing.CanCreatePage(userID)
		if err != nil {
			writeMicrositeError(c, err)
			return
		}
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "plan_limit_reached", "message": msg})
			return
		}
	}
	m, err := h.sites.Create(userID, req)
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, h.wrap(m))
}

// GetMicrosite handles GET /api/v1/microsites/:id
func (h *MicrositeHandlers) GetMicrosite(c *gin.Context) {
	m, err := h.sites.Get(c.GetInt64("microsite_id"))
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"microsite":    h.wrap(m),
		"entitlements": h.entitlementsWithUsage(m.UserID),
	})
}

// UpdateMicrosite handles PUT /api/v1/microsites/:id (draft save)
func (h *MicrositeHandlers) UpdateMicrosite(c *gin.Context) {
	var req models.UpdateMicrositeRequest
	if !bindMicrositeJSON(c, &req) {
		return
	}
	m, err := h.sites.UpdateDraft(c.GetInt64("microsite_id"), req)
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.wrap(m))
}

// DeleteMicrosite handles DELETE /api/v1/microsites/:id
func (h *MicrositeHandlers) DeleteMicrosite(c *gin.Context) {
	if err := h.sites.Delete(c.GetInt64("microsite_id")); err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Page deleted"})
}

// PublishMicrosite handles POST /api/v1/microsites/:id/publish
func (h *MicrositeHandlers) PublishMicrosite(c *gin.Context) {
	m, err := h.sites.Publish(c.GetInt64("microsite_id"))
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.wrap(m))
}

// UnpublishMicrosite handles POST /api/v1/microsites/:id/unpublish
func (h *MicrositeHandlers) UnpublishMicrosite(c *gin.Context) {
	m, err := h.sites.Unpublish(c.GetInt64("microsite_id"))
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.wrap(m))
}

// CheckMicrositeSlug handles GET /api/v1/microsites/slug-available?slug=&exclude_id=
func (h *MicrositeHandlers) CheckMicrositeSlug(c *gin.Context) {
	exclude, _ := strconv.ParseInt(c.Query("exclude_id"), 10, 64)
	slug, available, err := h.sites.SlugAvailable(c.Query("slug"), exclude)
	var verr *services.MicrositeValidationError
	if errors.As(err, &verr) {
		c.JSON(http.StatusOK, gin.H{"slug": slug, "available": false, "reason": verr.Message})
		return
	}
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"slug": slug, "available": available})
}

type micrositePreviewRequest struct {
	Title   string                  `json:"title"`
	Slug    string                  `json:"slug"`
	Content models.MicrositeContent `json:"content"`
}

// PreviewMicrosite handles POST /api/v1/microsites/preview. It renders unsaved
// content for the builder's sandboxed srcdoc iframe. Invalid blocks are skipped
// and the first validation error is returned alongside the HTML.
func (h *MicrositeHandlers) PreviewMicrosite(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req micrositePreviewRequest
	if !bindMicrositeJSON(c, &req) {
		return
	}
	var validation interface{}
	if _, err := services.NormalizeMicrositeDraft(req.Content); err != nil {
		var verr *services.MicrositeValidationError
		if errors.As(err, &verr) {
			validation = gin.H{"field": verr.Field, "message": verr.Message}
		}
	}
	ent := h.sites.Entitlements(userID)
	html, err := services.RenderMicrosite(req.Content, services.MicrositeRenderOptions{
		BaseURL:      h.baseURL,
		Slug:         strings.ToLower(strings.TrimSpace(req.Slug)),
		Title:        req.Title,
		ShowBadge:    !(req.Content.Settings.HideBadge && ent.CanRemoveBadge),
		Preview:      true,
		FormsEnabled: ent.CanUseForms,
	})
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"html": string(html), "validation_error": validation})
}

// GetMicrositeAnalytics handles GET /api/v1/microsites/:id/analytics?days=30
func (h *MicrositeHandlers) GetMicrositeAnalytics(c *gin.Context) {
	m, err := h.sites.Get(c.GetInt64("microsite_id"))
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	days, err := strconv.Atoi(c.DefaultQuery("days", "30"))
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: "days must be between 1 and 365"})
		return
	}
	if maxDays := h.sites.Entitlements(m.UserID).AnalyticsDays; maxDays > 0 && days > maxDays {
		days = maxDays
	}
	report, err := h.analytics.GetReport(m.ID, days, services.MicrositeBlockLabels(m.Draft))
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, report)
}

// UploadMicrositeAsset handles POST /api/v1/microsites/:id/assets (multipart "file")
func (h *MicrositeHandlers) UploadMicrositeAsset(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	micrositeID := c.GetInt64("microsite_id")

	// Cap the whole request body; multipart overhead is small.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.assets.MaxBytes()+64*1024)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeMicrositeError(c, services.ErrAssetTooLarge)
			return
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: "Upload an image in the \"file\" field"})
		return
	}
	defer file.Close()

	asset, err := h.assets.Upload(userID, &micrositeID, file)
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, asset)
}

// ListMicrositeAssets handles GET /api/v1/microsites/assets
func (h *MicrositeHandlers) ListMicrositeAssets(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	assets, err := h.assets.List(userID)
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"assets": assets})
}

// DeleteMicrositeAsset handles DELETE /api/v1/microsites/assets/:assetId
func (h *MicrositeHandlers) DeleteMicrositeAsset(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	assetID, err := strconv.ParseInt(c.Param("assetId"), 10, 64)
	if err != nil || assetID <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: "Invalid image ID"})
		return
	}
	if err := h.assets.Delete(userID, assetID); err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Image deleted"})
}

// CreateMicrositeShortLink handles POST /api/v1/microsites/:id/short-link.
// It creates (once) a short link that redirects to the page's /m/<slug> URL.
func (h *MicrositeHandlers) CreateMicrositeShortLink(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	m, err := h.sites.Get(c.GetInt64("microsite_id"))
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	if m.ShortCode != nil && *m.ShortCode != "" {
		c.JSON(http.StatusOK, gin.H{"short_code": *m.ShortCode, "short_url": h.baseURL + "/" + *m.ShortCode})
		return
	}
	if h.shortener == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Error: "unavailable", Message: "Short links are unavailable"})
		return
	}
	if h.billing != nil {
		ok, msg, err := h.billing.CanCreateURL(userID)
		if err != nil {
			writeMicrositeError(c, err)
			return
		}
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "plan_limit_reached", "message": msg})
			return
		}
	}
	mapping, err := h.shortener.CreateInternalShortURL(h.publicURL(m.Slug), userID, getClientIP(c))
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	if err := h.sites.SetShortCode(m.ID, mapping.ShortCode); err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"short_code": mapping.ShortCode, "short_url": h.baseURL + "/" + mapping.ShortCode})
}

// bindMicrositeJSON decodes a size-capped JSON body, writing a 400 on failure.
func bindMicrositeJSON(c *gin.Context, dst interface{}) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, micrositeMaxJSONBody)
	if err := json.NewDecoder(c.Request.Body).Decode(dst); err != nil {
		msg := "Invalid JSON body"
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			msg = "Request body is too large"
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: msg})
		return false
	}
	return true
}

// writeMicrositeError maps service errors to HTTP responses.
func writeMicrositeError(c *gin.Context, err error) {
	var verr *services.MicrositeValidationError
	switch {
	case errors.As(err, &verr):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation_error", "message": verr.Message, "field": verr.Field})
	case errors.Is(err, services.ErrMicrositeNotFound), errors.Is(err, services.ErrAssetNotFound),
		errors.Is(err, services.ErrSubmissionNotFound):
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: err.Error()})
	case errors.Is(err, services.ErrMicrositeSlugTaken):
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "slug_taken", Message: err.Error()})
	case errors.Is(err, services.ErrMicrositeVersion):
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "version_conflict", Message: err.Error()})
	case errors.Is(err, services.ErrMicrositeBadgeLocked), errors.Is(err, services.ErrMicrositeFormsLocked):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "plan_upgrade_required", Message: err.Error()})
	case errors.Is(err, services.ErrMicrositeInvalidTmpl):
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: err.Error()})
	case errors.Is(err, services.ErrAssetTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, models.ErrorResponse{Error: "file_too_large", Message: err.Error()})
	case errors.Is(err, services.ErrAssetType):
		c.JSON(http.StatusUnsupportedMediaType, models.ErrorResponse{Error: "unsupported_type", Message: err.Error()})
	case errors.Is(err, services.ErrAssetDimensions), errors.Is(err, services.ErrAssetCorrupt):
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_image", Message: err.Error()})
	case errors.Is(err, services.ErrAssetQuota):
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "quota_exceeded", Message: err.Error()})
	case errors.Is(err, services.ErrMicrositesUnavailable):
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Error: "unavailable", Message: err.Error()})
	default:
		log.Printf("microsites: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Something went wrong"})
	}
}
