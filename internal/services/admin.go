package services

import (
	"database/sql"
	"fmt"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type AdminService struct {
	db *storage.PostgresStorage
}

func NewAdminService(db *storage.PostgresStorage) *AdminService {
	return &AdminService{db: db}
}

// GetSystemStats returns system-wide statistics
func (s *AdminService) GetSystemStats() (*models.SystemStats, error) {
	stats := &models.SystemStats{
		SystemStatus: "healthy",
		StorageUsed:  "N/A", // Can be enhanced later
	}

	// Get active users count
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true AND deleted_at IS NULL").Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users count: %w", err)
	}

	// Get total users count
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users count: %w", err)
	}

	// Get total URLs count
	err = s.db.QueryRow("SELECT COUNT(*) FROM url_mappings").Scan(&stats.TotalURLs)
	if err != nil {
		return nil, fmt.Errorf("failed to get total URLs count: %w", err)
	}

	// Get total clicks
	err = s.db.QueryRow("SELECT COALESCE(SUM(click_count), 0) FROM url_mappings").Scan(&stats.TotalClicks)
	if err != nil {
		return nil, fmt.Errorf("failed to get total clicks: %w", err)
	}

	// Get today's signups
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURRENT_DATE AND deleted_at IS NULL").Scan(&stats.DailySignups)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily signups: %w", err)
	}

	// Get today's URLs created
	err = s.db.QueryRow("SELECT COUNT(*) FROM url_mappings WHERE DATE(created_at) = CURRENT_DATE").Scan(&stats.DailyURLsCreated)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily URLs created: %w", err)
	}

	return stats, nil
}

// GetUserList returns paginated user list with aggregated data
func (s *AdminService) GetUserList(limit, offset int) ([]*models.AdminUser, error) {
	query := `
		SELECT
			u.id, u.name, u.email, u.is_active, u.email_verified,
			u.account_type, u.created_at, u.last_login_at,
			COUNT(DISTINCT url.id) as urls_count,
			COALESCE(SUM(url.click_count), 0) as total_clicks
		FROM users u
		LEFT JOIN url_mappings url ON url.user_id = u.id
		WHERE u.deleted_at IS NULL
		GROUP BY u.id, u.name, u.email, u.is_active, u.email_verified, u.account_type, u.created_at, u.last_login_at
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.AdminUser
	for rows.Next() {
		user := &models.AdminUser{}
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.IsActive,
			&user.IsEmailVerified,
			&user.AccountType,
			&user.CreatedAt,
			&lastLoginAt,
			&user.URLsCount,
			&user.TotalClicks,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

// GetTotalUsersCount returns total count of users (for pagination)
func (s *AdminService) GetTotalUsersCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total users count: %w", err)
	}
	return count, nil
}

// ToggleUserStatus activates or deactivates a user
func (s *AdminService) ToggleUserStatus(userID int64, isActive bool) error {
	query := "UPDATE users SET is_active = $1 WHERE id = $2 AND deleted_at IS NULL"
	result, err := s.db.Exec(query, isActive, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// DeleteUser soft-deletes a user
func (s *AdminService) DeleteUser(userID int64) error {
	query := "UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL"
	result, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// GetUserByIDWithStats returns user details with aggregated stats
func (s *AdminService) GetUserByIDWithStats(userID int64) (*models.AdminUser, error) {
	query := `
		SELECT
			u.id, u.name, u.email, u.is_active, u.email_verified,
			u.account_type, u.created_at, u.last_login_at,
			COUNT(DISTINCT url.id) as urls_count,
			COALESCE(SUM(url.click_count), 0) as total_clicks
		FROM users u
		LEFT JOIN url_mappings url ON url.user_id = u.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
		GROUP BY u.id, u.name, u.email, u.is_active, u.email_verified, u.account_type, u.created_at, u.last_login_at
	`

	user := &models.AdminUser{}
	var lastLoginAt sql.NullTime

	err := s.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.IsActive,
		&user.IsEmailVerified,
		&user.AccountType,
		&user.CreatedAt,
		&lastLoginAt,
		&user.URLsCount,
		&user.TotalClicks,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// Activity Logging Methods

// LogActivity creates an activity log entry
func (s *AdminService) LogActivity(userID *int64, userName, activityType, description, ipAddress, userAgent string) error {
	query := `
		INSERT INTO activity_logs (user_id, user_name, activity_type, description, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.db.Exec(query, userID, userName, activityType, description, ipAddress, userAgent)
	if err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}
	return nil
}

