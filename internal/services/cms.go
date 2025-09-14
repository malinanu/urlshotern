package services

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type CMSService struct {
	db *storage.PostgresStorage
}

func NewCMSService(db *storage.PostgresStorage) *CMSService {
	return &CMSService{
		db: db,
	}
}

// Static Pages Methods

func (s *CMSService) CreateStaticPage(req *models.CreateStaticPageRequest, authorID int64) (*models.StaticPage, error) {
	// Check if slug already exists
	exists, err := s.PageSlugExists(req.Slug, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("page with slug '%s' already exists", req.Slug)
	}

	now := time.Now()
	var publishedAt *time.Time
	if req.IsPublished {
		publishedAt = &now
	}

	query := `
		INSERT INTO static_pages (slug, title, content, meta_description, meta_keywords, 
			is_published, sort_order, template, author_id, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	var page models.StaticPage
	err = s.db.QueryRow(query,
		req.Slug, req.Title, req.Content, req.MetaDescription, req.MetaKeywords,
		req.IsPublished, req.SortOrder, req.Template, authorID, publishedAt, now, now,
	).Scan(&page.ID, &page.CreatedAt, &page.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create static page: %w", err)
	}

	// Populate the page with the request data
	page.Slug = req.Slug
	page.Title = req.Title
	page.Content = req.Content
	page.MetaDescription = req.MetaDescription
	page.MetaKeywords = req.MetaKeywords
	page.IsPublished = req.IsPublished
	page.SortOrder = req.SortOrder
	page.Template = req.Template
	page.AuthorID = &authorID
	page.PublishedAt = publishedAt

	// Create initial revision
	if err := s.CreatePageRevision(page.ID, &page, authorID); err != nil {
		// Log error but don't fail the page creation
		fmt.Printf("Warning: failed to create initial revision for page %d: %v\n", page.ID, err)
	}

	return &page, nil
}

func (s *CMSService) GetStaticPageBySlug(slug string) (*models.StaticPage, error) {
	query := `
		SELECT p.id, p.slug, p.title, p.content, p.meta_description, p.meta_keywords,
			p.is_published, p.sort_order, p.template, p.author_id, p.published_at,
			p.created_at, p.updated_at, COALESCE(u.first_name || ' ' || u.last_name, u.email) as author_name
		FROM static_pages p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.slug = $1`

	var page models.StaticPage
	var authorName sql.NullString
	
	err := s.db.QueryRow(query, slug).Scan(
		&page.ID, &page.Slug, &page.Title, &page.Content, &page.MetaDescription,
		&page.MetaKeywords, &page.IsPublished, &page.SortOrder, &page.Template,
		&page.AuthorID, &page.PublishedAt, &page.CreatedAt, &page.UpdatedAt, &authorName,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("page with slug '%s' not found", slug)
		}
		return nil, fmt.Errorf("failed to get static page: %w", err)
	}

	if authorName.Valid {
		page.AuthorName = authorName.String
	}

	return &page, nil
}

func (s *CMSService) GetStaticPageByID(id int64) (*models.StaticPage, error) {
	query := `
		SELECT p.id, p.slug, p.title, p.content, p.meta_description, p.meta_keywords,
			p.is_published, p.sort_order, p.template, p.author_id, p.published_at,
			p.created_at, p.updated_at, COALESCE(u.first_name || ' ' || u.last_name, u.email) as author_name
		FROM static_pages p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.id = $1`

	var page models.StaticPage
	var authorName sql.NullString
	
	err := s.db.QueryRow(query, id).Scan(
		&page.ID, &page.Slug, &page.Title, &page.Content, &page.MetaDescription,
		&page.MetaKeywords, &page.IsPublished, &page.SortOrder, &page.Template,
		&page.AuthorID, &page.PublishedAt, &page.CreatedAt, &page.UpdatedAt, &authorName,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("page with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get static page: %w", err)
	}

	if authorName.Valid {
		page.AuthorName = authorName.String
	}

	return &page, nil
}

func (s *CMSService) UpdateStaticPage(id int64, req *models.UpdateStaticPageRequest, userID int64) (*models.StaticPage, error) {
	// Get current page
	currentPage, err := s.GetStaticPageByID(id)
	if err != nil {
		return nil, err
	}

	// Build update query dynamically
	setParts := []string{"updated_at = $1"}
	args := []interface{}{time.Now()}
	argIndex := 2

	if req.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}
	if req.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, *req.Content)
		argIndex++
	}
	if req.MetaDescription != nil {
		setParts = append(setParts, fmt.Sprintf("meta_description = $%d", argIndex))
		args = append(args, *req.MetaDescription)
		argIndex++
	}
	if req.MetaKeywords != nil {
		setParts = append(setParts, fmt.Sprintf("meta_keywords = $%d", argIndex))
		args = append(args, *req.MetaKeywords)
		argIndex++
	}
	if req.IsPublished != nil {
		setParts = append(setParts, fmt.Sprintf("is_published = $%d", argIndex))
		args = append(args, *req.IsPublished)
		argIndex++
		
		// Update published_at if publishing status changes
		if *req.IsPublished && !currentPage.IsPublished {
			setParts = append(setParts, fmt.Sprintf("published_at = $%d", argIndex))
			args = append(args, time.Now())
			argIndex++
		} else if !*req.IsPublished {
			setParts = append(setParts, fmt.Sprintf("published_at = $%d", argIndex))
			args = append(args, nil)
			argIndex++
		}
	}
	if req.SortOrder != nil {
		setParts = append(setParts, fmt.Sprintf("sort_order = $%d", argIndex))
		args = append(args, *req.SortOrder)
		argIndex++
	}
	if req.Template != nil {
		setParts = append(setParts, fmt.Sprintf("template = $%d", argIndex))
		args = append(args, *req.Template)
		argIndex++
	}

	// Add WHERE clause
	args = append(args, id)
	whereClause := fmt.Sprintf("WHERE id = $%d", argIndex)

	query := fmt.Sprintf("UPDATE static_pages SET %s %s", strings.Join(setParts, ", "), whereClause)

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update static page: %w", err)
	}

	// Get updated page
	updatedPage, err := s.GetStaticPageByID(id)
	if err != nil {
		return nil, err
	}

	// Create revision
	if err := s.CreatePageRevision(id, updatedPage, userID); err != nil {
		// Log error but don't fail the update
		fmt.Printf("Warning: failed to create revision for page %d: %v\n", id, err)
	}

	return updatedPage, nil
}

func (s *CMSService) DeleteStaticPage(id int64) error {
	query := `DELETE FROM static_pages WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete static page: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("page with ID %d not found", id)
	}

	return nil
}

