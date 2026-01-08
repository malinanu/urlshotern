package services

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/services/email"
)

// MonitoringService handles system monitoring operations
type MonitoringService struct {
	dbManager      *storage.DatabaseManager
	redisStorage   *storage.RedisStorage
	polarService   *PolarService
	emailProvider  email.Provider
	smsService     *SMSService
	healthCache    map[string]cachedHealth
	healthCacheMux sync.RWMutex
}

// cachedHealth stores cached health check results
type cachedHealth struct {
	status    models.ServiceStatus
	timestamp time.Time
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(dbManager *storage.DatabaseManager, redisStorage *storage.RedisStorage) *MonitoringService {
	return &MonitoringService{
		dbManager:    dbManager,
		redisStorage: redisStorage,
		healthCache:  make(map[string]cachedHealth),
	}
}

// SetPolarService sets the Polar service for health checks
func (s *MonitoringService) SetPolarService(polarService *PolarService) {
	s.polarService = polarService
}

// SetEmailProvider sets the email provider for health checks
func (s *MonitoringService) SetEmailProvider(emailProvider email.Provider) {
	s.emailProvider = emailProvider
}

// SetSMSService sets the SMS service for health checks
func (s *MonitoringService) SetSMSService(smsService *SMSService) {
	s.smsService = smsService
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

// GetServicesHealth retrieves health status of all external services
func (s *MonitoringService) GetServicesHealth(ctx context.Context) (map[string]models.ServiceStatus, error) {
	services := make(map[string]models.ServiceStatus)

	// Check database
	dbStatus := s.checkServiceHealth(ctx, "database", func(ctx context.Context) error {
		dbHealth, err := s.GetDatabaseHealth(ctx)
		if err != nil || dbHealth.Status != "healthy" {
			return fmt.Errorf("database unhealthy")
		}
		return nil
	})
	services["database"] = dbStatus

	// Check Redis
	redisStatus := s.checkServiceHealth(ctx, "redis", func(ctx context.Context) error {
		redisHealth, err := s.GetRedisHealth(ctx)
		if err != nil || redisHealth.Status != "healthy" {
			return fmt.Errorf("redis unhealthy")
		}
		return nil
	})
	services["redis"] = redisStatus

	// Check Polar API
	if s.polarService != nil {
		polarStatus := s.checkServiceHealth(ctx, "polar_api", func(ctx context.Context) error {
			return s.polarService.HealthCheck(ctx)
		})
		services["polar_api"] = polarStatus
	}

	// Check SMTP
	if s.emailProvider != nil {
		smtpStatus := s.checkServiceHealth(ctx, "smtp", func(ctx context.Context) error {
			// Check if emailProvider supports HealthCheck
			if provider, ok := s.emailProvider.(interface {
				HealthCheck(context.Context) error
			}); ok {
				return provider.HealthCheck(ctx)
			}
			return nil // Consider healthy if no HealthCheck method
		})
		services["smtp"] = smtpStatus
	}

	// Check SMS
	if s.smsService != nil {
		smsStatus := s.checkServiceHealth(ctx, "sms", func(ctx context.Context) error {
			return s.smsService.HealthCheck(ctx)
		})
		services["sms"] = smsStatus
	}

	return services, nil
}

// checkServiceHealth checks a service's health with caching (30s TTL)
func (s *MonitoringService) checkServiceHealth(ctx context.Context, serviceName string, healthCheck func(context.Context) error) models.ServiceStatus {
	// Check cache first
	s.healthCacheMux.RLock()
	if cached, ok := s.healthCache[serviceName]; ok {
		if time.Since(cached.timestamp) < 30*time.Second {
			s.healthCacheMux.RUnlock()
			return cached.status
		}
	}
	s.healthCacheMux.RUnlock()

	// Perform health check with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status := models.ServiceStatus{
		Name:        s.formatServiceName(serviceName),
		LastChecked: time.Now(),
	}

	start := time.Now()
	err := healthCheck(checkCtx)
	status.ResponseTimeMs = time.Since(start).Milliseconds()

	if err != nil {
		status.Status = "unavailable"
		status.Message = err.Error()
	} else {
		status.Status = "available"
		status.Message = "Service is healthy"
	}

	// Cache result
	s.healthCacheMux.Lock()
	s.healthCache[serviceName] = cachedHealth{
		status:    status,
		timestamp: time.Now(),
	}
	s.healthCacheMux.Unlock()

	return status
}

// formatServiceName converts service key to human-readable name
func (s *MonitoringService) formatServiceName(serviceName string) string {
	names := map[string]string{
		"database":  "PostgreSQL",
		"redis":     "Redis Cache",
		"polar_api": "Polar API",
		"smtp":      "SMTP Email",
		"sms":       "TextLK SMS",
	}
	if name, ok := names[serviceName]; ok {
		return name
	}
	return serviceName
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

	// CPU usage (goroutine-based proxy)
	metrics.CPUUsagePercent = s.getCPUUsage()

	// Disk usage
	diskUsed, diskTotal := s.getDiskUsage()
	metrics.DiskUsedMB = diskUsed
	metrics.DiskTotalMB = diskTotal

	// Request metrics (placeholder - would require middleware integration)
	metrics.RequestsPerSecond = 0
	metrics.AvgResponseTimeMs = 0

	return metrics, nil
}

// getCPUUsage estimates CPU usage based on goroutine count
// This is a simplified proxy metric without external dependencies
func (s *MonitoringService) getCPUUsage() float64 {
	goroutines := float64(runtime.NumGoroutine())
	baseline := 50.0 // Typical idle goroutines

	if goroutines < baseline {
		return 0.0
	}

	// Simple heuristic: goroutines above baseline indicate load
	// Cap at 100%
	usage := (goroutines - baseline) / baseline * 100
	return math.Min(100.0, usage)
}

// getDiskUsage retrieves disk usage statistics
func (s *MonitoringService) getDiskUsage() (used int64, total int64) {
	// Platform-specific implementation
	return getDiskUsagePlatform()
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