// GetRecentActivity returns recent activity logs
func (s *AdminService) GetRecentActivity(limit int) ([]*models.RecentActivity, error) {
	query := `
		SELECT id, user_id, user_name, activity_type, description, created_at, ip_address
		FROM activity_logs
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query activity logs: %w", err)
	}
	defer rows.Close()

	var activities []*models.RecentActivity
	for rows.Next() {
		activity := &models.RecentActivity{}
		var userID sql.NullInt64
		var ipAddress sql.NullString

		err := rows.Scan(
			&activity.ID,
			&userID,
			&activity.UserName,
			&activity.ActivityType,
			&activity.Description,
			&activity.CreatedAt,
			&ipAddress,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity row: %w", err)
		}

		if userID.Valid {
			activity.UserID = &userID.Int64
		}
		if ipAddress.Valid {
			activity.IPAddress = &ipAddress.String
		}

		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activity rows: %w", err)
	}

	return activities, nil
}

// CMS Management Methods

// GetCMSPages returns list of CMS pages
func (s *AdminService) GetCMSPages(limit, offset int) ([]*models.CMSPage, error) {
	query := `
		SELECT id, slug, title, is_published, author_id, view_count, created_at, updated_at
		FROM static_pages
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query CMS pages: %w", err)
	}
	defer rows.Close()

	var pages []*models.CMSPage
	for rows.Next() {
		page := &models.CMSPage{}
		var authorID sql.NullInt64

		err := rows.Scan(
			&page.ID,
			&page.Slug,
			&page.Title,
			&page.IsPublished,
			&authorID,
			&page.ViewCount,
			&page.CreatedAt,
			&page.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan page row: %w", err)
		}

		if authorID.Valid {
			page.AuthorID = &authorID.Int64
		}

		pages = append(pages, page)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating page rows: %w", err)
	}

	return pages, nil
}

// GetCMSBlocks returns list of content blocks
func (s *AdminService) GetCMSBlocks(limit, offset int) ([]*models.CMSBlock, error) {
	query := `
		SELECT id, name, identifier, type, content, is_active, sort_order, created_at, updated_at
		FROM content_blocks
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query CMS blocks: %w", err)
	}
	defer rows.Close()

	var blocks []*models.CMSBlock
	for rows.Next() {
		block := &models.CMSBlock{}

		err := rows.Scan(
			&block.ID,
			&block.Name,
			&block.Identifier,
			&block.Type,
			&block.Content,
			&block.IsActive,
			&block.SortOrder,
			&block.CreatedAt,
			&block.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan block row: %w", err)
		}

		blocks = append(blocks, block)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating block rows: %w", err)
	}

	return blocks, nil
}

// GetCMSTemplates returns list of page templates
func (s *AdminService) GetCMSTemplates(limit, offset int) ([]*models.CMSTemplate, error) {
	query := `
		SELECT id, name, slug, description, is_active, created_at, updated_at
		FROM page_templates
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query CMS templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.CMSTemplate
	for rows.Next() {
		template := &models.CMSTemplate{}
		var description sql.NullString

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Slug,
			&description,
			&template.IsActive,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template row: %w", err)
		}

		if description.Valid {
			template.Description = &description.String
		}

		templates = append(templates, template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating template rows: %w", err)
	}

	return templates, nil
}

// GetTotalCMSPagesCount returns total count of CMS pages
func (s *AdminService) GetTotalCMSPagesCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM static_pages").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total pages count: %w", err)
	}
	return count, nil
}

// GetTotalCMSBlocksCount returns total count of content blocks
func (s *AdminService) GetTotalCMSBlocksCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM content_blocks").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total blocks count: %w", err)
	}
	return count, nil
}

// GetTotalCMSTemplatesCount returns total count of page templates
func (s *AdminService) GetTotalCMSTemplatesCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM page_templates").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total templates count: %w", err)
	}
	return count, nil
}

// SEO Tools Methods

