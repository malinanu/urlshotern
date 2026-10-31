package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// Microsite errors
var (
	ErrMicrositeNotFound     = errors.New("page not found")
	ErrMicrositeSlugTaken    = errors.New("that address is already taken")
	ErrMicrositeVersion      = errors.New("this page was changed elsewhere; reload to get the latest version")
	ErrMicrositesUnavailable = errors.New("pages are temporarily unavailable")
	ErrMicrositeBadgeLocked  = errors.New("removing the Trunc badge requires the Ultra plan")
	ErrMicrositeInvalidTmpl  = errors.New("unknown template")
)

const (
	micrositeCacheTTL         = 5 * time.Minute
	micrositeNegativeCacheTTL = 60 * time.Second
	micrositeCachePrefix      = "microsite:pub:"
	micrositeCacheMissing     = "-"
)

// micrositeDB is the subset of *storage.PostgresStorage the service uses.
type micrositeDB interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// micrositeCache is the subset of *storage.RedisStorage the service uses.
type micrositeCache interface {
	Get(key string) (string, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Del(keys ...string) error
}

// MicrositeEntitlementsFunc resolves a page owner's plan capabilities.
type MicrositeEntitlementsFunc func(userID int64) models.MicrositeEntitlements

// PublishedMicrosite is the cached, render-ready view of a published page.
type PublishedMicrosite struct {
	ID          int64                   `json:"id"`
	UserID      int64                   `json:"user_id"`
	Slug        string                  `json:"slug"`
	Title       string                  `json:"title"`
	Content     models.MicrositeContent `json:"content"`
	ShowBadge   bool                    `json:"show_badge"`
	PublishedAt time.Time               `json:"published_at"`
}

// MicrositeService manages micro-landing pages.
type MicrositeService struct {
	db           micrositeDB
	cache        micrositeCache
	entitlements MicrositeEntitlementsFunc
}

// NewMicrositeService creates a microsite service. redis may be nil.
func NewMicrositeService(db *storage.PostgresStorage, redis *storage.RedisStorage) *MicrositeService {
	s := &MicrositeService{db: db}
	if redis != nil {
		s.cache = redis
	}
	return s
}

// SetEntitlementsFunc wires the plan lookup used for badge enforcement.
func (s *MicrositeService) SetEntitlementsFunc(f MicrositeEntitlementsFunc) {
	s.entitlements = f
}

// Entitlements returns the plan capabilities for a user (free if unknown).
func (s *MicrositeService) Entitlements(userID int64) models.MicrositeEntitlements {
	if s.entitlements == nil {
		return models.MicrositeEntitlements{MaxPages: 1, AnalyticsDays: 30}
	}
	return s.entitlements(userID)
}

// Ready reports whether the microsite tables exist (migration 046 applied).
func (s *MicrositeService) Ready() error {
	var one int
	err := s.db.QueryRow(`SELECT 1 FROM microsites LIMIT 1`).Scan(&one)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return mapMicrositeDBError(err)
	}
	return nil
}

