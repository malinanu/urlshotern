package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// Form spam limits
const (
	formMinFillTime   = 3 * time.Second
	formTokenLifetime = 24 * time.Hour
	formIPWindow      = 10 * time.Minute
	formIPLimit       = 5
	formDailyLimit    = 500
	formNotifyEvery   = time.Minute
)

// Submit error codes (shown to visitors via FormErrorMessages).
const (
	FormErrInvalid        = "invalid"
	FormErrInvalidEmail   = "invalid_email"
	FormErrMissingMessage = "missing_message"
	FormErrTooFast        = "too_fast"
	FormErrExpired        = "expired"
	FormErrRateLimited    = "rate_limited"
	FormErrClosed         = "closed"
)

// ErrSubmissionNotFound is returned when deleting a missing submission.
var ErrSubmissionNotFound = errors.New("submission not found")

// formLimiter is the subset of *storage.RedisStorage used for rate limits and
// notification throttling.
type formLimiter interface {
	IncrWindow(key string, window time.Duration) (int64, error)
	SetNX(key string, value interface{}, ttl time.Duration) (bool, error)
	Get(key string) (string, error)
	Del(keys ...string) error
}

// formMailer sends the owner's alert email (*EmailService satisfies it).
type formMailer interface {
	SendEmail(to, subject, htmlBody, textBody string) error
}

// MicrositeFormService handles newsletter and contact form submissions.
type MicrositeFormService struct {
	db      micrositeDB
	limiter formLimiter
	mailer  formMailer
	key     []byte
	baseURL string
}

// NewMicrositeFormService creates the form service. secret keys the form
// tokens; redis and mailer may be nil (no rate limits / no alerts).
func NewMicrositeFormService(db *storage.PostgresStorage, redis *storage.RedisStorage, mailer *EmailService, secret, baseURL string) *MicrositeFormService {
	sum := sha256.Sum256([]byte("microsite-form:" + secret))
	s := &MicrositeFormService{db: db, key: sum[:], baseURL: strings.TrimRight(baseURL, "/")}
	if redis != nil {
		s.limiter = redis
	}
	if mailer != nil {
		s.mailer = mailer
	}
	return s
}

// Token signs a render time for a form, so submissions can be checked for a
// plausible fill time and a bounded age without server-side state.
func (s *MicrositeFormService) Token(micrositeID int64, blockID string, now time.Time) string {
	ts := now.Unix()
	return strconv.FormatInt(ts, 10) + "." + s.sign(micrositeID, blockID, ts)
}

// VerifyToken returns "" when the token is valid, or an error code.
func (s *MicrositeFormService) VerifyToken(micrositeID int64, blockID, token string, now time.Time) string {
	tsPart, sig, ok := strings.Cut(token, ".")
	if !ok {
		return FormErrInvalid
	}
	ts, err := strconv.ParseInt(tsPart, 10, 64)
	if err != nil || !hmac.Equal([]byte(sig), []byte(s.sign(micrositeID, blockID, ts))) {
		return FormErrInvalid
	}
	age := now.Sub(time.Unix(ts, 0))
	switch {
	case age < formMinFillTime:
		return FormErrTooFast
	case age > formTokenLifetime:
		return FormErrExpired
	}
	return ""
}

func (s *MicrositeFormService) sign(micrositeID int64, blockID string, ts int64) string {
	mac := hmac.New(sha256.New, s.key)
	fmt.Fprintf(mac, "form|%d|%s|%d", micrositeID, blockID, ts)
	return hex.EncodeToString(mac.Sum(nil))[:32]
}

// SubmissionInput is one visitor submission, as posted.
type SubmissionInput struct {
	BlockID     string
	Token       string
	Honeypot    string
	Name        string
	Email       string
	Message     string
	VisitorHash string
	Now         time.Time
}

// SubmitResult describes an accepted submission.
type SubmitResult struct {
	Stored     bool // false when silently dropped (honeypot)
	Submission models.MicrositeSubmission
	Form       models.FormData
}

