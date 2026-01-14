package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// AdminUserService handles admin operations on users
type AdminUserService struct {
	dbManager    *storage.DatabaseManager
	emailService *EmailService
	userService  *UserService
	jwtService   *JWTService
}

// NewAdminUserService creates a new admin user service
func NewAdminUserService(dbManager *storage.DatabaseManager, emailService *EmailService, userService *UserService, jwtService *JWTService) *AdminUserService {
	return &AdminUserService{
		dbManager:    dbManager,
		emailService: emailService,
		userService:  userService,
		jwtService:   jwtService,
	}
}

// ListUsers retrieves a paginated list of users with filtering
func (s *AdminUserService) ListUsers(ctx context.Context, filter models.UserFilter, page, limit int) ([]models.User, int64, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if filter.Email != "" {
		whereClause += fmt.Sprintf(" AND email = $%d", argCount)
		args = append(args, filter.Email)
		argCount++
	}

	if filter.AccountType != "" {
		whereClause += fmt.Sprintf(" AND account_type = $%d", argCount)
		args = append(args, filter.AccountType)
		argCount++
	}

	if filter.Verified != nil {
		whereClause += fmt.Sprintf(" AND email_verified = $%d", argCount)
		args = append(args, *filter.Verified)
		argCount++
	}

	if filter.Search != "" {
		whereClause += fmt.Sprintf(" AND (email ILIKE $%d OR name ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	var total int64
	err := s.dbManager.AdminQueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Build ORDER BY clause
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Query users
	query := fmt.Sprintf(`
		SELECT id, email, name, password_hash, phone, email_verified, phone_verified,
		       account_type, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := s.dbManager.AdminQueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.Phone,
			&user.EmailVerified, &user.PhoneVerified, &user.AccountType,
			&user.IsActive, &user.IsAdmin,
			&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// ListUsersWithStats retrieves a paginated list of users with usage statistics
func (s *AdminUserService) ListUsersWithStats(ctx context.Context, filter models.UserFilter, page, limit int) ([]models.UserWithStats, int64, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if filter.Email != "" {
		whereClause += fmt.Sprintf(" AND u.email = $%d", argCount)
		args = append(args, filter.Email)
		argCount++
	}

	if filter.AccountType != "" {
		whereClause += fmt.Sprintf(" AND u.account_type = $%d", argCount)
		args = append(args, filter.AccountType)
		argCount++
	}

	if filter.Verified != nil {
		whereClause += fmt.Sprintf(" AND u.email_verified = $%d", argCount)
		args = append(args, *filter.Verified)
		argCount++
	}

	if filter.Search != "" {
		whereClause += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.name ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
		argCount++
	}

	// Count total users
	countQuery := "SELECT COUNT(*) FROM users u " + whereClause
	var total int64
	err := s.dbManager.AdminQueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Build ORDER BY clause
	sortBy := "u.created_at"
	if filter.SortBy != "" {
		// Map sort fields to aliased columns
		switch filter.SortBy {
		case "email":
			sortBy = "u.email"
		case "created_at":
			sortBy = "u.created_at"
		case "last_login_at":
			sortBy = "u.last_login_at"
		case "url_count":
			sortBy = "url_count"
		case "total_clicks":
			sortBy = "total_clicks"
		default:
			sortBy = "u.created_at"
		}
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Query users with stats using LEFT JOIN to include users with no URLs
	query := fmt.Sprintf(`
		SELECT
			u.id, u.email, u.name, u.account_type,
			u.email_verified, u.phone_verified, u.is_active, u.is_admin,
			u.created_at, u.updated_at, u.last_login_at,
			COUNT(DISTINCT um.id) as url_count,
			COALESCE(SUM(um.click_count), 0) as total_clicks
		FROM users u
		LEFT JOIN url_mappings um ON u.id = um.user_id AND um.is_active = true
		%s
		GROUP BY u.id, u.email, u.name, u.account_type, u.email_verified, u.phone_verified,
		         u.is_active, u.is_admin, u.created_at, u.updated_at, u.last_login_at
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := s.dbManager.AdminQueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users with stats: %w", err)
	}
	defer rows.Close()

	users := []models.UserWithStats{}
	for rows.Next() {
		var user models.UserWithStats
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.AccountType,
			&user.EmailVerified, &user.PhoneVerified, &user.IsActive, &user.IsAdmin,
			&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
			&user.URLCount, &user.TotalClicks,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user with stats: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// GetUserDetailedStats retrieves comprehensive statistics for a single user
func (s *AdminUserService) GetUserDetailedStats(ctx context.Context, userID int64) (*models.UserDetailedStats, error) {
	// Get user basic information
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user statistics
	statsQuery := `
		SELECT
			COUNT(DISTINCT um.id) as total_urls,
			COALESCE(SUM(um.click_count), 0) as total_clicks,
			CASE
				WHEN COUNT(DISTINCT um.id) > 0
				THEN COALESCE(SUM(um.click_count), 0)::float / COUNT(DISTINCT um.id)
				ELSE 0
			END as avg_clicks_per_url,
			COUNT(DISTINCT CASE WHEN um.is_active = true THEN um.id END) as active_urls,
			COUNT(DISTINCT CASE WHEN um.expires_at < NOW() THEN um.id END) as expired_urls
		FROM url_mappings um
		WHERE um.user_id = $1
	`

	var stats models.UserStats
	err = s.dbManager.AdminQueryRowContext(ctx, statsQuery, userID).Scan(
		&stats.TotalURLs,
		&stats.TotalClicks,
		&stats.AvgClicksPerURL,
		&stats.ActiveURLs,
		&stats.ExpiredURLs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	// Get top 5 URLs by clicks
	topURLsQuery := `
		SELECT id, short_code, original_url, click_count, created_at, expires_at
		FROM url_mappings
		WHERE user_id = $1 AND is_active = true
		ORDER BY click_count DESC
		LIMIT 5
	`

	topURLsRows, err := s.dbManager.AdminQueryContext(ctx, topURLsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get top URLs: %w", err)
	}
	defer topURLsRows.Close()

	topURLs := []models.URLWithClicks{}
	for topURLsRows.Next() {
		var url models.URLWithClicks
		err := topURLsRows.Scan(
			&url.ID,
			&url.ShortCode,
			&url.OriginalURL,
			&url.ClickCount,
			&url.CreatedAt,
			&url.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top URL: %w", err)
		}
		topURLs = append(topURLs, url)
	}

	// Get recent activity (last 10 actions)
	activityQuery := `
		SELECT id, user_id, action, resource_type, resource_id, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`

	activityRows, err := s.dbManager.AdminQueryContext(ctx, activityQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	defer activityRows.Close()

	activities := []models.ActivityLog{}
	for activityRows.Next() {
		var activity models.ActivityLog
		err := activityRows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Action,
			&activity.ResourceType,
			&activity.ResourceID,
			&activity.IPAddress,
			&activity.UserAgent,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		activities = append(activities, activity)
	}

	// Get URLs created over time (last 30 days)
	urlsOverTimeQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM url_mappings
		WHERE user_id = $1 AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	urlsOverTimeRows, err := s.dbManager.AdminQueryContext(ctx, urlsOverTimeQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get URLs over time: %w", err)
	}
	defer urlsOverTimeRows.Close()

	urlsOverTime := []models.TimeSeriesDataPoint{}
	for urlsOverTimeRows.Next() {
		var dataPoint models.TimeSeriesDataPoint
		err := urlsOverTimeRows.Scan(&dataPoint.Date, &dataPoint.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan URL time series: %w", err)
		}
		urlsOverTime = append(urlsOverTime, dataPoint)
	}

	// Get clicks over time (last 30 days) - aggregated from URL click counts
	// Note: This is a simplified version. For accurate time-series data, you'd need a clicks history table
	clicksOverTimeQuery := `
		SELECT DATE(created_at) as date, COALESCE(SUM(click_count), 0) as count
		FROM url_mappings
		WHERE user_id = $1 AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date
	`

	clicksOverTimeRows, err := s.dbManager.AdminQueryContext(ctx, clicksOverTimeQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clicks over time: %w", err)
	}
	defer clicksOverTimeRows.Close()

	clicksOverTime := []models.TimeSeriesDataPoint{}
	for clicksOverTimeRows.Next() {
		var dataPoint models.TimeSeriesDataPoint
		err := clicksOverTimeRows.Scan(&dataPoint.Date, &dataPoint.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan clicks time series: %w", err)
		}
		clicksOverTime = append(clicksOverTime, dataPoint)
	}

	return &models.UserDetailedStats{
		User:           *user,
		Stats:          stats,
		TopURLs:        topURLs,
		RecentActivity: activities,
		URLsOverTime:   urlsOverTime,
		ClicksOverTime: clicksOverTime,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *AdminUserService) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	query := `
		SELECT id, email, name, password_hash, phone, email_verified, phone_verified,
		       account_type, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := s.dbManager.AdminQueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.Phone,
		&user.EmailVerified, &user.PhoneVerified, &user.AccountType,
		&user.IsActive, &user.IsAdmin,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates user information
func (s *AdminUserService) UpdateUser(ctx context.Context, userID int64, req models.AdminUpdateUserRequest) (*models.User, error) {
	// Build dynamic UPDATE query
	setClauses := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argCount))
		args = append(args, *req.Email)
		argCount++
	}
	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argCount))
		args = append(args, *req.Name)
		argCount++
	}
	if req.EmailVerified != nil {
		setClauses = append(setClauses, fmt.Sprintf("email_verified = $%d", argCount))
		args = append(args, *req.EmailVerified)
		argCount++
	}
	if req.PhoneVerified != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone_verified = $%d", argCount))
		args = append(args, *req.PhoneVerified)
		argCount++
	}
	if req.AccountType != nil {
		setClauses = append(setClauses, fmt.Sprintf("account_type = $%d", argCount))
		args = append(args, *req.AccountType)
		argCount++
	}

	if len(setClauses) == 0 {
		return s.GetUserByID(ctx, userID)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())
	argCount++

	args = append(args, userID)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d",
		setClauses[0], argCount)
	for i := 1; i < len(setClauses); i++ {
		query = fmt.Sprintf("%s, %s", query, setClauses[i])
	}

	_, err := s.dbManager.AdminExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetUserByID(ctx, userID)
}

// BanUser bans a user account
func (s *AdminUserService) BanUser(ctx context.Context, userID int64, req models.BanUserRequest, bannedBy int64) error {
	var bannedUntil *time.Time
	if req.Duration != nil && *req.Duration > 0 {
		until := time.Now().AddDate(0, 0, *req.Duration)
		bannedUntil = &until
	}

	query := `
		UPDATE users
		SET is_active = FALSE,
		    banned_at = $1,
		    banned_until = $2,
		    ban_reason = $3,
		    updated_at = $4
		WHERE id = $5
	`

	_, err := s.dbManager.AdminExecContext(ctx, query, time.Now(), bannedUntil, req.Reason, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	return nil
}

// UnbanUser unbans a user account
func (s *AdminUserService) UnbanUser(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET is_active = TRUE,
		    banned_at = NULL,
		    banned_until = NULL,
		    ban_reason = NULL,
		    updated_at = $1
		WHERE id = $2
	`

	_, err := s.dbManager.AdminExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to unban user: %w", err)
	}

	return nil
}

// ForceVerifyEmail forces email verification for a user
func (s *AdminUserService) ForceVerifyEmail(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET email_verified = TRUE,
		    updated_at = $1
		WHERE id = $2
	`

	_, err := s.dbManager.AdminExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

// SendPasswordReset sends a password reset email to a user
func (s *AdminUserService) SendPasswordReset(ctx context.Context, userID int64) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Use existing email service to send password reset
	if s.emailService != nil {
		// This would trigger the actual password reset flow
		return s.emailService.SendPasswordReset(user.Email, "admin-initiated-reset")
	}

	return nil
}

// DeleteUser soft deletes a user account
func (s *AdminUserService) DeleteUser(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET deleted_at = $1,
		    is_active = FALSE,
		    updated_at = $2
		WHERE id = $3
	`

	_, err := s.dbManager.AdminExecContext(ctx, query, time.Now(), time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// UpdateUserLimits updates user rate limits and quotas
func (s *AdminUserService) UpdateUserLimits(ctx context.Context, userID int64, limits models.UpdateUserLimitsRequest) error {
	// This would update a user_limits table or JSON field
	// For now, placeholder implementation
	query := `
		UPDATE users
		SET updated_at = $1
		WHERE id = $2
	`

	_, err := s.dbManager.AdminExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user limits: %w", err)
	}

	return nil
}

// UpdateUserPlan changes a user's subscription plan
func (s *AdminUserService) UpdateUserPlan(ctx context.Context, userID int64, req models.UpdateUserPlanRequest) error {
	query := `
		UPDATE users
		SET subscription_tier = $1,
		    updated_at = $2
		WHERE id = $3
	`

	_, err := s.dbManager.AdminExecContext(ctx, query, req.PlanID, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user plan: %w", err)
	}

	return nil
}

// NotifyUser sends a notification to a user
func (s *AdminUserService) NotifyUser(ctx context.Context, userID int64, req models.NotifyUserRequest) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if s.emailService != nil && req.Type == "email" {
		return s.emailService.SendEmail(user.Email, req.Subject, req.Message, "")
	}

	return nil
}

// GetUserActivity retrieves user activity history
func (s *AdminUserService) GetUserActivity(ctx context.Context, userID int64, days int) ([]models.AuditLog, error) {
	// This would query from audit logs or activity tracking table
	// Placeholder implementation
	return []models.AuditLog{}, nil
}

// GetUserURLs retrieves all URLs created by a user
func (s *AdminUserService) GetUserURLs(ctx context.Context, userID int64, page, limit int) ([]models.URLMapping, int64, error) {
	offset := (page - 1) * limit

	// Count total
	countQuery := "SELECT COUNT(*) FROM url_mappings WHERE user_id = $1"
	var total int64
	err := s.dbManager.AppQueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user URLs: %w", err)
	}

	// Query URLs
	query := `
		SELECT id, short_code, original_url, user_id, created_at, expires_at, click_count, is_active
		FROM url_mappings
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.dbManager.AppQueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user URLs: %w", err)
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

// ExportUsers exports user data (placeholder)
func (s *AdminUserService) ExportUsers(ctx context.Context, filter models.UserFilter) ([]byte, error) {
	// This would generate CSV/JSON export
	// Placeholder implementation
	return []byte{}, nil
}

// CreateImpersonationToken generates an impersonation JWT token for admins to impersonate users
func (s *AdminUserService) CreateImpersonationToken(ctx context.Context, targetUserID, adminID int64) (string, time.Time, error) {
	// Get target user
	targetUser, err := s.GetUserByID(ctx, targetUserID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("target user not found: %w", err)
	}

	// Check if target user is an admin (prevent impersonating other admins)
	if targetUser.IsAdmin {
		return "", time.Time{}, fmt.Errorf("cannot impersonate admin users")
	}

	// Generate session ID for impersonation
	sessionID := fmt.Sprintf("impersonate_%d_%d_%d", adminID, targetUserID, time.Now().Unix())

	// Create impersonation token with 15 minute expiration
	token, expiresAt, err := s.jwtService.CreateImpersonationToken(targetUser, adminID, sessionID, 15*time.Minute)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create impersonation token: %w", err)
	}

	return token, expiresAt, nil
}