// GetSEOSettings returns SEO-related global settings
func (s *AdminService) GetSEOSettings() (map[string]interface{}, error) {
	query := `
		SELECT key, value, display_name, description
		FROM global_settings
		WHERE category = 'seo' OR key LIKE 'meta_%' OR key LIKE 'google_%'
		ORDER BY sort_order ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query SEO settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]interface{})
	var settingsList []map[string]interface{}

	for rows.Next() {
		var key, value, displayName string
		var description sql.NullString

		err := rows.Scan(&key, &value, &displayName, &description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SEO setting row: %w", err)
		}

		setting := map[string]interface{}{
			"key":          key,
			"value":        value,
			"display_name": displayName,
		}

		if description.Valid {
			setting["description"] = description.String
		}

		settingsList = append(settingsList, setting)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating SEO setting rows: %w", err)
	}

	settings["settings"] = settingsList
	settings["total"] = len(settingsList)

	return settings, nil
}

// GetSEOMetrics returns SEO-related metrics
func (s *AdminService) GetSEOMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Total published pages
	var publishedPages int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM static_pages WHERE is_published = true").Scan(&publishedPages)
	if err != nil {
		return nil, fmt.Errorf("failed to get published pages count: %w", err)
	}

	// Total page views
	var totalPageViews int64
	err = s.db.QueryRow("SELECT COALESCE(SUM(view_count), 0) FROM static_pages").Scan(&totalPageViews)
	if err != nil {
		return nil, fmt.Errorf("failed to get total page views: %w", err)
	}

	// Pages missing meta descriptions
	var pagesMissingMeta int64
	err = s.db.QueryRow("SELECT COUNT(*) FROM static_pages WHERE is_published = true AND (meta_description IS NULL OR meta_description = '')").Scan(&pagesMissingMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages missing meta: %w", err)
	}

	// Average page views
	var avgPageViews float64
	if publishedPages > 0 {
		avgPageViews = float64(totalPageViews) / float64(publishedPages)
	}

	metrics["published_pages"] = publishedPages
	metrics["total_page_views"] = totalPageViews
	metrics["pages_missing_meta"] = pagesMissingMeta
	metrics["avg_page_views"] = avgPageViews
	metrics["seo_health_score"] = calculateSEOHealth(publishedPages, pagesMissingMeta)

	return metrics, nil
}

// calculateSEOHealth calculates SEO health score (0-100)
func calculateSEOHealth(totalPages, pagesMissingMeta int64) int {
	if totalPages == 0 {
		return 100 // No pages, no problems
	}

	pagesWithMeta := totalPages - pagesMissingMeta
	healthScore := (pagesWithMeta * 100) / totalPages

	return int(healthScore)
}

// GetSEORecommendations returns SEO improvement recommendations
func (s *AdminService) GetSEORecommendations() ([]map[string]interface{}, error) {
	recommendations := []map[string]interface{}{}

	// Check for pages missing meta descriptions
	var pagesMissingMeta int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM static_pages WHERE is_published = true AND (meta_description IS NULL OR meta_description = '')").Scan(&pagesMissingMeta)
	if err == nil && pagesMissingMeta > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "warning",
			"title":       "Missing Meta Descriptions",
			"description": fmt.Sprintf("%d published pages are missing meta descriptions", pagesMissingMeta),
			"action":      "Add meta descriptions to improve SEO",
			"priority":    "high",
		})
	}

	// Check for pages missing meta keywords
	var pagesMissingKeywords int64
	err = s.db.QueryRow("SELECT COUNT(*) FROM static_pages WHERE is_published = true AND (meta_keywords IS NULL OR meta_keywords = '')").Scan(&pagesMissingKeywords)
	if err == nil && pagesMissingKeywords > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "info",
			"title":       "Missing Meta Keywords",
			"description": fmt.Sprintf("%d published pages are missing meta keywords", pagesMissingKeywords),
			"action":      "Add relevant keywords to pages",
			"priority":    "medium",
		})
	}

	// Check for duplicate slugs (should never happen, but good to check)
	var duplicateSlugs int64
	err = s.db.QueryRow("SELECT COUNT(*) FROM (SELECT slug FROM static_pages GROUP BY slug HAVING COUNT(*) > 1) AS duplicates").Scan(&duplicateSlugs)
	if err == nil && duplicateSlugs > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "error",
			"title":       "Duplicate Page Slugs",
			"description": fmt.Sprintf("%d duplicate slugs found", duplicateSlugs),
			"action":      "Fix duplicate slugs immediately",
			"priority":    "critical",
		})
	}

	// If no issues found
	if len(recommendations) == 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "success",
			"title":       "SEO Looks Good!",
			"description": "No critical SEO issues detected",
			"action":      "Keep monitoring your SEO health",
			"priority":    "info",
		})
	}

	return recommendations, nil
}

// GetSEOAnalyses returns list of SEO analysis results with pagination
func (s *AdminService) GetSEOAnalyses(limit, offset int) ([]*models.SEOAnalysisDB, error) {
	query := `
		SELECT id, page_id, url, title, meta_description, h1_tags, h2_tags,
		       image_alt_count, word_count, internal_links, external_links,
		       mobile_friendly, loading_speed, ssl_enabled, schema_markup,
		       score, issues, recommendations, created_by, created_at
		FROM seo_analyses
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query SEO analyses: %w", err)
	}
	defer rows.Close()

	var analyses []*models.SEOAnalysisDB
	for rows.Next() {
		analysis := &models.SEOAnalysisDB{}
		var pageID sql.NullInt64

		err := rows.Scan(
			&analysis.ID,
			&pageID,
			&analysis.URL,
			&analysis.Title,
			&analysis.MetaDescription,
			&analysis.H1Tags,
			&analysis.H2Tags,
			&analysis.ImageAltCount,
			&analysis.WordCount,
			&analysis.InternalLinks,
			&analysis.ExternalLinks,
			&analysis.MobileFriendly,
			&analysis.LoadingSpeed,
			&analysis.SSLEnabled,
			&analysis.SchemaMarkup,
			&analysis.Score,
			&analysis.Issues,
			&analysis.Recommendations,
			&analysis.CreatedBy,
			&analysis.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SEO analysis row: %w", err)
		}

		if pageID.Valid {
			analysis.PageID = &pageID.Int64
		}

		analyses = append(analyses, analysis)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating SEO analysis rows: %w", err)
	}

	return analyses, nil
}