func (s *CMSService) ListStaticPages(query *models.ListPagesQuery) (*models.PaginatedResponse, error) {
	baseQuery := `
		SELECT p.id, p.slug, p.title, p.content, p.meta_description, p.meta_keywords,
			p.is_published, p.sort_order, p.template, p.author_id, p.published_at,
			p.created_at, p.updated_at, COALESCE(u.first_name || ' ' || u.last_name, u.email) as author_name
		FROM static_pages p
		LEFT JOIN users u ON p.author_id = u.id`

	countQuery := `SELECT COUNT(*) FROM static_pages p`

	// Build WHERE conditions
	var conditions []string
	var args []interface{}
	argIndex := 1

	if query.Published != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_published = $%d", argIndex))
		args = append(args, *query.Published)
		argIndex++
	}
	if query.Template != nil {
		conditions = append(conditions, fmt.Sprintf("p.template = $%d", argIndex))
		args = append(args, *query.Template)
		argIndex++
	}
	if query.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("p.author_id = $%d", argIndex))
		args = append(args, *query.AuthorID)
		argIndex++
	}
	if query.Search != nil && *query.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.title ILIKE $%d OR p.content ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*query.Search+"%")
		argIndex++
	}

	// Add WHERE clause if needed
	var whereClause string
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count pages: %w", err)
	}

	// Add ORDER BY
	orderBy := "p.sort_order ASC, p.created_at DESC"
	if query.SortBy != "" {
		direction := "ASC"
		if query.SortOrder == "desc" {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("p.%s %s", query.SortBy, direction)
	}
	baseQuery += " ORDER BY " + orderBy

	// Add LIMIT and OFFSET
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, query.Limit, query.Offset)

	// Execute query
	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list pages: %w", err)
	}
	defer rows.Close()

	var pages []models.StaticPage
	for rows.Next() {
		var page models.StaticPage
		var authorName sql.NullString

		err := rows.Scan(
			&page.ID, &page.Slug, &page.Title, &page.Content, &page.MetaDescription,
			&page.MetaKeywords, &page.IsPublished, &page.SortOrder, &page.Template,
			&page.AuthorID, &page.PublishedAt, &page.CreatedAt, &page.UpdatedAt, &authorName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan page: %w", err)
		}

		if authorName.Valid {
			page.AuthorName = authorName.String
		}

		pages = append(pages, page)
	}

	return &models.PaginatedResponse{
		Data:    pages,
		Total:   total,
		Limit:   query.Limit,
		Offset:  query.Offset,
		HasMore: query.Offset+query.Limit < int(total),
	}, nil
}

