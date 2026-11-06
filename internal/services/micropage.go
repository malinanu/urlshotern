package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

var (
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	reservedSlugs = map[string]bool{
		"api": true, "admin": true, "login": true, "register": true,
		"dashboard": true, "pages": true, "m": true, "micropages": true,
		"features": true, "pricing": true, "analytics": true, "health": true,
		"static": true, "assets": true, "vcard": true, "contact": true,
		"about": true, "terms": true, "privacy": true, "settings": true,
	}
)

type MicroPageService struct {
	db *storage.PostgresStorage
}

func NewMicroPageService(db *storage.PostgresStorage) *MicroPageService {
	return &MicroPageService{db: db}
}

// ValidateSlug checks format and ensures it is not reserved
func (s *MicroPageService) ValidateSlug(slug string) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if len(slug) < 2 || len(slug) > 64 {
		return errors.New("slug must be between 2 and 64 characters")
	}
	if !slugRegex.MatchString(slug) {
		return errors.New("slug can only contain lowercase letters, numbers, and single hyphens")
	}
	if reservedSlugs[slug] {
		return fmt.Errorf("the slug '%s' is a reserved system path", slug)
	}
	return nil
}

// SlugExists checks whether a slug is already taken by another micropage
func (s *MicroPageService) SlugExists(slug string, excludeID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM micropages WHERE slug = $1 AND id != $2`
	var count int
	err := s.db.QueryRow(query, strings.ToLower(slug), excludeID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateMicroPage creates a new digital vCard / micro landing page
func (s *MicroPageService) CreateMicroPage(userID int64, req *models.CreateMicroPageRequest) (*models.MicroPageResponse, error) {
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if err := s.ValidateSlug(slug); err != nil {
		return nil, err
	}

	taken, err := s.SlugExists(slug, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check slug: %w", err)
	}
	if taken {
		return nil, fmt.Errorf("slug '%s' is already in use", slug)
	}

	theme := req.Theme
	if theme == "" {
		theme = "executive-obsidian"
	}

	configJSON, _ := json.Marshal(req.Config)
	if req.Config == nil {
		configJSON = []byte("{}")
	}

	vcardJSON, _ := json.Marshal(req.VCard)
	if req.VCard == nil {
		vcardJSON = []byte("{}")
	}

	socialJSON, _ := json.Marshal(req.SocialLinks)
	if req.SocialLinks == nil {
		socialJSON = []byte("[]")
	}

	linksJSON, _ := json.Marshal(req.Links)
	if req.Links == nil {
		linksJSON = []byte("[]")
	}

	now := time.Now()
	query := `
		INSERT INTO micropages (
			user_id, slug, title, bio, avatar_url, cover_url,
			theme, custom_css, config, vcard_data, social_links, links,
			is_published, view_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, 0, $14, $15
		) RETURNING id, created_at, updated_at`

	var page models.MicroPage
	page.UserID = userID
	page.Slug = slug
	page.Title = strings.TrimSpace(req.Title)
	page.Bio = req.Bio
	page.AvatarURL = req.AvatarURL
	page.CoverURL = req.CoverURL
	page.Theme = theme
	page.CustomCSS = req.CustomCSS
	page.Config = string(configJSON)
	page.VCardData = string(vcardJSON)
	page.SocialLinks = string(socialJSON)
	page.Links = string(linksJSON)
	page.IsPublished = req.IsPublished

	err = s.db.QueryRow(
		query,
		userID, slug, page.Title, page.Bio, page.AvatarURL, page.CoverURL,
		page.Theme, page.CustomCSS, page.Config, page.VCardData, page.SocialLinks, page.Links,
		page.IsPublished, now, now,
	).Scan(&page.ID, &page.CreatedAt, &page.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create micropage: %w", err)
	}

	return s.modelToResponse(&page), nil
}

// GetMicroPageByID retrieves a micro page by ID for its owner
func (s *MicroPageService) GetMicroPageByID(id int64, userID int64) (*models.MicroPageResponse, error) {
	query := `
		SELECT id, user_id, slug, title, bio, avatar_url, cover_url,
		       theme, custom_css, config, vcard_data, social_links, links,
		       is_published, view_count, created_at, updated_at
		FROM micropages
		WHERE id = $1 AND user_id = $2`

	var m models.MicroPage
	err := s.db.QueryRow(query, id, userID).Scan(
		&m.ID, &m.UserID, &m.Slug, &m.Title, &m.Bio, &m.AvatarURL, &m.CoverURL,
		&m.Theme, &m.CustomCSS, &m.Config, &m.VCardData, &m.SocialLinks, &m.Links,
		&m.IsPublished, &m.ViewCount, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("micropage not found")
		}
		return nil, fmt.Errorf("failed to get micropage: %w", err)
	}

	return s.modelToResponse(&m), nil
}

// GetMicroPageBySlug retrieves a published or owner's micro page by slug
func (s *MicroPageService) GetMicroPageBySlug(slug string, incrementView bool) (*models.MicroPageResponse, error) {
	cleanSlug := strings.ToLower(strings.TrimSpace(slug))
	query := `
		SELECT id, user_id, slug, title, bio, avatar_url, cover_url,
		       theme, custom_css, config, vcard_data, social_links, links,
		       is_published, view_count, created_at, updated_at
		FROM micropages
		WHERE slug = $1`

	var m models.MicroPage
	err := s.db.QueryRow(query, cleanSlug).Scan(
		&m.ID, &m.UserID, &m.Slug, &m.Title, &m.Bio, &m.AvatarURL, &m.CoverURL,
		&m.Theme, &m.CustomCSS, &m.Config, &m.VCardData, &m.SocialLinks, &m.Links,
		&m.IsPublished, &m.ViewCount, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("micropage not found")
		}
		return nil, fmt.Errorf("failed to get micropage: %w", err)
	}

	if incrementView {
		go func(pageID int64) {
			_, _ = s.db.Exec(`UPDATE micropages SET view_count = view_count + 1 WHERE id = $1`, pageID)
		}(m.ID)
		m.ViewCount++
	}

	return s.modelToResponse(&m), nil
}

// GetUserMicroPages returns a paginated list of micro pages for a user
func (s *MicroPageService) GetUserMicroPages(userID int64, limit, offset int) ([]*models.MicroPageResponse, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	countQuery := `SELECT COUNT(*) FROM micropages WHERE user_id = $1`
	var total int
	if err := s.db.QueryRow(countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count micropages: %w", err)
	}

	listQuery := `
		SELECT id, user_id, slug, title, bio, avatar_url, cover_url,
		       theme, custom_css, config, vcard_data, social_links, links,
		       is_published, view_count, created_at, updated_at
		FROM micropages
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(listQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list micropages: %w", err)
	}
	defer rows.Close()

	var pages []*models.MicroPageResponse
	for rows.Next() {
		var m models.MicroPage
		err := rows.Scan(
			&m.ID, &m.UserID, &m.Slug, &m.Title, &m.Bio, &m.AvatarURL, &m.CoverURL,
			&m.Theme, &m.CustomCSS, &m.Config, &m.VCardData, &m.SocialLinks, &m.Links,
			&m.IsPublished, &m.ViewCount, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan micropage: %w", err)
		}
		pages = append(pages, s.modelToResponse(&m))
	}

	return pages, total, nil
}