// GetTotalSEOAnalysesCount returns total count of SEO analyses
func (s *AdminService) GetTotalSEOAnalysesCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM seo_analyses").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total SEO analyses count: %w", err)
	}
	return count, nil
}

// AnalyzeURL performs a basic SEO analysis on a URL (placeholder implementation)
func (s *AdminService) AnalyzeURL(url string, userID int64) (*models.SEOAnalysisDB, error) {
	// Placeholder implementation - creates a basic analysis record
	// Future enhancement: Fetch URL, parse HTML, analyze meta tags, etc.

	query := `
		INSERT INTO seo_analyses (
			url, title, meta_description, h1_tags, h2_tags,
			image_alt_count, word_count, internal_links, external_links,
			mobile_friendly, loading_speed, ssl_enabled, schema_markup,
			score, issues, recommendations, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING id, created_at
	`

	analysis := &models.SEOAnalysisDB{
		URL:             url,
		Title:           "Pending Analysis",
		MetaDescription: "",
		H1Tags:          "[]",
		H2Tags:          "[]",
		ImageAltCount:   0,
		WordCount:       0,
		InternalLinks:   0,
		ExternalLinks:   0,
		MobileFriendly:  false,
		LoadingSpeed:    0.0,
		SSLEnabled:      false,
		SchemaMarkup:    "{}",
		Score:           0,
		Issues:          "[]",
		Recommendations: `["Full SEO analysis coming soon"]`,
		CreatedBy:       userID,
	}

	err := s.db.QueryRow(
		query,
		analysis.URL,
		analysis.Title,
		analysis.MetaDescription,
		analysis.H1Tags,
		analysis.H2Tags,
		analysis.ImageAltCount,
		analysis.WordCount,
		analysis.InternalLinks,
		analysis.ExternalLinks,
		analysis.MobileFriendly,
		analysis.LoadingSpeed,
		analysis.SSLEnabled,
		analysis.SchemaMarkup,
		analysis.Score,
		analysis.Issues,
		analysis.Recommendations,
		analysis.CreatedBy,
	).Scan(&analysis.ID, &analysis.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create SEO analysis: %w", err)
	}

	return analysis, nil
}