// Submit validates and stores a submission. It returns an error code for
// the visitor ("" on success) and, on success, the stored row.
func (s *MicrositeFormService) Submit(page *PublishedMicrosite, in SubmissionInput) (string, *SubmitResult, error) {
	if !page.FormsEnabled {
		return FormErrClosed, nil, nil
	}
	var form *models.FormData
	for _, b := range page.Content.Blocks {
		if b.Type == models.BlockTypeForm && b.ID == in.BlockID && !b.Hidden {
			if d, err := DecodeMicrositeBlockData(b.Type, b.Data, "block", true); err == nil {
				fd := d.(models.FormData)
				form = &fd
			}
		}
	}
	if form == nil {
		return FormErrInvalid, nil, nil
	}
	if strings.TrimSpace(in.Honeypot) != "" {
		return "", &SubmitResult{Stored: false, Form: *form}, nil // bots get a success page and nothing is kept
	}
	if code := s.VerifyToken(page.ID, in.BlockID, in.Token, in.Now); code != "" {
		return code, nil, nil
	}

	email, err := cleanEmail(in.Email, "email")
	if err != nil || email == "" {
		return FormErrInvalidEmail, nil, nil
	}
	sub := models.MicrositeSubmission{BlockID: in.BlockID, Kind: form.Kind, Email: email}
	if form.Kind == formKindContact {
		if sub.Name, err = cleanText(in.Name, 100, false, "name"); err != nil {
			return FormErrInvalid, nil, nil
		}
		if sub.Message, err = cleanText(in.Message, 2000, true, "message"); err != nil {
			return FormErrInvalid, nil, nil
		}
		if sub.Message == "" {
			return FormErrMissingMessage, nil, nil
		}
	}

	if s.limiter != nil {
		pageKey := strconv.FormatInt(page.ID, 10)
		if n, err := s.limiter.IncrWindow("microsite:form:ip:"+pageKey+":"+in.VisitorHash, formIPWindow); err == nil && n > formIPLimit {
			return FormErrRateLimited, nil, nil
		}
		day := in.Now.UTC().Format("20060102")
		if n, err := s.limiter.IncrWindow("microsite:form:day:"+pageKey+":"+day, 25*time.Hour); err == nil && n > formDailyLimit {
			return FormErrRateLimited, nil, nil
		}
	}

	err = s.db.QueryRow(`
		INSERT INTO microsite_submissions (microsite_id, block_id, kind, email, name, message, visitor_hash)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''))
		RETURNING id, created_at`,
		page.ID, sub.BlockID, sub.Kind, sub.Email, sub.Name, sub.Message, in.VisitorHash).Scan(&sub.ID, &sub.CreatedAt)
	if err != nil {
		return FormErrInvalid, nil, mapMicrositeDBError(err)
	}
	return "", &SubmitResult{Stored: true, Submission: sub, Form: *form}, nil
}

