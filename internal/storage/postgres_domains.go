package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

// ---- Custom Domain DB Methods ----

// CreateCustomDomain inserts a new custom domain record and returns the created domain.
func (p *PostgresStorage) CreateCustomDomain(domain *models.CustomDomain) error {
	return p.db.QueryRow(`
		INSERT INTO custom_domains
		  (user_id, domain, verification_method, verification_token, ssl_enabled, ssl_status, dns_status, is_active, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at`,
		domain.UserID,
		domain.Domain,
		"dns",
		domain.VerificationKey,
		domain.SSLEnabled,
		domain.SSLStatus,
		domain.DNSStatus,
		domain.IsActive,
		domain.IsVerified,
	).Scan(&domain.ID, &domain.CreatedAt, &domain.UpdatedAt)
}

// GetCustomDomainsByUser returns all domains owned by a user.
// The verification_token is intentionally excluded from the list to avoid
// unnecessary exposure; it is only returned by GetCustomDomainByID (used by
// the dedicated verification-info endpoint).
func (p *PostgresStorage) GetCustomDomainsByUser(userID int64) ([]*models.CustomDomain, error) {
	rows, err := p.db.Query(`
		SELECT id, user_id, domain, is_verified, verified_at,
		       ssl_enabled, ssl_status, dns_status, health_status, is_active, created_at, updated_at
		FROM custom_domains
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user domains: %w", err)
	}
	defer rows.Close()

	var domains []*models.CustomDomain
	for rows.Next() {
		d := &models.CustomDomain{}
		var verifiedAt sql.NullTime
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.Domain,
			&d.IsVerified, &verifiedAt,
			&d.SSLEnabled, &d.SSLStatus, &d.DNSStatus, &d.HealthStatus,
			&d.IsActive, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if verifiedAt.Valid {
			d.VerifiedAt = verifiedAt.Time
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

// GetCustomDomainByID returns a domain by its primary key.
func (p *PostgresStorage) GetCustomDomainByID(id int64) (*models.CustomDomain, error) {
	d := &models.CustomDomain{}
	var verifiedAt sql.NullTime
	err := p.db.QueryRow(`
		SELECT id, user_id, domain, verification_token, is_verified, verified_at,
		       ssl_enabled, ssl_status, dns_status, health_status, is_active, created_at, updated_at
		FROM custom_domains WHERE id = $1`, id).
		Scan(&d.ID, &d.UserID, &d.Domain, &d.VerificationKey,
			&d.IsVerified, &verifiedAt,
			&d.SSLEnabled, &d.SSLStatus, &d.DNSStatus, &d.HealthStatus,
			&d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("domain not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get domain by id: %w", err)
	}
	if verifiedAt.Valid {
		d.VerifiedAt = verifiedAt.Time
	}
	return d, nil
}

// GetCustomDomainByName returns a domain by its hostname.
func (p *PostgresStorage) GetCustomDomainByName(domainName string) (*models.CustomDomain, error) {
	d := &models.CustomDomain{}
	var verifiedAt sql.NullTime
	err := p.db.QueryRow(`
		SELECT id, user_id, domain, verification_token, is_verified, verified_at,
		       ssl_enabled, ssl_status, dns_status, health_status, is_active, created_at, updated_at
		FROM custom_domains WHERE domain = $1`, domainName).
		Scan(&d.ID, &d.UserID, &d.Domain, &d.VerificationKey,
			&d.IsVerified, &verifiedAt,
			&d.SSLEnabled, &d.SSLStatus, &d.DNSStatus, &d.HealthStatus,
			&d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("domain not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get domain by name: %w", err)
	}
	if verifiedAt.Valid {
		d.VerifiedAt = verifiedAt.Time
	}
	return d, nil
}

// UpdateCustomDomain saves updated fields for a domain.
func (p *PostgresStorage) UpdateCustomDomain(d *models.CustomDomain) error {
	_, err := p.db.Exec(`
		UPDATE custom_domains
		SET is_active = $1, ssl_enabled = $2, ssl_status = $3, dns_status = $4,
		    health_status = $5, is_verified = $6, verified_at = $7, updated_at = NOW()
		WHERE id = $8`,
		d.IsActive, d.SSLEnabled, d.SSLStatus, d.DNSStatus,
		d.HealthStatus, d.IsVerified, nullTimeFromTime(d.VerifiedAt),
		d.ID)
	return err
}

// DeleteCustomDomain removes a domain (enforcing ownership).
func (p *PostgresStorage) DeleteCustomDomain(id, userID int64) error {
	res, err := p.db.Exec(
		`DELETE FROM custom_domains WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete domain: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("domain not found or access denied")
	}
	return nil
}