// GetSitemapEntries returns sitemap entries from published pages and active URLs
func (s *AdminService) GetSitemapEntries() ([]*models.SitemapEntry, error) {
	var entries []*models.SitemapEntry

	// Get published static pages
	pagesQuery := `
		SELECT slug, updated_at
		FROM static_pages
		WHERE is_published = true
		ORDER BY updated_at DESC
	`

	rows, err := s.db.Query(pagesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query static pages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var slug string
		var updatedAt sql.NullTime

		if err := rows.Scan(&slug, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan page row: %w", err)
		}

		lastMod := ""
		if updatedAt.Valid {
			lastMod = updatedAt.Time.Format("2006-01-02")
		}

		entries = append(entries, &models.SitemapEntry{
			URL:        "/" + slug,
			LastMod:    lastMod,
			ChangeFreq: "weekly",
			Priority:   0.8,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating page rows: %w", err)
	}

	return entries, nil
}

// GenerateSitemap generates an XML sitemap
func (s *AdminService) GenerateSitemap() (string, error) {
	entries, err := s.GetSitemapEntries()
	if err != nil {
		return "", fmt.Errorf("failed to get sitemap entries: %w", err)
	}

	// Build XML sitemap
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`

	for _, entry := range entries {
		xml += fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>%s</changefreq>
    <priority>%.1f</priority>
  </url>
`, entry.URL, entry.LastMod, entry.ChangeFreq, entry.Priority)
	}

	xml += "</urlset>"

	return xml, nil
}

// GetMetaTags returns list of meta tags with pagination
func (s *AdminService) GetMetaTags(limit, offset int) ([]*models.MetaTag, error) {
	query := `
		SELECT id, page_id, name, content, property, created_by, created_at, updated_by, updated_at
		FROM meta_tags
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query meta tags: %w", err)
	}
	defer rows.Close()

	var metaTags []*models.MetaTag
	for rows.Next() {
		tag := &models.MetaTag{}
		var pageID sql.NullInt64
		var property sql.NullString
		var updatedBy sql.NullInt64

		err := rows.Scan(
			&tag.ID,
			&pageID,
			&tag.Name,
			&tag.Content,
			&property,
			&tag.CreatedBy,
			&tag.CreatedAt,
			&updatedBy,
			&tag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meta tag row: %w", err)
		}

		if pageID.Valid {
			tag.PageID = &pageID.Int64
		}
		if property.Valid {
			propertyStr := property.String
			tag.Property = &propertyStr
		}
		if updatedBy.Valid {
			tag.UpdatedBy = &updatedBy.Int64
		}

		metaTags = append(metaTags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating meta tag rows: %w", err)
	}

	return metaTags, nil
}

// GetTotalMetaTagsCount returns total count of meta tags
func (s *AdminService) GetTotalMetaTagsCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM meta_tags").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total meta tags count: %w", err)
	}
	return count, nil
}

// CreateMetaTag creates a new meta tag
func (s *AdminService) CreateMetaTag(req *models.CreateMetaTagRequest, userID int64) (*models.MetaTag, error) {
	// Check for duplicates (page_id + name + property combination)
	var existingID int64
	checkQuery := `
		SELECT id FROM meta_tags
		WHERE (page_id = $1 OR ($1 IS NULL AND page_id IS NULL))
		  AND name = $2
		  AND (property = $3 OR ($3 IS NULL AND property IS NULL))
	`
	err := s.db.QueryRow(checkQuery, req.PageID, req.Name, req.Property).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err == nil {
			return nil, fmt.Errorf("meta tag already exists with this combination")
		}
		return nil, fmt.Errorf("failed to check for duplicate meta tag: %w", err)
	}

	// Insert new meta tag
	query := `
		INSERT INTO meta_tags (page_id, name, content, property, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	tag := &models.MetaTag{
		PageID:    req.PageID,
		Name:      req.Name,
		Content:   req.Content,
		Property:  req.Property,
		CreatedBy: userID,
	}

	err = s.db.QueryRow(
		query,
		tag.PageID,
		tag.Name,
		tag.Content,
		tag.Property,
		tag.CreatedBy,
	).Scan(&tag.ID, &tag.CreatedAt, &tag.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create meta tag: %w", err)
	}

	return tag, nil
}

// DeleteMetaTag deletes a meta tag by ID
func (s *AdminService) DeleteMetaTag(id int64) error {
	query := "DELETE FROM meta_tags WHERE id = $1"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete meta tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("meta tag not found")
	}

	return nil
}
