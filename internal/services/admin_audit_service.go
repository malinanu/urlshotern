package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
)

// AdminAuditService handles admin audit logging operations
type AdminAuditService struct {
	dbManager *storage.DatabaseManager
	snowflake *utils.SnowflakeGenerator
}

// NewAdminAuditService creates a new admin audit service
func NewAdminAuditService(dbManager *storage.DatabaseManager, snowflake *utils.SnowflakeGenerator) *AdminAuditService {
	return &AdminAuditService{
		dbManager: dbManager,
		snowflake: snowflake,
	}
}

// LogAction logs an admin action to the audit trail
func (s *AdminAuditService) LogAction(ctx context.Context, req models.CreateAuditLogRequest) error {
	// Generate a unique ID for the audit log
	id, err := s.snowflake.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate audit log ID: %w", err)
	}

	// Set defaults
	if req.Status == "" {
		req.Status = models.AuditStatusSuccess
	}
	if req.Severity == "" {
		req.Severity = models.AuditSeverityInfo
	}

	// Convert details to JSONB
	var detailsJSON []byte
	if req.Details != nil {
		detailsJSON, err = json.Marshal(req.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
	} else {
		detailsJSON = []byte("{}")
	}

	// Convert changes to JSONB
	var changesJSON []byte
	if req.Changes != nil {
		changesJSON, err = json.Marshal(req.Changes)
		if err != nil {
			return fmt.Errorf("failed to marshal changes: %w", err)
		}
	}

	// Insert into admin.audit_logs table using admin database connection
	query := `
		INSERT INTO admin.audit_logs (
			id, user_id, username, action, resource_type, resource_id,
			details, changes, ip_address, user_agent, session_id,
			status, error_message, severity, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = s.dbManager.AdminExecContext(ctx, query,
		id,
		req.UserID,
		req.Username,
		req.Action,
		req.ResourceType,
		req.ResourceID,
		detailsJSON,
		changesJSON,
		req.IPAddress,
		req.UserAgent,
		req.SessionID,
		req.Status,
		req.ErrorMessage,
		req.Severity,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	return nil
}

// GetAuditLog retrieves a single audit log by ID
func (s *AdminAuditService) GetAuditLog(ctx context.Context, id int64) (*models.AdminAuditLog, error) {
	query := `
		SELECT id, user_id, username, action, resource_type, resource_id,
		       details, changes, ip_address, user_agent, session_id,
		       status, error_message, severity, created_at
		FROM admin.audit_logs
		WHERE id = $1
	`

	var log models.AdminAuditLog
	var detailsJSON, changesJSON []byte

	err := s.dbManager.AdminQueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.UserID,
		&log.Username,
		&log.Action,
		&log.ResourceType,
		&log.ResourceID,
		&detailsJSON,
		&changesJSON,
		&log.IPAddress,
		&log.UserAgent,
		&log.SessionID,
		&log.Status,
		&log.ErrorMessage,
		&log.Severity,
		&log.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("audit log not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	// Parse JSONB fields
	if len(detailsJSON) > 0 {
		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal details: %w", err)
		}
	}

	if len(changesJSON) > 0 {
		log.Changes = &models.AuditLogChange{}
		if err := json.Unmarshal(changesJSON, log.Changes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
		}
	}

	return &log, nil
}

// ListAuditLogs retrieves audit logs with filtering and pagination
func (s *AdminAuditService) ListAuditLogs(ctx context.Context, filter models.AuditLogFilter) (*models.AuditLogListResponse, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.Username != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("username ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Username+"%")
		argIndex++
	}

	if filter.Action != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, filter.Action)
		argIndex++
	}

	if filter.ResourceType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("resource_type = $%d", argIndex))
		args = append(args, filter.ResourceType)
		argIndex++
	}

	if filter.ResourceID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("resource_id = $%d", argIndex))
		args = append(args, filter.ResourceID)
		argIndex++
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.Severity != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("severity = $%d", argIndex))
		args = append(args, filter.Severity)
		argIndex++
	}

	if filter.IPAddress != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("ip_address = $%d", argIndex))
		args = append(args, filter.IPAddress)
		argIndex++
	}

	if !filter.StartDate.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.StartDate)
		argIndex++
	}

	if !filter.EndDate.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filter.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin.audit_logs %s", whereClause)
	var totalCount int64
	err := s.dbManager.AdminQueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Set pagination defaults
	limit := filter.Limit
	if limit == 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	// Get audit logs
	query := fmt.Sprintf(`
		SELECT id, user_id, username, action, resource_type, resource_id,
		       details, changes, ip_address, user_agent, session_id,
		       status, error_message, severity, created_at
		FROM admin.audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := s.dbManager.AdminQueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	logs := []models.AdminAuditLog{}
	for rows.Next() {
		var log models.AdminAuditLog
		var detailsJSON, changesJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Username,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&detailsJSON,
			&changesJSON,
			&log.IPAddress,
			&log.UserAgent,
			&log.SessionID,
			&log.Status,
			&log.ErrorMessage,
			&log.Severity,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Parse JSONB fields
		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
				return nil, fmt.Errorf("failed to unmarshal details: %w", err)
			}
		}

		if len(changesJSON) > 0 {
			log.Changes = &models.AuditLogChange{}
			if err := json.Unmarshal(changesJSON, log.Changes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
			}
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit logs: %w", err)
	}

	// Calculate pagination
	page := (offset / limit) + 1
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	return &models.AuditLogListResponse{
		Logs:       logs,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	}, nil
}

