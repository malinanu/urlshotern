package handlers

import (
	"encoding/json"
	"errors"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/utils"
)

var micrositeUploadPathPattern = regexp.MustCompile(`^/m/[0-9]{1,20}/[a-f0-9]{64}\.(jpg|png)$`)

// ServeMicrositePage handles GET /m/:slug — the public page.
func (h *MicrositeHandlers) ServeMicrositePage(c *gin.Context) {
	page, err := h.sites.GetPublishedBySlug(c.Param("slug"))
	if err != nil {
		h.publicError(c, err)
		return
	}
	body, err := services.RenderMicrosite(page.Content, services.MicrositeRenderOptions{
		BaseURL:   h.baseURL,
		Slug:      page.Slug,
		Title:     page.Title,
		ShowBadge: page.ShowBadge,
	})
	if err != nil {
		log.Printf("microsites: render %s: %v", page.Slug, err)
		h.publicError(c, err)
		return
	}

	ev := h.buildEvent(c, page, models.MicrositeEventView, "", "")
	utils.Go("microsite_view", func() error { return h.analytics.RecordEvent(ev) })

	// Not cached at shared caches so every view reaches the server and is counted.
	c.Header("Cache-Control", "private, no-cache")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Data(http.StatusOK, "text/html; charset=utf-8", body)
}

// ServeMicrositeVCard handles GET /m/:slug/vcard?b=<blockID> — the .vcf download.
func (h *MicrositeHandlers) ServeMicrositeVCard(c *gin.Context) {
	page, err := h.sites.GetPublishedBySlug(c.Param("slug"))
	if err != nil {
		h.publicError(c, err)
		return
	}

	var (
		card    *models.VCardActionData
		blockID string
		avatar  string
	)
	want := c.Query("b")
	for _, b := range page.Content.Blocks {
		if b.Hidden {
			continue
		}
		data, err := services.DecodeMicrositeBlockData(b.Type, b.Data, "block")
		if err != nil {
			continue
		}
		switch d := data.(type) {
		case models.ProfileHeaderData:
			if avatar == "" {
				avatar = d.AvatarURL
			}
		case models.VCardActionData:
			if card == nil && (want == "" || want == b.ID) {
				dd := d
				card, blockID = &dd, b.ID
			}
		}
	}
	if card == nil {
		h.publicError(c, services.ErrMicrositeNotFound)
		return
	}

	photoURL := firstNonEmptyStr(card.PhotoURL, avatar)
	photo := h.assets.LoadVCardPhoto(photoURL)
	if photo == nil && strings.HasPrefix(photoURL, "/") && strings.HasPrefix(h.baseURL, "https://") {
		card.PhotoURL = h.baseURL + photoURL
	} else {
		card.PhotoURL = photoURL
	}
	vcf := services.BuildVCard(*card, photo)

	ev := h.buildEvent(c, page, models.MicrositeEventVCardDownload, blockID, "")
	utils.Go("microsite_vcard", func() error { return h.analytics.RecordEvent(ev) })

	filename := services.VCardFilename(*card, page.Slug)
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "text/vcard; charset=utf-8", vcf)
}

type micrositeEventRequest struct {
	Type    string `json:"type"`
	BlockID string `json:"block_id"`
	ItemID  string `json:"item_id"`
}

// RecordMicrositeEvent handles POST /api/v1/m/:slug/event — the click beacon.
// Views and vCard downloads are recorded server-side, so only tap events are
// accepted here, and only for blocks/items that exist on the published page.
func (h *MicrositeHandlers) RecordMicrositeEvent(c *gin.Context) {
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, micrositeMaxEventBody))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	var req micrositeEventRequest
	if err := json.Unmarshal(raw, &req); err != nil || !services.IsMicrositeEventType(req.Type) ||
		req.Type == models.MicrositeEventView || req.Type == models.MicrositeEventVCardDownload {
		c.Status(http.StatusBadRequest)
		return
	}
	page, err := h.sites.GetPublishedBySlug(c.Param("slug"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !micrositeHasTarget(page.Content, req.BlockID, req.ItemID) {
		c.Status(http.StatusNoContent) // ignore stale or forged targets silently
		return
	}
	ev := h.buildEvent(c, page, req.Type, req.BlockID, req.ItemID)
	utils.Go("microsite_event", func() error { return h.analytics.RecordEvent(ev) })
	c.Status(http.StatusNoContent)
}

// ServeMicrositeBeacon handles GET /m-static/beacon.js
func (h *MicrositeHandlers) ServeMicrositeBeacon(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "application/javascript; charset=utf-8", services.MicrositeBeaconJS)
}

// ServeMicrositeUpload handles GET /uploads/m/*path. Only content-addressed,
// re-encoded JPEG/PNG files are served, with a fixed content type.
func (h *MicrositeHandlers) ServeMicrositeUpload(c *gin.Context) {
	rel := "/m" + c.Param("path")
	if !micrositeUploadPathPattern.MatchString(rel) || h.uploadRoot == "" {
		c.Status(http.StatusNotFound)
		return
	}
	full := filepath.Join(h.uploadRoot, filepath.FromSlash(rel))
	if _, err := os.Stat(full); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	contentType := "image/jpeg"
	if strings.HasSuffix(rel, ".png") {
		contentType = "image/png"
	}
	c.Header("Content-Type", contentType)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	c.Header("Cache-Control", "public, max-age=31536000, immutable") // file names are content hashes
	c.File(full)
}