const micrositeColumns = `id, user_id, slug, title, template, status, draft, published, version,
	short_code, published_at, created_at, updated_at`

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanMicrosite(row rowScanner) (*models.Microsite, error) {
	var (
		m         models.Microsite
		draft     []byte
		published []byte
		shortCode sql.NullString
		pubAt     sql.NullTime
	)
	if err := row.Scan(&m.ID, &m.UserID, &m.Slug, &m.Title, &m.Template, &m.Status, &draft, &published,
		&m.Version, &shortCode, &pubAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	if len(draft) > 0 {
		if err := json.Unmarshal(draft, &m.Draft); err != nil {
			return nil, fmt.Errorf("decode draft: %w", err)
		}
	}
	if m.Draft.Blocks == nil {
		m.Draft.Blocks = []models.MicrositeBlock{}
	}
	if len(published) > 0 {
		var p models.MicrositeContent
		if err := json.Unmarshal(published, &p); err != nil {
			return nil, fmt.Errorf("decode published: %w", err)
		}
		m.Published = &p
	}
	if shortCode.Valid {
		m.ShortCode = &shortCode.String
	}
	if pubAt.Valid {
		m.PublishedAt = &pubAt.Time
	}
	return &m, nil
}

// List returns a user's pages, newest first, with 30-day view counts.
func (s *MicrositeService) List(userID int64) ([]models.MicrositeSummary, error) {
	rows, err := s.db.Query(`
		SELECT m.id, m.slug, m.title, m.template, m.status, m.short_code, m.published_at, m.updated_at,
		       (SELECT COUNT(*) FROM microsite_events e
		         WHERE e.microsite_id = m.id AND e.event_type = 'view' AND e.is_bot = FALSE
		           AND e.occurred_at >= NOW() - INTERVAL '30 days') AS views_30d
		  FROM microsites m
		 WHERE m.user_id = $1 AND m.deleted_at IS NULL
		 ORDER BY m.updated_at DESC`, userID)
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	defer rows.Close()

	out := []models.MicrositeSummary{}
	for rows.Next() {
		var (
			ms        models.MicrositeSummary
			shortCode sql.NullString
			pubAt     sql.NullTime
		)
		if err := rows.Scan(&ms.ID, &ms.Slug, &ms.Title, &ms.Template, &ms.Status, &shortCode, &pubAt, &ms.UpdatedAt, &ms.Views30d); err != nil {
			return nil, err
		}
		if shortCode.Valid {
			ms.ShortCode = &shortCode.String
		}
		if pubAt.Valid {
			ms.PublishedAt = &pubAt.Time
		}
		out = append(out, ms)
	}
	return out, rows.Err()
}

// Get returns a live (not deleted) page by ID.
func (s *MicrositeService) Get(id int64) (*models.Microsite, error) {
	m, err := scanMicrosite(s.db.QueryRow(
		`SELECT `+micrositeColumns+` FROM microsites WHERE id = $1 AND deleted_at IS NULL`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMicrositeNotFound
	}
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	return m, nil
}

// GetOwnerID returns the owning user of a live page (used by the ownership middleware).
func (s *MicrositeService) GetOwnerID(id int64) (int64, error) {
	var userID int64
	err := s.db.QueryRow(`SELECT user_id FROM microsites WHERE id = $1 AND deleted_at IS NULL`, id).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrMicrositeNotFound
	}
	if err != nil {
		return 0, mapMicrositeDBError(err)
	}
	return userID, nil
}

// Create creates a draft page from a template. Plan limits are checked by the caller.
func (s *MicrositeService) Create(userID int64, req models.CreateMicrositeRequest) (*models.Microsite, error) {
	slug, err := NormalizeMicrositeSlug(req.Slug)
	if err != nil {
		return nil, err
	}
	title, err := cleanText(req.Title, 120, false, "title")
	if err != nil {
		return nil, err
	}
	tmpl := req.Template
	if tmpl == "" {
		tmpl = MicrositeTemplateBlank
	}
	if !IsMicrositeTemplate(tmpl) {
		return nil, ErrMicrositeInvalidTmpl
	}
	if title == "" {
		title = slug
	}

	content, err := NormalizeMicrositeDraft(BuildMicrositeTemplate(tmpl, title))
	if err != nil {
		return nil, err
	}
	draft, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	m, err := scanMicrosite(s.db.QueryRow(`
		INSERT INTO microsites (user_id, slug, title, template, draft)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+micrositeColumns, userID, slug, title, tmpl, draft))
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	return m, nil
}

// UpdateDraft saves title/slug/content changes to the draft. The stored
// version must equal req.Version, otherwise ErrMicrositeVersion is returned.
func (s *MicrositeService) UpdateDraft(id int64, req models.UpdateMicrositeRequest) (*models.Microsite, error) {
	current, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if req.Version != current.Version {
		return nil, ErrMicrositeVersion
	}

	title := current.Title
	if req.Title != nil {
		if title, err = cleanText(*req.Title, 120, false, "title"); err != nil {
			return nil, err
		}
		if title == "" {
			return nil, validationErr("title", "is required")
		}
	}
	slug := current.Slug
	if req.Slug != nil {
		if slug, err = NormalizeMicrositeSlug(*req.Slug); err != nil {
			return nil, err
		}
	}
	content := current.Draft
	if req.Content != nil {
		if content, err = NormalizeMicrositeDraft(*req.Content); err != nil {
			return nil, err
		}
		if content.Settings.HideBadge && !s.Entitlements(current.UserID).CanRemoveBadge {
			// Turning the badge off requires the entitlement. A previously
			// saved hide_badge (e.g. before a downgrade) is quietly cleared
			// so other edits are not blocked.
			if !current.Draft.Settings.HideBadge {
				return nil, ErrMicrositeBadgeLocked
			}
			content.Settings.HideBadge = false
		}
	}
	draft, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	updated, err := scanMicrosite(s.db.QueryRow(`
		UPDATE microsites
		   SET title = $1, slug = $2, draft = $3, version = version + 1, updated_at = NOW()
		 WHERE id = $4 AND version = $5 AND deleted_at IS NULL
		RETURNING `+micrositeColumns, title, slug, draft, id, req.Version))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMicrositeVersion
	}
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	if slug != current.Slug {
		s.invalidate(current.Slug, slug)
	}
	return updated, nil
}

// Publish validates the draft and makes it the publicly served snapshot.
func (s *MicrositeService) Publish(id int64) (*models.Microsite, error) {
	current, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	content, err := NormalizeMicrositeContent(current.Draft)
	if err != nil {
		return nil, err
	}
	if content.Settings.HideBadge && !s.Entitlements(current.UserID).CanRemoveBadge {
		content.Settings.HideBadge = false
	}
	snapshot, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	m, err := scanMicrosite(s.db.QueryRow(`
		UPDATE microsites
		   SET published = $1, status = 'published', published_at = NOW(), updated_at = NOW()
		 WHERE id = $2 AND deleted_at IS NULL
		RETURNING `+micrositeColumns, snapshot, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMicrositeNotFound
	}
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	s.invalidate(m.Slug)
	return m, nil
}

// Unpublish takes a page offline (it keeps its draft and slug).
func (s *MicrositeService) Unpublish(id int64) (*models.Microsite, error) {
	m, err := scanMicrosite(s.db.QueryRow(`
		UPDATE microsites
		   SET published = NULL, status = 'draft', updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+micrositeColumns, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMicrositeNotFound
	}
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	s.invalidate(m.Slug)
	return m, nil
}

// Delete soft-deletes a page; its slug becomes available again.
func (s *MicrositeService) Delete(id int64) error {
	var slug string
	err := s.db.QueryRow(`
		UPDATE microsites
		   SET deleted_at = NOW(), published = NULL, status = 'draft', updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL
		RETURNING slug`, id).Scan(&slug)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrMicrositeNotFound
	}
	if err != nil {
		return mapMicrositeDBError(err)
	}
	s.invalidate(slug)
	return nil
}

// SetShortCode links a short link to a page.
func (s *MicrositeService) SetShortCode(id int64, shortCode string) error {
	res, err := s.db.Exec(`UPDATE microsites SET short_code = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`, shortCode, id)
	if err != nil {
		return mapMicrositeDBError(err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrMicrositeNotFound
	}
	return nil
}

// SlugAvailable reports whether a slug is valid and unclaimed. excludeID lets a
// page keep its own slug.
func (s *MicrositeService) SlugAvailable(raw string, excludeID int64) (string, bool, error) {
	slug, err := NormalizeMicrositeSlug(raw)
	if err != nil {
		return "", false, err
	}
	var taken bool
	err = s.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM microsites WHERE lower(slug) = $1 AND deleted_at IS NULL AND id <> $2)`,
		slug, excludeID).Scan(&taken)
	if err != nil {
		return slug, false, mapMicrositeDBError(err)
	}
	return slug, !taken, nil
}

// GetPublishedBySlug returns the render-ready published page for a slug.
// Results (including misses) are cached briefly in Redis.
func (s *MicrositeService) GetPublishedBySlug(rawSlug string) (*PublishedMicrosite, error) {
	slug := strings.ToLower(strings.TrimSpace(rawSlug))
	if !slugPattern.MatchString(slug) {
		return nil, ErrMicrositeNotFound
	}
	key := micrositeCachePrefix + slug
	if s.cache != nil {
		if cached, err := s.cache.Get(key); err == nil {
			if cached == micrositeCacheMissing {
				return nil, ErrMicrositeNotFound
			}
			var p PublishedMicrosite
			if json.Unmarshal([]byte(cached), &p) == nil {
				return &p, nil
			}
		}
	}

	var (
		p         PublishedMicrosite
		published []byte
	)
	err := s.db.QueryRow(`
		SELECT id, user_id, slug, title, published, published_at
		  FROM microsites
		 WHERE lower(slug) = $1 AND status = 'published' AND published IS NOT NULL AND deleted_at IS NULL`,
		slug).Scan(&p.ID, &p.UserID, &p.Slug, &p.Title, &published, &p.PublishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		if s.cache != nil {
			_ = s.cache.Set(key, micrositeCacheMissing, micrositeNegativeCacheTTL)
		}
		return nil, ErrMicrositeNotFound
	}
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	if err := json.Unmarshal(published, &p.Content); err != nil {
		return nil, fmt.Errorf("decode published: %w", err)
	}
	p.ShowBadge = !(p.Content.Settings.HideBadge && s.Entitlements(p.UserID).CanRemoveBadge)

	if s.cache != nil {
		if encoded, err := json.Marshal(p); err == nil {
			_ = s.cache.Set(key, encoded, micrositeCacheTTL)
		}
	}
	return &p, nil
}

func (s *MicrositeService) invalidate(slugs ...string) {
	if s.cache == nil || len(slugs) == 0 {
		return
	}
	keys := make([]string, 0, len(slugs))
	for _, slug := range slugs {
		keys = append(keys, micrositeCachePrefix+strings.ToLower(slug))
	}
	_ = s.cache.Del(keys...)
}

// mapMicrositeDBError converts Postgres errors into service errors: a missing
// table (migration 046 not applied) becomes ErrMicrositesUnavailable, a unique
// violation becomes ErrMicrositeSlugTaken.
func mapMicrositeDBError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "42P01": // undefined_table
			return ErrMicrositesUnavailable
		case "23505": // unique_violation
			return ErrMicrositeSlugTaken
		}
	}
	return err
}