// UpdateMicroPage updates an existing micro page
func (s *MicroPageService) UpdateMicroPage(id int64, userID int64, req *models.UpdateMicroPageRequest) (*models.MicroPageResponse, error) {
	// First fetch existing
	existing, err := s.GetMicroPageByID(id, userID)
	if err != nil {
		return nil, err
	}

	targetSlug := existing.Slug
	if req.Slug != "" && strings.ToLower(req.Slug) != existing.Slug {
		targetSlug = strings.ToLower(strings.TrimSpace(req.Slug))
		if err := s.ValidateSlug(targetSlug); err != nil {
			return nil, err
		}
		taken, err := s.SlugExists(targetSlug, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check slug: %w", err)
		}
		if taken {
			return nil, fmt.Errorf("slug '%s' is already in use", targetSlug)
		}
	}

	title := existing.Title
	if req.Title != "" {
		title = strings.TrimSpace(req.Title)
	}

	bio := existing.Bio
	if req.Bio != nil {
		bio = *req.Bio
	}

	avatarURL := existing.AvatarURL
	if req.AvatarURL != nil {
		avatarURL = *req.AvatarURL
	}

	coverURL := existing.CoverURL
	if req.CoverURL != nil {
		coverURL = *req.CoverURL
	}

	theme := existing.Theme
	if req.Theme != "" {
		theme = req.Theme
	}

	customCSS := existing.CustomCSS
	if req.CustomCSS != nil {
		customCSS = *req.CustomCSS
	}

	configJSON, _ := json.Marshal(req.Config)
	if req.Config == nil {
		configJSON, _ = json.Marshal(existing.Config)
	}

	vcardJSON, _ := json.Marshal(req.VCard)
	if req.VCard == nil {
		vcardJSON, _ = json.Marshal(existing.VCard)
	}

	socialJSON, _ := json.Marshal(req.SocialLinks)
	if req.SocialLinks == nil {
		socialJSON, _ = json.Marshal(existing.SocialLinks)
	}

	linksJSON, _ := json.Marshal(req.Links)
	if req.Links == nil {
		linksJSON, _ = json.Marshal(existing.Links)
	}

	isPublished := existing.IsPublished
	if req.IsPublished != nil {
		isPublished = *req.IsPublished
	}

	now := time.Now()
	query := `
		UPDATE micropages SET
			slug = $1, title = $2, bio = $3, avatar_url = $4, cover_url = $5,
			theme = $6, custom_css = $7, config = $8, vcard_data = $9,
			social_links = $10, links = $11, is_published = $12, updated_at = $13
		WHERE id = $14 AND user_id = $15
		RETURNING updated_at`

	var updatedAt time.Time
	err = s.db.QueryRow(
		query,
		targetSlug, title, bio, avatarURL, coverURL,
		theme, customCSS, string(configJSON), string(vcardJSON),
		string(socialJSON), string(linksJSON), isPublished, now,
		id, userID,
	).Scan(&updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update micropage: %w", err)
	}

	return s.GetMicroPageByID(id, userID)
}

