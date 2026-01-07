package services

import (
	"context"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// DomainBlocklistService handles domain blocking operations
type DomainBlocklistService struct {
	dbManager *storage.DatabaseManager
}

// NewDomainBlocklistService creates a new domain blocklist service
func NewDomainBlocklistService(dbManager *storage.DatabaseManager) *DomainBlocklistService {
	return &DomainBlocklistService{
		dbManager: dbManager,
	}
}

// GetBlockedDomains retrieves all blocked domains with pagination
func (s *DomainBlocklistService) GetBlockedDomains(ctx context.Context, page, limit int) ([]models.BlockedDomain, int64, error) {
	offset := (page - 1) * limit

	// Count total
	countQuery := "SELECT COUNT(*) FROM admin.blocked_domains"
	var total int64
	err := s.dbManager.AdminQueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count blocked domains: %w", err)
	}

	// Query domains
	query := `
		SELECT id, domain, reason, blocked_by, blocked_at
		FROM admin.blocked_domains
		ORDER BY blocked_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.dbManager.AdminQueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query blocked domains: %w", err)
	}
	defer rows.Close()

	domains := []models.BlockedDomain{}
	for rows.Next() {
		var domain models.BlockedDomain
		err := rows.Scan(
			&domain.ID, &domain.Domain, &domain.Reason,
			&domain.BlockedBy, &domain.BlockedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan blocked domain: %w", err)
		}
		domains = append(domains, domain)
	}

	return domains, total, nil
}

// BlockDomain adds a domain to the blocklist
func (s *DomainBlocklistService) BlockDomain(ctx context.Context, domain, reason string, blockedBy int64) error {
	query := `
		INSERT INTO admin.blocked_domains (domain, reason, blocked_by, blocked_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (domain) DO UPDATE
		SET reason = EXCLUDED.reason,
		    blocked_by = EXCLUDED.blocked_by,
		    blocked_at = EXCLUDED.blocked_at
	`

	_, err := s.dbManager.AdminExecContext(ctx, query, domain, reason, blockedBy, time.Now())
	if err != nil {
		return fmt.Errorf("failed to block domain: %w", err)
	}

	return nil
}

// UnblockDomain removes a domain from the blocklist
func (s *DomainBlocklistService) UnblockDomain(ctx context.Context, domain string) error {
	query := "DELETE FROM admin.blocked_domains WHERE domain = $1"

	result, err := s.dbManager.AdminExecContext(ctx, query, domain)
	if err != nil {
		return fmt.Errorf("failed to unblock domain: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("domain not found in blocklist")
	}

	return nil
}

// IsDomainBlocked checks if a domain is blocked
func (s *DomainBlocklistService) IsDomainBlocked(ctx context.Context, domain string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM admin.blocked_domains WHERE domain = $1)"

	var exists bool
	err := s.dbManager.AdminQueryRowContext(ctx, query, domain).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check domain: %w", err)
	}

	return exists, nil
}
