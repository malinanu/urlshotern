package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// AdminURLService handles admin operations on URLs
type AdminURLService struct {
	dbManager *storage.DatabaseManager
}

// NewAdminURLService creates a new admin URL service
func NewAdminURLService(dbManager *storage.DatabaseManager) *AdminURLService {
	return &AdminURLService{
		dbManager: dbManager,
	}
}

// ListURLs retrieves a paginated list of URLs with filtering
func (s *AdminURLService) ListURLs(ctx context.Context, filter models.URLFilter, page, limit int) ([]models.URLMapping, int64, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if filter.UserID != nil {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.Status != "" {
		if filter.Status == "active" {
			whereClause += fmt.Sprintf(" AND is_active = TRUE AND (expires_at IS NULL OR expires_at > NOW())")
		} else if filter.Status == "expired" {
			whereClause += fmt.Sprintf(" AND expires_at IS NOT NULL AND expires_at <= NOW()")
		} else if filter.Status == "suspended" {
			whereClause += fmt.Sprintf(" AND is_active = FALSE")
		}
	}

	if filter.Search != "" {
		whereClause += fmt.Sprintf(" AND (short_code ILIKE $%d OR original_url ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	if filter.CreatedFrom != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filter.CreatedFrom)
		argCount++
	}

	if filter.CreatedTo != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filter.CreatedTo)
		argCount++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM url_mappings " + whereClause
	var total int64
	err := s.dbManager.AppQueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count URLs: %w", err)
	}

	// Build ORDER BY
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Query URLs
	query := fmt.Sprintf(`
		SELECT id, short_code, original_url, user_id, created_at, expires_at, click_count, is_active
		FROM url_mappings
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := s.dbManager.AppQueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query URLs: %w", err)
	}
	defer rows.Close()

	urls := []models.URLMapping{}
	for rows.Next() {
		var url models.URLMapping
		err := rows.Scan(
			&url.ID, &url.ShortCode, &url.OriginalURL, &url.UserID,
			&url.CreatedAt, &url.ExpiresAt, &url.ClickCount, &url.IsActive,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, url)
	}

	return urls, total, nil
}

// GetURLByShortCode retrieves a URL by short code
func (s *AdminURLService) GetURLByShortCode(ctx context.Context, shortCode string) (*models.URLMapping, error) {
	query := `
		SELECT id, short_code, original_url, user_id, created_at, expires_at, click_count, is_active
		FROM url_mappings
		WHERE short_code = $1
	`

	var url models.URLMapping
	err := s.dbManager.AppQueryRowContext(ctx, query, shortCode).Scan(
		&url.ID, &url.ShortCode, &url.OriginalURL, &url.UserID,
		&url.CreatedAt, &url.ExpiresAt, &url.ClickCount, &url.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("URL not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}

	return &url, nil
}

// UpdateURLStatus updates the status of a URL
func (s *AdminURLService) UpdateURLStatus(ctx context.Context, shortCode string, req models.UpdateURLStatusRequest) error {
	isActive := req.Status == "active"

	query := `
		UPDATE url_mappings
		SET is_active = $1,
		    updated_at = $2
		WHERE short_code = $3
	`

	result, err := s.dbManager.AdminExecContext(ctx, query, isActive, time.Now(), shortCode)
	if err != nil {
		return fmt.Errorf("failed to update URL status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("URL not found")
	}

	return nil
}

// DeleteURL deletes a URL permanently
func (s *AdminURLService) DeleteURL(ctx context.Context, shortCode string) error {
	query := "DELETE FROM url_mappings WHERE short_code = $1"

	result, err := s.dbManager.AdminExecContext(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("URL not found")
	}

	return nil
}

// TransferURLOwnership transfers URL to a different owner
func (s *AdminURLService) TransferURLOwnership(ctx context.Context, shortCode string, newOwnerID int64) error {
	query := `
		UPDATE url_mappings
		SET user_id = $1,
		    updated_at = $2
		WHERE short_code = $3
	`

	result, err := s.dbManager.AdminExecContext(ctx, query, newOwnerID, time.Now(), shortCode)
	if err != nil {
		return fmt.Errorf("failed to transfer URL ownership: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("URL not found")
	}

	return nil
}

// ExtendExpiration extends the expiration date of a URL
func (s *AdminURLService) ExtendExpiration(ctx context.Context, shortCode string, newExpiry time.Time) error {
	query := `
		UPDATE url_mappings
		SET expires_at = $1,
		    updated_at = $2
		WHERE short_code = $3
	`

	result, err := s.dbManager.AdminExecContext(ctx, query, newExpiry, time.Now(), shortCode)
	if err != nil {
		return fmt.Errorf("failed to extend URL expiration: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("URL not found")
	}

	return nil
}

// GetFlaggedURLs retrieves all flagged URLs
func (s *AdminURLService) GetFlaggedURLs(ctx context.Context, page, limit int) ([]models.URLMapping, int64, error) {
	// This assumes a flagged_urls table exists from migration
	offset := (page - 1) * limit

	countQuery := `
		SELECT COUNT(DISTINCT u.id)
		FROM url_mappings u
		INNER JOIN admin.flagged_urls f ON u.short_code = f.short_code
		WHERE f.resolved_at IS NULL
	`

	var total int64
	err := s.dbManager.AdminQueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count flagged URLs: %w", err)
	}

	query := `
		SELECT DISTINCT u.id, u.short_code, u.original_url, u.user_id, u.created_at, u.expires_at, u.click_count, u.is_active
		FROM url_mappings u
		INNER JOIN admin.flagged_urls f ON u.short_code = f.short_code
		WHERE f.resolved_at IS NULL
		ORDER BY f.flagged_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.dbManager.AdminQueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query flagged URLs: %w", err)
	}
	defer rows.Close()

	urls := []models.URLMapping{}
	for rows.Next() {
		var url models.URLMapping
		err := rows.Scan(
			&url.ID, &url.ShortCode, &url.OriginalURL, &url.UserID,
			&url.CreatedAt, &url.ExpiresAt, &url.ClickCount, &url.IsActive,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, url)
	}

	return urls, total, nil
}

// FlagURL flags a URL as problematic
func (s *AdminURLService) FlagURL(ctx context.Context, shortCode string, req models.FlagURLRequest, flaggedBy int64) error {
	query := `
		INSERT INTO admin.flagged_urls (short_code, reason, category, flagged_by, flagged_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (short_code) DO UPDATE
		SET reason = EXCLUDED.reason,
		    category = EXCLUDED.category,
		    flagged_by = EXCLUDED.flagged_by,
		    flagged_at = EXCLUDED.flagged_at,
		    resolved_at = NULL
	`

	_, err := s.dbManager.AdminExecContext(ctx, query, shortCode, req.Reason, req.Category, flaggedBy, time.Now())
	if err != nil {
		return fmt.Errorf("failed to flag URL: %w", err)
	}

	return nil
}

// UnflagURL removes flag from a URL
func (s *AdminURLService) UnflagURL(ctx context.Context, shortCode string, unflaggedBy int64) error {
	query := `
		UPDATE admin.flagged_urls
		SET resolved_at = $1,
		    resolved_by = $2
		WHERE short_code = $3
	`

	result, err := s.dbManager.AdminExecContext(ctx, query, time.Now(), unflaggedBy, shortCode)
	if err != nil {
		return fmt.Errorf("failed to unflag URL: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("flagged URL not found")
	}

	return nil
}