// GetUserAuditTrail retrieves all audit logs for a specific user
func (s *AdminAuditService) GetUserAuditTrail(ctx context.Context, userID int64, limit, offset int) (*models.AuditLogListResponse, error) {
	filter := models.AuditLogFilter{
		UserID: &userID,
		Limit:  limit,
		Offset: offset,
	}

	return s.ListAuditLogs(ctx, filter)
}

// GetResourceAuditTrail retrieves all audit logs for a specific resource
func (s *AdminAuditService) GetResourceAuditTrail(ctx context.Context, resourceType, resourceID string, limit, offset int) (*models.AuditLogListResponse, error) {
	filter := models.AuditLogFilter{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Limit:        limit,
		Offset:       offset,
	}

	return s.ListAuditLogs(ctx, filter)
}

// GetRecentCriticalActions retrieves recent critical audit logs
func (s *AdminAuditService) GetRecentCriticalActions(ctx context.Context, limit int) ([]models.AdminAuditLog, error) {
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	filter := models.AuditLogFilter{
		Severity: models.AuditSeverityCritical,
		Limit:    limit,
		Offset:   0,
	}

	response, err := s.ListAuditLogs(ctx, filter)
	if err != nil {
		return nil, err
	}

	return response.Logs, nil
}

// CleanupOldLogs deletes audit logs older than the specified retention period
func (s *AdminAuditService) CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays < 30 {
		retentionDays = 30 // Minimum 30 days retention
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	query := `
		DELETE FROM admin.audit_logs
		WHERE created_at < $1 AND severity != $2
	`

	result, err := s.dbManager.AdminExecContext(ctx, query, cutoffDate, models.AuditSeverityCritical)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// GetAuditStatistics returns statistics about audit logs
func (s *AdminAuditService) GetAuditStatistics(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_actions,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_actions,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_actions,
			COUNT(CASE WHEN action LIKE 'user.%' THEN 1 END) as user_actions,
			COUNT(CASE WHEN action LIKE 'url.%' THEN 1 END) as url_actions,
			COUNT(CASE WHEN action LIKE 'subscription.%' THEN 1 END) as billing_actions
		FROM admin.audit_logs
		WHERE created_at BETWEEN $1 AND $2
	`

	var stats struct {
		TotalActions    int64
		UniqueUsers     int64
		FailedActions   int64
		CriticalActions int64
		UserActions     int64
		URLActions      int64
		BillingActions  int64
	}

	err := s.dbManager.AdminQueryRowContext(ctx, query, startDate, endDate).Scan(
		&stats.TotalActions,
		&stats.UniqueUsers,
		&stats.FailedActions,
		&stats.CriticalActions,
		&stats.UserActions,
		&stats.URLActions,
		&stats.BillingActions,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get audit statistics: %w", err)
	}

	return map[string]interface{}{
		"total_actions":    stats.TotalActions,
		"unique_users":     stats.UniqueUsers,
		"failed_actions":   stats.FailedActions,
		"critical_actions": stats.CriticalActions,
		"user_actions":     stats.UserActions,
		"url_actions":      stats.URLActions,
		"billing_actions":  stats.BillingActions,
	}, nil
}

// LogActionHelper is a convenience method for logging actions with individual parameters
// This simplifies calling LogAction from handlers
func (s *AdminAuditService) LogActionHelper(
	ctx context.Context,
	userID int64,
	action string,
	resourceType string,
	resourceID *int64,
	details map[string]interface{},
) error {
	var resourceIDStr string
	if resourceID != nil {
		resourceIDStr = fmt.Sprintf("%d", *resourceID)
	}

	req := models.CreateAuditLogRequest{
		UserID:       &userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceIDStr,
		Details:      details,
		Status:       models.AuditStatusSuccess,
		Severity:     models.AuditSeverityInfo,
	}

	return s.LogAction(ctx, req)
}

// GetRecentActivity retrieves recent activity logs for dashboard activity feed
func (s *AdminAuditService) GetRecentActivity(ctx context.Context, limit int) ([]models.AdminAuditLog, error) {
	if limit == 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	query := `
		SELECT id, user_id, username, action, resource_type, resource_id,
		       details, changes, ip_address, user_agent, session_id,
		       status, error_message, severity, created_at
		FROM admin.audit_logs
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := s.dbManager.AdminQueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent activity: %w", err)
	}
	defer rows.Close()

	logs := []models.AdminAuditLog{}
	for rows.Next() {
		var log models.AdminAuditLog
		var detailsJSON, changesJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Username,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&detailsJSON,
			&changesJSON,
			&log.IPAddress,
			&log.UserAgent,
			&log.SessionID,
			&log.Status,
			&log.ErrorMessage,
			&log.Severity,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity log: %w", err)
		}

		// Parse JSONB fields
		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
				return nil, fmt.Errorf("failed to unmarshal details: %w", err)
			}
		}

		if len(changesJSON) > 0 {
			log.Changes = &models.AuditLogChange{}
			if err := json.Unmarshal(changesJSON, log.Changes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
			}
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activity logs: %w", err)
	}

	return logs, nil
}
