package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(config *configs.Config) (*PostgresStorage, error) {
	// Use SSL mode from configuration
	sslMode := config.DBSSLMode
	if sslMode == "" {
		sslMode = "disable" // Default to disable if not specified
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool with granular settings
	db.SetMaxOpenConns(config.DBMaxConns)
	db.SetMaxIdleConns(config.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(config.DBConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(config.DBConnMaxIdleTime) * time.Second)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{db: db}

	// Initialize tables if they don't exist
	if err := storage.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

// GetPoolStats returns database connection pool statistics
func (p *PostgresStorage) GetPoolStats() map[string]interface{} {
	stats := p.db.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// Database access methods for authentication services

// QueryRow executes a query that returns at most one row
func (p *PostgresStorage) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRow(query, args...)
}

// QueryRowContext executes a query with context that returns at most one row
func (p *PostgresStorage) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows
func (p *PostgresStorage) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(query, args...)
}

// Query executes a query that returns multiple rows
func (p *PostgresStorage) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.Query(query, args...)
}

// Begin starts a database transaction
func (p *PostgresStorage) Begin() (*sql.Tx, error) {
	return p.db.Begin()
}

// GetRawDB returns the underlying *sql.DB for use with sqlx wrappers.
func (p *PostgresStorage) GetRawDB() *sql.DB {
	return p.db
}

// initTables creates the necessary tables if they don't exist
func (p *PostgresStorage) initTables() error {
	// Initialize authentication tables first (required by url_mappings and analytics)
	if err := p.initAuthenticationTables(); err != nil {
		return fmt.Errorf("failed to initialize authentication tables: %w", err)
	}

	// Create core URL shortener tables (depends on users table)
	queries := []string{
		`CREATE TABLE IF NOT EXISTS url_mappings (
			id BIGINT PRIMARY KEY,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			expires_at TIMESTAMP WITH TIME ZONE,
			click_count BIGINT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_by_ip INET,
			user_id BIGINT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_short_code ON url_mappings(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_created_at ON url_mappings(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_active ON url_mappings(is_active)`,
		`ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255)`,
		`CREATE TABLE IF NOT EXISTS click_events (
			id BIGINT PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			clicked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			ip_address INET,
			user_agent TEXT,
			referrer TEXT,
			country_code CHAR(2)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_short_code_time ON click_events(short_code, clicked_at)`,
		// Bot detection columns — added after initial schema
		`ALTER TABLE click_events ADD COLUMN IF NOT EXISTS is_bot    BOOLEAN DEFAULT FALSE`,
		`ALTER TABLE click_events ADD COLUMN IF NOT EXISTS is_unique BOOLEAN DEFAULT FALSE`,
		`ALTER TABLE click_events ADD COLUMN IF NOT EXISTS ip_hash   VARCHAR(64)`,
		`CREATE INDEX IF NOT EXISTS idx_click_events_is_bot ON click_events(is_bot)`,
		// Fix: coupons.id was created without a sequence (BIGINT PRIMARY KEY, no default).
		// GORM's autoIncrement omits id from INSERT, so we need the DB to supply a default.
		`CREATE SEQUENCE IF NOT EXISTS coupons_id_seq`,
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='coupons' AND column_name='id'
				  AND column_default IS NOT NULL
			) THEN
				PERFORM setval('coupons_id_seq', COALESCE((SELECT MAX(id) FROM coupons), 0) + 1, false);
				ALTER TABLE coupons ALTER COLUMN id SET DEFAULT nextval('coupons_id_seq');
			END IF;
		END $$`,
		`CREATE SEQUENCE IF NOT EXISTS coupon_redemptions_id_seq`,
		`DO $$ BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name='coupon_redemptions' AND column_name='id'
				  AND column_default IS NOT NULL
			) THEN
				PERFORM setval('coupon_redemptions_id_seq', COALESCE((SELECT MAX(id) FROM coupon_redemptions), 0) + 1, false);
				ALTER TABLE coupon_redemptions ALTER COLUMN id SET DEFAULT nextval('coupon_redemptions_id_seq');
			END IF;
		END $$`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	// Initialize enhanced analytics tables (depends on url_mappings and users)
	// custom_domains and teams tables are created here
	if err := p.initAdvancedAnalyticsTables(); err != nil {
		return fmt.Errorf("failed to initialize advanced analytics tables: %w", err)
	}

	// Add custom_domain_id and team_id to url_mappings after custom_domains and teams exist
	p6Queries := []string{
		`ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS custom_domain_id INTEGER REFERENCES custom_domains(id) ON DELETE SET NULL`,
		`ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS team_id INTEGER REFERENCES teams(id) ON DELETE SET NULL`,
		`CREATE INDEX IF NOT EXISTS idx_url_mappings_custom_domain ON url_mappings(custom_domain_id, short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_url_mappings_team ON url_mappings(team_id)`,
	}
	for _, q := range p6Queries {
		if _, err := p.db.Exec(q); err != nil {
			return fmt.Errorf("failed to execute p6 migration %s: %w", q, err)
		}
	}

	// Initialize CMS tables (depends on users and media_files)
	if err := p.initCMSTables(); err != nil {
		return fmt.Errorf("failed to initialize CMS tables: %w", err)
	}

	return nil
}

// initCMSTables creates CMS-related tables
func (p *PostgresStorage) initCMSTables() error {
	queries := []string{
		// Media files table
		`CREATE TABLE IF NOT EXISTS media_files (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			filename VARCHAR(255) NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			file_path VARCHAR(500) NOT NULL,
			file_url VARCHAR(500) NOT NULL,
			file_size BIGINT NOT NULL,
			mime_type VARCHAR(100) NOT NULL,
			file_type VARCHAR(50) NOT NULL,
			width INTEGER,
			height INTEGER,
			alt TEXT,
			caption TEXT,
			is_public BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_media_files_user_id ON media_files(user_id)`,

		// Static pages table
		`CREATE TABLE IF NOT EXISTS static_pages (
			id BIGSERIAL PRIMARY KEY,
			slug VARCHAR(100) UNIQUE NOT NULL,
			title VARCHAR(200) NOT NULL,
			content TEXT NOT NULL,
			content_blocks JSONB,
			meta_description TEXT,
			meta_keywords TEXT,
			meta_image VARCHAR(255),
			meta_title VARCHAR(200),
			canonical_url VARCHAR(255),
			schema_markup JSONB,
			is_published BOOLEAN DEFAULT FALSE,
			sort_order INTEGER DEFAULT 0,
			template VARCHAR(50) DEFAULT 'default',
			author_id BIGINT REFERENCES users(id),
			published_at TIMESTAMP WITH TIME ZONE,
			scheduled_at TIMESTAMP WITH TIME ZONE,
			expires_at TIMESTAMP WITH TIME ZONE,
			featured_image_id BIGINT,
			category VARCHAR(50),
			tags JSONB,
			custom_fields JSONB,
			view_count BIGINT DEFAULT 0,
			is_sticky BOOLEAN DEFAULT FALSE,
			allow_comments BOOLEAN DEFAULT TRUE,
			language VARCHAR(10) DEFAULT 'en',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (featured_image_id) REFERENCES media_files(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_static_pages_slug ON static_pages(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_static_pages_published ON static_pages(is_published)`,

		// Global settings table
		`CREATE TABLE IF NOT EXISTS global_settings (
			id BIGSERIAL PRIMARY KEY,
			key VARCHAR(100) UNIQUE NOT NULL,
			value TEXT NOT NULL,
			type VARCHAR(20) DEFAULT 'text',
			category VARCHAR(50) DEFAULT 'general',
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			is_public BOOLEAN DEFAULT FALSE,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_global_settings_key ON global_settings(key)`,

		// Content blocks table
		`CREATE TABLE IF NOT EXISTS content_blocks (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			identifier VARCHAR(100) UNIQUE NOT NULL,
			content TEXT NOT NULL,
			type VARCHAR(50) DEFAULT 'generic',
			is_active BOOLEAN DEFAULT TRUE,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_content_blocks_identifier ON content_blocks(identifier)`,

		// Page templates table
		`CREATE TABLE IF NOT EXISTS page_templates (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			slug VARCHAR(100) UNIQUE NOT NULL,
			description TEXT,
			content TEXT NOT NULL,
			fields JSONB,
			preview_url VARCHAR(255),
			is_active BOOLEAN DEFAULT TRUE,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_page_templates_slug ON page_templates(slug)`,

		// Navigation menus table
		`CREATE TABLE IF NOT EXISTS navigation_menus (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			location VARCHAR(50) NOT NULL,
			items JSONB NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_navigation_menus_location ON navigation_menus(location)`,

		// Page revisions table
		`CREATE TABLE IF NOT EXISTS page_revisions (
			id BIGSERIAL PRIMARY KEY,
			page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
			title VARCHAR(200) NOT NULL,
			content TEXT NOT NULL,
			meta_description TEXT,
			meta_keywords TEXT,
			revision_number INTEGER NOT NULL,
			author_id BIGINT REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_page_revisions_page_id ON page_revisions(page_id)`,

		// Page visits table
		`CREATE TABLE IF NOT EXISTS page_visits (
			id BIGSERIAL PRIMARY KEY,
			page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
			ip_address INET,
			user_agent TEXT,
			referrer TEXT,
			country_code CHAR(2),
			visited_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_page_visits_page_id ON page_visits(page_id)`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

// initAuthenticationTables creates authentication and user management tables
func (p *PostgresStorage) initAuthenticationTables() error {
	queries := []string{
		// Users table - core user information
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			phone VARCHAR(20) UNIQUE,
			password_hash VARCHAR(255),
			email_verified BOOLEAN DEFAULT FALSE,
			phone_verified BOOLEAN DEFAULT FALSE,
			provider VARCHAR(50) DEFAULT 'local', -- local, google, etc.
			provider_id VARCHAR(255),
			avatar_url TEXT,
			account_type VARCHAR(20) DEFAULT 'free', -- free, pro, enterprise
			is_active BOOLEAN DEFAULT TRUE,
			is_admin BOOLEAN DEFAULT FALSE,
			last_login_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone)`,
		`CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id)`,

		// User sessions table
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id UUID PRIMARY KEY,
			user_id BIGINT NOT NULL,
			device_info TEXT,
			ip_address INET,
			user_agent TEXT,
			refresh_token_hash VARCHAR(255),
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user ON user_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON user_sessions(expires_at)`,

		// User preferences table
		`CREATE TABLE IF NOT EXISTS user_preferences (
			user_id BIGINT PRIMARY KEY,
			default_expiration INTEGER, -- days
			analytics_public BOOLEAN DEFAULT FALSE,
			email_notifications BOOLEAN DEFAULT TRUE,
			marketing_emails BOOLEAN DEFAULT FALSE,
			timezone VARCHAR(50) DEFAULT 'UTC',
			theme VARCHAR(20) DEFAULT 'light',
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Email verification table
		`CREATE TABLE IF NOT EXISTS email_verifications (
			id UUID PRIMARY KEY,
			user_id BIGINT NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			verified_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token)`,
		`CREATE INDEX IF NOT EXISTS idx_email_verifications_user ON email_verifications(user_id)`,

		// Phone verification table
		`CREATE TABLE IF NOT EXISTS phone_verifications (
			id UUID PRIMARY KEY,
			user_id BIGINT NOT NULL,
			phone VARCHAR(20) NOT NULL,
			otp_code VARCHAR(6) NOT NULL,
			attempts INTEGER DEFAULT 0,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			verified_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_phone_verifications_user ON phone_verifications(user_id)`,

		// Password reset table
		`CREATE TABLE IF NOT EXISTS password_resets (
			id UUID PRIMARY KEY,
			user_id BIGINT NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			used_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token)`,
		`CREATE INDEX IF NOT EXISTS idx_password_resets_user ON password_resets(user_id)`,

		// Audit log table
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT,
			action VARCHAR(100) NOT NULL,
			resource_type VARCHAR(50),
			resource_id VARCHAR(100),
			details TEXT,
			ip_address INET,
			user_agent TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at)`,

		// API Keys table - for programmatic access
		`CREATE TABLE IF NOT EXISTS api_keys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id BIGINT NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			key_hash VARCHAR(255) NOT NULL UNIQUE,
			key_prefix VARCHAR(100) NOT NULL,
			permissions JSONB DEFAULT '[]'::jsonb,
			rate_limit INTEGER DEFAULT 1000,
			is_active BOOLEAN DEFAULT TRUE,
			last_used_at TIMESTAMP WITH TIME ZONE,
			last_used_ip VARCHAR(45),
			expires_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys(key_prefix)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active)`,

		// API Key Usage table - for tracking and analytics
		`CREATE TABLE IF NOT EXISTS api_key_usage (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			api_key_id UUID NOT NULL,
			endpoint VARCHAR(255) NOT NULL,
			method VARCHAR(10) NOT NULL,
			status_code INTEGER NOT NULL,
			response_time_ms INTEGER NOT NULL,
			ip_address VARCHAR(45),
			user_agent TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_api_key_usage_api_key_id ON api_key_usage(api_key_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_key_usage_created_at ON api_key_usage(created_at)`,

		// COPPA age gate (also in migration 048)
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS date_of_birth   DATE`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS is_minor        BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS age_verified_at TIMESTAMP WITH TIME ZONE`,

		// CAN-SPAM suppression list (also in migration 048)
		`CREATE TABLE IF NOT EXISTS email_suppressions (
			email VARCHAR(320) PRIMARY KEY,
			reason VARCHAR(50) NOT NULL DEFAULT 'unsubscribe',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

// initAdvancedAnalyticsTables creates enhanced analytics tables
func (p *PostgresStorage) initAdvancedAnalyticsTables() error {
	queries := []string{
		// Enhanced geographic analytics table
		`CREATE TABLE IF NOT EXISTS geographic_analytics (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			country_code CHAR(2),
			country_name VARCHAR(100),
			region VARCHAR(100),
			city VARCHAR(100),
			latitude DECIMAL(10,8),
			longitude DECIMAL(11,8),
			clicks INTEGER DEFAULT 1,
			unique_ips INTEGER DEFAULT 1,
			last_click TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(short_code, country_code, region, city),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_geographic_short_code ON geographic_analytics(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_geographic_country ON geographic_analytics(country_code)`,
		`CREATE INDEX IF NOT EXISTS idx_geographic_location ON geographic_analytics(country_code, region, city)`,

		// Enhanced device analytics table
		`CREATE TABLE IF NOT EXISTS device_analytics (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			device_type VARCHAR(50), -- mobile, desktop, tablet
			device_brand VARCHAR(50), -- Apple, Samsung, etc.
			device_model VARCHAR(100),
			os_name VARCHAR(50),
			os_version VARCHAR(50),
			browser_name VARCHAR(50),
			browser_version VARCHAR(50),
			screen_resolution VARCHAR(20),
			user_agent_hash VARCHAR(64), -- SHA-256 hash for grouping
			clicks INTEGER DEFAULT 1,
			unique_ips INTEGER DEFAULT 1,
			last_click TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(short_code, user_agent_hash),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_device_short_code ON device_analytics(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_device_type ON device_analytics(device_type)`,
		`CREATE INDEX IF NOT EXISTS idx_device_brand ON device_analytics(device_brand, device_model)`,
		`CREATE INDEX IF NOT EXISTS idx_device_os ON device_analytics(os_name, os_version)`,
		`CREATE INDEX IF NOT EXISTS idx_device_browser ON device_analytics(browser_name, browser_version)`,

		// Time-based analytics table
		`CREATE TABLE IF NOT EXISTS time_analytics (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			click_date DATE NOT NULL,
			hour_of_day INTEGER CHECK (hour_of_day >= 0 AND hour_of_day <= 23),
			day_of_week INTEGER CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0=Sunday
			day_of_month INTEGER CHECK (day_of_month >= 1 AND day_of_month <= 31),
			month INTEGER CHECK (month >= 1 AND month <= 12),
			year INTEGER,
			clicks INTEGER DEFAULT 1,
			unique_ips INTEGER DEFAULT 1,
			conversions INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE,
			UNIQUE(short_code, click_date, hour_of_day)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_time_short_code ON time_analytics(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_time_date ON time_analytics(click_date)`,
		`CREATE INDEX IF NOT EXISTS idx_time_hour ON time_analytics(hour_of_day)`,
		`CREATE INDEX IF NOT EXISTS idx_time_day_week ON time_analytics(day_of_week)`,
		`CREATE INDEX IF NOT EXISTS idx_time_heatmap ON time_analytics(short_code, click_date, hour_of_day)`,

		// Enhanced referrer analytics table
		`CREATE TABLE IF NOT EXISTS referrer_analytics (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			referrer_domain VARCHAR(255),
			referrer_url TEXT,
			referrer_type VARCHAR(50), -- social, search, direct, email, etc.
			campaign_source VARCHAR(100), -- UTM tracking
			campaign_medium VARCHAR(100),
			campaign_name VARCHAR(100),
			campaign_term VARCHAR(100),
			campaign_content VARCHAR(100),
			clicks INTEGER DEFAULT 1,
			unique_clicks INTEGER DEFAULT 1,
			conversions INTEGER DEFAULT 0,
			last_click TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(short_code, referrer_domain, campaign_source, campaign_medium),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_referrer_short_code ON referrer_analytics(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_referrer_domain ON referrer_analytics(referrer_domain)`,
		`CREATE INDEX IF NOT EXISTS idx_referrer_type ON referrer_analytics(referrer_type)`,
		`CREATE INDEX IF NOT EXISTS idx_referrer_campaign ON referrer_analytics(campaign_source, campaign_medium)`,

		// Conversion tracking tables
		`CREATE TABLE IF NOT EXISTS conversion_goals (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			goal_name VARCHAR(100) NOT NULL,
			goal_type VARCHAR(50) NOT NULL, -- url_visit, custom_event, form_submit, purchase
			target_url TEXT,
			custom_event_name VARCHAR(100),
			goal_value DECIMAL(10,2) DEFAULT 0, -- monetary value if applicable
			attribution_window INTEGER DEFAULT 30, -- days
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conversion_goals_user ON conversion_goals(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversion_goals_active ON conversion_goals(is_active)`,

		// Conversions tracking table
		`CREATE TABLE IF NOT EXISTS conversions (
			id BIGSERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			goal_id BIGINT NOT NULL,
			conversion_id VARCHAR(100) UNIQUE NOT NULL, -- unique identifier for deduplication
			conversion_type VARCHAR(50) NOT NULL,
			conversion_value DECIMAL(10,2) DEFAULT 0,
			user_ip INET,
			user_agent TEXT,
			referrer TEXT,
			session_id VARCHAR(100), -- for session tracking
			click_id BIGINT, -- link to original click event
			conversion_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			attribution_model VARCHAR(50) DEFAULT 'last_click', -- first_click, last_click, linear
			time_to_conversion INTEGER, -- minutes from click to conversion
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE,
			FOREIGN KEY (goal_id) REFERENCES conversion_goals(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conversions_short_code ON conversions(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_conversions_goal ON conversions(goal_id)`,
		`CREATE INDEX IF NOT EXISTS idx_conversions_time ON conversions(conversion_time)`,
		`CREATE INDEX IF NOT EXISTS idx_conversions_session ON conversions(session_id)`,

		// A/B testing tables
		`CREATE TABLE IF NOT EXISTS ab_tests (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			test_name VARCHAR(100) NOT NULL,
			description TEXT,
			short_code_a VARCHAR(10) NOT NULL, -- Control variant
			short_code_b VARCHAR(10) NOT NULL, -- Test variant
			traffic_split INTEGER DEFAULT 50, -- percentage to variant B (0-100)
			start_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			end_date TIMESTAMP WITH TIME ZONE,
			status VARCHAR(20) DEFAULT 'active', -- draft, active, paused, completed
			winner VARCHAR(1), -- 'A', 'B', or NULL if inconclusive
			confidence_level DECIMAL(5,2), -- statistical confidence percentage
			min_sample_size INTEGER DEFAULT 100,
			conversion_goal_id BIGINT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (short_code_a) REFERENCES url_mappings(short_code) ON DELETE CASCADE,
			FOREIGN KEY (short_code_b) REFERENCES url_mappings(short_code) ON DELETE CASCADE,
			FOREIGN KEY (conversion_goal_id) REFERENCES conversion_goals(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_tests_user ON ab_tests(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_tests_status ON ab_tests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_tests_dates ON ab_tests(start_date, end_date)`,

		// A/B test results tracking
		`CREATE TABLE IF NOT EXISTS ab_test_results (
			id BIGSERIAL PRIMARY KEY,
			test_id BIGINT NOT NULL,
			variant VARCHAR(1) NOT NULL, -- 'A' or 'B'
			date DATE DEFAULT CURRENT_DATE,
			clicks INTEGER DEFAULT 0,
			conversions INTEGER DEFAULT 0,
			conversion_rate DECIMAL(6,4) DEFAULT 0, -- stored as percentage (0-100.00)
			revenue DECIMAL(10,2) DEFAULT 0,
			unique_visitors INTEGER DEFAULT 0,
			bounce_rate DECIMAL(5,2) DEFAULT 0,
			avg_time_on_site INTEGER DEFAULT 0, -- seconds
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (test_id) REFERENCES ab_tests(id) ON DELETE CASCADE,
			UNIQUE(test_id, variant, date)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_test_results_test ON ab_test_results(test_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ab_test_results_date ON ab_test_results(date)`,

		// Attribution tracking for better conversion analysis
		`CREATE TABLE IF NOT EXISTS attribution_touchpoints (
			id BIGSERIAL PRIMARY KEY,
			session_id VARCHAR(100) NOT NULL,
			short_code VARCHAR(10) NOT NULL,
			user_ip INET,
			user_agent TEXT,
			referrer TEXT,
			campaign_source VARCHAR(100),
			campaign_medium VARCHAR(100),
			campaign_name VARCHAR(100),
			touchpoint_order INTEGER NOT NULL, -- 1st touch, 2nd touch, etc.
			touchpoint_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			conversion_id VARCHAR(100), -- link to final conversion if exists
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_attribution_session ON attribution_touchpoints(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_attribution_conversion ON attribution_touchpoints(conversion_id)`,
		`CREATE INDEX IF NOT EXISTS idx_attribution_time ON attribution_touchpoints(touchpoint_time)`,

		// AI Insights table — one active insight per type per link
		`CREATE TABLE IF NOT EXISTS ai_insights (
			id            BIGSERIAL PRIMARY KEY,
			short_code    VARCHAR(10) NOT NULL REFERENCES url_mappings(short_code) ON DELETE CASCADE,
			user_id       BIGINT REFERENCES users(id),
			insight_type  VARCHAR(50) NOT NULL,
			content       TEXT NOT NULL,
			data_snapshot JSONB,
			generated_at  TIMESTAMPTZ DEFAULT NOW(),
			expires_at    TIMESTAMPTZ,
			is_read       BOOLEAN DEFAULT FALSE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_insights_short_code ON ai_insights(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_insights_type ON ai_insights(insight_type)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_insights_code_type ON ai_insights(short_code, insight_type)`,

		// Custom Domains — premium/enterprise feature
		`CREATE TABLE IF NOT EXISTS custom_domains (
			id                  BIGSERIAL PRIMARY KEY,
			user_id             BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			domain              VARCHAR(255) UNIQUE NOT NULL,
			verification_method VARCHAR(20) DEFAULT 'dns',
			verification_token  VARCHAR(255) UNIQUE NOT NULL,
			is_verified         BOOLEAN DEFAULT FALSE,
			verified_at         TIMESTAMPTZ,
			ssl_enabled         BOOLEAN DEFAULT TRUE,
			ssl_status          VARCHAR(20) DEFAULT 'pending',
			dns_status          VARCHAR(20) DEFAULT 'pending',
			health_status       VARCHAR(20) DEFAULT 'unknown',
			last_health_check   TIMESTAMPTZ,
			is_active           BOOLEAN DEFAULT TRUE,
			created_at          TIMESTAMPTZ DEFAULT NOW(),
			updated_at          TIMESTAMPTZ DEFAULT NOW()
		)`,
		// ALTER TABLE ensures columns added after migration 016 exist on older DBs.
		`ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS ssl_status    VARCHAR(20) DEFAULT 'pending'`,
		`ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS dns_status    VARCHAR(20) DEFAULT 'pending'`,
		`ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS health_status VARCHAR(20) DEFAULT 'unknown'`,
		`ALTER TABLE custom_domains ADD COLUMN IF NOT EXISTS last_health_check TIMESTAMPTZ`,
		`CREATE INDEX IF NOT EXISTS idx_custom_domains_user_id ON custom_domains(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_custom_domains_domain  ON custom_domains(domain)`,
		`CREATE INDEX IF NOT EXISTS idx_custom_domains_verified ON custom_domains(is_verified)`,

		// Brand Settings — enterprise white-label customisation
		`CREATE TABLE IF NOT EXISTS brand_settings (
			id             BIGSERIAL PRIMARY KEY,
			user_id        BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			company_name   VARCHAR(100) DEFAULT '',
			logo_url       TEXT DEFAULT '',
			primary_color  VARCHAR(7)  DEFAULT '#3b82f6',
			accent_color   VARCHAR(7)  DEFAULT '#10b981',
			custom_css     TEXT DEFAULT '',
			favicon_url    TEXT DEFAULT '',
			created_at     TIMESTAMPTZ DEFAULT NOW(),
			updated_at     TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_brand_settings_user_id ON brand_settings(user_id)`,

		// Teams — CREATE TABLE handles fresh installs; ALTER TABLE adds columns
		// that may be missing on existing DBs that ran migration 014_team_management_up.sql.
		`CREATE TABLE IF NOT EXISTS teams (
			id           BIGSERIAL PRIMARY KEY,
			name         VARCHAR(100) UNIQUE NOT NULL,
			display_name VARCHAR(255) NOT NULL DEFAULT '',
			description  TEXT,
			owner_id     BIGINT NOT NULL REFERENCES users(id),
			is_active    BOOLEAN NOT NULL DEFAULT TRUE,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at   TIMESTAMPTZ
		)`,
		`ALTER TABLE teams ADD COLUMN IF NOT EXISTS display_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE teams ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ`,
		`ALTER TABLE teams ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW()`,
		`ALTER TABLE teams ALTER COLUMN slug DROP NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams(owner_id)`,

		`CREATE TABLE IF NOT EXISTS team_members (
			id         BIGSERIAL PRIMARY KEY,
			team_id    BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role       VARCHAR(50) NOT NULL DEFAULT 'team_member',
			invited_by BIGINT REFERENCES users(id),
			joined_at  TIMESTAMPTZ,
			status     VARCHAR(50) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(team_id, user_id)
		)`,
		`ALTER TABLE team_members ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'active'`,
		`ALTER TABLE team_members ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW()`,
		`ALTER TABLE team_members ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW()`,
		`CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id)`,
		`CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id)`,

		`CREATE TABLE IF NOT EXISTS team_invitations (
			id         BIGSERIAL PRIMARY KEY,
			team_id    BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			email      VARCHAR(255) NOT NULL,
			role       VARCHAR(50) NOT NULL DEFAULT 'team_member',
			invited_by BIGINT NOT NULL REFERENCES users(id),
			token      VARCHAR(255) UNIQUE,
			status     VARCHAR(50) NOT NULL DEFAULT 'pending',
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE team_invitations ADD COLUMN IF NOT EXISTS token VARCHAR(255)`,
		`ALTER TABLE team_invitations ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'pending'`,
		`ALTER TABLE team_invitations ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW()`,
		`ALTER TABLE team_invitations ALTER COLUMN invitation_token DROP NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_team_invitations_team_id ON team_invitations(team_id)`,
		`CREATE INDEX IF NOT EXISTS idx_team_invitations_token ON team_invitations(token)`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

// SaveAIInsight upserts an AI insight (one per short_code + insight_type).
func (p *PostgresStorage) SaveAIInsight(insight *models.AIInsight) error {
	query := `
		INSERT INTO ai_insights (short_code, user_id, insight_type, content, data_snapshot, generated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (short_code, insight_type) DO UPDATE SET
			content       = EXCLUDED.content,
			data_snapshot = EXCLUDED.data_snapshot,
			generated_at  = EXCLUDED.generated_at,
			expires_at    = EXCLUDED.expires_at,
			is_read       = FALSE
		RETURNING id
	`
	return p.db.QueryRow(query,
		insight.ShortCode, insight.UserID, insight.InsightType,
		insight.Content, insight.DataSnapshot,
		insight.GeneratedAt, insight.ExpiresAt,
	).Scan(&insight.ID)
}

// GetAIInsight retrieves the latest AI insight for a given short code and type.
func (p *PostgresStorage) GetAIInsight(shortCode, insightType string) (*models.AIInsight, error) {
	query := `
		SELECT id, short_code, user_id, insight_type, content, data_snapshot, generated_at, expires_at, is_read
		FROM ai_insights
		WHERE short_code = $1 AND insight_type = $2
	`
	row := p.db.QueryRow(query, shortCode, insightType)

	var ins models.AIInsight
	var userID sql.NullInt64
	var expiresAt sql.NullTime
	var dataSnapshot []byte

	err := row.Scan(&ins.ID, &ins.ShortCode, &userID, &ins.InsightType,
		&ins.Content, &dataSnapshot, &ins.GeneratedAt, &expiresAt, &ins.IsRead)
	if err != nil {
		return nil, err
	}
	if userID.Valid {
		ins.UserID = &userID.Int64
	}
	if expiresAt.Valid {
		ins.ExpiresAt = &expiresAt.Time
	}
	ins.DataSnapshot = dataSnapshot
	return &ins, nil
}

// SaveURLMapping saves a URL mapping to the database
func (p *PostgresStorage) SaveURLMapping(mapping *models.URLMapping) error {
	query := `
		INSERT INTO url_mappings (id, short_code, original_url, created_at, expires_at, created_by_ip, user_id, password_hash, og_title, og_description, og_image, twitter_card, custom_domain_id, team_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := p.db.Exec(query, mapping.ID, mapping.ShortCode, mapping.OriginalURL,
		mapping.CreatedAt, mapping.ExpiresAt, mapping.CreatedByIP, mapping.UserID, mapping.PasswordHash,
		mapping.OGTitle, mapping.OGDescription, mapping.OGImage, mapping.TwitterCard,
		mapping.CustomDomainID, mapping.TeamID)

	if err != nil {
		return fmt.Errorf("failed to save URL mapping: %w", err)
	}
	return nil
}

// CountURLsByIP counts anonymous URLs created by a specific IP address
func (p *PostgresStorage) CountURLsByIP(ip string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM url_mappings WHERE created_by_ip = $1 AND user_id IS NULL`
	err := p.db.QueryRow(query, ip).Scan(&count)
	return count, err
}

// GetURLMappingByShortCode retrieves a URL mapping by its short code
func (p *PostgresStorage) GetURLMappingByShortCode(shortCode string) (*models.URLMapping, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, click_count, is_active, created_by_ip, user_id, password_hash,
		       og_title, og_description, og_image, twitter_card, custom_domain_id, team_id
		FROM url_mappings
		WHERE short_code = $1 AND is_active = TRUE
	`

	mapping := &models.URLMapping{}
	var expiresAt sql.NullTime
	var createdByIP sql.NullString
	var userID sql.NullInt64
	var passwordHash sql.NullString
	var ogTitle, ogDescription, ogImage, twitterCard sql.NullString
	var customDomainID, teamID sql.NullInt64

	err := p.db.QueryRow(query, shortCode).Scan(
		&mapping.ID, &mapping.ShortCode, &mapping.OriginalURL,
		&mapping.CreatedAt, &expiresAt, &mapping.ClickCount,
		&mapping.IsActive, &createdByIP, &userID, &passwordHash,
		&ogTitle, &ogDescription, &ogImage, &twitterCard,
		&customDomainID, &teamID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get URL mapping: %w", err)
	}

	if expiresAt.Valid {
		mapping.ExpiresAt = &expiresAt.Time
	}
	if createdByIP.Valid {
		mapping.CreatedByIP = createdByIP.String
	}
	if userID.Valid {
		mapping.UserID = &userID.Int64
	}
	if passwordHash.Valid {
		mapping.PasswordHash = &passwordHash.String
	}
	if ogTitle.Valid {
		mapping.OGTitle = &ogTitle.String
	}
	if ogDescription.Valid {
		mapping.OGDescription = &ogDescription.String
	}
	if ogImage.Valid {
		mapping.OGImage = &ogImage.String
	}
	if twitterCard.Valid {
		mapping.TwitterCard = &twitterCard.String
	}
	if customDomainID.Valid {
		mapping.CustomDomainID = &customDomainID.Int64
	}
	if teamID.Valid {
		mapping.TeamID = &teamID.Int64
	}

	// Check if URL has expired
	if mapping.ExpiresAt != nil && mapping.ExpiresAt.Before(time.Now()) {
		return nil, ErrURLExpired
	}

	return mapping, nil
}

// GetURLMappingByOriginalURLAndUser retrieves a URL mapping by original URL and user ID
func (p *PostgresStorage) GetURLMappingByOriginalURLAndUser(originalURL string, userID int64) (*models.URLMapping, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, click_count, is_active, created_by_ip, user_id, password_hash,
		       og_title, og_description, og_image, twitter_card, custom_domain_id, team_id
		FROM url_mappings
		WHERE original_url = $1 AND user_id = $2 AND is_active = TRUE
		LIMIT 1
	`

	mapping := &models.URLMapping{}
	var expiresAt sql.NullTime
	var createdByIP sql.NullString
	var dbUserID sql.NullInt64
	var passwordHash sql.NullString
	var ogTitle, ogDescription, ogImage, twitterCard sql.NullString
	var customDomainID, teamID sql.NullInt64

	err := p.db.QueryRow(query, originalURL, userID).Scan(
		&mapping.ID, &mapping.ShortCode, &mapping.OriginalURL,
		&mapping.CreatedAt, &expiresAt, &mapping.ClickCount,
		&mapping.IsActive, &createdByIP, &dbUserID, &passwordHash,
		&ogTitle, &ogDescription, &ogImage, &twitterCard,
		&customDomainID, &teamID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get URL mapping: %w", err)
	}

	if expiresAt.Valid {
		mapping.ExpiresAt = &expiresAt.Time
	}
	if createdByIP.Valid {
		mapping.CreatedByIP = createdByIP.String
	}
	if dbUserID.Valid {
		mapping.UserID = &dbUserID.Int64
	}
	if passwordHash.Valid {
		mapping.PasswordHash = &passwordHash.String
	}
	if ogTitle.Valid {
		mapping.OGTitle = &ogTitle.String
	}
	if ogDescription.Valid {
		mapping.OGDescription = &ogDescription.String
	}
	if ogImage.Valid {
		mapping.OGImage = &ogImage.String
	}
	if twitterCard.Valid {
		mapping.TwitterCard = &twitterCard.String
	}
	if customDomainID.Valid {
		mapping.CustomDomainID = &customDomainID.Int64
	}
	if teamID.Valid {
		mapping.TeamID = &teamID.Int64
	}

	// Check if URL has expired
	if mapping.ExpiresAt != nil && mapping.ExpiresAt.Before(time.Now()) {
		return nil, ErrURLExpired
	}

	return mapping, nil
}

// GetURLMappingByDomain retrieves a URL mapping by short code and custom domain hostname.
// Used when the Host header indicates a custom domain request.
func (p *PostgresStorage) GetURLMappingByDomain(shortCode, hostname string) (*models.URLMapping, error) {
	query := `
		SELECT um.id, um.short_code, um.original_url, um.created_at, um.expires_at, um.click_count,
		       um.is_active, um.created_by_ip, um.user_id, um.password_hash,
		       um.og_title, um.og_description, um.og_image, um.twitter_card, um.custom_domain_id, um.team_id
		FROM url_mappings um
		JOIN custom_domains cd ON um.custom_domain_id = cd.id
		WHERE um.short_code = $1 AND cd.domain = $2 AND um.is_active = TRUE
	`

	mapping := &models.URLMapping{}
	var expiresAt sql.NullTime
	var createdByIP sql.NullString
	var userID sql.NullInt64
	var passwordHash sql.NullString
	var ogTitle, ogDescription, ogImage, twitterCard sql.NullString
	var customDomainID, teamID sql.NullInt64

	err := p.db.QueryRow(query, shortCode, hostname).Scan(
		&mapping.ID, &mapping.ShortCode, &mapping.OriginalURL,
		&mapping.CreatedAt, &expiresAt, &mapping.ClickCount,
		&mapping.IsActive, &createdByIP, &userID, &passwordHash,
		&ogTitle, &ogDescription, &ogImage, &twitterCard,
		&customDomainID, &teamID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get URL mapping by domain: %w", err)
	}

	if expiresAt.Valid {
		mapping.ExpiresAt = &expiresAt.Time
	}
	if createdByIP.Valid {
		mapping.CreatedByIP = createdByIP.String
	}
	if userID.Valid {
		mapping.UserID = &userID.Int64
	}
	if passwordHash.Valid {
		mapping.PasswordHash = &passwordHash.String
	}
	if ogTitle.Valid {
		mapping.OGTitle = &ogTitle.String
	}
	if ogDescription.Valid {
		mapping.OGDescription = &ogDescription.String
	}
	if ogImage.Valid {
		mapping.OGImage = &ogImage.String
	}
	if twitterCard.Valid {
		mapping.TwitterCard = &twitterCard.String
	}
	if customDomainID.Valid {
		mapping.CustomDomainID = &customDomainID.Int64
	}
	if teamID.Valid {
		mapping.TeamID = &teamID.Int64
	}

	if mapping.ExpiresAt != nil && mapping.ExpiresAt.Before(time.Now()) {
		return nil, ErrURLExpired
	}

	return mapping, nil
}

// UpdateURLOGMetadata updates OG metadata fields for a URL mapping
func (p *PostgresStorage) UpdateURLOGMetadata(shortCode string, userID int64, req *models.UpdateOGRequest) error {
	query := `
		UPDATE url_mappings
		SET og_title = $1, og_description = $2, og_image = $3, twitter_card = $4
		WHERE short_code = $5 AND user_id = $6
	`
	result, err := p.db.Exec(query, req.OGTitle, req.OGDescription, req.OGImage, req.TwitterCard, shortCode, userID)
	if err != nil {
		return fmt.Errorf("failed to update OG metadata: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return ErrURLNotFound
	}
	return nil
}

// IncrementClickCount increments the click count for a URL mapping
func (p *PostgresStorage) IncrementClickCount(shortCode string) error {
	query := `UPDATE url_mappings SET click_count = click_count + 1 WHERE short_code = $1`
	_, err := p.db.Exec(query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}
	return nil
}

// SaveClickEvent saves a click event to the database
func (p *PostgresStorage) SaveClickEvent(event *models.ClickEvent) error {
	query := `
		INSERT INTO click_events (id, short_code, clicked_at, ip_address, user_agent, referrer, country_code, is_bot, is_unique, ip_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := p.db.Exec(query, event.ID, event.ShortCode, event.ClickedAt,
		event.IPAddress, event.UserAgent, event.Referrer, event.CountryCode,
		event.IsBot, event.IsUnique, event.IPHash)

	if err != nil {
		return fmt.Errorf("failed to save click event: %w", err)
	}
	return nil
}

// GetAnalytics retrieves analytics data for a short code
func (p *PostgresStorage) GetAnalytics(shortCode string, days int) (*models.AnalyticsResponse, error) {
	// Get basic URL info
	mapping, err := p.GetURLMappingByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	analytics := &models.AnalyticsResponse{
		ShortCode:   mapping.ShortCode,
		OriginalURL: mapping.OriginalURL,
		TotalClicks: mapping.ClickCount,
		CreatedAt:   mapping.CreatedAt,
	}

	// Get last click time
	var lastClickAt sql.NullTime
	err = p.db.QueryRow(
		`SELECT MAX(clicked_at) FROM click_events WHERE short_code = $1`,
		shortCode,
	).Scan(&lastClickAt)
	if err == nil && lastClickAt.Valid {
		analytics.LastClickAt = &lastClickAt.Time
	}

	// Get daily clicks for the specified number of days
	dailyClicksQuery := `
		SELECT DATE(clicked_at) as date, COUNT(*) as clicks
		FROM click_events
		WHERE short_code = $1 AND clicked_at >= NOW() - INTERVAL '%d days' AND is_bot = FALSE
		GROUP BY DATE(clicked_at)
		ORDER BY date DESC
	`
	rows, err := p.db.Query(fmt.Sprintf(dailyClicksQuery, days), shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var dailyClick models.DailyClick
			if err := rows.Scan(&dailyClick.Date, &dailyClick.Clicks); err == nil {
				analytics.DailyClicks = append(analytics.DailyClicks, dailyClick)
			}
		}
	}

	// Get country stats
	countryStatsQuery := `
		SELECT country_code, COUNT(*) as clicks
		FROM click_events
		WHERE short_code = $1 AND country_code IS NOT NULL AND is_bot = FALSE
		GROUP BY country_code
		ORDER BY clicks DESC
		LIMIT 10
	`
	rows, err = p.db.Query(countryStatsQuery, shortCode)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var countryStat models.CountryStat
			if err := rows.Scan(&countryStat.CountryCode, &countryStat.Clicks); err == nil {
				analytics.CountryStats = append(analytics.CountryStats, countryStat)
			}
		}
	}

	return analytics, nil
}

// ShortCodeExists checks if a short code already exists
func (p *PostgresStorage) ShortCodeExists(shortCode string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM url_mappings WHERE short_code = $1)`
	err := p.db.QueryRow(query, shortCode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if short code exists: %w", err)
	}
	return exists, nil
}

// User management methods

// CreateUser creates a new user in the database
func (p *PostgresStorage) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, name, email, phone, password_hash, email_verified, phone_verified,
			              provider, provider_id, avatar_url, account_type, is_active, is_admin, created_at, updated_at,
			              date_of_birth, is_minor, age_verified_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	_, err := p.Exec(query, user.ID, user.Name, user.Email, user.Phone, user.PasswordHash,
		user.EmailVerified, user.PhoneVerified, user.Provider, user.ProviderID, user.AvatarURL,
		user.AccountType, user.IsActive, user.IsAdmin, user.CreatedAt, user.UpdatedAt,
		user.DateOfBirth, user.IsMinor, user.AgeVerifiedAt)
	return err
}

// GetUserByID retrieves a user by ID
func (p *PostgresStorage) GetUserByID(id int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, name, email, phone, password_hash, email_verified, phone_verified,
		       provider, provider_id, avatar_url, account_type, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users WHERE id = $1
	`
	err := p.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.PasswordHash,
		&user.EmailVerified, &user.PhoneVerified, &user.Provider, &user.ProviderID,
		&user.AvatarURL, &user.AccountType, &user.IsActive, &user.IsAdmin,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// SyncAccountTypeFromSubscription checks polar_subscriptions for an active paid plan
// and updates users.account_type if the stored value is 'free'.
// Returns (newPlanType, wasUpdated, error).
func (p *PostgresStorage) SyncAccountTypeFromSubscription(userID int64) (string, bool, error) {
	var planType string
	err := p.QueryRow(`
		SELECT plan_type FROM polar_subscriptions
		WHERE user_id = $1
		  AND status IN ('active', 'trialing')
		  AND plan_type IN ('premium', 'enterprise')
		  AND (current_period_end IS NULL OR current_period_end > NOW())
		ORDER BY CASE plan_type WHEN 'enterprise' THEN 2 WHEN 'premium' THEN 1 END DESC
		LIMIT 1
	`, userID).Scan(&planType)

	if err == sql.ErrNoRows {
		return "free", false, nil
	}
	if err != nil {
		return "", false, err
	}

	_, err = p.Exec(
		"UPDATE users SET account_type = $1, updated_at = NOW() WHERE id = $2",
		planType, userID,
	)
	if err != nil {
		return "", false, err
	}

	log.Printf("[auth] auto-synced account_type to %q for user %d", planType, userID)
	return planType, true, nil
}

// GetUserByEmail retrieves a user by email
func (p *PostgresStorage) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, name, email, phone, password_hash, email_verified, phone_verified,
		       provider, provider_id, avatar_url, account_type, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users WHERE email = $1
	`
	err := p.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.PasswordHash,
		&user.EmailVerified, &user.PhoneVerified, &user.Provider, &user.ProviderID,
		&user.AvatarURL, &user.AccountType, &user.IsActive, &user.IsAdmin,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByPhone retrieves a user by phone
func (p *PostgresStorage) GetUserByPhone(phone string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, name, email, phone, password_hash, email_verified, phone_verified,
		       provider, provider_id, avatar_url, account_type, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users WHERE phone = $1
	`
	err := p.QueryRow(query, phone).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.PasswordHash,
		&user.EmailVerified, &user.PhoneVerified, &user.Provider, &user.ProviderID,
		&user.AvatarURL, &user.AccountType, &user.IsActive, &user.IsAdmin,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByProviderID retrieves a user by provider and provider_id
func (p *PostgresStorage) GetUserByProviderID(provider, providerID string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, name, email, phone, password_hash, email_verified, phone_verified,
		       provider, provider_id, avatar_url, account_type, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users WHERE provider = $1 AND provider_id = $2
	`
	err := p.QueryRow(query, provider, providerID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.PasswordHash,
		&user.EmailVerified, &user.PhoneVerified, &user.Provider, &user.ProviderID,
		&user.AvatarURL, &user.AccountType, &user.IsActive, &user.IsAdmin,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserLastLogin updates user's last login time
func (p *PostgresStorage) UpdateUserLastLogin(userID int64, loginTime time.Time) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, loginTime, time.Now(), userID)
	return err
}

// CreateUserPreferences creates user preferences
func (p *PostgresStorage) CreateUserPreferences(prefs *models.UserPreferences) error {
	query := `
		INSERT INTO user_preferences (user_id, default_expiration, analytics_public, 
		                            email_notifications, marketing_emails, timezone, theme)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.Exec(query, prefs.UserID, prefs.DefaultExpiration, prefs.AnalyticsPublic,
		prefs.EmailNotifications, prefs.MarketingEmails, prefs.Timezone, prefs.Theme)
	return err
}

// CreateSession creates a new user session
func (p *PostgresStorage) CreateSession(session *models.UserSession) error {
	query := `
		INSERT INTO user_sessions (id, user_id, device_info, ip_address, user_agent,
		                         refresh_token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := p.Exec(query, session.ID, session.UserID, session.DeviceInfo, session.IPAddress,
		session.UserAgent, session.RefreshTokenHash, session.ExpiresAt, session.CreatedAt)
	return err
}

// GetSessionByID retrieves a session by ID
func (p *PostgresStorage) GetSessionByID(id string) (*models.UserSession, error) {
	session := &models.UserSession{}
	query := `
		SELECT id, user_id, device_info, ip_address, user_agent, 
		       refresh_token_hash, expires_at, created_at
		FROM user_sessions WHERE id = $1
	`
	err := p.QueryRow(query, id).Scan(
		&session.ID, &session.UserID, &session.DeviceInfo, &session.IPAddress,
		&session.UserAgent, &session.RefreshTokenHash, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// DeleteUserSessions deletes all sessions for a user
func (p *PostgresStorage) DeleteUserSessions(userID int64) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1`
	_, err := p.Exec(query, userID)
	return err
}

// ReconcileAdminFlags clears the legacy is_admin flag for any user who no longer
// holds an admin or super_admin RBAC role. Called at startup to fix stale state.
func (p *PostgresStorage) ReconcileAdminFlags() error {
	query := `
		UPDATE users
		SET is_admin = FALSE
		WHERE is_admin = TRUE
		  AND id NOT IN (
			SELECT DISTINCT ur.user_id
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE r.name IN ('admin', 'super_admin')
			  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		  )
	`
	res, err := p.Exec(query)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("⚠️  ReconcileAdminFlags: cleared is_admin for %d user(s) with no admin RBAC role", n)
	}
	return err
}

// Email verification methods

// CreateEmailVerification creates an email verification record
func (p *PostgresStorage) CreateEmailVerification(verification *models.EmailVerification) error {
	query := `
		INSERT INTO email_verifications (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := p.Exec(query, verification.ID, verification.UserID, verification.Token,
		verification.ExpiresAt, verification.CreatedAt)
	return err
}

// GetEmailVerificationByToken retrieves email verification by token
func (p *PostgresStorage) GetEmailVerificationByToken(token string) (*models.EmailVerification, error) {
	verification := &models.EmailVerification{}
	query := `
		SELECT id, user_id, token, expires_at, verified_at, created_at
		FROM email_verifications WHERE token = $1
	`
	err := p.QueryRow(query, token).Scan(
		&verification.ID, &verification.UserID, &verification.Token,
		&verification.ExpiresAt, &verification.VerifiedAt, &verification.CreatedAt)
	if err != nil {
		return nil, err
	}
	return verification, nil
}

// UpdateEmailVerification updates email verification record
func (p *PostgresStorage) UpdateEmailVerification(id string, verifiedAt time.Time) error {
	query := `UPDATE email_verifications SET verified_at = $1 WHERE id = $2`
	_, err := p.Exec(query, verifiedAt, id)
	return err
}

// UpdateUserEmailVerified updates user email verification status
func (p *PostgresStorage) UpdateUserEmailVerified(userID int64, verified bool) error {
	query := `UPDATE users SET email_verified = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, verified, time.Now(), userID)
	return err
}

// Phone verification methods

// CreatePhoneVerification creates a phone verification record
func (p *PostgresStorage) CreatePhoneVerification(verification *models.PhoneVerification) error {
	query := `
		INSERT INTO phone_verifications (user_id, phone, otp_code, attempts, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err := p.QueryRow(query, verification.UserID, verification.Phone,
		verification.OTPCode, verification.Attempts, verification.ExpiresAt,
		verification.CreatedAt).Scan(&verification.ID)
	return err
}

// DeletePhoneVerificationByUserID deletes phone verification by user ID
func (p *PostgresStorage) DeletePhoneVerificationByUserID(userID int64) error {
	query := `DELETE FROM phone_verifications WHERE user_id = $1`
	_, err := p.Exec(query, userID)
	return err
}

// GetPhoneVerificationByUserID retrieves phone verification by user ID
func (p *PostgresStorage) GetPhoneVerificationByUserID(userID int64) (*models.PhoneVerification, error) {
	verification := &models.PhoneVerification{}
	query := `
		SELECT id, user_id, phone, otp_code, attempts, expires_at, verified_at, created_at
		FROM phone_verifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1
	`
	err := p.QueryRow(query, userID).Scan(
		&verification.ID, &verification.UserID, &verification.Phone,
		&verification.OTPCode, &verification.Attempts, &verification.ExpiresAt, &verification.VerifiedAt, &verification.CreatedAt)
	if err != nil {
		return nil, err
	}
	return verification, nil
}

// UpdatePhoneVerificationAttempts updates phone verification attempts
func (p *PostgresStorage) UpdatePhoneVerificationAttempts(userID int64, attempts int) error {
	query := `UPDATE phone_verifications SET attempts = $1 WHERE user_id = $2`
	_, err := p.Exec(query, attempts, userID)
	return err
}

// UpdatePhoneVerification updates phone verification record
func (p *PostgresStorage) UpdatePhoneVerification(id string, verifiedAt time.Time) error {
	query := `UPDATE phone_verifications SET verified_at = $1 WHERE id = $2`
	_, err := p.Exec(query, verifiedAt, id)
	return err
}

// UpdateUserPhoneVerified updates user phone verification status
func (p *PostgresStorage) UpdateUserPhoneVerified(userID int64, verified bool) error {
	query := `UPDATE users SET phone_verified = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, verified, time.Now(), userID)
	return err
}

// Password reset methods

// CreatePasswordReset creates a password reset record
func (p *PostgresStorage) CreatePasswordReset(reset *models.PasswordReset) error {
	query := `
		INSERT INTO password_resets (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := p.Exec(query, reset.ID, reset.UserID, reset.Token,
		reset.ExpiresAt, reset.CreatedAt)
	return err
}

// GetPasswordResetByToken retrieves password reset by token
func (p *PostgresStorage) GetPasswordResetByToken(token string) (*models.PasswordReset, error) {
	reset := &models.PasswordReset{}
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_resets WHERE token = $1
	`
	err := p.QueryRow(query, token).Scan(
		&reset.ID, &reset.UserID, &reset.Token,
		&reset.ExpiresAt, &reset.UsedAt, &reset.CreatedAt)
	if err != nil {
		return nil, err
	}
	return reset, nil
}

// UpdatePasswordReset updates password reset record
func (p *PostgresStorage) UpdatePasswordReset(id string, usedAt time.Time) error {
	query := `UPDATE password_resets SET used_at = $1 WHERE id = $2`
	_, err := p.Exec(query, usedAt, id)
	return err
}

// MarkPasswordResetAsUsed marks a password reset token as used
func (p *PostgresStorage) MarkPasswordResetAsUsed(id interface{}, usedAt time.Time) error {
	query := `UPDATE password_resets SET used_at = $1 WHERE id = $2`
	_, err := p.Exec(query, usedAt, id)
	return err
}

// UpdateUserPassword updates user password
func (p *PostgresStorage) UpdateUserPassword(userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, passwordHash, time.Now(), userID)
	return err
}

// User management methods

// UpdateUser updates user information
func (p *PostgresStorage) UpdateUser(user *models.User) error {
	query := `
		UPDATE users SET name = $1, email = $2, phone = $3, provider = $4,
		                provider_id = $5, avatar_url = $6, account_type = $7,
		                updated_at = $8
		WHERE id = $9
	`
	_, err := p.Exec(query, user.Name, user.Email, user.Phone, user.Provider,
		user.ProviderID, user.AvatarURL, user.AccountType, time.Now(), user.ID)
	return err
}

// UpdateUserProfile updates specific user profile fields
func (p *PostgresStorage) UpdateUserProfile(userID int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Build dynamic query
	var setClauses []string
	var args []interface{}
	argIndex := 1

	for field, value := range updates {
		// Validate field names to prevent SQL injection
		switch field {
		case "name", "phone", "avatar_url":
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		default:
			return fmt.Errorf("invalid field name: %s", field)
		}
	}

	// Always update updated_at
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add userID as final parameter
	args = append(args, userID)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIndex)

	result, err := p.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %d or no changes made", userID)
	}

	log.Printf("[INFO] Profile updated successfully for user %d, rows affected: %d", userID, rowsAffected)
	return nil
}

// UpdateUserStatus updates user status (active/inactive)
func (p *PostgresStorage) UpdateUserStatus(userID int64, isActive bool) error {
	query := `UPDATE users SET is_active = $1, updated_at = $2 WHERE id = $3`
	_, err := p.Exec(query, isActive, time.Now(), userID)
	return err
}

// GetUserPreferences retrieves user preferences
func (p *PostgresStorage) GetUserPreferences(userID int64) (*models.UserPreferences, error) {
	prefs := &models.UserPreferences{}
	query := `
		SELECT user_id, default_expiration, analytics_public, email_notifications,
		       marketing_emails, timezone, theme
		FROM user_preferences WHERE user_id = $1
	`
	err := p.QueryRow(query, userID).Scan(
		&prefs.UserID, &prefs.DefaultExpiration, &prefs.AnalyticsPublic,
		&prefs.EmailNotifications, &prefs.MarketingEmails, &prefs.Timezone, &prefs.Theme)
	if err != nil {
		return nil, err
	}
	return prefs, nil
}

// UpdateUserPreferences updates user preferences (creates if not exists)
func (p *PostgresStorage) UpdateUserPreferences(prefs *models.UserPreferences) error {
	query := `
		INSERT INTO user_preferences (
			user_id, default_expiration, analytics_public,
			email_notifications, marketing_emails, timezone, theme
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			default_expiration = EXCLUDED.default_expiration,
			analytics_public = EXCLUDED.analytics_public,
			email_notifications = EXCLUDED.email_notifications,
			marketing_emails = EXCLUDED.marketing_emails,
			timezone = EXCLUDED.timezone,
			theme = EXCLUDED.theme
	`
	_, err := p.Exec(query,
		prefs.UserID,
		prefs.DefaultExpiration,
		prefs.AnalyticsPublic,
		prefs.EmailNotifications,
		prefs.MarketingEmails,
		prefs.Timezone,
		prefs.Theme,
	)
	return err
}

func (s *PostgresStorage) SeedDefaultStaticPages() error {
	pages := []struct{ slug, title, content string }{
		{"about", "About Us", `[{"id":"block-about","type":"text","content":{"html":"<h1>About Us</h1><p>Tell your story here.</p>"},"styles":{"padding":"32px","margin":"0px","backgroundColor":"transparent","textColor":"#000000","borderRadius":"0px","border":"none","textAlign":"left","fontSize":"16px","fontWeight":"normal"},"settings":{"visible":true}}]`},
		{"contact", "Contact", `[{"id":"block-contact","type":"text","content":{"html":"<h1>Contact</h1><p>Get in touch.</p>"},"styles":{"padding":"32px","margin":"0px","backgroundColor":"transparent","textColor":"#000000","borderRadius":"0px","border":"none","textAlign":"left","fontSize":"16px","fontWeight":"normal"},"settings":{"visible":true}}]`},
		{"terms", "Terms of Service", `[{"id":"block-terms","type":"text","content":{"html":"<h1>Terms of Service</h1><p>Your terms go here.</p>"},"styles":{"padding":"32px","margin":"0px","backgroundColor":"transparent","textColor":"#000000","borderRadius":"0px","border":"none","textAlign":"left","fontSize":"16px","fontWeight":"normal"},"settings":{"visible":true}}]`},
		{"privacy", "Privacy Policy", `[{"id":"block-privacy","type":"text","content":{"html":"<h1>Privacy Policy</h1><p>Your privacy policy goes here.</p>"},"styles":{"padding":"32px","margin":"0px","backgroundColor":"transparent","textColor":"#000000","borderRadius":"0px","border":"none","textAlign":"left","fontSize":"16px","fontWeight":"normal"},"settings":{"visible":true}}]`},
		{"features", "Features", `[{"id":"block-features","type":"text","content":{"html":"<h1>Features</h1><p>Highlight your features here.</p>"},"styles":{"padding":"32px","margin":"0px","backgroundColor":"transparent","textColor":"#000000","borderRadius":"0px","border":"none","textAlign":"left","fontSize":"16px","fontWeight":"normal"},"settings":{"visible":true}}]`},
		{"pricing", "Pricing", `[{"id":"block-pricing","type":"text","content":{"html":"<h1>Pricing</h1><p>Show your pricing plans here.</p>"},"styles":{"padding":"32px","margin":"0px","backgroundColor":"transparent","textColor":"#000000","borderRadius":"0px","border":"none","textAlign":"left","fontSize":"16px","fontWeight":"normal"},"settings":{"visible":true}}]`},
	}
	for _, p := range pages {
		_, err := s.db.Exec(`
			INSERT INTO static_pages (slug, title, content, template, is_published, sort_order, language, created_at, updated_at)
			VALUES ($1, $2, $3, 'default', false, 0, 'en', NOW(), NOW())
			ON CONFLICT (slug) DO NOTHING`,
			p.slug, p.title, p.content,
		)
		if err != nil {
			return fmt.Errorf("seed static page %s: %w", p.slug, err)
		}
	}
	return nil
}

// RunRetentionJob executes the nightly data retention policy:
//  1. Null raw IP addresses older than 24 hours
//  2. Delete click events for free-plan users older than 30 days
//  3. Delete click events for premium-plan users older than 1 year
//  4. Enterprise users: never delete
//  5. Microsite page events follow the same free/premium windows
func (p *PostgresStorage) RunRetentionJob() error {
	// 1. Null raw IPs older than 24 hours
	res, err := p.db.Exec(`
		UPDATE click_events
		SET ip_address = NULL
		WHERE clicked_at < NOW() - INTERVAL '24 hours'
		  AND ip_address IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("null raw IPs: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("🗑️  Retention: nulled %d raw IPs older than 24 h", n)
	}

	// 2. Delete clicks for free users older than 30 days
	res, err = p.db.Exec(`
		DELETE FROM click_events ce
		USING url_mappings um, users u
		WHERE ce.short_code = um.short_code
		  AND um.user_id   = u.id
		  AND u.account_type = 'free'
		  AND ce.clicked_at < NOW() - INTERVAL '30 days'
	`)
	if err != nil {
		return fmt.Errorf("delete free-plan clicks: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("🗑️  Retention: deleted %d click events for free users (>30 days)", n)
	}

	// 3. Delete clicks for premium users older than 1 year
	res, err = p.db.Exec(`
		DELETE FROM click_events ce
		USING url_mappings um, users u
		WHERE ce.short_code = um.short_code
		  AND um.user_id   = u.id
		  AND u.account_type = 'premium'
		  AND ce.clicked_at < NOW() - INTERVAL '1 year'
	`)
	if err != nil {
		return fmt.Errorf("delete premium-plan clicks: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("🗑️  Retention: deleted %d click events for premium users (>1 year)", n)
	}

	// 5. Microsite page events — same plan windows (skipped if migration 046 is not applied)
	if err := p.runMicrositeEventRetention(); err != nil {
		return err
	}

	return nil
}

// runMicrositeEventRetention applies the click retention windows to
// microsite_events and microsite_submissions: free 30 days, premium 1 year,
// enterprise kept. A table whose migration is not applied is skipped.
func (p *PostgresStorage) runMicrositeEventRetention() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	tables := []struct {
		name, timeColumn string
	}{
		{"microsite_events", "occurred_at"},
		{"microsite_submissions", "created_at"},
	}
	windows := []struct {
		accountType string
		interval    string
	}{
		{"free", "30 days"},
		{"premium", "1 year"},
	}
	for _, t := range tables {
		var exists bool
		if err := p.db.QueryRowContext(ctx, `SELECT to_regclass('public.' || $1) IS NOT NULL`, t.name).Scan(&exists); err != nil || !exists {
			continue
		}
		for _, w := range windows {
			// Table and column names are constants above, never user input.
			res, err := p.db.ExecContext(ctx, /* #nosec G202 -- identifiers from the constant list above */ `
				DELETE FROM `+t.name+` me
				USING microsites m, users u
				WHERE me.microsite_id = m.id
				  AND m.user_id = u.id
				  AND u.account_type = $1
				  AND me.`+t.timeColumn+` < NOW() - $2::interval
			`, w.accountType, w.interval)
			if err != nil {
				return fmt.Errorf("delete %s-plan %s: %w", w.accountType, t.name, err)
			}
			if n, _ := res.RowsAffected(); n > 0 {
				log.Printf("🗑️  Retention: deleted %d %s rows for %s users (>%s)", n, t.name, w.accountType, w.interval)
			}
		}
	}
	return nil
}
