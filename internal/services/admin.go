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
		SELECT id, name, identifier, type, is_active, created_at, updated_at
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
			&block.IsActive,
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