// buildEvent captures request telemetry synchronously (the gin context must not
// be used after the handler returns); the insert itself runs asynchronously.
func (h *MicrositeHandlers) buildEvent(c *gin.Context, page *services.PublishedMicrosite, eventType, blockID, itemID string) models.MicrositeEvent {
	ip := getClientIP(c)
	ua := c.GetHeader("User-Agent")
	ev := models.MicrositeEvent{
		MicrositeID: page.ID,
		EventType:   eventType,
		BlockID:     blockID,
		ItemID:      itemID,
		VisitorHash: h.analytics.VisitorHash(ip, ua, time.Now()),
		Country:     c.GetHeader("CF-IPCountry"),
	}

	key := "m:" + page.Slug
	if eventType != models.MicrositeEventView {
		key += ":ev" // keep tap rate limits separate from page views
	}
	if h.botDetector != nil {
		ev.IsBot = h.botDetector.IsBot(ua, c.GetHeader("Accept-Language"), c.GetHeader("Accept"), ip, key)
		if eventType == models.MicrositeEventView && !ev.IsBot {
			ev.IsUnique = h.botDetector.IsUniqueVisitor(ev.VisitorHash, key)
		}
	}

	if dev := h.uaParser.ParseUserAgent(ua); dev != nil {
		ev.DeviceType, ev.OS, ev.Browser = dev.DeviceType, dev.OSName, dev.BrowserName
		if dev.IsBot {
			ev.IsBot = true
			ev.IsUnique = false
		}
	}

	if eventType == models.MicrositeEventView {
		ev.UTMSource = c.Query("utm_source")
		ev.UTMMedium = c.Query("utm_medium")
		ev.UTMCampaign = c.Query("utm_campaign")
		h.applyReferrer(&ev, c.GetHeader("Referer"), c.Query("src"))
	}
	return ev
}

// applyReferrer classifies the traffic source of a page view.
func (h *MicrositeHandlers) applyReferrer(ev *models.MicrositeEvent, referrer, src string) {
	if strings.EqualFold(src, "qr") || strings.EqualFold(ev.UTMMedium, "qr") {
		ev.ReferrerCategory, ev.ReferrerDomain = "qr", "QR code"
		return
	}
	if referrer != "" {
		if u, err := url.Parse(referrer); err == nil {
			if base, err := url.Parse(h.baseURL); err == nil && strings.EqualFold(u.Hostname(), base.Hostname()) {
				referrer = "" // internal navigation (e.g. a short link redirect) counts as direct
			}
		}
	}
	parsed := h.refParser.ParseReferrer(referrer)
	ev.ReferrerCategory = parsed.Category
	ev.ReferrerDomain = parsed.Domain
	if ev.ReferrerDomain == "" && ev.UTMSource != "" {
		ev.ReferrerDomain = ev.UTMSource
	}
}

// micrositeHasTarget reports whether blockID (and itemID, if given) exist on the page.
func micrositeHasTarget(content models.MicrositeContent, blockID, itemID string) bool {
	if blockID == "" {
		return false
	}
	for _, b := range content.Blocks {
		if b.ID != blockID || b.Hidden {
			continue
		}
		if itemID == "" {
			return true
		}
		data, err := services.DecodeMicrositeBlockData(b.Type, b.Data, "block")
		if err != nil {
			return false
		}
		switch d := data.(type) {
		case models.LinkListData:
			for _, it := range d.Items {
				if it.ID == itemID {
					return true
				}
			}
		case models.ContactButtonsData:
			for _, it := range d.Items {
				if it.ID == itemID {
					return true
				}
			}
		case models.SocialIconsData:
			for _, it := range d.Items {
				if it.ID == itemID {
					return true
				}
			}
		}
		return false
	}
	return false
}

// publicError renders a minimal HTML error page for public microsite routes.
func (h *MicrositeHandlers) publicError(c *gin.Context, err error) {
	status, title, msg := http.StatusNotFound, "Page not found", "This page doesn't exist or isn't published."
	if errors.Is(err, services.ErrMicrositesUnavailable) {
		status, title, msg = http.StatusServiceUnavailable, "Temporarily unavailable", "Please try again in a few minutes."
	} else if !errors.Is(err, services.ErrMicrositeNotFound) {
		log.Printf("microsites: public error: %v", err)
		status, title, msg = http.StatusInternalServerError, "Something went wrong", "Please try again later."
	}
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(status, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html lang="en"><head><meta charset="utf-8">`+
		`<meta name="viewport" content="width=device-width, initial-scale=1"><meta name="robots" content="noindex">`+
		`<title>`+html.EscapeString(title)+`</title><style>body{font-family:system-ui,sans-serif;display:flex;min-height:100vh;`+
		`align-items:center;justify-content:center;margin:0;background:#f3f4f6;color:#111827;text-align:center;padding:16px}`+
		`a{color:#2563eb}</style></head><body><main><h1>`+html.EscapeString(title)+`</h1><p>`+html.EscapeString(msg)+
		`</p><p><a href="`+html.EscapeString(h.baseURL)+`/">Go to Trunc</a></p></main></body></html>`))
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