// Notify emails the page owner about a submission, at most once a minute per
// form; submissions in between are counted and mentioned in the next alert.
func (s *MicrositeFormService) Notify(page *PublishedMicrosite, res *SubmitResult) error {
	if s.mailer == nil || res == nil || !res.Stored || !res.Form.Notify {
		return nil
	}
	key := fmt.Sprintf("microsite:form:notify:%d:%s", page.ID, res.Submission.BlockID)
	pendingKey := key + ":pending"
	extra := 0
	if s.limiter != nil {
		ok, err := s.limiter.SetNX(key, 1, formNotifyEvery)
		if err == nil && !ok {
			_, _ = s.limiter.IncrWindow(pendingKey, time.Hour)
			return nil
		}
		if v, err := s.limiter.Get(pendingKey); err == nil {
			extra, _ = strconv.Atoi(v)
			_ = s.limiter.Del(pendingKey)
		}
	}

	var to string
	if err := s.db.QueryRow(`SELECT email FROM users WHERE id = $1`, page.UserID).Scan(&to); err != nil || to == "" {
		return err
	}
	inbox := fmt.Sprintf("%s/microsites/edit?id=%d&tab=inbox", s.baseURL, page.ID)
	what := "newsletter signup"
	if res.Form.Kind == formKindContact {
		what = "message"
	}
	subject := fmt.Sprintf("New %s on your page: %s", what, page.Title)
	sub := res.Submission

	var text, body strings.Builder
	fmt.Fprintf(&text, "You have a new %s on %s/m/%s\n\nEmail: %s\n", what, s.baseURL, page.Slug, sub.Email)
	fmt.Fprintf(&body, "<p>You have a new %s on <strong>%s</strong>.</p><p><strong>Email:</strong> %s</p>",
		what, html.EscapeString(page.Title), html.EscapeString(sub.Email))
	if sub.Name != "" {
		fmt.Fprintf(&text, "Name: %s\n", sub.Name)
		fmt.Fprintf(&body, "<p><strong>Name:</strong> %s</p>", html.EscapeString(sub.Name))
	}
	if sub.Message != "" {
		fmt.Fprintf(&text, "\n%s\n", sub.Message)
		fmt.Fprintf(&body, "<p style=\"white-space:pre-line\">%s</p>", html.EscapeString(sub.Message))
	}
	if extra > 0 {
		fmt.Fprintf(&text, "\n…and %d more since the last alert.\n", extra)
		fmt.Fprintf(&body, "<p>…and %d more since the last alert.</p>", extra)
	}
	fmt.Fprintf(&text, "\nSee all submissions: %s\n", inbox)
	fmt.Fprintf(&body, "<p><a href=\"%s\">See all submissions</a></p>", html.EscapeString(inbox))
	return s.mailer.SendEmail(to, subject, body.String(), text.String())
}

// List returns a page's submissions, newest first, with the total count.
func (s *MicrositeFormService) List(micrositeID int64, limit, offset int) ([]models.MicrositeSubmission, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM microsite_submissions WHERE microsite_id = $1`, micrositeID).Scan(&total); err != nil {
		return nil, 0, mapMicrositeDBError(err)
	}
	rows, err := s.db.Query(`
		SELECT id, block_id, kind, email, COALESCE(name, ''), COALESCE(message, ''), created_at
		  FROM microsite_submissions WHERE microsite_id = $1
		 ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`, micrositeID, limit, offset)
	if err != nil {
		return nil, 0, mapMicrositeDBError(err)
	}
	defer rows.Close()
	out := []models.MicrositeSubmission{}
	for rows.Next() {
		var m models.MicrositeSubmission
		if err := rows.Scan(&m.ID, &m.BlockID, &m.Kind, &m.Email, &m.Name, &m.Message, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

// Delete removes one submission of a page.
func (s *MicrositeFormService) Delete(micrositeID, submissionID int64) error {
	res, err := s.db.Exec(`DELETE FROM microsite_submissions WHERE id = $1 AND microsite_id = $2`, submissionID, micrositeID)
	if err != nil {
		return mapMicrositeDBError(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrSubmissionNotFound
	}
	return nil
}

// ExportCSV returns all of a page's submissions as CSV, with every cell
// guarded against spreadsheet formula injection.
func (s *MicrositeFormService) ExportCSV(micrositeID int64) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"received_at", "kind", "email", "name", "message", "form_block"})
	offset := 0
	for {
		page, _, err := s.List(micrositeID, 200, offset)
		if err != nil {
			return nil, err
		}
		for _, m := range page {
			_ = w.Write([]string{
				m.CreatedAt.UTC().Format(time.RFC3339), m.Kind, CSVSafe(m.Email), CSVSafe(m.Name), CSVSafe(m.Message), CSVSafe(m.BlockID),
			})
		}
		if len(page) < 200 {
			break
		}
		offset += len(page)
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// CSVSafe neutralises values a spreadsheet would treat as a formula.
func CSVSafe(v string) string {
	if v != "" && strings.ContainsRune("=+-@\t\r", rune(v[0])) {
		return "'" + v
	}
	return v
}

// Ready reports whether migration 047 (microsite_submissions) is applied.
func (s *MicrositeFormService) Ready() error {
	var ok bool
	err := s.db.QueryRow(`SELECT to_regclass('public.microsite_submissions') IS NOT NULL`).Scan(&ok)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if !ok {
		return ErrMicrositesUnavailable
	}
	return nil
}