// VerifyCustomDomain marks a domain as verified.
func (p *PostgresStorage) VerifyCustomDomain(id, userID int64, token string) error {
	now := time.Now()
	res, err := p.db.Exec(`
		UPDATE custom_domains
		SET is_verified = TRUE, verified_at = $1, dns_status = 'configured', updated_at = $1
		WHERE id = $2 AND user_id = $3 AND verification_token = $4`,
		now, id, userID, token)
	if err != nil {
		return fmt.Errorf("verify domain: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("domain not found, access denied, or invalid verification token")
	}
	return nil
}

// CountCustomDomainsByUser returns how many domains a user owns.
func (p *PostgresStorage) CountCustomDomainsByUser(userID int64) (int, error) {
	var count int
	err := p.db.QueryRow(
		`SELECT COUNT(*) FROM custom_domains WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

// GetAllDomainsAdmin returns all custom domains with pagination (admin use).
func (p *PostgresStorage) GetAllDomainsAdmin(page, limit int, status string) ([]*models.CustomDomain, int, error) {
	offset := (page - 1) * limit

	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		whereClause = " WHERE health_status = $1"
		args = append(args, status)
		argIdx++
	}

	var total int
	if err := p.db.QueryRow("SELECT COUNT(*) FROM custom_domains"+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count all domains: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := p.db.Query(`
		SELECT id, user_id, domain, is_verified, verified_at,
		       ssl_enabled, ssl_status, dns_status, health_status, is_active, created_at, updated_at
		FROM custom_domains`+whereClause+
		` ORDER BY created_at DESC LIMIT $`+fmt.Sprintf("%d", argIdx)+` OFFSET $`+fmt.Sprintf("%d", argIdx+1),
		args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query all domains: %w", err)
	}
	defer rows.Close()

	var domains []*models.CustomDomain
	for rows.Next() {
		d := &models.CustomDomain{}
		var verifiedAt sql.NullTime
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.Domain,
			&d.IsVerified, &verifiedAt,
			&d.SSLEnabled, &d.SSLStatus, &d.DNSStatus, &d.HealthStatus,
			&d.IsActive, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if verifiedAt.Valid {
			d.VerifiedAt = verifiedAt.Time
		}
		domains = append(domains, d)
	}
	return domains, total, rows.Err()
}

// GetDomainStatsAdmin returns aggregate domain statistics (admin use).
func (p *PostgresStorage) GetDomainStatsAdmin() (*models.DomainStats, error) {
	stats := &models.DomainStats{}
	err := p.db.QueryRow(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE is_active = TRUE),
			COUNT(*) FILTER (WHERE is_verified = TRUE),
			COUNT(*) FILTER (WHERE ssl_enabled = TRUE),
			COUNT(*) FILTER (WHERE health_status = 'pending'),
			COUNT(*) FILTER (WHERE health_status = 'failed')
		FROM custom_domains`).Scan(
		&stats.TotalDomains,
		&stats.ActiveDomains,
		&stats.VerifiedDomains,
		&stats.SSLDomains,
		&stats.PendingDomains,
		&stats.FailedDomains,
	)
	if err != nil {
		return nil, fmt.Errorf("get domain stats: %w", err)
	}
	return stats, nil
}

// ---- Brand Settings DB Methods ----

// GetBrandSettings returns brand settings for a user, creating defaults if absent.
func (p *PostgresStorage) GetBrandSettings(userID int64) (*models.BrandSettings, error) {
	b := &models.BrandSettings{}
	err := p.db.QueryRow(`
		SELECT id, user_id, company_name, logo_url, primary_color, accent_color, custom_css, favicon_url, created_at, updated_at
		FROM brand_settings WHERE user_id = $1`, userID).
		Scan(&b.ID, &b.UserID, &b.CompanyName, &b.LogoURL,
			&b.PrimaryColor, &b.AccentColor, &b.CustomCSS, &b.FaviconURL,
			&b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		// Return defaults — row will be created on first save
		return &models.BrandSettings{
			UserID:       userID,
			PrimaryColor: "#3b82f6",
			AccentColor:  "#10b981",
		}, nil
	}
	return b, err
}

// SaveBrandSettings upserts brand settings for a user.
func (p *PostgresStorage) SaveBrandSettings(b *models.BrandSettings) error {
	return p.db.QueryRow(`
		INSERT INTO brand_settings (user_id, company_name, logo_url, primary_color, accent_color, custom_css, favicon_url, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (user_id) DO UPDATE
		  SET company_name  = EXCLUDED.company_name,
		      logo_url      = EXCLUDED.logo_url,
		      primary_color = EXCLUDED.primary_color,
		      accent_color  = EXCLUDED.accent_color,
		      custom_css    = EXCLUDED.custom_css,
		      favicon_url   = EXCLUDED.favicon_url,
		      updated_at    = NOW()
		RETURNING id, created_at, updated_at`,
		b.UserID, b.CompanyName, b.LogoURL,
		b.PrimaryColor, b.AccentColor, b.CustomCSS, b.FaviconURL,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
}

// nullTimeFromTime converts a time.Time to a nullable SQL value.
func nullTimeFromTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}
