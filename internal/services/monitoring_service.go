package services

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// MonitoringService handles system monitoring operations
type MonitoringService struct {
	dbManager    *storage.DatabaseManager
	redisStorage *storage.RedisStorage
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(dbManager *storage.DatabaseManager, redisStorage *storage.RedisStorage) *MonitoringService {
	return &MonitoringService{
		dbManager:    dbManager,
		redisStorage: redisStorage,
	}
}

// GetSystemHealth retrieves overall system health status
func (s *MonitoringService) GetSystemHealth(ctx context.Context) (*models.SystemHealthResponse, error) {
	health := &models.SystemHealthResponse{
		Timestamp: time.Now(),
		Version:   "1.0.0", // Should come from config
		Services:  make(map[string]models.ServiceStatus),
	}

	// Check database health
	dbHealth, err := s.GetDatabaseHealth(ctx)
	if err == nil {
		health.Database = *dbHealth
	}

	// Check Redis health
	redisHealth, err := s.GetRedisHealth(ctx)
	if err == nil {
		health.Redis = *redisHealth
	}

	// Determine overall status
	if health.Database.Status == "healthy" && health.Redis.Status == "healthy" {
		health.Status = "healthy"
	} else if health.Database.Status == "unhealthy" || health.Redis.Status == "unhealthy" {
		health.Status = "unhealthy"
	} else {
		health.Status = "degraded"
	}

	// System uptime (placeholder - should track actual start time)
	health.Uptime = 0

	return health, nil
}

// GetDatabaseHealth checks database connection health
func (s *MonitoringService) GetDatabaseHealth(ctx context.Context) (*models.DatabaseHealthStatus, error) {
	health := &models.DatabaseHealthStatus{}

	start := time.Now()

	// Try to ping database
	err := s.dbManager.AdminDB.PingContext(ctx)
	if err != nil {
		health.Status = "unhealthy"
		health.ResponseTimeMs = time.Since(start).Milliseconds()
		return health, err
	}

	health.Status = "healthy"
	health.ResponseTimeMs = time.Since(start).Milliseconds()

	// Get pool stats
	stats := s.dbManager.AdminDB.Stats()
	health.ConnectionsUsed = stats.InUse
	health.ConnectionsMax = stats.MaxOpenConnections

	return health, nil
}

// GetRedisHealth checks Redis connection health
func (s *MonitoringService) GetRedisHealth(ctx context.Context) (*models.RedisHealthStatus, error) {
	health := &models.RedisHealthStatus{}

	if s.redisStorage == nil {
		health.Status = "unavailable"
		return health, nil
	}

	start := time.Now()

	// Try to ping Redis
	err := s.redisStorage.GetClient().Ping(ctx).Err()
	if err != nil {
		health.Status = "unhealthy"
		health.ResponseTimeMs = time.Since(start).Milliseconds()
		return health, err
	}

	health.Status = "healthy"
	health.ResponseTimeMs = time.Since(start).Milliseconds()

	// Get memory stats
	_, err = s.redisStorage.GetClient().Info(ctx, "memory").Result()
	if err == nil {
		// Parse memory info (simplified)
		health.MemoryUsedMB = 0 // Parse from info string
	}

	// Get key count
	dbSize, err := s.redisStorage.GetClient().DBSize(ctx).Result()
	if err == nil {
		health.KeyCount = dbSize
	}

	return health, nil
}

// GetSystemMetrics retrieves system resource metrics
func (s *MonitoringService) GetSystemMetrics(ctx context.Context) (*models.SystemMetricsResponse, error) {
	metrics := &models.SystemMetricsResponse{}

	// Goroutine count
	metrics.GoroutineCount = runtime.NumGoroutine()

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	metrics.MemoryUsedMB = int64(m.Alloc / 1024 / 1024)
	metrics.MemoryTotalMB = int64(m.Sys / 1024 / 1024)

	// CPU, Disk, and request metrics would require additional packages
	metrics.CPUUsagePercent = 0
	metrics.DiskUsedMB = 0
	metrics.DiskTotalMB = 0
	metrics.RequestsPerSecond = 0
	metrics.AvgResponseTimeMs = 0

	return metrics, nil
}

// ListAlerts retrieves system alerts with filtering
func (s *MonitoringService) ListAlerts(ctx context.Context, filter models.AlertFilter, page, limit int) ([]models.Alert, int64, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if filter.Severity != "" {
		whereClause += fmt.Sprintf(" AND severity = $%d", argCount)
		args = append(args, filter.Severity)
		argCount++
	}

	if filter.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}

	if filter.Service != "" {
		whereClause += fmt.Sprintf(" AND service = $%d", argCount)
		args = append(args, filter.Service)
		argCount++
	}

	if !filter.StartDate.IsZero() {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, filter.StartDate)
		argCount++
	}

	if !filter.EndDate.IsZero() {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, filter.EndDate)
		argCount++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM admin.system_alerts " + whereClause
	var total int64
	err := s.dbManager.AdminQueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
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

	// Query alerts
	query := fmt.Sprintf(`
		SELECT id, severity, service, title, message, status,
		       acknowledged_by, acknowledged_at, resolved_by, resolved_at,
		       resolution_note, created_at
		FROM admin.system_alerts
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := s.dbManager.AdminQueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	alerts := []models.Alert{}
	for rows.Next() {
		var alert models.Alert
		err := rows.Scan(
			&alert.ID, &alert.Severity, &alert.Service, &alert.Title, &alert.Message,
			&alert.Status, &alert.AcknowledgedBy, &alert.AcknowledgedAt,
			&alert.ResolvedBy, &alert.ResolvedAt, &alert.ResolutionNote, &alert.CreatedAt,
		)
		if err != nil {
			continue
		}
		alerts = append(alerts, alert)
	}

	return alerts, total, nil
}

// GetAlert retrieves a specific alert by ID
func (s *MonitoringService) GetAlert(ctx context.Context, alertID int64) (*models.Alert, error) {
	query := `
		SELECT id, severity, service, title, message, status,
		       acknowledged_by, acknowledged_at, resolved_by, resolved_at,
		       resolution_note, created_at
		FROM admin.system_alerts
		WHERE id = $1
	`

	var alert models.Alert
	err := s.dbManager.AdminQueryRowContext(ctx, query, alertID).Scan(
		&alert.ID, &alert.Severity, &alert.Service, &alert.Title, &alert.Message,
		&alert.Status, &alert.AcknowledgedBy, &alert.AcknowledgedAt,
		&alert.ResolvedBy, &alert.ResolvedAt, &alert.ResolutionNote, &alert.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return &alert, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *MonitoringService) AcknowledgeAlert(ctx context.Context, alertID, acknowledgedBy int64, note string) error {
	query := `
		UPDATE admin.system_alerts
		SET status = 'acknowledged',
		    acknowledged_by = $1,
		    acknowledged_at = $2
		WHERE id = $3 AND status = 'active'
	`

	result, err := s.dbManager.AdminExecContext(ctx, query, acknowledgedBy, time.Now(), alertID)
	if err != nil {
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("alert not found or already acknowledged")
	}

	return nil
}

// ResolveAlert marks an alert as resolved
func (s *MonitoringService) ResolveAlert(ctx context.Context, alertID, resolvedBy int64, resolution, note string) error {
	query := `
		UPDATE admin.system_alerts
		SET status = 'resolved',
		    resolved_by = $1,
		    resolved_at = $2,
		    resolution_note = $3
		WHERE id = $4 AND status != 'resolved'
	`

	result, err := s.dbManager.AdminExecContext(ctx, query, resolvedBy, time.Now(), resolution, alertID)
	if err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("alert not found or already resolved")
	}

	return nil
}

// GetErrorLogs retrieves recent error logs (placeholder)
func (s *MonitoringService) GetErrorLogs(ctx context.Context, level string, limit int) ([]map[string]interface{}, error) {
	// This would query from a logs table or external logging system
	// Placeholder implementation
	return []map[string]interface{}{}, nil
}
