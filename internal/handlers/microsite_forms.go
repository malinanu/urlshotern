package handlers

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/utils"
)

const micrositeMaxFormBody = 8 * 1024

var formBlockIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)

// SubmitMicrositeForm handles POST /m/:slug/submit — a newsletter or contact
// form posted from a public page. It works without JavaScript: the visitor
// is redirected back to the page (post/redirect/get) with ?sent= or
// ?form_error= so the page renders the result in place.
func (h *MicrositeHandlers) SubmitMicrositeForm(c *gin.Context) {
	page, err := h.sites.GetPublishedBySlug(c.Param("slug"))
	if err != nil {
		h.publicError(c, err)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, micrositeMaxFormBody)
	blockID := ""
	code := services.FormErrInvalid
	var res *services.SubmitResult
	if err := c.Request.ParseForm(); err == nil {
		blockID = c.Request.PostForm.Get("b")
		if h.forms == nil {
			code = services.FormErrClosed
		} else {
			ip := getClientIP(c)
			in := services.SubmissionInput{
				BlockID:     blockID,
				Token:       c.Request.PostForm.Get("t"),
				Honeypot:    c.Request.PostForm.Get("website"),
				Name:        c.Request.PostForm.Get("name"),
				Email:       c.Request.PostForm.Get("email"),
				Message:     c.Request.PostForm.Get("message"),
				VisitorHash: h.analytics.VisitorHash(ip, c.GetHeader("User-Agent"), time.Now()),
				Now:         time.Now(),
			}
			code, res, err = h.forms.Submit(page, in)
			if err != nil {
				log.Printf("microsites: form submit on %s: %v", page.Slug, err)
				code = services.FormErrInvalid
			}
		}
	}

	if !formBlockIDPattern.MatchString(blockID) {
		blockID = ""
	}
	q := url.Values{}
	if code == "" {
		q.Set("sent", blockID)
		if res != nil && res.Stored {
			ev := h.buildEvent(c, page, models.MicrositeEventFormSubmit, blockID, "")
			utils.Go("microsite_form_event", func() error { return h.analytics.RecordEvent(ev) })
			utils.Go("microsite_form_notify", func() error { return h.forms.Notify(page, res) })
		}
	} else {
		q.Set("form_error", code)
		q.Set("b", blockID)
	}
	target := "/m/" + url.PathEscape(page.Slug) + "?" + q.Encode()
	if blockID != "" {
		target += "#" + blockID
	}
	c.Header("Cache-Control", "no-store")
	c.Redirect(http.StatusSeeOther, target)
}

// ListMicrositeSubmissions handles GET /api/v1/microsites/:id/submissions
func (h *MicrositeHandlers) ListMicrositeSubmissions(c *gin.Context) {
	if h.forms == nil {
		writeMicrositeError(c, services.ErrMicrositesUnavailable)
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	subs, total, err := h.forms.List(c.GetInt64("microsite_id"), limit, offset)
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"submissions": subs, "total": total})
}

// ExportMicrositeSubmissions handles GET /api/v1/microsites/:id/submissions/export (CSV)
func (h *MicrositeHandlers) ExportMicrositeSubmissions(c *gin.Context) {
	if h.forms == nil {
		writeMicrositeError(c, services.ErrMicrositesUnavailable)
		return
	}
	id := c.GetInt64("microsite_id")
	data, err := h.forms.ExportCSV(id)
	if err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="page-`+strconv.FormatInt(id, 10)+`-submissions.csv"`)
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// DeleteMicrositeSubmission handles DELETE /api/v1/microsites/:id/submissions/:sid
func (h *MicrositeHandlers) DeleteMicrositeSubmission(c *gin.Context) {
	if h.forms == nil {
		writeMicrositeError(c, services.ErrMicrositesUnavailable)
		return
	}
	sid, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil || sid <= 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid_request", Message: "Invalid submission ID"})
		return
	}
	if err := h.forms.Delete(c.GetInt64("microsite_id"), sid); err != nil {
		writeMicrositeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Submission deleted"})
}