// DeleteMicroPage deletes a micro page owned by the user
func (s *MicroPageService) DeleteMicroPage(id int64, userID int64) error {
	result, err := s.db.Exec(`DELETE FROM micropages WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete micropage: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("micropage not found or not authorized")
	}
	return nil
}

// TrackLinkClick records a click on an individual link button
func (s *MicroPageService) TrackLinkClick(slug string, linkID string, ip, ua, ref, country string) error {
	page, err := s.GetMicroPageBySlug(slug, false)
	if err != nil {
		return err
	}

	// Update the link click count inside page links
	updated := false
	for i := range page.Links {
		if page.Links[i].ID == linkID {
			page.Links[i].ClickCount++
			updated = true
			break
		}
	}

	if updated {
		linksJSON, _ := json.Marshal(page.Links)
		_, _ = s.db.Exec(`UPDATE micropages SET links = $1 WHERE id = $2`, string(linksJSON), page.ID)
	}

	// Insert into micropage_clicks if table exists
	go func() {
		var validIP *string
		cleanIP := strings.TrimSpace(ip)
		if cleanIP != "" {
			if parsed := net.ParseIP(cleanIP); parsed != nil {
				ipStr := parsed.String()
				validIP = &ipStr
			}
		}

		_, _ = s.db.Exec(`
			INSERT INTO micropage_clicks (micropage_id, link_id, clicked_at, ip_address, user_agent, referrer, country_code)
			VALUES ($1, $2, NOW(), $3::inet, $4, $5, $6)
		`, page.ID, linkID, validIP, ua, ref, country)
	}()

	return nil
}

// GenerateVCF generates the vCard content and filename for a slug
func (s *MicroPageService) GenerateVCF(slug string) (string, string, error) {
	page, err := s.GetMicroPageBySlug(slug, false)
	if err != nil {
		return "", "", err
	}

	vcard := page.VCard
	if vcard == nil {
		vcard = &models.VCardContact{
			FirstName: page.Title,
			Note:      page.Bio,
		}
	} else if vcard.FirstName == "" && vcard.LastName == "" {
		vcard.FirstName = page.Title
	}

	vcfContent := GenerateVCard(vcard)
	filename := fmt.Sprintf("%s.vcf", page.Slug)
	return vcfContent, filename, nil
}

// modelToResponse parses internal database JSON strings into typed structs
func (s *MicroPageService) modelToResponse(m *models.MicroPage) *models.MicroPageResponse {
	resp := &models.MicroPageResponse{
		ID:          m.ID,
		UserID:      m.UserID,
		Slug:        m.Slug,
		Title:       m.Title,
		Bio:         m.Bio,
		AvatarURL:   m.AvatarURL,
		CoverURL:    m.CoverURL,
		Theme:       m.Theme,
		CustomCSS:   m.CustomCSS,
		IsPublished: m.IsPublished,
		ViewCount:   m.ViewCount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Config:      make(map[string]any),
		SocialLinks: make([]models.MicroSocialLink, 0),
		Links:       make([]models.MicroLinkItem, 0),
	}

	if m.Config != "" && m.Config != "{}" {
		_ = json.Unmarshal([]byte(m.Config), &resp.Config)
	}

	if m.VCardData != "" && m.VCardData != "{}" {
		var vc models.VCardContact
		if err := json.Unmarshal([]byte(m.VCardData), &vc); err == nil {
			resp.VCard = &vc
		}
	}

	if m.SocialLinks != "" && m.SocialLinks != "[]" {
		_ = json.Unmarshal([]byte(m.SocialLinks), &resp.SocialLinks)
	}

	if m.Links != "" && m.Links != "[]" {
		_ = json.Unmarshal([]byte(m.Links), &resp.Links)
	}

	return resp
}