func (s *CMSService) PageSlugExists(slug string, excludeID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM static_pages WHERE slug = $1 AND id != $2)`
	var exists bool
	err := s.db.QueryRow(query, slug, excludeID).Scan(&exists)
	return exists, err
}

// Page Revision Methods

func (s *CMSService) CreatePageRevision(pageID int64, page *models.StaticPage, authorID int64) error {
	// Get current revision number
	var revisionNumber int
	query := `SELECT COALESCE(MAX(revision_number), 0) + 1 FROM page_revisions WHERE page_id = $1`
	err := s.db.QueryRow(query, pageID).Scan(&revisionNumber)
	if err != nil {
		return fmt.Errorf("failed to get revision number: %w", err)
	}

	// Create revision
	insertQuery := `
		INSERT INTO page_revisions (page_id, title, content, meta_description, meta_keywords, 
			revision_number, author_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = s.db.Exec(insertQuery,
		pageID, page.Title, page.Content, page.MetaDescription, page.MetaKeywords,
		revisionNumber, authorID, time.Now(),
	)
	
	return err
}

func (s *CMSService) GetPageRevisions(pageID int64, limit int) ([]models.PageRevision, error) {
	query := `
		SELECT r.id, r.page_id, r.title, r.content, r.meta_description, r.meta_keywords,
			r.revision_number, r.author_id, r.created_at, 
			COALESCE(u.first_name || ' ' || u.last_name, u.email) as author_name
		FROM page_revisions r
		LEFT JOIN users u ON r.author_id = u.id
		WHERE r.page_id = $1
		ORDER BY r.revision_number DESC
		LIMIT $2`

	rows, err := s.db.Query(query, pageID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get page revisions: %w", err)
	}
	defer rows.Close()

	var revisions []models.PageRevision
	for rows.Next() {
		var revision models.PageRevision
		var authorName sql.NullString

		err := rows.Scan(
			&revision.ID, &revision.PageID, &revision.Title, &revision.Content,
			&revision.MetaDescription, &revision.MetaKeywords, &revision.RevisionNumber,
			&revision.AuthorID, &revision.CreatedAt, &authorName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan revision: %w", err)
		}

		if authorName.Valid {
			revision.AuthorName = authorName.String
		}

		revisions = append(revisions, revision)
	}

	return revisions, nil
}

// Page Analytics Methods

func (s *CMSService) RecordPageVisit(pageID int64, ipAddress, userAgent, referrer string) error {
	query := `
		INSERT INTO page_visits (page_id, ip_address, user_agent, referrer, visited_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := s.db.Exec(query, pageID, ipAddress, userAgent, referrer, time.Now())
	return err
}

func (s *CMSService) GetPageAnalytics(pageID int64) (*models.PageAnalytics, error) {
	query := `
		SELECT 
			COUNT(*) as total_visits,
			COUNT(CASE WHEN visited_at >= CURRENT_DATE THEN 1 END) as today_visits,
			COUNT(CASE WHEN visited_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as week_visits,
			COUNT(CASE WHEN visited_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as month_visits,
			COUNT(DISTINCT ip_address) as unique_visitors
		FROM page_visits 
		WHERE page_id = $1`

	var analytics models.PageAnalytics
	err := s.db.QueryRow(query, pageID).Scan(
		&analytics.TotalVisits, &analytics.TodayVisits, &analytics.WeekVisits,
		&analytics.MonthVisits, &analytics.UniqueVisitors,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get page analytics: %w", err)
	}

	analytics.PageID = pageID
	// Calculate bounce rate (simplified - you might want a more sophisticated calculation)
	if analytics.TotalVisits > 0 {
		analytics.BounceRate = 0.3 // Placeholder - implement proper calculation
	}

	return &analytics, nil
}

// Helper function to generate API key
func generateAPIKey() (string, string, error) {
	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// Create key with prefix
	plainKey := "us_" + hex.EncodeToString(bytes)
	
	// Create hash for storage
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])
	
	return plainKey, keyHash, nil
}